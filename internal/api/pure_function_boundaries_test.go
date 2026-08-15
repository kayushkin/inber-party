package api

// Boundary VALUES for api.go's PURE functions — the corner of the file that
// needs no database, no fixture and no live port to reach.
//
// api_test.go never sends a query string and never calls any of these, so
// before this file every number below was free to drift. What is new here is
// not the technique but the shape of the answer: FOUR of these literals cannot
// be pinned at their own boundary at all, and each is unreachable for a
// different reason. They are carried in scripts/sabotage-purefunctions.py as
// declared known-negatives, with the enumeration that proves it.
//
//	the tool-call defaults 50 and 20, and the caps 100 and 50
//	    dominated by `i < 10` DOWNSTREAM in getRecentToolCalls. The response
//	    length is min(limit/2, 10), so every limit at or above 20 renders the
//	    same ten rows. Accepting 100 and falling back to 50 are the same
//	    observable answer.
//	the gold floor's guard, loosened by one
//	    `goldReward < 5` -> `< 6` only moves inputs where the raw reward is
//	    exactly 5 — and 5 is arithmetically unreachable, because the base is
//	    always even and the five multipliers are 0.8/1.0/1.5/2.0/3.0. `< 7` IS
//	    observable, so the guard is pinned here by a two-step straddle.
//
// The 207th's dominance rule arriving from two new directions: not from a
// ladder that cannot reach a tier, and not from a value being its own guard,
// but from a cap further down the call chain and from a gap in the arithmetic
// itself.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// --- calculateGoldReward ---------------------------------------------------
//
// Unexported, pure, and reached in production from exactly one call site:
// handleCreateTask at api.go:589, and only when the caller omits gold_reward.
// That call site needs a live PostgreSQL INSERT two lines later, so the
// function is pinned here directly and the caller stays unpinned. Said out
// loud because §5 of the onboarding note asks for it: the number that ships is
// the one this function computes, and nothing downstream of it is asserted.

// TestTheGoldBaseConversionIsTwoGoldPerExperiencePoint pins `xpReward * 2` and
// the medium multiplier together. They are separable: a mutated base moves
// every tier at once, a mutated multiplier moves one, and the ladder test
// below is what tells them apart.
func TestTheGoldBaseConversionIsTwoGoldPerExperiencePoint(t *testing.T) {
	const xp = 50
	if got := calculateGoldReward(xp, "medium"); got != xp*2 {
		t.Errorf("calculateGoldReward(%d, medium) = %d, want %d", xp, got, xp*2)
	}
}

// TestTheGoldDifficultyLadderIsPinnedAtEveryTier crosses all five multipliers.
// The XP is high enough that the minimum-gold floor never fires, so a failure
// here is the multiplier and not the floor.
func TestTheGoldDifficultyLadderIsPinnedAtEveryTier(t *testing.T) {
	const xp = 50 // base 100, so each tier's gold reads as its multiplier x100

	for _, c := range []struct {
		difficulty string
		want       int
	}{
		{"easy", 80},
		{"medium", 100},
		{"hard", 150},
		{"expert", 200},
		{"legendary", 300},

		// The switch accepts the canonical digit alongside the name, so these
		// six rows never reach the strconv.Atoi arm below them.
		{"1", 80},
		{"2", 100},
		{"3", 150},
		{"4", 200},
		{"5", 300},

		// An unrecognised difficulty keeps multiplier 1.0. This pins the
		// default arm's value, which is the one tier with no case of its own.
		{"trivial", 100},
		{"", 100},
	} {
		if got := calculateGoldReward(xp, c.difficulty); got != c.want {
			t.Errorf("calculateGoldReward(%d, %q) = %d, want %d", xp, c.difficulty, got, c.want)
		}
	}
}

