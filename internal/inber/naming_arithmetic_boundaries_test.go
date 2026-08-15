package inber

// Boundary VALUES for this package's pure text and naming arithmetic — the
// numbers reached by calling a function directly, with no database, no HTTP
// server and no fixture.
//
// The truncation family at these same sites is pinned by
// truncation_budgets_test.go and is deliberately not repeated here:
// generateQuestName's 80/60/57 and truncateText's maxLen-3 belong to that file.
// What is left is the arithmetic around them — the complexity ladder that
// decides which quest-name pattern is used, the word-length floors and term
// limits in the key-term extractor, the spawn-name floor, and the two XP and
// energy formulas.
//
// Every number below was measured UNNOTICED or half-pinned before this file
// existed. None of them was wrong. They were correct values with nothing
// holding them in place, which is what this sweep exists to change.
//
// ⚠️ A CLAMP IS TWO BOUNDARIES AND MOST OF THESE ARE REACHABLE FROM ONE SIDE.
// analyzeTaskForNaming applies five clamps in a fixed order, and the ceiling of
// a clamp can only be observed when the sum actually overruns it. No task type
// starts above 4, so `min(5, complexity+1)` on the urgent line can never bind —
// its 5 is pinnable from below (drift to 4 changes an answer) and unpinnable
// from above (drift to 6 changes nothing, ever). `max(1, complexity-1)` is the
// mirror image. Each test below says which side it holds, so a later reader
// does not record an unpinnable side as an untested one.
//
// Assertions are on lengths, term counts and complexity numbers rather than on
// whole quest-name strings: sibling branches rewrite the capitalisation helper
// under this cluster, and exact-string assertions would break on merge. All
// inputs here are plain ASCII with no apostrophes for the same reason.

