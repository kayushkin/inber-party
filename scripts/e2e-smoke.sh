#!/usr/bin/env bash
# Boot-and-answer smoke test for inber-party.
#
# Runs the binary TWICE, because inber-party has two genuinely different boot
# modes and only one of them can be tested without a database:
#
#   PHASE 1 — DB-less boot (fully hermetic, no daemons, no network).
#     Proves the binary boots, binds, registers ~80 routes without a panic, and
#     serialises. Asserts the DEGRADED contract: /health is 503 "unhealthy",
#     GET /api/agents is 200 [], POST /api/agents is 503. This is the part that
#     catches the failure class this guard exists for.
#
#   PHASE 2 — real boot against a THROWAWAY PostgreSQL in docker.
#     Proves the migrations and the seed actually run (main.go:48-57, both behind
#     logger.Fatal), and round-trips an agent write->read-back through Postgres.
#
# Why this exists: `go build` proves the tree compiles, not that the binary is
# alive. Go 1.22+ http.ServeMux panics on a conflicting route pattern at
# REGISTRATION time — the compiler never sees it, `go vet` never sees it, and the
# process dies on its first breath. Tier 1 of the repo guard calls such a tree
# green. This is tier 2: it boots the thing and makes it answer.
#
# WHY DOCKER, AND WHY IT IS REQUIRED RATHER THAN OPTIONAL:
# inber-party is the only Postgres-backed service in the fleet. Its write path is
# hard-gated on a real Postgres (POST /api/agents 503s without one, api.go:326-330)
# and its dialect is Postgres-specific (SERIAL, $1, RETURNING), so SQLite cannot
# stand in. This host has NO postgres binaries at all — initdb, pg_ctl, postgres
# and pg_isready are all absent — so the usual "spin up a private cluster in a temp
# dir" answer is not available, and the repo ships no compose file. A container is
# the only way to give this binary the database it actually runs against.
#
# It is REQUIRED, not best-effort, on purpose. A smoke that silently skipped phase 2
# whenever docker happened to be down would report GREEN while never once running
# the migrations — and "the guard didn't actually check" is indistinguishable from
# "the guard passed" precisely when it matters. If docker cannot provide a database
# this script FAILS and says so, naming docker rather than blaming inber-party.
# Set E2E_SKIP_POSTGRES=1 to deliberately run phase 1 only.
#
# 🔴 INBER_URL IS NOT OPTIONAL. It defaults to http://127.0.0.1:8200 — the LIVE
# inber-server on this box. Left at its default, this smoke would attach a
# QuestMonitor to the live gateway and start a 30s polling goroutine against it
# (main.go:163-192). It is pinned to a closed port so inberSource ends up nil.
#
# 🔴 READINESS IS /health, NOT /api/health. /api/health (api.go:1679) performs an
# unconditional http.Get("http://localhost:8080/status") with the DEFAULT http
# client — no timeout — so it phones whatever is listening on the host's :8080
# (currently kayushkin-serve). Never probe it from a smoke.
#
# PORT must move off 8080 for the same reason: that is inber-party's default and
# it is occupied on this host; a collision is fatal (main.go:390).
#
# HOME is redirected: DefaultDBPaths() reads ~/.inber/{sessions,gateway}.db and the
# logstack client reads ~/.openclaw/agents. With a temp HOME both resolve to
# nothing, which is exactly what we want inberSource to be.
#
# Exits 0 on success, non-zero on the FIRST failing assertion, dumping the server
# log to stderr.
#
# Env:
#   E2E_PORT           — inber-party HTTP port     (default 19135)
#   E2E_PG_PORT        — throwaway postgres port   (default 15432)
#   E2E_PG_IMAGE       — postgres image            (default postgres:16-alpine)
#   E2E_SKIP_POSTGRES=1— run phase 1 only, and say so loudly
#   E2E_KEEP=1         — leave $TMP_DIR in place after the run

set -euo pipefail

REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PORT="${E2E_PORT:-19135}"
PG_PORT="${E2E_PG_PORT:-15432}"
PG_IMAGE="${E2E_PG_IMAGE:-postgres:16-alpine}"
BASE="http://127.0.0.1:$PORT"

