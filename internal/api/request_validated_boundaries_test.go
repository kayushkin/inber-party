package api

// Boundary VALUES that api.go's request handlers validate before they query.
//
// Card 98ebdd04-0b0a-4eb7-9591-25b10906efa1, worked by the 211th nightly pass.
// Onboarding note e72c6e27. Scored by scripts/sabotage-request-boundaries.py.
//
// The sibling files in this package pin what happens to text once a handler has
// decided to act: spawn_message_test.go and spawn_message_budget_test.go hold
// the spawn message's 100-byte cut. This file pins the numbers a handler
// compares a REQUEST against — the pagination defaults and caps, the date-format
// length, and the required-skills ceiling. Before it existed, api_test.go sent
// no query string at all, so not one numeric axis in this file was varied.
//
// # Three things measured here that the card got wrong, and one it warned about
//
//  1. The card scoped api.go's bounty tier ladder (":2536-2540, probably
//     reachable today"). It is reachable and it is DEAD. See
//     TestTheApiTierLadderIsOverwrittenBeforeAnyoneReadsIt below.
//  2. The card said "api.go:2570 encodes the api.go tier into the response
//     while the row stores the db tier". Measured: response and row both carry
//     the DB ladder. They never disagree, because there is only one ladder left
//     by the time either is written.
//  3. The card's trap 6 — "`parsed > 0` guards are zero guards, not
//     boundaries: a bad limit silently keeps the default, so a mutated default
//     and a rejected parameter are indistinguishable. Vary the valid limit."
//     Honoured: every default here is asserted with NO limit parameter, and
//     separately with a valid one, so the two cannot be confused.
//
// # What is deliberately NOT here
//
//   - logstack's own `if limit <= 0 { limit = 20 }` (logstack.go:95) is a second
//     spelling of 20 that this handler can never reach, because
//     handleConversations only calls through with a positive limit. It belongs
//     to the logstack child of this sweep, not to this file.
//   - validation.ValidateLimit's own mechanics are already held by
//     validation_test.go:460. What that test does NOT hold is the ARGUMENTS the
//     two call sites pass, which are the numbers that ship. Those are pinned
//     here, through the handler. (The 184th's rule: a parameter every test
//     supplies is not the value that ships.)

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kayushkin/inber-party/internal/db"
	"github.com/kayushkin/inber-party/internal/logstack"
	"github.com/kayushkin/inber-party/internal/ws"

	_ "github.com/mattn/go-sqlite3"
)

