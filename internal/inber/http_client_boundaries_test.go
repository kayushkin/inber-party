package inber

// Boundary values in HTTPClient.GetQuests and HTTPClient.GetQuestHistory.
//
// WHY THIS FILE EXISTS AT ALL, WHEN store_path_boundaries_test.go ALREADY
// TESTS "the difficulty ladder":
//
// The ladder is spelled TWICE, character for character, in two files:
//
//	internal/inber/inber.go:685       (*Store).GetQuests      — SQLite-backed
//	internal/inber/http_client.go:180 (*HTTPClient).GetQuests — HTTP-backed
//
// store_path_boundaries_test.go:187 pins the FIRST one, and says so in its own
// header (`inber.go:684-693`). Nothing reached the second. Both are shipped:
// api.go's Handler takes either as its `source`, so which ladder runs is a
// deployment question, and until this file existed only one of the two answers
// was held in place.
//
// The same is true of the progress constants, the XP floor, and both pagination
// guards. The sweep that produced these children split the work by FILE; the
// mechanism spans two files, so one copy was closed and its twin was left free
// while a grep for "difficulty ladder" reported the mechanism covered.
//
// Rules this file follows, inherited from the sibling boundary tests:
//
//   - Every threshold T is pinned by the straddle pair (T, T+1). Asserting one
//     side lets the number move in the unasserted direction, green throughout.
//   - The expected values are SPELLED OUT here, never derived from the constants
//     under test. A table built by ranging over the thing it pins asserts only
//     that the constant equals itself.
//   - Fixture reach is asserted separately from the value, so a fixture that
//     stops reaching the code reads as a fixture failure and not as a pass.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// session builds one /api/sessions row. Only the fields these tests vary are
// named; the rest default, which is what the real endpoint does for a session
// that has not finished.
func session(agent, status string, inTokens, outTokens int) map[string]any {
	return map[string]any{
		"key":        fmt.Sprintf("sess-%s-%d", agent, inTokens+outTokens),
		"agent":      agent,
		"status":     status,
		"in_tokens":  inTokens,
		"out_tokens": outTokens,
		"input_text": "task",
	}
}

// clientOver stands the whole GetQuests path up behind httptest and returns a
// client pointed at it, plus a reach counter. The counter is the reach guard:
// GetQuests answers an empty slice both when the fixture served no sessions and
// when it was never asked, and those are a passing test and a broken one.
func clientOver(t *testing.T, sessions ...map[string]any) (*HTTPClient, *int) {
	t.Helper()
	if sessions == nil {
		sessions = []map[string]any{}
	}
	served := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/sessions" {
			http.NotFound(w, r)
			return
		}
		served++
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(sessions); err != nil {
			t.Errorf("encoding the fake sessions failed: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	return NewHTTPClient(server.URL), &served
}

// oneHTTPQuestWithTokens drives a single session of exactly totalTokens through
// HTTPClient.GetQuests. Deliberately the HTTP-path twin of store_path_-
// boundaries_test.go's oneQuestWithTokens, so the two ladders are asserted the
// same way and a reader can diff the two tables.
func oneHTTPQuestWithTokens(t *testing.T, totalTokens int) RPGQuest {
	t.Helper()
	client, served := clientOver(t, session("claxon", "running", totalTokens, 0))
	quests, err := client.GetQuests(50)
	if err != nil {
		t.Fatalf("GetQuests: %v", err)
	}
	if *served != 1 {
		t.Fatalf("the fixture no longer reaches GetQuests: /api/sessions served %d times, want 1", *served)
	}
	if len(quests) != 1 {
		t.Fatalf("want exactly 1 quest, got %d", len(quests))
	}
	return quests[0]
}

// ---------------------------------------------------------------------------
// 1. GetQuests — the difficulty ladder (http_client.go:180-188)
// ---------------------------------------------------------------------------

// The ladder is four strictly-greater comparisons, so each threshold T is
// pinned by the pair (T, T+1): at T the lower tier still applies, at T+1 the
// higher one takes over.
//
// totalTokens is InTokens+OutTokens, both free integers off the wire, so every
// threshold is arithmetically reachable from either side — there is no
// dominated direction here, unlike the gold guard in internal/api.
func TestHTTPQuestDifficultyLadderStraddlesEveryThreshold(t *testing.T) {
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
			if got := oneHTTPQuestWithTokens(t, tc.tokens).Difficulty; got != tc.want {
				t.Errorf("tokens=%d: Difficulty = %d, want %d", tc.tokens, got, tc.want)
			}
		})
	}
}