for bin in go curl jq; do
  if ! command -v "$bin" >/dev/null 2>&1; then
    echo "ERROR: required tool '$bin' not found on PATH" >&2
    exit 2
  fi
done

TMP_DIR="$(mktemp -d -t inber-party-e2e.XXXXXX)"
BIN_DIR="$TMP_DIR/bin"
FAKE_HOME="$TMP_DIR/home"
RUN_DIR="$TMP_DIR/run"
SERVER_LOG="$TMP_DIR/server.log"
mkdir -p "$BIN_DIR" "$FAKE_HOME" "$RUN_DIR"

PG_CONTAINER="inber-party-e2e-$$"
SERVER_PID=""
PG_STARTED=""
cleanup() {
  if [ -n "$SERVER_PID" ] && kill -0 "$SERVER_PID" 2>/dev/null; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
  # Tear the container down on EVERY exit path, including a failed assertion and
  # an interrupt. --rm alone is not enough: it only fires when the container stops.
  if [ -n "$PG_STARTED" ]; then
    docker rm -f "$PG_CONTAINER" >/dev/null 2>&1 || true
  fi
  if [ "${E2E_KEEP:-}" = "1" ]; then
    echo "[e2e] keeping $TMP_DIR"
  else
    rm -rf "$TMP_DIR"
  fi
}
trap cleanup EXIT INT TERM

step() { printf '\n==> %s\n' "$*"; }
dump_logs() {
  echo "----- server.log -----" >&2
  cat "$SERVER_LOG" >&2 2>/dev/null || true
}
fail() { echo "FAIL: $*" >&2; dump_logs; exit 1; }

# The name must match ^[a-zA-Z0-9_-]+$ (validation.ValidateAgentData) — no spaces,
# no dots. $$ keeps it unique to this run so an assertion can only pass on a row
# THIS run created.
AGENT_NAME="e2e-agent-$$"

# boot <label> [env assignments...] — launch the server, poll /health, leave it in
# $SERVER_PID. Returns the /health body in $HEALTH, and its status code in $HCODE.
# Polls; never sleeps and hopes. Aborts the instant the pid dies, which is how a
# route-registration panic surfaces as a named failure instead of a mystery timeout.
#
# ⚠ THE BINARY IS LAUNCHED DIRECTLY, NOT INSIDE A `( … ) &` SUBSHELL. This script
# boots the server twice on the SAME port, so phase 1 must be genuinely dead before
# phase 2 binds. With a subshell, `$!` is the SUBSHELL's pid — killing it reaps the
# wrapper and leaves the real server orphaned and still holding the port. Phase 2
# then fails to bind and dies, and every subsequent assertion is silently answered
# by the STALE phase-1 process, which happily reports database=disconnected. That
# is not a hypothetical: it is what the first draft of this script did, and it read
# as an inber-party bug rather than a test bug. Keep the launch un-nested so $! is
# the server itself.
boot() {
  local label="$1"; shift
  : >"$SERVER_LOG"
  cd "$RUN_DIR"
  env -i PATH="$PATH" HOME="$FAKE_HOME" PORT="$PORT" \
      INBER_URL="http://127.0.0.1:1" "$@" "$BIN_DIR/inber-party" >"$SERVER_LOG" 2>&1 &
  SERVER_PID=$!
  HEALTH=""; HCODE=""
  local i
  for i in $(seq 1 75); do
    if ! kill -0 "$SERVER_PID" 2>/dev/null; then
      fail "[$label] inber-party exited during startup (route registration panic? migration fatal? port in use?)"
    fi
    # -f would swallow the body on the 503 that phase 1 EXPECTS, so read the code
    # and the body separately and assert on both.
    HCODE=$(curl -sS -o "$TMP_DIR/health.json" -w '%{http_code}' --max-time 2 "$BASE/health" 2>/dev/null || echo "")
    if [ -n "$HCODE" ] && [ "$HCODE" != "000" ]; then
      HEALTH=$(cat "$TMP_DIR/health.json")
      break
    fi
    HCODE=""
    sleep 0.2
  done
  [ -n "$HCODE" ] || fail "[$label] inber-party did not answer $BASE/health within 15s"
}

