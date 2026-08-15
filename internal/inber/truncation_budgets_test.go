package inber

// Boundary VALUES for this package's four truncation sites.
//
// truncation_test.go and http_client_test.go pin the rune-boundary *mechanism*
// — the cut never splits a rune. Neither varies an input length across a guard,
// and both assert "a prefix of at least N bytes survived", which is a lower
// bound: it stays true when the budget grows and when the guard moves. So every
// number in the four sites below was free to drift with the suite green.
//
// Each site spells its budget twice — a guard deciding whether to cut, and the
// number handed to the cut — and the two are pinned separately, because a
// disagreement between them is exactly what neither test would show.
//
// ⛔ The ellipsis convention is inconsistent across the six sites in this repo:
// generateQuestName and truncateText count the "..." inside the byte budget,
// the two 200-byte quest descriptions and api.go's spawn message put it
// outside. textutil.go's doc comment records that as deliberate and unsettled.
// These tests pin each site to what it does today. Making them agree is a
// product change and is not this sweep's job.

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// ---------------------------------------------------------------------------
// generateQuestName's fallback: guard 60, budget 57, ellipsis inside.
// ---------------------------------------------------------------------------

// The fallback branch is only reached when the procedural name overruns 80
// bytes, and which naming pattern is used is chosen by a hash of the input —
// so reaching it is a property of the exact input string, not of its length.
// "bug" repeated is a stem for which BOTH the 60-byte and the 61-byte input
// land in the fallback, which is what makes them a straddle pair rather than
// two unrelated cases. Found by search over three-letter stems, not by guess.
//
// Both outputs are 68 bytes long — 8 bytes of emoji and space, then either 60
// bytes of text or 57 plus "...". So a length assertion cannot tell them apart
// and these tests compare the text itself.
const questNameStem = "bug"

func questNameFallbackInput(t *testing.T, n int) string {
	t.Helper()
	in := strings.Repeat(questNameStem, n/len(questNameStem)+2)[:n]
	out := generateQuestName(in, "running")
	// The reach guard, not an assertion. If a change to the naming patterns or
	// the key-term extractor stops this input overrunning 80 bytes, the test
	// below would pass against every mutation without ever running the cut.
	if !strings.Contains(out, questNameStem+questNameStem) {
		t.Fatalf("a %d-byte input no longer reaches generateQuestName's truncating "+
			"fallback, so this test proves nothing; got the procedural name %q", n, out)
	}
	return in
}

func TestAQuestNameFallbackOfExactlyTheGuardIsLeftWhole(t *testing.T) {
	const guard = 60
	in := questNameFallbackInput(t, guard)

	got := generateQuestName(in, "running")
	if !strings.HasSuffix(got, in) {
		t.Errorf("a %d-byte fallback text was not carried through whole:\n got  (%d) %q\n want it to end with (%d) %q",
			guard, len(got), got, len(in), in)
	}
	if strings.HasSuffix(got, "...") {
		t.Errorf("a fallback text of exactly the guard was truncated: %q", got)
	}
}

func TestAQuestNameFallbackOneByteOverTheGuardIsCutToTheBudget(t *testing.T) {
	const (
		guard  = 60
		budget = 57
	)
	in := questNameFallbackInput(t, guard+1)

	got := generateQuestName(in, "running")
	want := in[:budget] + "..."
	if !strings.HasSuffix(got, want) {
		t.Errorf("a %d-byte fallback text was not cut to the budget:\n got  (%d) %q\n want it to end with (%d) %q",
			len(in), len(got), got, len(want), want)
	}
	// The emoji prefix varies with the task type and the sibling naming
	// branches rewrite the procedural names, so the assertion above is a
	// suffix rather than an equality. This pins the rest: the text after the
	// emoji is exactly budget+3 bytes, i.e. this site counts the ellipsis
	// inside the 60.
	if _, text, ok := strings.Cut(got, " "); !ok || len(text) != guard {
		t.Errorf("the fallback text is %q, want %d bytes after the emoji", text, guard)
	}
}