// boundarySchema is deliberately NOT createTestServer's schema.
//
// createTestServer (api_test.go:20) defines two tables, and four sibling tests
// assert exact counts against them. The 206th pass's rule: a fixture shared by
// count-asserting tests is not a place to add rows, and this file inserts
// upwards of a thousand. It also needs three columns createTestServer's
// `bounties` omits — work_submission, verification_notes and submitted_at,
// all three of which GetBounties SELECTs — plus a `cost_entries` table that
// does not exist there at all. A local builder costs thirty lines.
const boundarySchema = `
CREATE TABLE bounties (
	id                 INTEGER PRIMARY KEY AUTOINCREMENT,
	title              TEXT NOT NULL,
	description        TEXT NOT NULL,
	requirements       TEXT NOT NULL DEFAULT '',
	payout_amount      INTEGER NOT NULL,
	status             TEXT NOT NULL DEFAULT 'open',
	deadline           TIMESTAMP,
	creator_id         INTEGER NOT NULL,
	claimer_id         INTEGER,
	work_submission    TEXT,
	verification_notes TEXT,
	required_skills    TEXT,
	tier               TEXT NOT NULL,
	claimed_at         TIMESTAMP,
	submitted_at       TIMESTAMP,
	completed_at       TIMESTAMP,
	created_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE cost_entries (
	id             INTEGER PRIMARY KEY AUTOINCREMENT,
	agent_id       INTEGER,
	task_id        INTEGER,
	session_id     TEXT NOT NULL DEFAULT '',
	tokens_used    INTEGER NOT NULL DEFAULT 0,
	cost_usd       REAL NOT NULL DEFAULT 0,
	model_name     TEXT NOT NULL DEFAULT '',
	operation_type TEXT NOT NULL DEFAULT '',
	date           TEXT NOT NULL DEFAULT '',
	created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	metadata       TEXT
);
CREATE TABLE disputes (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	bounty_id   INTEGER NOT NULL,
	claimer_id  INTEGER NOT NULL,
	creator_id  INTEGER NOT NULL,
	reason      TEXT NOT NULL DEFAULT '',
	evidence    TEXT NOT NULL DEFAULT '',
	status      TEXT NOT NULL DEFAULT 'open',
	admin_notes TEXT,
	resolution  TEXT,
	resolved_by INTEGER,
	resolved_at TIMESTAMP,
	created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

func newBoundaryServer(t *testing.T) *Server {
	t.Helper()

	raw, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "boundaries.db"))
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { raw.Close() })

	if _, err := raw.Exec(boundarySchema); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return NewServer(&db.DB{DB: raw}, ws.NewHub(), nil, nil)
}

// seedBounties inserts count rows in one transaction. `requirements` is NOT
// NULL on purpose: GetBounties scans it into a plain string, so a NULL row
// makes every list request 500 and every count assertion below trivially zero.
func seedBounties(t *testing.T, s *Server, count int) {
	t.Helper()

	tx, err := s.DB.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	for i := 0; i < count; i++ {
		_, err := tx.Exec(
			`INSERT INTO bounties (title, description, requirements, payout_amount, creator_id, tier)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			fmt.Sprintf("bounty %d", i), "description", "requirements", 10, 1, "bronze")
		if err != nil {
			tx.Rollback()
			t.Fatalf("seed bounty %d: %v", i, err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
}

func seedDisputes(t *testing.T, s *Server, count int) {
	t.Helper()

	tx, err := s.DB.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	for i := 0; i < count; i++ {
		_, err := tx.Exec(
			`INSERT INTO disputes (bounty_id, claimer_id, creator_id, reason, evidence)
			 VALUES (?, ?, ?, ?, ?)`,
			i+1, 1, 2, "reason", "evidence")
		if err != nil {
			tx.Rollback()
			t.Fatalf("seed dispute %d: %v", i, err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
}

func seedCostEntries(t *testing.T, s *Server, count int) {
	t.Helper()

	tx, err := s.DB.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	for i := 0; i < count; i++ {
		_, err := tx.Exec(
			`INSERT INTO cost_entries (session_id, tokens_used, cost_usd, model_name, operation_type, date)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			fmt.Sprintf("session-%d", i), 100, 0.5, "claude", "task", "2026-01-01")
		if err != nil {
			tx.Rollback()
			t.Fatalf("seed cost entry %d: %v", i, err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
}

// seedSessionFiles writes count parseable JSONL session logs for one agent and
// returns a LogstackClient rooted at the temp directory holding them.
//
// parseSessionFile rejects a file with no `session` event or no non-empty
// message, so both lines are required for the file to be counted at all.
func seedSessionFiles(t *testing.T, agentID string, count int) *logstack.LogstackClient {
	t.Helper()

	root := t.TempDir()
	sessionsDir := filepath.Join(root, agentID, "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatalf("create sessions dir: %v", err)
	}

	for i := 0; i < count; i++ {
		lines := fmt.Sprintf(
			`{"type":"session","id":"session-%d","timestamp":"2026-01-01T00:00:00Z"}`+"\n"+
				`{"type":"message","id":"message-%d","timestamp":"2026-01-01T00:00:01Z",`+
				`"message":{"role":"user","content":[{"type":"text","text":"hello"}]}}`+"\n", i, i)
		path := filepath.Join(sessionsDir, fmt.Sprintf("session-%d.jsonl", i))
		if err := os.WriteFile(path, []byte(lines), 0o644); err != nil {
			t.Fatalf("write session file %d: %v", i, err)
		}
	}

	return &logstack.LogstackClient{AgentsDir: root}
}

// getJSON calls one handler with a query string and decodes the response.
func getJSON(t *testing.T, handler http.HandlerFunc, path string, into interface{}) int {
	t.Helper()

	recorder := httptest.NewRecorder()
	handler(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	if into != nil && recorder.Body.Len() > 0 {
		if err := json.Unmarshal(recorder.Body.Bytes(), into); err != nil {
			t.Fatalf("decode response for %s (status %d): %v\nbody: %s",
				path, recorder.Code, err, recorder.Body.String())
		}
	}
	return recorder.Code
}

// --------------------------------------------------------------------------
// handleConversations: the default limit, the cap guard and the cap value.
//
// This handler has no database dependency at all — it reads LogstackClient,
// which walks a filesystem directory. api.go:1509-1515 spells the ceiling
// TWICE: `if limit > 100` decides whether to cut, and `limit = 100` is what it
// cuts to. The 184th's rule — a budget spelled twice is two boundaries — so
// each is separated below rather than assumed to move together.
// --------------------------------------------------------------------------

func TestConversationLimitDefaultsToTwenty(t *testing.T) {
	server := NewServer(nil, ws.NewHub(), nil, nil)
	server.LogstackClient = seedSessionFiles(t, "elf", 25)

	var conversations []logstack.Conversation
	if code := getJSON(t, server.handleConversations, "/api/conversations?agent=elf", &conversations); code != http.StatusOK {
		t.Fatalf("status = %d, want 200", code)
	}

	// 25 files are available, so anything other than 20 is the default moving.
	if len(conversations) != 20 {
		t.Errorf("no limit parameter returned %d conversations, want the default of 20",
			len(conversations))
	}
}

func TestConversationDefaultIsNotJustAnyRejectedLimit(t *testing.T) {
	// Trap 6 from the card: `parsed > 0` is a zero guard, not a boundary. A
	// rejected limit silently keeps the default, so the test above cannot tell
	// "the default is 20" from "any bad input yields 20". Varying a VALID limit
	// is what separates them: if the handler honours 7, then the 20 above is
	// the default rather than a fallback everything lands in.
	server := NewServer(nil, ws.NewHub(), nil, nil)
	server.LogstackClient = seedSessionFiles(t, "elf", 25)

	for _, valid := range []int{1, 7, 19, 21} {
		var conversations []logstack.Conversation
		path := fmt.Sprintf("/api/conversations?agent=elf&limit=%d", valid)
		if code := getJSON(t, server.handleConversations, path, &conversations); code != http.StatusOK {
			t.Fatalf("limit=%d: status = %d, want 200", valid, code)
		}
		if len(conversations) != valid {
			t.Errorf("limit=%d returned %d conversations, want %d",
				valid, len(conversations), valid)
		}
	}
}

func TestConversationLimitIsCappedAtOneHundred(t *testing.T) {
	// 105 files, so every case below is limited by the handler and not by how
	// many conversations exist.
	server := NewServer(nil, ws.NewHub(), nil, nil)
	server.LogstackClient = seedSessionFiles(t, "elf", 105)

	cases := []struct {
		limit int
		want  int
		why   string
	}{
		{100, 100, "100 is not over the cap, so it passes through untouched"},
		{101, 100, "101 is the first limit the cap has to cut — moving the " +
			"GUARD to `> 101` shows up here and nowhere else"},
		{500, 100, "far above the cap, so this separates the cap's VALUE " +
			"from the guard that applies it"},
	}

	for _, testCase := range cases {
		var conversations []logstack.Conversation
		path := fmt.Sprintf("/api/conversations?agent=elf&limit=%d", testCase.limit)
		if code := getJSON(t, server.handleConversations, path, &conversations); code != http.StatusOK {
			t.Fatalf("limit=%d: status = %d, want 200", testCase.limit, code)
		}
		if len(conversations) != testCase.want {
			t.Errorf("limit=%d returned %d conversations, want %d (%s)",
				testCase.limit, len(conversations), testCase.want, testCase.why)
		}
	}
}

// --------------------------------------------------------------------------
// listCosts: the YYYY-MM-DD length check, and ValidateLimit's call arguments.
// --------------------------------------------------------------------------

func TestCostDateParametersMustBeExactlyTenCharacters(t *testing.T) {
	// api.go:1972 and :1975 check `len(date) != 10` — not that the date parses.
	// So the boundary is a STRING LENGTH, and only a 9/10/11 straddle pins it.
	// Both parameters are checked separately and are scored separately: a
	// length that drifted on one and not the other would be invisible to a test
	// that only sent start_date.
	server := newBoundaryServer(t)
	seedCostEntries(t, server, 3)

	cases := []struct {
		parameter string
		value     string
		wantCode  int
		why       string
	}{
		{"start_date", "2026-01-01", http.StatusOK, "exactly 10 characters"},
		{"start_date", "2026-01-0", http.StatusBadRequest, "9 characters, one short"},
		{"start_date", "2026-01-011", http.StatusBadRequest, "11 characters, one long"},
		{"end_date", "2026-12-31", http.StatusOK, "exactly 10 characters"},
		{"end_date", "2026-12-3", http.StatusBadRequest, "9 characters, one short"},
		{"end_date", "2026-12-311", http.StatusBadRequest, "11 characters, one long"},
	}

	for _, testCase := range cases {
		path := fmt.Sprintf("/api/costs?%s=%s", testCase.parameter, testCase.value)
		code := getJSON(t, server.listCosts, path, nil)
		if code != testCase.wantCode {
			t.Errorf("%s=%q (%s): status = %d, want %d",
				testCase.parameter, testCase.value, testCase.why, code, testCase.wantCode)
		}
	}
}

func TestCostLimitDefaultsToOneHundredAndCapsAtOneThousand(t *testing.T) {
	// api.go:1979 is `validator.ValidateLimit(..., 100, 1000)`. Both numbers are
	// CALL ARGUMENTS, so the grep that sized this sweep cannot see either — the
	// 204th's point about TruncateAtRuneBoundary(task, 100), recurring here.
	// validation_test.go:460 pins how ValidateLimit behaves given (10, 100);
	// nothing pinned what listCosts actually passes it.
	server := newBoundaryServer(t)
	seedCostEntries(t, server, 1001)

	cases := []struct {
		query string
		want  int
		why   string
	}{
		{"", 100, "no limit parameter, so the default argument applies"},
		{"?limit=7", 7, "a valid limit is honoured, so 100 above is the " +
			"default and not a fallback that every request lands in"},
		{"?limit=1000", 1000, "1000 is the max itself, so it is not capped"},
		{"?limit=1001", 1000, "the first limit over the max — this is where a " +
			"max of 1001 would show"},
		{"?limit=99999", 1000, "far above the max"},
	}

	for _, testCase := range cases {
		var costs []db.CostEntry
		if code := getJSON(t, server.listCosts, "/api/costs"+testCase.query, &costs); code != http.StatusOK {
			t.Fatalf("%q: status = %d, want 200", testCase.query, code)
		}
		if len(costs) != testCase.want {
			t.Errorf("%q returned %d cost entries, want %d (%s)",
				testCase.query, len(costs), testCase.want, testCase.why)
		}
	}
}

// --------------------------------------------------------------------------
// listBounties: the other ValidateLimit call site, with different arguments.
// --------------------------------------------------------------------------

func TestBountyLimitDefaultsToFiftyAndCapsAtOneHundred(t *testing.T) {
	// api.go:2474 is `validator.ValidateLimit(..., 50, 100)`. Different numbers
	// from listCosts', at the same function, in the same file. A test that
	// pinned only one call site would leave the other free — which is the whole
	// reason both are here.
	server := newBoundaryServer(t)
	seedBounties(t, server, 101)

	cases := []struct {
		query string
		want  int
		why   string
	}{
		{"", 50, "no limit parameter, so the default argument applies"},
		{"?limit=7", 7, "a valid limit is honoured, so 50 above is the default"},
		{"?limit=100", 100, "100 is the max itself, so it is not capped"},
		{"?limit=101", 100, "the first limit over the max"},
		{"?limit=500", 100, "far above the max"},
	}

	for _, testCase := range cases {
		var bounties []db.Bounty
		if code := getJSON(t, server.listBounties, "/api/bounties"+testCase.query, &bounties); code != http.StatusOK {
			t.Fatalf("%q: status = %d, want 200", testCase.query, code)
		}
		if len(bounties) != testCase.want {
			t.Errorf("%q returned %d bounties, want %d (%s)",
				testCase.query, len(bounties), testCase.want, testCase.why)
		}
	}
}

// --------------------------------------------------------------------------
// listDisputes: the SAME pagination block again, at a third site.
//
// The card named two pagination sites. There are three: api.go:2926-2933 is a
// character-for-character copy of handleConversations' block, differing only in
// the default (50 rather than 20), and it was not in scope because the grep that
// sized the sweep saw `if limit > 100 { // Cap at 100` as one line rather than
// two sites. The 209th's rule, one level over: when a defect class is found in
// one copy of a block, sweep the whole file for the other copies.
//
// This matters mechanically as well as morally. `limit > 100` and `limit = 100`
// each occur TWICE in api.go, so any scorer needle written for the
// conversations site without an anchor mutates the disputes site and reports the
// result under the conversations row — the exact failure the exactly-once rule
// exists to prevent, and one an anchorless needle slips straight past.
// --------------------------------------------------------------------------

func TestDisputeLimitDefaultsToFiftyAndCapsAtOneHundred(t *testing.T) {
	server := newBoundaryServer(t)
	seedDisputes(t, server, 105)

	cases := []struct {
		query string
		want  int
		why   string
	}{
		{"", 50, "no limit parameter, so the default applies — note this is 50 " +
			"here and 20 at the otherwise identical conversations site"},
		{"?limit=7", 7, "a valid limit is honoured, so 50 above is the default"},
		{"?limit=100", 100, "100 is not over the cap"},
		{"?limit=101", 100, "the first limit the cap has to cut"},
		{"?limit=500", 100, "far above the cap"},
	}

	for _, testCase := range cases {
		var disputes []db.Dispute
		if code := getJSON(t, server.listDisputes, "/api/disputes"+testCase.query, &disputes); code != http.StatusOK {
			t.Fatalf("%q: status = %d, want 200", testCase.query, code)
		}
		if len(disputes) != testCase.want {
			t.Errorf("%q returned %d disputes, want %d (%s)",
				testCase.query, len(disputes), testCase.want, testCase.why)
		}
	}
}

// --------------------------------------------------------------------------
// createBounty: the required-skills ceiling, and the ladder that is dead.
// --------------------------------------------------------------------------

// newBountyRequest builds a request body that passes every OTHER validation
// rule, so that a 400 can only have come from the axis under test.
func newBountyRequest(payout int, skills []string) string {
	quoted := make([]string, len(skills))
	for i, skill := range skills {
		quoted[i] = `"` + skill + `"`
	}
	return fmt.Sprintf(
		`{"title":"a bounty title that is long enough",`+
			`"description":"a description that is comfortably long enough to validate",`+
			`"requirements":"some requirements","payout_amount":%d,"creator_id":1,`+
			`"required_skills":[%s]}`, payout, strings.Join(quoted, ","))
}

func postBounty(t *testing.T, server *Server, body string) (int, db.Bounty) {
	t.Helper()

	recorder := httptest.NewRecorder()
	server.createBounty(recorder, httptest.NewRequest(http.MethodPost, "/api/bounties", strings.NewReader(body)))

	var created db.Bounty
	if recorder.Code == http.StatusCreated {
		if err := json.Unmarshal(recorder.Body.Bytes(), &created); err != nil {
			t.Fatalf("decode created bounty: %v\nbody: %s", err, recorder.Body.String())
		}
	}
	return recorder.Code, created
}

func TestBountyAcceptsTenRequiredSkillsAndRejectsEleven(t *testing.T) {
	// api.go:2520, `len(bounty.RequiredSkills) > 10`. It short-circuits to 400
	// before any query runs, so both sides of the straddle are cheap.
	server := newBoundaryServer(t)

	skills := func(n int) []string {
		out := make([]string, n)
		for i := range out {
			out[i] = fmt.Sprintf("skill-%d", i)
		}
		return out
	}

	if code, _ := postBounty(t, server, newBountyRequest(100, skills(10))); code != http.StatusCreated {
		t.Errorf("10 required skills: status = %d, want 201 — 10 is the "+
			"largest accepted count, so a ceiling of 9 shows here", code)
	}
	if code, _ := postBounty(t, server, newBountyRequest(100, skills(11))); code != http.StatusBadRequest {
		t.Errorf("11 required skills: status = %d, want 400 — 11 is the "+
			"smallest rejected count, so a ceiling of 11 shows here", code)
	}
}

func TestTheApiTierLadderIsOverwrittenBeforeAnyoneReadsIt(t *testing.T) {
	// ⛔ THIS TEST PINS THE DB LADDER. api.go's IS DEAD. Read this before
	// filing a gap for api.go:2536-2540.
	//
	// api.go computes a default tier from payout — 1000/500/100 — then calls
	// db.CreateBounty, whose LAST act (bounties.go:143) is `b.Tier = tier`,
	// recomputed unconditionally from its own ladder of 500/200/50. `bounty` is
	// passed by pointer and the response is encoded afterwards, so BOTH the
	// response and the stored row carry the db ladder and api.go's result is
	// discarded without ever being read.
	//
	// Measured, not inferred: replacing api.go's entire ladder with a single
	// `bounty.Tier = "COMPLETELY-BOGUS-TIER"` leaves every response and every
	// row below byte-identical. That is why sabotage-request-boundaries.py
	// carries api.go's three thresholds as declared KNOWN-NEGATIVE controls
	// rather than as open gaps: no test can catch a mutation to a value that
	// nothing reads, and filing them would send the next pass after a test that
	// cannot exist.
	//
	// The card said the response carries api.go's tier while the row carries
	// the db's. It does not; they agree, because only one ladder survives.
	//
	// ⛔ Do NOT "fix" the disagreement between the ladders here. Three of them
	// exist (api.go, internal/db/bounties.go, internal/bounty/types.go) and
	// reconciling them is a product change, not a test-only sweep. This file
	// records which one ships.
	server := newBoundaryServer(t)

	cases := []struct {
		payout int
		want   string
		why    string
	}{
		{49, "bronze", "below every db threshold"},
		{50, "silver", "the db silver floor; api.go would still say bronze here"},
		{100, "silver", "api.go's silver floor, which changes nothing"},
		{199, "silver", "one below the db gold floor"},
		{200, "gold", "the db gold floor; api.go would still say silver here"},
		{499, "gold", "one below the db legendary floor"},
		{500, "legendary", "the db legendary floor; api.go would say gold here"},
		{1000, "legendary", "api.go's legendary floor, which changes nothing"},
	}

	for _, testCase := range cases {
		code, created := postBounty(t, server, newBountyRequest(testCase.payout, nil))
		if code != http.StatusCreated {
			t.Fatalf("payout=%d: status = %d, want 201", testCase.payout, code)
		}

		if created.Tier != testCase.want {
			t.Errorf("payout=%d: response tier = %q, want %q (%s)",
				testCase.payout, created.Tier, testCase.want, testCase.why)
		}

		var storedTier string
		if err := server.DB.QueryRow(`SELECT tier FROM bounties WHERE id = ?`, created.ID).Scan(&storedTier); err != nil {
			t.Fatalf("payout=%d: read back stored tier: %v", testCase.payout, err)
		}
		if storedTier != created.Tier {
			t.Errorf("payout=%d: stored tier %q disagrees with response tier %q — "+
				"if this ever fires, CreateBounty stopped overwriting the caller's "+
				"tier and api.go's ladder is live again",
				testCase.payout, storedTier, created.Tier)
		}
	}
}

// --------------------------------------------------------------------------
// The base branch, asserted as a side effect.
//
// The 208th's rule: when a card's scope names something a neighbour already
// covers, put it in the case table and let it score `already held` — do not
// delete it with a comment. The spawn message's 100-byte cut (api.go:1608-1609)
// is held by spawn_message_budget_test.go, which exists only on
// test/the-truncation-budgets-are-unpinned. That branch is in this one's base.
//
// This test does not add coverage. It asserts the PREMISE: if the neighbour is
// missing from the base, the two rows the scorer carries for that cut would
// report `gap closed` and inflate this sweep's number. Here that shows up as a
// plain failure instead.
// --------------------------------------------------------------------------

func TestTheTruncationNeighbourIsPresentInThisBranchesBase(t *testing.T) {
	const neighbour = "spawn_message_budget_test.go"

	if _, err := os.Stat(neighbour); err != nil {
		t.Fatalf("%s is missing from this working tree: %v\n"+
			"This file's scorer carries the spawn-message cut as two "+
			"`already held` rows. Without the neighbour they would score as gaps "+
			"this sweep closed, and every other number in the run would be "+
			"inflated. Base on test/the-truncation-budgets-are-unpinned or a "+
			"branch that merges it, not on main.", neighbour, err)
	}
}