# Kill phase 1 and WAIT for the port to actually come free. A dead pid does not
# mean a closed listener: the socket lingers for a moment after exit, and phase 2
# binding into that window dies with "address already in use". Poll, don't sleep.
stop_server() {
  if [ -n "$SERVER_PID" ] && kill -0 "$SERVER_PID" 2>/dev/null; then
    kill "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
  SERVER_PID=""
  local i
  for i in $(seq 1 50); do
    curl -sS -o /dev/null --max-time 1 "$BASE/health" 2>/dev/null || return 0
    sleep 0.2
  done
  fail "something is STILL answering on :$PORT after phase 1 was killed — phase 2 would be asserting against a stale process"
}

step "build cmd/server from $REPO_DIR"
cd "$REPO_DIR"
# Explicit -o: a bare `go build ./cmd/server` writes ./server into the CWD.
# CGO stays ENABLED on purpose — although the primary store is Postgres via the
# pure-Go lib/pq, mattn/go-sqlite3 is imported by non-test production code
# (internal/inber/inber.go:17, internal/bounty/repository.go:10), so CGO_ENABLED=0
# does not build. Better to fail at the build than ship a binary that dies later.
go build -o "$BIN_DIR/inber-party" ./cmd/server
echo "    inber-party: $(ls -lh "$BIN_DIR/inber-party" | awk '{print $5}')"

################################################################################
step "PHASE 1 — DB-less boot: the binary must come up and answer, degraded"
################################################################################
# No DATABASE_URL at all. db.Connect is never called; `database` stays nil and the
# server runs in what main.go calls "inber-only mode". This is the hermetic half:
# no docker, no network, no daemons.
boot "phase1"

[ "$HCODE" = "503" ] \
  || fail "[phase1] /health returned $HCODE, expected 503 — with neither Postgres nor inber reachable, overall status must be 'unhealthy' (main.go:246-283). If this is now 200, the health contract changed."
[ "$(jq -r '.status' <<<"$HEALTH")" = "unhealthy" ] || fail "[phase1] /health status != unhealthy: $HEALTH"
[ "$(jq -r '.services.database.status' <<<"$HEALTH")" = "disconnected" ] || fail "[phase1] database is not reported disconnected: $HEALTH"
# The websocket hub is the one subsystem that must be alive even with no DB. If it
# is not, the process is up but structurally broken.
[ "$(jq -r '.services.websocket.status' <<<"$HEALTH")" = "running" ] || fail "[phase1] the websocket hub is not running: $HEALTH"
echo "    503 unhealthy, database=disconnected, websocket=running"

step "PHASE 1 — GET /api/agents must be 200 [] (the nil-DB guard, not a crash)"
AGENTS=$(curl -fsS --max-time 10 "$BASE/api/agents") || fail "[phase1] GET /api/agents failed — the nil-DB guard (api.go:230-235) is missing or panicked"
[ "$(jq -r 'type' <<<"$AGENTS")" = "array" ] || fail "[phase1] /api/agents is not an array: $AGENTS"
[ "$(jq -r 'length' <<<"$AGENTS")" = "0" ] || fail "[phase1] /api/agents is not empty without a DB: $AGENTS"
echo "    200 []"

step "PHASE 1 — POST /api/agents must be REFUSED with 503, not silently accepted"
# Drive the failure branch. A write path that appears to succeed with no database
# behind it is far worse than one that refuses: it would look exactly like a
# working service while persisting nothing.
CODE=$(curl -sS -o /dev/null -w '%{http_code}' --max-time 10 -X POST "$BASE/api/agents" \
  -H 'Content-Type: application/json' \
  -d "{\"name\":\"$AGENT_NAME\",\"class\":\"rogue\"}")
[ "$CODE" = "503" ] || fail "[phase1] POST /api/agents returned $CODE with NO database configured — expected 503. A write that is accepted without a DB persists nothing."
echo "    503 (write correctly refused)"