// ---------------------------------------------------------------------------
// truncateText: guard maxLen (a parameter), budget maxLen-3, ellipsis inside.
// ---------------------------------------------------------------------------

// ⚠️ The value that SHIPS is not pinned here, and cannot be pinned cheaply.
// truncateText has exactly one production call site — inber.go's
// (*Store).GetAgentJournal — and it passes 100. Reaching that call needs
// r.output_text, which createTestGatewayDB's requests table does not define, so
// pinning the shipping 100 means a fixture schema change. This file pins the
// function's own arithmetic; the caller's 100 is unpinned and this comment
// exists so the next reader does not assume otherwise.
func TestTruncateTextCutsAtMaxLenMinusTheEllipsis(t *testing.T) {
	const maxLen = 40
	in := strings.Repeat("a", maxLen+1)

	got := truncateText(in, maxLen)
	want := in[:maxLen-3] + "..."
	if got != want {
		t.Errorf("truncateText(%d bytes, %d):\n got  (%d) %q\n want (%d) %q",
			len(in), maxLen, len(got), got, len(want), want)
	}
	// The ellipsis is inside the budget at this site: the result is maxLen, not
	// maxLen+3.
	if len(got) != maxLen {
		t.Errorf("truncateText returned %d bytes for maxLen=%d — this site counts "+
			"the ellipsis inside the budget", len(got), maxLen)
	}
}

func TestTruncateTextLeavesTextOfExactlyMaxLenAlone(t *testing.T) {
	const maxLen = 40
	in := strings.Repeat("a", maxLen-1) + "Z" // a distinct final byte

	if got := truncateText(in, maxLen); got != in {
		t.Errorf("truncateText(%d bytes, %d) = %q, want it returned unchanged",
			len(in), maxLen, got)
	}
}

// ---------------------------------------------------------------------------
// HTTPClient.GetQuests: guard 200, budget 200, ellipsis OUTSIDE (ships 203).
// ---------------------------------------------------------------------------

// questsFromFakeInber stands the whole HTTP path up the way
// http_client_test.go does, so the description under test is the one the
// dashboard would receive rather than the output of the cut in isolation.
func questsFromFakeInber(t *testing.T, inputText string) []RPGQuest {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/sessions" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode([]map[string]any{{
			"key":        "session-1",
			"agent":      "claxon",
			"status":     "running",
			"input_text": inputText,
		}}); err != nil {
			t.Errorf("encoding the fake session failed: %v", err)
		}
	}))
	defer server.Close()

	quests, err := NewHTTPClient(server.URL).GetQuests(10)
	if err != nil {
		t.Fatalf("GetQuests failed: %v", err)
	}
	if len(quests) != 1 {
		t.Fatalf("expected 1 quest, got %d", len(quests))
	}
	return quests
}

func TestAnHTTPQuestDescriptionOfExactlyTheGuardIsLeftWhole(t *testing.T) {
	const guard = 200
	in := strings.Repeat("a", guard-1) + "Z"

	if got := questsFromFakeInber(t, in)[0].Description; got != in {
		t.Errorf("a description of exactly %d bytes was not returned whole:\n got  (%d) %q",
			guard, len(got), got)
	}
}

func TestAnHTTPQuestDescriptionOneByteOverTheGuardIsCutToTheBudget(t *testing.T) {
	const (
		guard  = 200
		budget = 200
	)
	in := strings.Repeat("a", guard+1)

	got := questsFromFakeInber(t, in)[0].Description
	want := in[:budget] + "..."
	if got != want {
		t.Errorf("a description of %d bytes was not cut to the budget:\n got  (%d) %q\n want (%d) %q",
			len(in), len(got), got, len(want), want)
	}
	// This site puts the ellipsis outside the budget, so it ships 203 bytes.
	if len(got) != budget+3 {
		t.Errorf("a cut description is %d bytes, want %d — this site counts the "+
			"ellipsis outside the budget", len(got), budget+3)
	}
}

// ---------------------------------------------------------------------------
// (*Store).GetQuests: the same 200/200 cut against the SQLite gateway.
// ---------------------------------------------------------------------------