// TestTheSecondGoldLadderIsReachableOnlyByANonCanonicalSpelling pins the
// duplicate ladder inside the default arm.
//
// calculateGoldReward spells the 0.8/1.0/1.5/2.0/3.0 ladder TWICE: once over
// the names and canonical digits, and once over strconv.Atoi's result. The
// second copy looks like the digit path and is not — "3" is caught by
// `case "hard", "3"` before Atoi is ever called. The only inputs that reach it
// are spellings Atoi accepts and the switch does not: a sign, or a leading
// zero. So the second ladder ships, is separately mutable, and is reached by
// no canonical request at all.
//
// The onboarding note's §6 warns that a grep undercounts literals passed as
// call arguments. This is the other half of that: five literals that a grep
// sees twice and a reader assumes are the same five.
func TestTheSecondGoldLadderIsReachableOnlyByANonCanonicalSpelling(t *testing.T) {
	const xp = 50

	for _, c := range []struct {
		difficulty string
		want       int
	}{
		{"+1", 80},
		{"+2", 100},
		{"+3", 150},
		{"+4", 200},
		{"+5", 300},

		{"01", 80},
		{"03", 150},
		{"005", 300},

		// Out of the ladder's range: Atoi succeeds, no case matches, and the
		// multiplier stays 1.0. This pins that the second ladder has the same
		// five rungs as the first and no sixth.
		{"6", 100},
		{"+9", 100},
		{"-3", 100},
	} {
		if got := calculateGoldReward(xp, c.difficulty); got != c.want {
			t.Errorf("calculateGoldReward(%d, %q) = %d, want %d — the Atoi ladder drifted",
				xp, c.difficulty, got, c.want)
		}
	}
}

// TestTheGoldFloorLiftsAnyRewardBelowFive straddles the minimum-gold floor.
//
// ⚠️ The straddle is TWO steps wide on the loosening side, deliberately. The
// guard `goldReward < 5` and the value it assigns are the same number, so
// widening it to `< 6` only captures a raw reward of exactly 5 — and no input
// produces 5: the base is `xp * 2`, always even, and the five multipliers take
// it to {2x, 1.6x, 3x, 4x, 6x}, whose floors below 6 are 0,1,2,3,4 and 6. The
// row at raw 6 is what separates `< 5` from `< 7`, and it is the only reason
// this guard is pinned in that direction at all.
func TestTheGoldFloorLiftsAnyRewardBelowFive(t *testing.T) {
	for _, c := range []struct {
		xp         int
		difficulty string
		raw        int // what the arithmetic produces before the floor
		want       int
	}{
		{0, "medium", 0, 5}, // nothing earned still pays the minimum
		{2, "medium", 4, 5}, // one below the floor, lifted
		{1, "expert", 4, 5}, // reached down a different multiplier
		{3, "easy", 4, 5},   // 6 * 0.8 = 4.8, truncated to 4
		{3, "medium", 6, 6}, // first raw reward above the floor: NOT lifted
		{2, "hard", 6, 6},   // same value, different tier
		{1, "legendary", 6, 6},
	} {
		got := calculateGoldReward(c.xp, c.difficulty)
		if got != c.want {
			t.Errorf("calculateGoldReward(%d, %q) = %d, want %d (raw arithmetic gives %d)",
				c.xp, c.difficulty, got, c.want, c.raw)
		}
	}
}

// --- the tavern banter thresholds ------------------------------------------
//
// generateBanterMessage and getWorkloadMessage are methods on *Server that
// dereference no field, so a bare &Server{} reaches both with no database.
//
// Every message pool is indexed by `time.Now().Unix() % len(pool)`, so no test
// may assert an exact string. What each pool DOES share is the quest count it
// interpolates, and the three pools differ in whose count that is — which is
// enough to identify the branch without depending on the clock.

// workloadVoice reports which arm of getWorkloadMessage produced a message,
// using the one thing all three variants in an arm agree on.
//
// The busy arms interpolate a quest count; the steady arm interpolates none.
// Callers therefore pass counts that cannot be confused with each other.
func workloadVoice(t *testing.T, message string, agentACount, agentBCount int) string {
	t.Helper()
	if !strings.ContainsAny(message, "0123456789") {
		return "steady"
	}
	a := strings.Contains(message, strconv.Itoa(agentACount))
	b := strings.Contains(message, strconv.Itoa(agentBCount))
	switch {
	case a && !b:
		return "agentA is busy"
	case b && !a:
		return "agentB is busy"
	}
	t.Fatalf("cannot tell which arm produced %q with counts %d and %d — pick counts "+
		"whose decimal spellings do not overlap", message, agentACount, agentBCount)
	return ""
}

