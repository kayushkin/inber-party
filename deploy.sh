#!/usr/bin/env bash
set -euo pipefail

# Builds inber-party (frontend + backend), installs the binary AND the systemd
# user unit, restarts the service, and refuses to report success unless the thing
# actually answers.
#
# WHY THIS EXISTS: until now inber-party had no deploy path at all. Its binary was
# hand-built and hand-copied, and by the time this script was written the RUNNING
# binary was 229 commits behind the repo's own committed HEAD -- a gap nobody chose,
# nobody could see, and nobody could close without remembering the steps. That is the
# same state forge was in when its binary sat 4 months behind (six shipped fixes were
# stranded in the repo the whole time), and healthcheck's when its sat 2.5 months
# stale. "There is no deploy path" is what that costs.
#
# THIS IS A USER UNIT, NOT A SYSTEM ONE. inber-party.service lives in
# ~/.config/systemd/user and runs inside user@1000.service -- so everything here is
# `systemctl --user`, and NOTHING here needs sudo. (The sibling argraphments/deploy.sh
# this is modelled on is a system-unit deploy and sudos throughout; copying it
# wholesale would have you editing /etc/systemd/system, where inber-party is
# LoadState=not-found -- which reads exactly like a service that does not exist.)
#
# The unit template in this repo is the single source of truth and deploying
# installs it, so the live unit cannot drift from the template without the next
# deploy pulling it back.

REPO_DIR="$(cd "$(dirname "$0")" && pwd)"
UNIT_NAME="inber-party.service"
REPO_UNIT="$REPO_DIR/deploy/$UNIT_NAME"
LIVE_UNIT="$HOME/.config/systemd/user/$UNIT_NAME"
BACKUP_DIR="$HOME/.local/share/inber-party-deploy-backup"
STAMP="$(date +%Y%m%dT%H%M%S)"

cd "$REPO_DIR"

# systemctl --user talks to the user manager over DBus, and from a non-login shell
# (an agent, a cron job) those two variables are unset -- in which case systemctl
# returns EMPTY OUTPUT AND EXIT 0 rather than an error. Silence then reads as "the
# unit isn't there", which is indistinguishable from a real finding. Set them.
export XDG_RUNTIME_DIR="${XDG_RUNTIME_DIR:-/run/user/$(id -u)}"
export DBUS_SESSION_BUS_ADDRESS="${DBUS_SESSION_BUS_ADDRESS:-unix:path=${XDG_RUNTIME_DIR}/bus}"

# go, node and npm exist ONLY under mise on this host (there is no /usr/bin/node at
# all), and a stripped environment -- systemd, the scheduler -- carries HOME but no
# PATH. Resolve the real binaries rather than trusting the caller's PATH, and fail
# loudly naming the tool if we cannot. Prefer `mise which` over the shims: a shim
# re-reads mise config from the CWD and refuses to run from an untrusted path.
MISE="$HOME/.local/bin/mise"
resolve_tool() {
  local tool="$1" path=""
  if [ -x "$MISE" ]; then path="$("$MISE" which "$tool" 2>/dev/null || true)"; fi
  [ -n "$path" ] || path="$(command -v "$tool" 2>/dev/null || true)"
  [ -n "$path" ] || { echo "ERROR: cannot resolve '$tool' (not via mise, not on PATH)"; exit 1; }
  printf '%s\n' "$path"
}
GO="$(resolve_tool go)"
NPM="$(resolve_tool npm)"
NODE="$(resolve_tool node)"
export PATH="$(dirname "$NODE"):$PATH"   # vite/tsc shell out to node

[ -f "$REPO_UNIT" ] || { echo "ERROR: no unit template at $REPO_UNIT"; exit 1; }

# Everything the deploy needs is already declared in the unit, so read it from
# there rather than restating it here -- a second copy of a value is a second thing
# to drift. (A sibling repo's deploy.sh hardcoded a BIN_NAME it had inherited from
# the repo it was cloned from, and would have overwritten another live service's
# binary.) %h is systemd's specifier for the user's home; expand it the same way
# systemd will.
unit_field() { grep -E "^$1=" "$REPO_UNIT" | head -1 | cut -d= -f2- | sed "s|%h|$HOME|g"; }
BINARY_PATH="$(unit_field ExecStart)"
BINARY_NAME="$(basename "$BINARY_PATH")"
WORK_DIR="$(unit_field WorkingDirectory)"
PORT="$(grep -E '^Environment=PORT=' "$REPO_UNIT" | head -1 | cut -d= -f3)"

echo "==> Unit declares:"
echo "    ExecStart        = $BINARY_PATH"
echo "    WorkingDirectory = $WORK_DIR"
echo "    PORT             = $PORT"