stop_server
echo "    phase 1 complete — binary boots, routes register, degraded contract holds"

################################################################################
if [ "${E2E_SKIP_POSTGRES:-}" = "1" ]; then
  echo ""
  echo "==> PHASE 2 SKIPPED (E2E_SKIP_POSTGRES=1)"
  echo "    The migrations, the seed and the entire write path were NOT exercised."
  echo "    This run proved only that the binary boots and serves its degraded contract."
  printf '\nsmoke test OK (inber-party boots; PHASE 2 SKIPPED — postgres path unverified)\n'
  exit 0
fi

step "PHASE 2 — start a THROWAWAY PostgreSQL in docker (never a shared one)"
################################################################################
if ! command -v docker >/dev/null 2>&1; then
  echo "ERROR: docker not found on PATH." >&2
  echo "       inber-party's write path REQUIRES a real PostgreSQL (POST /api/agents" >&2
  echo "       503s without one, and the dialect is Postgres-specific so SQLite cannot" >&2
  echo "       stand in). This host has no postgres binaries — initdb, pg_ctl, postgres" >&2
  echo "       and pg_isready are all absent — so a container is the only way to give" >&2
  echo "       this binary the database it actually runs against." >&2
  echo "       Failing rather than skipping: a green run that never touched the" >&2
  echo "       migrations would be indistinguishable from one that verified them." >&2
  echo "       Set E2E_SKIP_POSTGRES=1 to deliberately accept that gap." >&2
  exit 2
fi
if ! docker info >/dev/null 2>&1; then
  echo "ERROR: the docker daemon is not reachable (docker info failed)." >&2
  echo "       See the note above: this is a DOCKER failure, not an inber-party one." >&2
  echo "       Set E2E_SKIP_POSTGRES=1 to run the boot-only half." >&2
  exit 2
fi

PG_USER=postgres
PG_PASS="e2e-$$"
PG_DB=inberparty_e2e
docker rm -f "$PG_CONTAINER" >/dev/null 2>&1 || true
docker run -d --rm --name "$PG_CONTAINER" \
  -e POSTGRES_PASSWORD="$PG_PASS" \
  -e POSTGRES_USER="$PG_USER" \
  -e POSTGRES_DB="$PG_DB" \
  -p "127.0.0.1:$PG_PORT:5432" \
  "$PG_IMAGE" >/dev/null \
  || fail "could not start $PG_IMAGE (is the image pulled? port $PG_PORT free?)"
PG_STARTED=1
echo "    container: $PG_CONTAINER ($PG_IMAGE) on 127.0.0.1:$PG_PORT"

# Poll with the container's OWN psql over TCP+auth — the host has no client.
# CRITICAL: gate on the exact path the app uses (a real TCP connect + auth +
# query), NOT the unix socket. The postgres entrypoint runs a temporary
# socket-only server during initdb, so socket `pg_isready` reports "ready" ~250ms
# before the final server accepts external TCP. The app pings the DB exactly once
# at boot (main.go:43) and falls into permanent degraded mode on failure, so a
# ping that lands in that gap dies with "connection reset by peer" and the whole
# phase-2 health assertion fails — which is the flake this smoke exhibited. A
# SELECT over TCP is ready only when the real path the app takes is ready.
pg_up=0
for _ in $(seq 1 60); do
  if docker exec -e PGPASSWORD="$PG_PASS" "$PG_CONTAINER" \
       psql -h 127.0.0.1 -p 5432 -U "$PG_USER" -d "$PG_DB" -tAc 'select 1' >/dev/null 2>&1; then
    pg_up=1; break
  fi
  sleep 0.5
done
[ "$pg_up" = "1" ] || fail "the throwaway postgres never became ready within 30s"
echo "    postgres ready (TCP + auth + query)"

step "PHASE 2 — boot against it: migrations and seed must run (both are Fatal)"
DSN="postgres://$PG_USER:$PG_PASS@127.0.0.1:$PG_PORT/$PG_DB?sslmode=disable"
boot "phase2" DATABASE_URL="$DSN"