func workSummary(name string, recentQuests int) AgentWorkSummary {
	return AgentWorkSummary{Name: name, Class: "warrior", Level: 3, RecentQuests: recentQuests}
}

// TestWorkloadBanterNeedsMoreThanTwoRecentQuestsFromEitherAgent pins the gate
// at api.go:4247, which decides whether the workload conversation happens at
// all. It is a different number from the one inside getWorkloadMessage — three
// recent quests open the conversation and produce the steady, unimpressed
// variant of it.
func TestWorkloadBanterNeedsMoreThanTwoRecentQuestsFromEitherAgent(t *testing.T) {
	s := &Server{}

	for _, c := range []struct {
		agentA, agentB int
		wantBanter     bool
	}{
		{0, 0, false},
		{2, 2, false}, // exactly at the threshold on both sides: still silent
		{3, 0, true},  // one over, on agentA
		{0, 3, true},  // one over, on agentB
		{2, 3, true},
	} {
		got := s.generateBanterMessage(
			workSummary("Alpha", c.agentA), workSummary("Bravo", c.agentB),
			"workload_discussion")

		if (got != nil) != c.wantBanter {
			t.Errorf("generateBanterMessage(recent %d vs %d) returned %v, want banter = %v",
				c.agentA, c.agentB, got != nil, c.wantBanter)
			continue
		}
		if got != nil && got["context"] != "workload" {
			t.Errorf("recent %d vs %d produced context %v, want workload",
				c.agentA, c.agentB, got["context"])
		}
	}
}

// TestTheWorkloadMessageChangesVoiceAboveThreeRecentQuests pins the two
// thresholds inside getWorkloadMessage (api.go:4327 and :4334). They are a
// second, higher pair than the gate above: at exactly 3 the conversation opens
// and both agents still read as taking it easy.
func TestTheWorkloadMessageChangesVoiceAboveThreeRecentQuests(t *testing.T) {
	s := &Server{}

	for _, c := range []struct {
		agentA, agentB int
		want           string
	}{
		{3, 0, "steady"},         // exactly at the threshold: not busy yet
		{4, 0, "agentA is busy"}, // one over
		{0, 3, "steady"},         // the second threshold, from below
		{0, 4, "agentB is busy"},
		{3, 3, "steady"},
		{4, 7, "agentA is busy"}, // agentA is asked first and wins the tie
		{3, 7, "agentB is busy"}, // agentA below, so agentB's arm answers
	} {
		message := s.getWorkloadMessage(workSummary("Alpha", c.agentA), workSummary("Bravo", c.agentB))
		if got := workloadVoice(t, message, c.agentA, c.agentB); got != c.want {
			t.Errorf("getWorkloadMessage(recent %d vs %d) spoke as %q, want %q: %q",
				c.agentA, c.agentB, got, c.want, message)
		}
	}
}

// --- the recent-tool-call feed ---------------------------------------------
//
// getRecentToolCalls is declared mock data and touches no field of *Server, so
// both handlers run against &Server{} and an httptest recorder. The response
// length is the whole observable surface: min(limit/2, 10).

func toolCallsFrom(t *testing.T, path string, handler http.HandlerFunc) []ToolCall {
	t.Helper()
	recorder := httptest.NewRecorder()
	handler(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET %s = %d, want 200: %s", path, recorder.Code, recorder.Body.String())
	}
	var calls []ToolCall
	if err := json.Unmarshal(recorder.Body.Bytes(), &calls); err != nil {
		t.Fatalf("GET %s returned undecodable body %q: %v", path, recorder.Body.String(), err)
	}
	return calls
}