# Preflight the directory systemd must chdir into. A WorkingDirectory that does not
# exist fails the start with an opaque `200/CHDIR`, and with Restart= set that
# becomes a silent crash loop -- three services on this box ran up 1.8 MILLION
# failed starts that way before anyone looked. Check it here, where we can name it.
echo "==> Preflight..."
[ -d "$WORK_DIR" ] || { echo "ERROR: WorkingDirectory $WORK_DIR does not exist -- systemd cannot chdir and will crash-loop"; exit 1; }
# The unit serves the UI from $WORK_DIR/frontend/dist (a relative http.Dir), so the
# checkout the unit points at must be the one we are building in. If it is not, this
# script would build a frontend nobody serves and "succeed".
[ "$WORK_DIR" = "$REPO_DIR" ] || { echo "ERROR: unit's WorkingDirectory ($WORK_DIR) is not this checkout ($REPO_DIR) -- the frontend built here would not be the one served"; exit 1; }
echo "    WorkingDirectory exists and is this checkout"

echo "==> Building frontend..."
# Build into a staging dir and swap, rather than letting vite empty and refill the
# live dist/ in place: dist/ IS the directory nginx-fronted traffic is being served
# from right now, and an in-place rebuild serves a half-written site for the seconds
# it takes. Staging also gives the rollback below something to restore.
FRONTEND_DIST="$REPO_DIR/frontend/dist"
STAGING_DIST="$REPO_DIR/frontend/.dist-staging"
rm -rf "$STAGING_DIST"
(
  cd frontend
  # ALWAYS `npm ci`, never "npm ci only if node_modules is missing". A present
  # node_modules is not a correct one: this checkout had one, four months stale,
  # missing react-icons entirely -- a dependency that has been in package.json AND
  # in the lockfile the whole time. The build failed with "Cannot find module
  # 'react-icons/fa'", which reads like a broken tree and is really a broken
  # sandbox. npm ci installs exactly the lockfile and removes anything else, so the
  # deploy's dependency set stops depending on what happens to be lying around.
  "$NPM" ci
  "$NPM" run build -- --outDir "$STAGING_DIST" --emptyOutDir
)
# Assert the build actually produced a site. `npm run build` can exit 0 having
# written nothing useful, and an empty dist/ is served as 404s, not as an error.
[ -f "$STAGING_DIST/index.html" ] || { echo "ERROR: frontend build produced no index.html"; exit 1; }
ASSET_COUNT="$(find "$STAGING_DIST/assets" -type f 2>/dev/null | wc -l)"
[ "$ASSET_COUNT" -gt 0 ] || { echo "ERROR: frontend build produced no assets"; exit 1; }
echo "    frontend built: index.html + $ASSET_COUNT assets"

echo "==> Building $BINARY_NAME..."
# CGO must stay ENABLED. inber-party reads inber's session DBs through
# mattn/go-sqlite3, which is a cgo driver: built with CGO_ENABLED=0 it still
# compiles, still links, still vets green -- and then every sql.Open("sqlite3")
# fails at RUNTIME with "requires cgo to work". A tree that builds into a binary
# that cannot do its job is the exact failure this repo's smoke test exists to
# catch; do not "harden" this by turning cgo off.
CGO_ENABLED=1 "$GO" build -ldflags "-X github.com/kayushkin/inber-party/internal/version.Version=$(git describe --tags --always --dirty 2>/dev/null || echo dev) -X github.com/kayushkin/inber-party/internal/version.GitCommit=$(git rev-parse --short HEAD) -X github.com/kayushkin/inber-party/internal/version.BuildTime=$STAMP" -o "$BINARY_NAME" ./cmd/server
echo "    built: $(ls -lh "$BINARY_NAME" | awk '{print $5}')"

# Provenance. Asserted here, immediately after the build and long before the
# backups, the unit install or the binary rename -- nothing destructive has
# happened yet at this point. `go build` writes no VCS stamp when it cannot find
# a .git DIRECTORY, and it does not fail when that happens, not even with
# -buildvcs=true; the usual cause is building from a git worktree, whose .git is
# a pointer file. Such a binary compiles, links, vets green and reads clean in
# the log -- it simply cannot be traced back to a commit, which is the whole
# point of the ldflags stamp two lines above.
#
# `$GO` and not a bare `go`: this script resolves its own toolchain, and a bare
# `go` here would read the stamp with a different binary than the one that wrote
# it -- or with none at all, failing the deploy for the wrong reason.
echo "==> Checking provenance..."
buildinfo="$("$GO" version -m "$BINARY_NAME")"
vcs_revision="$(printf '%s\n' "$buildinfo" | awk -F= '$1 ~ /[[:space:]]vcs\.revision$/ {print $2}')"
vcs_modified="$(printf '%s\n' "$buildinfo" | awk -F= '$1 ~ /[[:space:]]vcs\.modified$/ {print $2}')"
if [ -z "$vcs_revision" ]; then
  echo "    REFUSING TO INSTALL: this binary carries no vcs.revision, so nothing can tie" >&2
  echo "    it back to a commit. Build from a real clone or checkout, not a worktree." >&2
  exit 1