// The tier VALUES are a separate question from the thresholds: a ladder
// returning 5/4/3/2/1 and one returning 6/4/3/2/1 cross at exactly the same
// token counts, so the straddle table above cannot tell them apart. Each token
// count here sits mid-tier, away from every threshold, so this test moves only
// when a tier's assigned value moves.
func TestHTTPQuestDifficultyTierValuesAreExact(t *testing.T) {
	for tokens, want := range map[int]int{
		100: 1, 600: 2, 1500: 3, 3000: 4, 9000: 5,
	} {
		if got := oneHTTPQuestWithTokens(t, tokens).Difficulty; got != want {
			t.Errorf("tokens=%d: Difficulty = %d, want exactly %d", tokens, got, want)
		}
	}
}

// ---------------------------------------------------------------------------
// 2. GetQuests — the XP floor (http_client.go:159-161)
// ---------------------------------------------------------------------------

// xpForTokens is tokens/100, so every session under 100 tokens earns 0 XP and
// the floor lifts it to 1. Nothing asserted XPReward anywhere in this package
// before this test, in either GetQuests.
//
// The pair that pins the floor is (99, 100): at 99 the floor is doing the work
// and at 100 the division is. A test at 99 alone cannot tell a floor of 1 from
// a floor that fires at every token count.
func TestHTTPQuestXPFloorLiftsAZeroRewardToOne(t *testing.T) {
	for _, tc := range []struct {
		tokens int
		want   int
	}{
		{0, 1},   // 0/100 == 0, floored
		{99, 1},  // still 0, floored
		{100, 1}, // 100/100 == 1 on its own merits, floor inactive
		{199, 1}, // 1, floor inactive
		{200, 2}, // the floor must not clamp anything above 1
		{5000, 50},
	} {
		t.Run(fmt.Sprintf("tokens=%d", tc.tokens), func(t *testing.T) {
			if got := oneHTTPQuestWithTokens(t, tc.tokens).XPReward; got != tc.want {
				t.Errorf("tokens=%d: XPReward = %d, want %d", tc.tokens, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 3. GetQuests — the progress constants (http_client.go:190-197)
// ---------------------------------------------------------------------------

// Four assigned constants selected by status, not a computation, so there is no
// ordering to lean on and each value has to be named. The status strings are
// the raw inber spellings, which GetQuests maps before it reaches this ladder.
func TestHTTPQuestProgressConstantsAreExactPerStatus(t *testing.T) {
	for _, tc := range []struct {
		status string
		want   int
	}{
		{"running", 50}, // the default arm
		{"completed", 100},
		{"success", 100}, // same quest status, same constant
		{"error", 30},
		{"timeout", 30},
		{"interrupted", 30},
		{"pending", 0},
	} {
		t.Run(tc.status, func(t *testing.T) {
			client, served := clientOver(t, session("claxon", tc.status, 10, 0))
			quests, err := client.GetQuests(50)
			if err != nil {
				t.Fatalf("GetQuests: %v", err)
			}
			if *served != 1 {
				t.Fatalf("the fixture no longer reaches GetQuests: served %d times, want 1", *served)
			}
			if len(quests) != 1 {
				t.Fatalf("want exactly 1 quest, got %d", len(quests))
			}
			if got := quests[0].Progress; got != tc.want {
				t.Errorf("status=%q: Progress = %d, want exactly %d", tc.status, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 4. GetQuests — the pagination guard (http_client.go:154)
// ---------------------------------------------------------------------------

func fiveSessions() []map[string]any {
	return []map[string]any{
		session("claxon", "running", 1, 0),
		session("claxon", "running", 2, 0),
		session("claxon", "running", 3, 0),
		session("claxon", "running", 4, 0),
		session("claxon", "running", 5, 0),
	}
}

// `if limit > 0 && i >= limit { break }` is two boundaries on one line, and they
// fail in different directions:
//
//   - `i >= limit` decides WHERE the cut lands. Pinned by asking for fewer
//     quests than the fixture holds: at limit=k the answer is exactly k, and a
//     drift to `i > limit` answers k+1.
//   - `limit > 0` decides whether there is a cut at all. Pinned in the test
//     below, because it is only observable at limit == 0.
func TestGetQuestsReturnsExactlyLimitQuests(t *testing.T) {
	for _, tc := range []struct {
		limit int
		want  int
	}{
		{1, 1},
		{2, 2},
		{4, 4},
		{5, 5}, // exactly the fixture size: the break fires on the last index
		{6, 5}, // more than the fixture holds
		{50, 5},
	} {
		t.Run(fmt.Sprintf("limit=%d", tc.limit), func(t *testing.T) {
			client, served := clientOver(t, fiveSessions()...)
			quests, err := client.GetQuests(tc.limit)
			if err != nil {
				t.Fatalf("GetQuests: %v", err)
			}
			if *served != 1 {
				t.Fatalf("the fixture no longer reaches GetQuests: served %d times, want 1", *served)
			}
			if len(quests) != tc.want {
				t.Errorf("GetQuests(%d) returned %d quests, want %d", tc.limit, len(quests), tc.want)
			}
		})
	}
}

// The `limit > 0` half of the guard, which has exactly one discriminating
// input. At limit=0 the guard is false and every session is returned; widen it
// to `limit >= 0` and the break fires on the first index, answering zero.
//
// ⚠️ A NEGATIVE limit does not discriminate: -1 fails both `> 0` and `>= 0`, so
// the mutant and the original agree there. Enumerated over -5..5 before this
// test was written; 0 is the only input that separates them, which is why it is
// asserted on its own rather than folded into the table above.
//
// ⚠️ And no production caller passes 0. api.go's two handlers substitute their
// defaults unless the parsed limit is `> 0`, and the four in-package callers
// pass 1000 or 100 literally. So this branch ships unreachable — the test pins
// what the function promises, not a path the product takes.
func TestGetQuestsTreatsAZeroLimitAsUnlimited(t *testing.T) {
	for _, limit := range []int{0, -1, -5} {
		t.Run(fmt.Sprintf("limit=%d", limit), func(t *testing.T) {
			client, _ := clientOver(t, fiveSessions()...)
			quests, err := client.GetQuests(limit)
			if err != nil {
				t.Fatalf("GetQuests: %v", err)
			}
			if len(quests) != 5 {
				t.Errorf("GetQuests(%d) returned %d quests, want all 5", limit, len(quests))
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 5. GetQuestHistory — the pagination guard (http_client.go:314)
// ---------------------------------------------------------------------------

// GetQuestHistory filters GetQuests(1000) down to one agent and then cuts. The
// cut is checked AFTER the append, so `len(entries) >= limit` and the drift to
// `> limit` differ by exactly one entry.
//
// The fixture interleaves two agents so the filter and the cut are separable: a
// cut that counted loop iterations rather than kept entries would answer
// differently at every limit below 3.
func mixedAgentSessions() []map[string]any {
	return []map[string]any{
		session("claxon", "completed", 1, 0),
		session("brigid", "completed", 2, 0),
		session("claxon", "error", 3, 0),
		session("brigid", "running", 4, 0),
		session("claxon", "running", 5, 0),
		session("claxon", "pending", 6, 0),
	}
}

func TestGetQuestHistoryReturnsExactlyLimitEntries(t *testing.T) {
	for _, tc := range []struct {
		limit int
		want  int
	}{
		{1, 1},
		{2, 2},
		{3, 3},
		{4, 4}, // exactly the number of claxon quests
		{5, 4}, // more than the agent has
		{20, 4},
	} {
		t.Run(fmt.Sprintf("limit=%d", tc.limit), func(t *testing.T) {
			client, served := clientOver(t, mixedAgentSessions()...)
			entries, err := client.GetQuestHistory("claxon", tc.limit)
			if err != nil {
				t.Fatalf("GetQuestHistory: %v", err)
			}
			if *served != 1 {
				t.Fatalf("the fixture no longer reaches GetQuestHistory: served %d times, want 1", *served)
			}
			if len(entries) != tc.want {
				t.Errorf("GetQuestHistory(claxon, %d) returned %d entries, want %d",
					tc.limit, len(entries), tc.want)
			}
		})
	}
}

// The `limit > 0` half, same shape as GetQuests's and same single
// discriminating input. Widened to `limit >= 0`, the break fires after the
// first append and answers one entry instead of all four.
//
// ⚠️ Unreachable in production for the same reason: handleQuestHistory
// substitutes 20 unless the parsed limit is `> 0`, and it is the only caller.
func TestGetQuestHistoryTreatsAZeroLimitAsUnlimited(t *testing.T) {
	for _, limit := range []int{0, -1} {
		t.Run(fmt.Sprintf("limit=%d", limit), func(t *testing.T) {
			client, _ := clientOver(t, mixedAgentSessions()...)
			entries, err := client.GetQuestHistory("claxon", limit)
			if err != nil {
				t.Fatalf("GetQuestHistory: %v", err)
			}
			if len(entries) != 4 {
				t.Errorf("GetQuestHistory(claxon, %d) returned %d entries, want all 4",
					limit, len(entries))
			}
		})
	}
}

// The filter itself, asserted separately so a cut that silently stopped
// filtering could not pass the tables above by returning the right COUNT of the
// wrong agent's quests.
func TestGetQuestHistoryReturnsOnlyTheNamedAgentsQuests(t *testing.T) {
	client, _ := clientOver(t, mixedAgentSessions()...)
	entries, err := client.GetQuestHistory("brigid", 20)
	if err != nil {
		t.Fatalf("GetQuestHistory: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("want brigid's 2 quests, got %d", len(entries))
	}
	// brigid's rows carry 2 and 4 tokens; claxon's carry 1, 3, 5 and 6.
	for _, e := range entries {
		if e.Tokens != 2 && e.Tokens != 4 {
			t.Errorf("entry with %d tokens is not one of brigid's", e.Tokens)
		}
	}
}