// TestTheToolCallFeedReturnsHalfTheLimitCappedAtTen pins the derived cap
// `i < limit/2 && i < 10`, which the onboarding note names as invisible to the
// grep that sized this card: both numbers live in a loop condition, not in a
// comparison against a request field.
//
// Integer division truncates, so 7 and 6 agree — the pair at 6/7 is what pins
// `limit/2` as division rather than a subtraction or a shift, and the pair at
// 18/20 is what pins the cap.
func TestTheToolCallFeedReturnsHalfTheLimitCappedAtTen(t *testing.T) {
	s := &Server{}

	for _, c := range []struct{ limit, want int }{
		{1, 0},
		{2, 1},
		{6, 3},
		{7, 3}, // 7/2 truncates to 3, same as 6
		{8, 4},
		{18, 9}, // one below the cap
		{19, 9},
		{20, 10}, // exactly the cap
		{21, 10}, // and it holds above it
		{100, 10},
	} {
		calls := toolCallsFrom(t, fmt.Sprintf("/api/spectator/recent-tool-calls?limit=%d", c.limit),
			s.handleRecentToolCalls)
		if len(calls) != c.want {
			t.Errorf("?limit=%d returned %d tool calls, want %d", c.limit, len(calls), c.want)
		}
	}
}

// TestTheToolCallLimitGateRejectsZeroAndNegativeLimits pins `parsedLimit > 0`
// from both sides, which is the one comparison in this handler that IS
// separable at its own boundary: a rejected limit falls back to the default
// ten rows, and an accepted 1 renders none.
func TestTheToolCallLimitGateRejectsZeroAndNegativeLimits(t *testing.T) {
	s := &Server{}
	const defaultRows = 10 // what the fallback renders, whatever the default is

	for _, c := range []struct {
		query string
		want  int
	}{
		{"", defaultRows},           // no parameter at all
		{"?limit=", defaultRows},    // present and empty
		{"?limit=abc", defaultRows}, // unparseable
		{"?limit=0", defaultRows},   // rejected by `> 0`, so the default stands
		{"?limit=-4", defaultRows},
		{"?limit=1", 0}, // accepted, and half of 1 is no rows at all
	} {
		calls := toolCallsFrom(t, "/api/spectator/recent-tool-calls"+c.query, s.handleRecentToolCalls)
		if len(calls) != c.want {
			t.Errorf("GET recent-tool-calls%s returned %d tool calls, want %d",
				c.query, len(calls), c.want)
		}
	}
}

// TestTheToolCallCeilingIsObservableOnlyWellBelowItsOwnValue is the honest
// statement of what can be pinned about `parsedLimit <= 100`.
//
// ⚠️ Nothing separates `<= 100` from `<= 99` or `<= 101`. At 100 the accepted
// path renders min(50, 10) = 10 rows and the rejected path falls back to the
// default 50, which renders min(25, 10) = 10 — the same answer. Enumerated over
// every limit from -5 to 204 plus the empty and unparseable spellings, the two
// handlers agree at every input. So the ceiling has no straddle pair, and the
// same is true of the default 50: neither 49 nor 51 changes any response.
//
// What IS observable is a ceiling that drifts far enough to reject a limit
// below 20, because only there does the cap stop hiding the difference. This
// test asserts that band, and the rest is carried as a declared known-negative
// in the scorer rather than left looking like a gap someone could close.
func TestTheToolCallCeilingIsObservableOnlyWellBelowItsOwnValue(t *testing.T) {
	s := &Server{}

	for _, c := range []struct {
		limit, want int
		why         string
	}{
		{12, 6, "inside the ceiling and below the cap, so the limit is visible"},
		{19, 9, "the highest limit whose exact value still shows in the response"},
		{101, 10, "over the ceiling: rejected, and the default renders ten rows"},
		{1000, 10, "far over: same ten rows, which is why the ceiling cannot be straddled"},
	} {
		calls := toolCallsFrom(t, fmt.Sprintf("/api/spectator/recent-tool-calls?limit=%d", c.limit),
			s.handleRecentToolCalls)
		if len(calls) != c.want {
			t.Errorf("?limit=%d returned %d tool calls, want %d (%s)",
				c.limit, len(calls), c.want, c.why)
		}
	}
}

