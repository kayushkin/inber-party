package inber

// The XP floor on (*Store).GetQuests' SQLite read path (inber.go:656-658) —
// the last unheld row of a mechanism that is written twice.
//
// Card d1682133-eeee-4b20-baca-8a08e7a7c6de, worked by the 220th nightly pass.
// Parent 6a6625a4, grandparent 23be5012. Onboarding note e72c6e27 carries the
// reachability map. Sibling test/the-difficulty-ladder-and-pagination-are-
// unpinned closed the twin in http_client.go and left this one row open.
//
// GetQuests is written once per data source and the two bodies are
// character-for-character identical through the difficulty ladder, the tier
// values, the progress constants and this floor:
//
//	internal/inber/inber.go:656        (*Store).GetQuests       SQLite-backed
//	internal/inber/http_client.go:159  (*HTTPClient).GetQuests  HTTP-backed
//
// Three of the four clusters were held on both sides. The floor was held only
// on the HTTP side, by TestHTTPQuestXPFloorLiftsAZeroRewardToOne, whose six
// rows this file copies onto the Store path. The two implementations differ by
// one level of indentation and by nothing else.
//
// This is insurance, not a bug fix. The floor is a correct value that nothing
// on this path was holding in place; scripts/sabotage-the-xp-floor.py scored
// every row here UNNOTICED against the suite as it stood. Nothing here is a
// live defect.
//
// ⛔ What this file deliberately does NOT do:
//
//  1. It does not score the one-step widening of the guard. `if xpReward < 1`
//     spells a bound one apart from the value it substitutes, so `< 2` only
//     additionally captures xpReward == 1 and assigns it 1 — the value it
//     already had. Enumerated over every token count 0..99999, `< 2` and `< 1`
//     agree at EVERY input; `< 3` first differs at 200 tokens. The scorer
//     carries `< 2` as a declared known-negative and pins the upward direction
//     with the two-step widening.
//
//  2. It does not re-pin xpForTokens' /100 divisor.
//     TestXPForTokensCountsTheHundredthTokenAndNotTheNinetyNinth
//     (naming_arithmetic_boundaries_test.go:47) holds it with the same straddle
//     pair. The scorer carries it as a neighbour row rather than as a claim, so
//     "already covered" is a measurement here and not a reading.
//
//  3. It does not touch createTestGatewayDB, for the reason
//     store_path_boundaries_test.go gives: four sibling tests assert exact
//     counts against that fixture. This file builds its quest through
//     oneQuestWithTokens, which stands up its own database per case.

import (
	"fmt"
	"testing"
)

// ---------------------------------------------------------------------------
// GetQuests — the XP floor (inber.go:656-658)
// ---------------------------------------------------------------------------

// xpForTokens is tokens/100, so every request under 100 tokens earns 0 XP and
// the floor lifts it to 1. XPReward was asserted nowhere in this package until
// the sibling pinned the HTTP copy, and nothing has ever asserted it on the
// Store path.
//
// The pair that pins the floor is (99, 100): at 99 the floor is doing the work
// and at 100 the division is. A case at 99 alone cannot tell a floor of 1 from
// a floor that fires at every token count, and a case at 100 alone cannot see
// the floor at all.
//
// The row at 200 is what makes the guard's bound observable upward. The floor
// and the value it substitutes are one apart, so a bound that drifts to 2
// changes nothing anywhere; the first input any drift can reach is 200 tokens,
// where a widened guard would clamp a legitimately-earned 2 back down to 1.
func TestQuestXPFloorLiftsAZeroRewardToOne(t *testing.T) {
	for _, tc := range []struct {
		tokens int
		want   int
	}{
		{0, 1},   // 0/100 == 0, floored
		{99, 1},  // still 0, floored — the floor is doing the work
		{100, 1}, // 100/100 == 1 on its own merits, floor inactive
		{199, 1}, // 1, floor inactive
		{200, 2}, // the floor must not clamp anything above 1
		{5000, 50},
	} {
		t.Run(fmt.Sprintf("tokens=%d", tc.tokens), func(t *testing.T) {
			if got := oneQuestWithTokens(t, tc.tokens).XPReward; got != tc.want {
				t.Errorf("tokens=%d: XPReward = %d, want %d", tc.tokens, got, tc.want)
			}
		})
	}
}

// The XP the floor guards is computed from inTokens + outTokens, and every
// fixture that reaches this code — oneQuestWithTokens above, and the ladder and
// tier tables in store_path_boundaries_test.go — sets outTokens to 0. A summand
// that is always zero is not summed by any test, so the whole output column
// could drop out of the arithmetic unnoticed.
//
// 100 and 100 is the pair that separates them: the sum earns 2 XP, while the
// input column alone earns 1. The two token counts are equal so that dropping
// EITHER summand gives the same wrong answer — this pins the addition, not the
// order of its operands. Both values sit above the floor, so nothing here is
// answered by the floor lifting a zero.
func TestQuestXPRewardCountsBothTokenColumns(t *testing.T) {
	gw := gatewayDBWith(t,
		map[string]string{"sess-1": "claxon"},
		[]requestRow{{sessionKey: "sess-1", status: "running", inputText: "task",
			inTokens: 100, outTokens: 100}})
	quests, err := storeOver(t, "", gw).GetQuests(50)
	if err != nil {
		t.Fatalf("GetQuests: %v", err)
	}
	if len(quests) != 1 {
		t.Fatalf("the fixture no longer reaches GetQuests: want exactly 1 quest, got %d", len(quests))
	}
	if got := quests[0].XPReward; got != 2 {
		t.Errorf("100 input + 100 output tokens: XPReward = %d, want 2 — earned by the "+
			"SUM of both columns; either column alone earns 1", got)
	}
}

// The HTTP copy sums two columns as well (http_client.go:157), and its own
// fixture has the same blind spot: oneHTTPQuestWithTokens passes outTokens 0 on
// every case, and no test in the package has ever sent a non-zero out_tokens
// down either path.
//
// ⚠️ This twin lives here rather than beside its sibling in
// http_client_boundaries_test.go, where it belongs by subject. The scorer builds
// its PRIOR column by moving THIS pass's file aside; a test written into a
// committed sibling would survive that move, so the HTTP row would report
// "already held" when this pass is what held it. Keeping both halves in one
// file is what makes the neighbour column mean anything.
//
// The 218th pass's rule is why the row is here at all: a mechanism spelled twice
// is one job, and closing one copy makes the other invisible, because a grep for
// the mechanism finds the sibling's test and reports it covered.
func TestHTTPQuestXPRewardCountsBothTokenColumns(t *testing.T) {
	client, served := clientOver(t, session("claxon", "running", 100, 100))
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
	if got := quests[0].XPReward; got != 2 {
		t.Errorf("100 in + 100 out tokens: XPReward = %d, want 2 — earned by the "+
			"SUM of both columns; either column alone earns 1", got)
	}
}