// ⛔ scripts/sabotage-truncation.py declared this row a KNOWN GAP because
// "GetQuests reads PostgreSQL". It does not — internal/inber opens sqlite3, and
// createTestGatewayDB already builds a real one. The function was reached all
// along; the 200-byte *branch* was not, because every input_text in the shared
// fixture is shorter than 200 bytes. The fix was one longer row, not a
// database, and the scorer's note has been corrected.
//
// This builds its own gateway DB rather than lengthening createTestGatewayDB,
// because four sibling tests assert exact counts against that fixture (6
// quests, 4 completed, 3 agents) and a seventh row would move all of them. A
// fixture shared by count-asserting tests is not a place to add a row.
func gatewayDBWithInputText(t *testing.T, inputText string) string {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "gateway.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE sessions (
			key TEXT PRIMARY KEY,
			agent TEXT NOT NULL,
			label TEXT,
			last_active DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_key TEXT NOT NULL REFERENCES sessions(key),
			status TEXT NOT NULL DEFAULT 'pending',
			input_text TEXT,
			turns INTEGER DEFAULT 0,
			input_tokens INTEGER DEFAULT 0,
			output_tokens INTEGER DEFAULT 0,
			cost REAL DEFAULT 0,
			started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			error_text TEXT,
			parent_request_id INTEGER
		);
		INSERT INTO sessions (key, agent) VALUES ('sess-1', 'claxon');
	`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(
		`INSERT INTO requests (session_key, status, input_text) VALUES ('sess-1', 'running', ?)`,
		inputText); err != nil {
		t.Fatal(err)
	}
	return dbPath
}

func storeQuestDescription(t *testing.T, inputText string) string {
	t.Helper()
	store, err := NewStore("", gatewayDBWithInputText(t, inputText), "")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	quests, err := store.GetQuests(50)
	if err != nil {
		t.Fatal(err)
	}
	if len(quests) != 1 {
		t.Fatalf("expected 1 quest from the fixture, got %d", len(quests))
	}
	return quests[0].Description
}

func TestAStoreQuestDescriptionOfExactlyTheGuardIsLeftWhole(t *testing.T) {
	const guard = 200
	in := strings.Repeat("a", guard-1) + "Z"

	if got := storeQuestDescription(t, in); got != in {
		t.Errorf("a description of exactly %d bytes was not returned whole:\n got  (%d) %q",
			guard, len(got), got)
	}
}

func TestAStoreQuestDescriptionOneByteOverTheGuardIsCutToTheBudget(t *testing.T) {
	const (
		guard  = 200
		budget = 200
	)
	in := strings.Repeat("a", guard+1)

	got := storeQuestDescription(t, in)
	want := in[:budget] + "..."
	if got != want {
		t.Errorf("a description of %d bytes was not cut to the budget:\n got  (%d) %q\n want (%d) %q",
			len(in), len(got), got, len(want), want)
	}
	if len(got) != budget+3 {
		t.Errorf("a cut description is %d bytes, want %d — this site counts the "+
			"ellipsis outside the budget", len(got), budget+3)
	}
}

// TestAStoreQuestDescriptionNeverSplitsARune closes the last rune-safety gap in
// this family. scripts/sabotage-truncation.py listed this site as a KNOWN GAP
// on the strength of a reason that was never true; the branch was one fixture
// row away the whole time, and gatewayDBWithInputText supplies it.
//
// The straddle tests above cannot cover this: their inputs are ASCII, and over
// ASCII a byte cut and a rune-safe cut return the same string. Pinning a budget
// and pinning the cut are separate jobs even at one site.
func TestAStoreQuestDescriptionNeverSplitsARune(t *testing.T) {
	const cutAt = 200

	for offset := 0; offset <= 4; offset++ {
		label := labelFor("Store.GetQuests description", offset)
		in := strings.Repeat("a", cutAt-4+offset) + fourByteRune + strings.Repeat("b", 40)
		assertCutIsRuneSafe(t, label, storeQuestDescription(t, in), cutAt-4)
	}
}