import (
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// xpForTokens: the /100 divisor
// ---------------------------------------------------------------------------

// TestXPForTokens (inber_test.go) checks 0→0, 100→1 and 1500→15. That is
// half a pin: 1500/99 == 15 and 100/99 == 1, so moving the divisor DOWN to 99
// survives all three cases, while moving it up to 101 dies at 100→1. One more
// case either side of the first whole XP closes it.
func TestXPForTokensCountsTheHundredthTokenAndNotTheNinetyNinth(t *testing.T) {
	if got := xpForTokens(99); got != 0 {
		t.Errorf("xpForTokens(99) = %d, want 0 — the last token before the "+
			"first whole XP; this case is what separates a /100 divisor from /99", got)
	}
	if got := xpForTokens(100); got != 1 {
		t.Errorf("xpForTokens(100) = %d, want 1 — the token that earns the "+
			"first whole XP; this case is what separates /100 from /101", got)
	}
}

// ---------------------------------------------------------------------------
// energyFromActivity: the nil answer 100, the cap 100, the floor 20, the
// recovery rate 10 per hour
// ---------------------------------------------------------------------------
//
// TestEnergyFromActivity (inber_test.go:349) asserts only that a recent
// timestamp lands in the range 20..30 and that an old one is 100. A range is
// not a value: the floor can move within it, and the rate can move by 2 without
// leaving it.
//
// ⚠️ There is no clock seam here — the function calls time.Since directly — so
// every timestamp below is built as an offset from time.Now(). Elapsed time is
// always slightly MORE than the nominal offset, so the product is always
// slightly above the integer the case expects and truncation lands on it.
// Offsets whose expected energy sits just below an integer would be flaky; the
// ones chosen do not.

func hoursAgo(h float64) *time.Time {
	ts := time.Now().Add(-time.Duration(h * float64(time.Hour)))
	return &ts
}

func TestEnergyFromActivityAnswersAFullHundredForAnAgentNeverSeen(t *testing.T) {
	// A distinct literal from the cap below, reached by a different branch.
	if got := energyFromActivity(nil); got != 100 {
		t.Errorf("energyFromActivity(nil) = %d, want 100", got)
	}
}

func TestEnergyFromActivityRecoversTenPerHourFromAFloorOfTwenty(t *testing.T) {
	// Two points pin two numbers: one hour gives floor+rate, two hours gives
	// floor+2*rate. Either alone leaves the pair free to trade against each
	// other, which is how a floor of 20 and a rate of 10 survived a range check.
	for _, c := range []struct {
		hours float64
		want  int
	}{
		{1, 30},
		{2, 40},
		{7, 90},
	} {
		if got := energyFromActivity(hoursAgo(c.hours)); got != c.want {
			t.Errorf("energyFromActivity(%gh ago) = %d, want %d (floor 20 + %g × 10)",
				c.hours, got, c.want, c.hours)
		}
	}
}

func TestEnergyFromActivityCapsAtAHundred(t *testing.T) {
	// A straddle across the cap. 7.9 hours is the last point below it: 20 +
	// 79 = 99. Nine hours would reach 110 uncapped, so this pair separates the
	// cap from the arithmetic feeding it, and moving the cap either way changes
	// exactly one of the two answers.
	if got := energyFromActivity(hoursAgo(7.9)); got != 99 {
		t.Errorf("energyFromActivity(7.9h ago) = %d, want 99 — one below the cap", got)
	}
	if got := energyFromActivity(hoursAgo(9)); got != 100 {
		t.Errorf("energyFromActivity(9h ago) = %d, want the cap 100 (110 uncapped)", got)
	}
}

// ---------------------------------------------------------------------------
// analyzeTaskForNaming: the per-type complexity ladder
// ---------------------------------------------------------------------------

// The complexity this returns chooses which prefix table generateProceduralQuestName
// reads, so every one of these numbers is visible in a quest name. Nothing
// varied them before this test: the package's naming tests all drive one input.
//
// Each keyword below is the FIRST case in the switch that its text matches, and
// the filler "zzz" matches no keyword in any arm — containsAnyKeyword is a
// substring test, so a filler word sharing three letters with any keyword would
// silently select a different arm.
func TestAnalyzeTaskForNamingReturnsOneComplexityPerTaskType(t *testing.T) {
	for _, c := range []struct {
		text           string
		wantType       string
		wantComplexity int
	}{
		{"bug the zzz", "debugging", 3},
		{"build the zzz", "development", 4},
		{"deploy the zzz", "deployment", 3},
		{"test the zzz", "testing", 2},
		{"document the zzz", "documentation", 2},
		{"refactor the zzz", "optimization", 3},
		{"design the zzz", "design", 4},
		{"monitor the zzz", "monitoring", 2},
		{"zzz the zzz", "general", 2}, // the default, matching no arm
	} {
		gotType, gotComplexity := analyzeTaskForNaming(c.text, "running")
		if gotType != c.wantType || gotComplexity != c.wantComplexity {
			t.Errorf("analyzeTaskForNaming(%q) = (%q, %d), want (%q, %d)",
				c.text, gotType, gotComplexity, c.wantType, c.wantComplexity)
		}
	}
}

// ---------------------------------------------------------------------------
// analyzeTaskForNaming: the 100-byte length bump
// ---------------------------------------------------------------------------

// A long task is treated as one degree more complex. The threshold is a byte
// length, so the straddle is two inputs one byte apart that are otherwise
// identical — and both must match no keyword, or the arm they select moves with
// them and the case measures the switch instead of the bump.
func TestAnalyzeTaskForNamingBumpsComplexityPastAHundredBytes(t *testing.T) {
	if _, got := analyzeTaskForNaming(strings.Repeat("z", 100), "running"); got != 2 {
		t.Errorf("a 100-byte keyword-free task got complexity %d, want the "+
			"unbumped default 2 — 100 bytes is not yet over the threshold", got)
	}
	if _, got := analyzeTaskForNaming(strings.Repeat("z", 101), "running"); got != 3 {
		t.Errorf("a 101-byte keyword-free task got complexity %d, want 3 — "+
			"one byte over the threshold is what triggers the bump", got)
	}
}

// ---------------------------------------------------------------------------
// analyzeTaskForNaming: the five clamps, each from the side it can be seen
// ---------------------------------------------------------------------------

// The three clamps that CAN bind. "develop a complex thing" is development (4)
// plus the complex bump (+2) = 6, which the ceiling cuts to 5. Adding a long
// body or an error status pushes an already-clamped 5 past the ceiling again,
// which is the only way to reach the ceilings on those two lines at all.
func TestAnalyzeTaskForNamingHoldsComplexityAtItsCeiling(t *testing.T) {
	const long = " " // filler joined below; keyword-free

	for _, c := range []struct {
		name   string
		text   string
		status string
	}{
		{"the complex bump alone overruns it (4+2)", "develop a complex thing", "running"},
		{"the length bump overruns an already-clamped 5",
			"develop a complex thing" + long + strings.Repeat("z", 101), "running"},
		{"the error bump overruns an already-clamped 5", "develop a complex thing", "error"},
	} {
		if _, got := analyzeTaskForNaming(c.text, c.status); got != 5 {
			t.Errorf("%s: complexity = %d, want the ceiling 5", c.name, got)
		}
	}
}

// ⚠️ The two clamps that CANNOT bind, each held from its one reachable side.
//
// No arm of the switch returns more than 4, and the urgent line runs before any
// other bump, so `min(5, complexity+1)` there sees at most 4+1 = 5 and never
// cuts anything. Its ceiling is therefore unobservable from above — raising it
// to 6 cannot change any answer this function can produce — but lowering it to
// 4 changes design+urgent from 5 to 4, and that is what this pins.
//
// `max(1, complexity-1)` is the mirror: complexity is never below 2 when it
// runs, so the floor never lifts anything, but raising it to 2 changes
// simple+testing from 1 to 2.
//
// Recording this is the point. A later reader measuring these two lines will
// find one direction permanently UNNOTICED and should not file it as a gap: it
// is a dominated no-op, and the only honest fix would be a product change to
// the ladder feeding it.
func TestAnalyzeTaskForNamingPinsItsDominatedClampsFromTheReachableSide(t *testing.T) {
	if _, got := analyzeTaskForNaming("urgent design work", "running"); got != 5 {
		t.Errorf("urgent + design = %d, want 5 — design is 4 and urgent adds one; "+
			"this holds the urgent ceiling from BELOW only, which is the only "+
			"side it has", got)
	}
	if _, got := analyzeTaskForNaming("simple testing job", "running"); got != 1 {
		t.Errorf("simple + testing = %d, want 1 — testing is 2 and simple takes "+
			"one off; this holds the floor from ABOVE only, which is the only "+
			"side it has", got)
	}
}

// ---------------------------------------------------------------------------
// extractKeyTermsForNaming: two word-length floors and two term limits
// ---------------------------------------------------------------------------
//
// extract_key_terms_characterisation_test.go pins this function's ANSWERS
// against a shared corpus, and — measured, not assumed — that corpus already
// catches all five mutations these four tests are aimed at. The scorer reports
// every key-term row as "already held by a neighbour" for exactly that reason.
//
// They are kept anyway, for two reasons worth stating rather than leaving a
// later reader to wonder why a redundant test exists:
//
//  1. A golden file is regenerated by hand when an answer changes on purpose.
//     Regenerating it after an ACCIDENTAL drift records the drift as the new
//     truth, silently — the golden moves with the code. An assertion that names
//     the number cannot be updated by regeneration, so it still fails.
//  2. The corpus lives on a sibling branch that is not merged to main. A worker
//     grepping main for coverage of these floors would find none and would file
//     them as gaps. Saying here that they are held, and by what, is what stops
//     that measurement being taken a third time.
//
// The two passes disagree about which end of the text they read: the priority
// pass scans forwards, the fallback scan runs backwards from the last word. So
// an input whose priority word is NOT last tells the two apart, which is what
// makes the priority floor observable at all.

func TestExtractKeyTermsPriorityPassNeedsMoreThanTwoLetters(t *testing.T) {
	// "api" is the shortest priority word in the table at three letters. If the
	// floor moved up, the priority pass would skip it and the backward scan
	// would answer instead — and the backward scan reads "zzz" too, so the term
	// COUNT changes and the case does not depend on capitalisation.
	got := extractKeyTermsForNaming("api zzz")
	if len(got) != 1 || got[0] != "Api" {
		t.Errorf("extractKeyTermsForNaming(%q) = %q, want exactly [Api] from the "+
			"priority pass; two terms means the three-letter priority word was "+
			"skipped and the backward scan answered instead", "api zzz", got)
	}
}

func TestExtractKeyTermsPriorityPassStopsAtTwoTerms(t *testing.T) {
	got := extractKeyTermsForNaming("api database server zzz")
	if len(got) != 2 {
		t.Errorf("extractKeyTermsForNaming(%q) returned %d terms %q, want exactly 2 — "+
			"three priority words are present and the pass is limited to two",
			"api database server zzz", len(got), got)
	}
}

func TestExtractKeyTermsBackwardScanNeedsMoreThanTwoLetters(t *testing.T) {
	// "zz" is two letters and must be skipped; "qqq" is three and must not be.
	// Neither is a stop word, so the floor is the only thing separating them.
	got := extractKeyTermsForNaming("zz qqq")
	if len(got) != 1 || got[0] != "Qqq" {
		t.Errorf("extractKeyTermsForNaming(%q) = %q, want exactly [Qqq] — a "+
			"two-letter word is below the floor and a three-letter one is above it",
			"zz qqq", got)
	}
}

func TestExtractKeyTermsBackwardScanStopsAtTwoTerms(t *testing.T) {
	got := extractKeyTermsForNaming("zzz qqq www")
	if len(got) != 2 {
		t.Errorf("extractKeyTermsForNaming(%q) returned %d terms %q, want exactly 2 — "+
			"three eligible words are present and the scan is limited to two",
			"zzz qqq www", len(got), got)
	}
}

// ---------------------------------------------------------------------------
// extractAgentNameFromSpawn: the candidate word-length floor
// ---------------------------------------------------------------------------

// The named-agent list is checked first, so the floor is only reachable through
// the "spawn X to" pattern with a word that is not one of the nine known names.
// A straddle needs both sides: a three-letter candidate that IS returned and a
// two-letter one that is NOT, or a floor moving down would look identical to a
// floor holding.
func TestExtractAgentNameFromSpawnNeedsMoreThanTwoLetters(t *testing.T) {
	if got := extractAgentNameFromSpawn("spawn zed to do zzz"); got != "zed" {
		t.Errorf("extractAgentNameFromSpawn with a three-letter candidate = %q, "+
			"want %q — three letters clears the floor", got, "zed")
	}
	if got := extractAgentNameFromSpawn("spawn zo to do zzz"); got != "" {
		t.Errorf("extractAgentNameFromSpawn with a two-letter candidate = %q, "+
			"want \"\" — two letters is below the floor", got)
	}
}
