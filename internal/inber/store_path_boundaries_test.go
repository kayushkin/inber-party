package inber

// Boundary VALUES on (*Store)'s SQLite read paths: the quest difficulty ladder,
// the quest progress constants, the `limit <= 0` substituted defaults, the
// GetAgents status heuristics, and the two tool-call skill gates.
//
// Card 6b8fb2ca-0bfe-4ce8-bb01-79823bbd8182, worked by the 208th nightly pass.
// Onboarding note e72c6e27-b036-4854-992b-c083f94f2650 carries the reachability
// map this file is built from.
//
// Every number below was a correct value with nothing holding it in place —
// scripts/sabotage-store-paths.py scored each one UNNOTICED against the suite
// as it stood. These tests are insurance, not bug fixes. Nothing here is a
// live defect.
//
// ⛔ Two things this file deliberately does NOT do:
//
//  1. It does not touch createTestGatewayDB. Four sibling tests assert exact
//     counts against that fixture (6 quests, 4 completed, 3 agents, a 2-row
//     limit check), so a seventh row moves all four. Every test here builds its
//     own database from gatewayDBWith / sessionsDBWith.
//
//  2. It does not re-pin the 200-byte quest-description cut at inber.go:678-679.
//     truncation_budgets_test.go already holds that guard and budget with a
//     straddle pair, on test/the-truncation-budgets-are-unpinned, which is in
//     this branch's base. The card listed it as in scope; it was done one pass
//     earlier and the whole-repo baseline caught that. See the write-up.

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// ---------------------------------------------------------------------------
// Local fixture builders
// ---------------------------------------------------------------------------

type requestRow struct {
	sessionKey string
	status     string
	inputText  string
	inTokens   int
	outTokens  int
}

// gatewayDBWith builds a gateway.db holding exactly the given agents and
// requests. Sessions are derived from the request rows' session keys plus any
// bare agents named in agentsBySession.
func gatewayDBWith(t *testing.T, agentsBySession map[string]string, requests []requestRow) string {
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
	`); err != nil {
		t.Fatal(err)
	}

	for key, agent := range agentsBySession {
		if _, err := db.Exec(
			`INSERT INTO sessions (key, agent) VALUES (?, ?)`, key, agent); err != nil {
			t.Fatal(err)
		}
	}
	for _, r := range requests {
		if _, err := db.Exec(
			`INSERT INTO requests (session_key, status, input_text, input_tokens, output_tokens)
			 VALUES (?, ?, ?, ?, ?)`,
			r.sessionKey, r.status, r.inputText, r.inTokens, r.outTokens); err != nil {
			t.Fatal(err)
		}
	}
	return dbPath
}

// sessionsDBWith builds a sessions.db giving each named agent one session whose
// turns sum to the given tool-call count.
func sessionsDBWith(t *testing.T, toolCallsByAgent map[string]int) string {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "sessions.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			agent TEXT NOT NULL,
			status TEXT DEFAULT 'completed',
			started_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE turns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL REFERENCES sessions(id),
			in_tokens INTEGER DEFAULT 0,
			out_tokens INTEGER DEFAULT 0,
			cost REAL DEFAULT 0,
			tool_calls INTEGER DEFAULT 0
		);
	`); err != nil {
		t.Fatal(err)
	}

	i := 0
	for agent, toolCalls := range toolCallsByAgent {
		i++
		id := fmt.Sprintf("s%d", i)
		if _, err := db.Exec(
			`INSERT INTO sessions (id, agent, status) VALUES (?, ?, 'completed')`, id, agent); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(
			`INSERT INTO turns (session_id, in_tokens, out_tokens, cost, tool_calls)
			 VALUES (?, 0, 0, 0, ?)`, id, toolCalls); err != nil {
			t.Fatal(err)
		}
	}
	return dbPath
}