// TestTheAgentToolCallDefaultIsPinnedFromBelowOnly covers the second handler.
//
// Its default is 20 and 20 sits exactly at the knee: 20/2 is 10, the cap. So a
// default that drifts DOWN to 19 renders nine rows and is caught here, while
// one that drifts UP to 21 renders the same ten and cannot be. That asymmetry
// is the whole difference between this handler and its sibling, whose default
// of 50 is above the knee and is therefore unpinnable in both directions.
func TestTheAgentToolCallDefaultIsPinnedFromBelowOnly(t *testing.T) {
	s := &Server{}
	const path = "/api/spectator/recent-tool-calls/claxon"

	for _, c := range []struct {
		query string
		want  int
		why   string
	}{
		{"", 10, "the default is 20 and half of it is exactly the cap"},
		{"?limit=0", 10, "rejected by `> 0`, so the default stands"},
		{"?limit=1", 0, "accepted, and half of 1 is no rows — the other side of the gate"},
		{"?limit=12", 6, "inside the ceiling and below the cap"},
		{"?limit=51", 10, "over this handler's lower ceiling: rejected, default stands"},
	} {
		calls := toolCallsFrom(t, path+c.query, s.handleRecentToolCallsForAgent)
		if len(calls) != c.want {
			t.Errorf("GET %s%s returned %d tool calls, want %d (%s)",
				path, c.query, len(calls), c.want, c.why)
		}
	}
}

// TestTheToolCallRowsCycleThroughTheirOwnTables pins the arithmetic inside
// getRecentToolCalls that the response exposes alongside the count: the tool
// name cycles over a six-entry table, the synthetic agent name over three, and
// a duration is attached on every other row at half a second per step.
//
// These are mock values, so nothing downstream depends on them being right —
// but they are the part of this function a reader is likeliest to "tidy", and
// the count assertions above would not notice.
func TestTheToolCallRowsCycleThroughTheirOwnTables(t *testing.T) {
	s := &Server{}
	wantTools := []string{"exec", "read", "write", "web_search", "browser", "nodes"}

	calls := toolCallsFrom(t, "/api/spectator/recent-tool-calls?limit=20", s.handleRecentToolCalls)
	if len(calls) != 10 {
		t.Fatalf("?limit=20 returned %d tool calls, want 10", len(calls))
	}

	for i, call := range calls {
		if want := wantTools[i%len(wantTools)]; call.ToolName != want {
			t.Errorf("row %d has tool_name %q, want %q — the tool table drifted", i, call.ToolName, want)
		}
		if want := fmt.Sprintf("agent_%d", i%3); call.AgentName != want {
			t.Errorf("row %d has agent_name %q, want %q — the agent cycle drifted", i, call.AgentName, want)
		}

		if i%2 != 0 {
			if call.Duration != nil {
				t.Errorf("row %d carries a duration; only even rows should", i)
			}
			continue
		}
		if call.Duration == nil {
			t.Fatalf("row %d carries no duration; every even row should", i)
		}
		if want := float64(i+1) * 0.5; *call.Duration != want {
			t.Errorf("row %d has duration %v, want %v", i, *call.Duration, want)
		}
		if call.Success == nil || !*call.Success {
			t.Errorf("row %d is not marked successful", i)
		}
	}
}

// TestAnAgentScopedFeedNamesTheRequestedAgent pins the one branch in
// getRecentToolCalls that the agent-scoped handler reaches and the other does
// not: a non-empty agent id replaces the synthetic agent_N cycle outright.
func TestAnAgentScopedFeedNamesTheRequestedAgent(t *testing.T) {
	s := &Server{}
	calls := toolCallsFrom(t, "/api/spectator/recent-tool-calls/claxon?limit=8",
		s.handleRecentToolCallsForAgent)

	if len(calls) != 4 {
		t.Fatalf("?limit=8 returned %d tool calls, want 4", len(calls))
	}
	for i, call := range calls {
		if call.AgentName != "claxon" {
			t.Errorf("row %d has agent_name %q, want claxon", i, call.AgentName)
		}
	}
}