# Now the DB is reachable and inber is not, so overall must be healthy (200).
# A 206/"degraded" here would mean Ping succeeded but something else errored.
[ "$HCODE" = "200" ] \
  || fail "[phase2] /health returned $HCODE, expected 200 — with Postgres up, overall status must be 'healthy': $HEALTH"
[ "$(jq -r '.status' <<<"$HEALTH")" = "healthy" ] || fail "[phase2] /health status != healthy: $HEALTH"
[ "$(jq -r '.services.database.status' <<<"$HEALTH")" = "connected" ] || fail "[phase2] database is not reported connected: $HEALTH"
echo "    200 healthy, database=connected (migrations + seed survived boot)"

step "PHASE 2 — POST /api/agents — write into Postgres"
CREATED=$(curl -fsS --max-time 15 -X POST "$BASE/api/agents" \
  -H 'Content-Type: application/json' \
  -d "{\"name\":\"$AGENT_NAME\",\"title\":\"Smoke Tester\",\"class\":\"rogue\",\"avatar_emoji\":\"🧪\"}") \
  || fail "[phase2] POST /api/agents failed — this is the path that 503'd in phase 1, so a failure here is the DB, not the guard"
NEW_ID=$(jq -r '.id' <<<"$CREATED")
case "$NEW_ID" in
  ''|null|*[!0-9]*) fail "[phase2] POST /api/agents returned no numeric id: $CREATED" ;;
esac
echo "    created agent id=$NEW_ID name=$AGENT_NAME"

step "PHASE 2 — GET /api/agents — read it back out of Postgres"
# The seed (db.Seed(), main.go:53) populates rows at boot, so this list is NOT
# empty and its length is NOT a meaningful assertion. Select OUR row by name.
AGENTS=$(curl -fsS --max-time 10 "$BASE/api/agents")
[ "$(jq -r 'type' <<<"$AGENTS")" = "array" ] || fail "[phase2] /api/agents is not an array: $AGENTS"
MINE=$(jq -c --arg n "$AGENT_NAME" '.[] | select(.name == $n)' <<<"$AGENTS")
[ -n "$MINE" ] || fail "[phase2] the agent just written is not in /api/agents: $AGENTS"
[ "$(jq -r '.id'    <<<"$MINE")" = "$NEW_ID" ] || fail "[phase2] read-back id != $NEW_ID: $MINE"
[ "$(jq -r '.class' <<<"$MINE")" = "rogue"   ] || fail "[phase2] read-back class != rogue: $MINE"
[ "$(jq -r '.title' <<<"$MINE")" = "Smoke Tester" ] || fail "[phase2] read-back title wrong: $MINE"
# These come from the COLUMN DEFAULTS the migration declared, not from our request
# body — so they only hold if Migrate() really created the table it claims to.
[ "$(jq -r '.gold'   <<<"$MINE")" = "100" ] || fail "[phase2] gold default (100) did not apply — check the migration: $MINE"
[ "$(jq -r '.level'  <<<"$MINE")" = "1"   ] || fail "[phase2] level default (1) did not apply — check the migration: $MINE"
[ "$(jq -r '.status' <<<"$MINE")" = "idle" ] || fail "[phase2] status default ('idle') did not apply — this value comes from the DB column default, so the migration did not create the table it claims: $MINE"
echo "    read back: $MINE"

step "PHASE 2 — the seed really ran (the list holds more than just our row)"
TOTAL=$(jq -r 'length' <<<"$AGENTS")
[ "$TOTAL" -gt 1 ] 2>/dev/null \
  || fail "[phase2] /api/agents holds only $TOTAL row(s) — db.Seed() (main.go:53) did not populate anything, so the boot path silently skipped it"
echo "    $TOTAL agents total (ours + seed)"

step "process still alive after serving every route"
kill -0 "$SERVER_PID" 2>/dev/null || fail "inber-party died while serving"

printf '\nsmoke test OK (inber-party boots degraded AND against a real postgres, on :%s)\n' "$PORT"