fi
echo "    vcs.revision=$vcs_revision"
if [ "$vcs_modified" = "true" ]; then
  echo "    WARNING: built from a DIRTY tree (vcs.modified=true). $vcs_revision names the" >&2
  echo "    commit this binary was built NEAR, not the source it was built FROM, and that" >&2
  echo "    source is not recoverable from any commit. Commit first for a reproducible build." >&2
fi

# Gate on vet, not `go test ./...`: several suites here want a live Postgres and
# would go red on a perfectly good tree, and a deploy gate that fails for reasons
# unrelated to the deploy is one people learn to ignore. vet still typechecks the
# test files, so a test that stops compiling is caught here.
echo "==> Vetting..."
"$GO" vet ./...
echo "    vet clean"

# --- Backups. Timestamped, never a fixed filename. -----------------------------
# Verification necessarily runs AFTER the new artifacts are installed -- that is
# the point of it -- so the backup is the only way back. A fixed-filename backup is
# NOT a rollback point: the second run overwrites the good copy with the bad one,
# and the one thing you needed is gone. (A sibling repo's deploy.sh nearly
# destroyed the only existing copy of an irreplaceable binary exactly this way.)
mkdir -p "$BACKUP_DIR"
BACKUP_BIN="$BACKUP_DIR/$BINARY_NAME.$STAMP"
BACKUP_UNIT="$BACKUP_DIR/$UNIT_NAME.$STAMP"
BACKUP_DIST="$BACKUP_DIR/dist.$STAMP"
# Written as if/then and not `[ -f x ] && cp ...`: in the one-liner form a FAILING cp
# is not the final command of the && list, so `set -e` exempts it and the backup
# silently does not happen -- and we would find that out at rollback time, which is
# the worst possible moment to learn a backup is missing.
if [ -f "$BINARY_PATH" ]; then cp -p "$BINARY_PATH" "$BACKUP_BIN"; echo "==> Backed up binary -> $BACKUP_BIN"; fi
if [ -f "$LIVE_UNIT" ]; then cp -p "$LIVE_UNIT" "$BACKUP_UNIT"; echo "==> Backed up unit   -> $BACKUP_UNIT"; fi
if [ -d "$FRONTEND_DIST" ]; then cp -a "$FRONTEND_DIST" "$BACKUP_DIST"; echo "==> Backed up dist   -> $BACKUP_DIST"; fi

rollback() {
  echo "!!! Rolling back..."
  if [ -f "$BACKUP_BIN" ]; then cp -p "$BACKUP_BIN" "$BINARY_PATH.rb"; mv "$BINARY_PATH.rb" "$BINARY_PATH"; echo "    restored binary"; fi
  if [ -f "$BACKUP_UNIT" ]; then cp -p "$BACKUP_UNIT" "$LIVE_UNIT"; echo "    restored unit"; fi
  if [ -d "$BACKUP_DIST" ]; then rm -rf "$FRONTEND_DIST"; cp -a "$BACKUP_DIST" "$FRONTEND_DIST"; echo "    restored frontend"; fi
  systemctl --user daemon-reload
  systemctl --user restart "$UNIT_NAME" || true
  echo "!!! Rolled back to the previous binary/unit/frontend. Deploy FAILED."
}

echo "==> Installing unit -> $LIVE_UNIT..."
mkdir -p "$(dirname "$LIVE_UNIT")"
if cmp -s "$REPO_UNIT" "$LIVE_UNIT" 2>/dev/null; then
  echo "    live unit already matches the template"
else
  echo "    live unit differs from the template (or is absent) -- installing:"
  diff "$LIVE_UNIT" "$REPO_UNIT" 2>/dev/null | sed 's/^/      /' || true
  cp "$REPO_UNIT" "$LIVE_UNIT"
fi
systemctl --user daemon-reload

echo "==> Installing frontend -> $FRONTEND_DIST..."
rm -rf "$FRONTEND_DIST"
mv "$STAGING_DIST" "$FRONTEND_DIST"

echo "==> Installing binary -> $BINARY_PATH..."
# Stage next to the target and rename into place. A plain `cp` over the live binary
# fails with ETXTBSY ("text file busy") because the running service still holds it
# open for execution; rename(2) is atomic, swaps the inode, and leaves the running
# process on the old one until the restart below.
cp "$BINARY_NAME" "$BINARY_PATH.new"
chmod 755 "$BINARY_PATH.new"
mv "$BINARY_PATH.new" "$BINARY_PATH"