func storeOver(t *testing.T, sessionsDB, gatewayDB string) *Store {
	t.Helper()
	store, err := NewStore(sessionsDB, gatewayDB, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

// ---------------------------------------------------------------------------
// 1. GetQuests — the difficulty ladder (inber.go:684-693)
// ---------------------------------------------------------------------------

// oneQuestWithTokens drives a single request of exactly totalTokens through
// GetQuests and returns the quest it produced.
func oneQuestWithTokens(t *testing.T, totalTokens int) RPGQuest {
	t.Helper()
	gw := gatewayDBWith(t,
		map[string]string{"sess-1": "claxon"},
		[]requestRow{{sessionKey: "sess-1", status: "running", inputText: "task",
			inTokens: totalTokens, outTokens: 0}})
	quests, err := storeOver(t, "", gw).GetQuests(50)
	if err != nil {
		t.Fatalf("GetQuests: %v", err)
	}
	if len(quests) != 1 {
		t.Fatalf("want exactly 1 quest, got %d", len(quests))
	}
	return quests[0]
}

// The ladder is four strictly-greater comparisons, so each threshold T is
// pinned by the pair (T, T+1): at T the lower tier still applies, at T+1 the
// higher one takes over. Asserting only one side lets the threshold move in
// the unasserted direction with every test still green.
func TestQuestDifficultyLadderStraddlesEveryThreshold(t *testing.T) {
	for _, tc := range []struct {
		tokens int
		want   int
	}{
		{0, 1},
		{500, 1},     // at the threshold, not over it
		{501, 2},     // over
		{1000, 2},    // at
		{1001, 3},    // over
		{2000, 3},    // at
		{2001, 4},    // over
		{5000, 4},    // at
		{5001, 5},    // over
		{1000000, 5}, // the ladder has no sixth tier
	} {
		t.Run(fmt.Sprintf("tokens=%d", tc.tokens), func(t *testing.T) {
			if got := oneQuestWithTokens(t, tc.tokens).Difficulty; got != tc.want {
				t.Errorf("tokens=%d: Difficulty = %d, want %d", tc.tokens, got, tc.want)
			}
		})
	}
}

// The tier VALUES are a separate question from the thresholds. A ladder that
// returns 5/4/3/2/1 and one that returns 6/4/3/2/1 cross at exactly the same
// token counts, so a test that only checks "difficulty rises with tokens"
// cannot tell them apart.
func TestQuestDifficultyTierValuesAreExact(t *testing.T) {
	for tokens, want := range map[int]int{
		100: 1, 600: 2, 1500: 3, 3000: 4, 9000: 5,
	} {
		if got := oneQuestWithTokens(t, tokens).Difficulty; got != want {
			t.Errorf("tokens=%d: Difficulty = %d, want exactly %d", tokens, got, want)
		}
	}
}

// ---------------------------------------------------------------------------
// 2. GetQuests — the progress constants (inber.go:696-703)
// ---------------------------------------------------------------------------

// Progress is four assigned constants selected by status, not a computation, so
// there is no ordering for a test to lean on: each value has to be named.
// TestGetQuests asserts count, status mix and Children and never reads
// Progress, so before this test all four were free to move.
func TestQuestProgressConstantsAreExactPerStatus(t *testing.T) {
	for _, tc := range []struct {
		dbStatus     string
		wantStatus   string
		wantProgress int
	}{
		{"completed", "completed", 100},
		{"success", "completed", 100},
		{"error", "failed", 30},
		{"timeout", "failed", 30},
		{"interrupted", "failed", 30},
		{"pending", "available", 0},
		{"running", "in_progress", 50}, // the un-switched default
	} {
		t.Run(tc.dbStatus, func(t *testing.T) {
			gw := gatewayDBWith(t,
				map[string]string{"sess-1": "claxon"},
				[]requestRow{{sessionKey: "sess-1", status: tc.dbStatus, inputText: "task"}})
			quests, err := storeOver(t, "", gw).GetQuests(50)
			if err != nil {
				t.Fatalf("GetQuests: %v", err)
			}
			if len(quests) != 1 {
				t.Fatalf("want exactly 1 quest, got %d", len(quests))
			}
			if got := quests[0].Status; got != tc.wantStatus {
				t.Errorf("db status %q: Status = %q, want %q", tc.dbStatus, got, tc.wantStatus)
			}
			if got := quests[0].Progress; got != tc.wantProgress {
				t.Errorf("db status %q: Progress = %d, want %d", tc.dbStatus, got, tc.wantProgress)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 3. The substituted `limit <= 0` defaults
// ---------------------------------------------------------------------------

// manyRequests builds n requests against one agent.
func manyRequests(t *testing.T, agent string, n int) string {
	t.Helper()
	rows := make([]requestRow, 0, n)
	for i := 0; i < n; i++ {
		rows = append(rows, requestRow{sessionKey: "sess-1", status: "completed",
			inputText: fmt.Sprintf("task %d", i)})
	}
	return gatewayDBWith(t, map[string]string{"sess-1": agent}, rows)
}

// The 184th's rule: a parameter every test supplies is not the value that
// ships. TestGetQuests_Limit passes 2 and TestGetQuests passes 50, so the
// limit axis is varied but never crossed below 1 — the `limit <= 0 -> 50`
// substitution is the number a caller who passes nothing actually gets, and
// nothing entered that branch.
//
// Pinning it needs more rows than the default, or the LIMIT never bites and
// every candidate default returns the same slice.
func TestGetQuestsSubstitutesFiftyForANonPositiveLimit(t *testing.T) {
	store := storeOver(t, "", manyRequests(t, "claxon", 51))

	for _, limit := range []int{0, -1} {
		quests, err := store.GetQuests(limit)
		if err != nil {
			t.Fatalf("GetQuests(%d): %v", limit, err)
		}
		if len(quests) != 50 {
			t.Errorf("GetQuests(%d) returned %d quests, want the substituted default of 50",
				limit, len(quests))
		}
	}

	// Control: a positive limit is passed through untouched, so the assertion
	// above is about the default and not about the LIMIT clause working at all.
	quests, err := store.GetQuests(3)
	if err != nil {
		t.Fatalf("GetQuests(3): %v", err)
	}
	if len(quests) != 3 {
		t.Errorf("GetQuests(3) returned %d quests, want 3", len(quests))
	}
}

// Same substitution, different number, in a function no test called at all.
func TestGetQuestHistorySubstitutesTwentyForANonPositiveLimit(t *testing.T) {
	store := storeOver(t, "", manyRequests(t, "claxon", 21))

	for _, limit := range []int{0, -1} {
		entries, err := store.GetQuestHistory("claxon", limit)
		if err != nil {
			t.Fatalf("GetQuestHistory(%d): %v", limit, err)
		}
		if len(entries) != 20 {
			t.Errorf("GetQuestHistory(claxon, %d) returned %d entries, want the substituted default of 20",
				limit, len(entries))
		}
	}

	entries, err := store.GetQuestHistory("claxon", 5)
	if err != nil {
		t.Fatalf("GetQuestHistory(5): %v", err)
	}
	if len(entries) != 5 {
		t.Errorf("GetQuestHistory(claxon, 5) returned %d entries, want 5", len(entries))
	}
}

// ⚠️ GetConversations' own `limit <= 0 -> 50` (inber.go:883-884) is NOT pinned
// here and cannot be cheaply. Its query needs sessions.parent_session_id,
// initial_message and last_message_at, none of which any fixture in this
// package defines, so the query errors before the substituted limit can affect
// a single returned row. Reaching it is a schema change, which the card put
// out of scope. Recorded so the next reader does not assume it was covered.

// ---------------------------------------------------------------------------
// 4. GetAgents — the status heuristics (inber.go:419-424)
// ---------------------------------------------------------------------------

func agentStatusFor(t *testing.T, statuses []string) string {
	t.Helper()
	rows := make([]requestRow, 0, len(statuses))
	for _, st := range statuses {
		rows = append(rows, requestRow{sessionKey: "sess-1", status: st, inputText: "task"})
	}
	gw := gatewayDBWith(t, map[string]string{"sess-1": "claxon"}, rows)
	agents, err := storeOver(t, "", gw).GetAgents()
	if err != nil {
		t.Fatalf("GetAgents: %v", err)
	}
	if len(agents) != 1 {
		t.Fatalf("want exactly 1 agent, got %d", len(agents))
	}
	return agents[0].Status
}

// The error ratio is the only float boundary on this path, and `> 0.5` means an
// agent exactly half of whose requests failed is NOT stuck. That is the
// interesting side and it is the one an inequality mutation flips first.
//
// TestGetAgents_GatewayOnly asserts Class, Tokens, ErrorCount and Name and
// never reads Status, so before this test the whole heuristic was free.
func TestGetAgentsStatusHeuristicsStraddleTheErrorRatio(t *testing.T) {
	for _, tc := range []struct {
		name     string
		statuses []string
		want     string
	}{
		// running > 0 wins outright, whatever the error ratio.
		{"one running", []string{"running"}, "working"},
		{"running alongside errors", []string{"running", "error", "error"}, "working"},

		// errors > 0 is a real gate: no errors is idle, not stuck.
		{"all completed", []string{"completed", "completed"}, "idle"},

		// The ratio straddle. 1/2 is exactly 0.5 and must NOT be stuck;
		// 2/3 is the smallest ratio over it that a two-value fixture reaches.
		{"ratio exactly one half", []string{"error", "completed"}, "idle"},
		{"ratio two thirds", []string{"error", "error", "completed"}, "stuck"},
		{"ratio one third", []string{"error", "completed", "completed"}, "idle"},
		{"every request errored", []string{"error", "error"}, "stuck"},

		// A single errored request is the only input that separates the
		// `errors > 0` gate from the ratio test beside it. At two errors the
		// gate survives being tightened to `errors > 1`, because two is still
		// over one — so a fixture that only ever errors twice pins the ratio
		// and leaves the gate free.
		{"one request, and it errored", []string{"error"}, "stuck"},

		// 3/5 = 0.6 and 2/5 = 0.4 bracket the constant more tightly, so a
		// mutation to 0.55 or 0.45 dies here even though it survives above.
		{"ratio three fifths", []string{"error", "error", "error", "completed", "completed"}, "stuck"},
		{"ratio two fifths", []string{"error", "error", "completed", "completed", "completed"}, "idle"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := agentStatusFor(t, tc.statuses); got != tc.want {
				t.Errorf("statuses %v: Status = %q, want %q", tc.statuses, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 5. GetAgents — the two tool-call skill gates (inber.go:485, 535)
// ---------------------------------------------------------------------------

func toolMasterySkill(t *testing.T, agents []RPGAgent, agentID string) (RPGSkill, bool) {
	t.Helper()
	for _, a := range agents {
		if a.ID != agentID {
			continue
		}
		for _, sk := range a.Skills {
			if sk.Name == "Tool Mastery" {
				return sk, true
			}
		}
		return RPGSkill{}, false
	}
	t.Fatalf("agent %q not in result", agentID)
	return RPGSkill{}, false
}

// `toolCalls > 0` is spelled twice — once where the agent is already in the map
// from the gateway DB, once where the sessions DB is the only source. They are
// the same rule in two branches, so a test that exercises one leaves the other
// free. Both are driven here, in one call, by an agent present in both
// databases and an agent present in only the sessions DB.
func TestToolMasterySkillGateIsPinnedOnBothGetAgentsBranches(t *testing.T) {
	for _, tc := range []struct {
		toolCalls int
		wantSkill bool
	}{
		{0, false}, // at the gate
		{1, true},  // one over it
		{7, true},
	} {
		t.Run(fmt.Sprintf("toolCalls=%d", tc.toolCalls), func(t *testing.T) {
			gw := gatewayDBWith(t,
				map[string]string{"sess-1": "claxon"},
				[]requestRow{{sessionKey: "sess-1", status: "completed", inputText: "task"}})
			sess := sessionsDBWith(t, map[string]int{
				"claxon": tc.toolCalls, // merge branch: already in agentMap
				"brigid": tc.toolCalls, // sessions-only branch
			})
			agents, err := storeOver(t, sess, gw).GetAgents()
			if err != nil {
				t.Fatalf("GetAgents: %v", err)
			}

			for _, agentID := range []string{"claxon", "brigid"} {
				skill, ok := toolMasterySkill(t, agents, agentID)
				if ok != tc.wantSkill {
					t.Fatalf("%s toolCalls=%d: Tool Mastery present = %v, want %v",
						agentID, tc.toolCalls, ok, tc.wantSkill)
				}
				if ok && skill.TaskCount != tc.toolCalls {
					t.Errorf("%s: Tool Mastery TaskCount = %d, want %d",
						agentID, skill.TaskCount, tc.toolCalls)
				}
			}
		})
	}
}

// The `* 10` in `levelForXP(toolCalls * 10)` is a call argument, so no grep for
// numeric literals on a comparison line can see it — and it is the number that
// decides the level the dashboard renders.
//
// Pinning it needs tool-call counts whose level actually moves when the
// multiplier does, and the two directions need different counts:
//
//	10 calls -> 100 XP -> level 2, but * 9 gives 90 XP -> level 1  (catches a decrease)
//	28 calls -> 280 XP -> level 2, but * 11 gives 308 XP -> level 3 (catches an increase)
//
// Either count alone leaves the multiplier free to move the other way.
func TestToolMasteryLevelPinsTheTenXPPerToolCallMultiplier(t *testing.T) {
	for _, tc := range []struct {
		toolCalls int
		wantLevel int
	}{
		{10, 2}, // 100 XP — exactly the level-2 threshold
		{28, 2}, // 280 XP — still short of the level-3 threshold at 300
		{30, 3}, // 300 XP — exactly the level-3 threshold
		{1, 1},  // 10 XP
	} {
		t.Run(fmt.Sprintf("toolCalls=%d", tc.toolCalls), func(t *testing.T) {
			gw := gatewayDBWith(t,
				map[string]string{"sess-1": "claxon"},
				[]requestRow{{sessionKey: "sess-1", status: "completed", inputText: "task"}})
			sess := sessionsDBWith(t, map[string]int{
				"claxon": tc.toolCalls,
				"brigid": tc.toolCalls,
			})
			agents, err := storeOver(t, sess, gw).GetAgents()
			if err != nil {
				t.Fatalf("GetAgents: %v", err)
			}
			for _, agentID := range []string{"claxon", "brigid"} {
				skill, ok := toolMasterySkill(t, agents, agentID)
				if !ok {
					t.Fatalf("%s: expected a Tool Mastery skill at toolCalls=%d", agentID, tc.toolCalls)
				}
				if skill.Level != tc.wantLevel {
					t.Errorf("%s toolCalls=%d: Tool Mastery Level = %d, want %d",
						agentID, tc.toolCalls, skill.Level, tc.wantLevel)
				}
			}
		})
	}
}