echo "==> Restarting $UNIT_NAME..."
systemctl --user reset-failed "$UNIT_NAME" 2>/dev/null || true
systemctl --user restart "$UNIT_NAME"

# Poll for readiness; do not sleep and hope. A fixed sleep either races a slow start
# (and then exits 1 on a good binary it has already installed) or pads every good
# deploy. `systemctl is-active` is NOT readiness either -- it says a process exists,
# not that the binary answers.
echo "==> Waiting for :$PORT to answer..."
READY_TIMEOUT=45
for i in $(seq 1 "$READY_TIMEOUT"); do
  if curl -s -o /dev/null --max-time 5 "http://127.0.0.1:$PORT/health"; then
    echo "    ready after ${i}s"
    break
  fi
  if ! systemctl --user is-active --quiet "$UNIT_NAME"; then
    echo "ERROR: $UNIT_NAME died while starting up"
    journalctl --user -u "$UNIT_NAME" -n 30 --no-pager || true
    rollback
    exit 1
  fi
  if [ "$i" -eq "$READY_TIMEOUT" ]; then
    echo "ERROR: $UNIT_NAME still not answering :$PORT after ${READY_TIMEOUT}s"
    journalctl --user -u "$UNIT_NAME" -n 30 --no-pager || true
    rollback
    exit 1
  fi
  sleep 1
done

echo "==> Smoke test..."
fail() { echo "ERROR: $1"; journalctl --user -u "$UNIT_NAME" -n 30 --no-pager || true; rollback; exit 1; }

# /health reports the two data sources separately and grades itself: "healthy",
# "degraded" (206) or "unhealthy" (503). Assert on the GRADE, not on a 200 -- and
# specifically refuse "unhealthy", which means database AND inber are both
# disconnected. Postgres is legitimately absent here, so inber is the only source
# this deployment has; losing it is a real regression that still binds the port and
# still serves the UI, i.e. exactly the kind of failure a "did the port open?" check
# waves through.
HEALTH="$(curl -s --max-time 10 "http://127.0.0.1:$PORT/health")"
# `jq -r .status` and not a grep: a grep for "healthy" also matches "unhealthy".
STATUS="$(printf '%s' "$HEALTH" | jq -r '.status // ""' 2>/dev/null || true)"
case "$STATUS" in
  healthy|degraded) echo "    /health: $STATUS" ;;
  unhealthy)        fail "/health reports 'unhealthy' -- database AND inber are both disconnected; the deployed binary cannot see any data source. Body: ${HEALTH:0:300}" ;;
  *)                fail "/health did not report a recognisable status; got: ${HEALTH:0:300}" ;;
esac

# Prove the deployed binary can actually reach a store and serialise from it. "The
# port is open" only proves a process bound it.
AGENTS="$(curl -fsS --max-time 10 "http://127.0.0.1:$PORT/api/agents")"
case "$AGENTS" in
  '['*) echo "    /api/agents returned a JSON array" ;;
  *)    fail "/api/agents did not return a JSON array; got: ${AGENTS:0:200}" ;;
esac

# The frontend is served by the Go binary out of WorkingDirectory/frontend/dist. If
# the chdir or the dist swap went wrong, THIS is what catches it -- the API would be
# perfectly healthy while the site served 404 for its own index.
INDEX_CODE="$(curl -s -o /dev/null -w '%{http_code}' --max-time 10 "http://127.0.0.1:$PORT/")"
[ "$INDEX_CODE" = "200" ] || fail "the frontend index did not serve (HTTP $INDEX_CODE) -- the binary is up but the site is not"
echo "    frontend index served (HTTP 200)"

# is-active is the last word: a unit that is up now but scheduled to restart is not
# a successful deploy.
systemctl --user is-active --quiet "$UNIT_NAME" || fail "$UNIT_NAME is not active"
RESTARTS="$(systemctl --user show "$UNIT_NAME" -p NRestarts --value)"
[ "$RESTARTS" = "0" ] || fail "$UNIT_NAME has restarted $RESTARTS time(s) since the deploy -- it is crash-looping, not running"
echo "    $UNIT_NAME active, NRestarts=$RESTARTS"
echo "    smoke test OK"

# The vhost is what puts party.kayushkin.com on the internet. It is part of the
# deploy, not something to hand-edit in /etc/nginx -- a vhost no repo tracks
# survives only as long as this box does.
echo "==> Installing nginx vhost..."
./deploy/nginx/install.sh

echo "==> Done. Deployed $(git rev-parse --short HEAD) ($(git rev-list --count HEAD) commits)."
