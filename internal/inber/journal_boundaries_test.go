package inber

// Boundary tests for the held-item, tool-duration and journal-prose tables.
//
// Until this file existed, not one of these numbers was held in place: no test
// in the package called getHeldItemsForAgent, estimateToolDuration,
// generateNarrative, generateJournalTitle or generateActivityDescription at
// all. Every value below could be moved by one and the suite stayed green.
//
// Nothing here is a bug report. Each number is believed correct; what was
// missing was anything that would notice it changing. See
// scripts/sabotage-journal-tables.py for the measurement.
//
// Two rules this file follows, both inherited:
//
//   - A boundary is pinned by a test that varies the input ACROSS it. Single
//     -sided assertions move with the constant they claim to check, so every
//     threshold here gets a straddle pair (n-1 rejected, n accepted).
//   - getHeldItemsForAgent ranks a map, and Go randomises map iteration. Ties
//     in score are broken nondeterministically, so every fixture below gives
//     its qualifying activities DISTINCT counts. A tie here would be flaky,
//     not wrong.
//
// The three generate* methods are declared on *Store but never dereference the
// receiver, so (&Store{}) is enough and no database is involved.

import (
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// getHeldItemsForAgent — activity thresholds
//
// Trap, from the onboarding note: do NOT drive this through GetAgents.
// analyzeRecentActivity fails silently on the package fixture (its query wants
// sessions.initial_message and turns.content, which the fixture does not
// define), so it returns a zero ActivityAnalysis and the held-item logic only
// ever sees zeros. Hand-build the analysis instead.
// ---------------------------------------------------------------------------

// The default threshold is 5, and the two lowered ones are 2. Each is straddled
// so moving it in either direction reddens the suite: at threshold-1 the
// activity must not qualify, at threshold it must.
func TestHeldItemActivityThresholdsAreStraddled(t *testing.T) {
	cases := []struct {
		name      string
		threshold int
		build     func(count int) *ActivityAnalysis
		wantID    string
	}{
		{
			name:      "spawn has a lowered threshold of 2",
			threshold: 2,
			build:     func(c int) *ActivityAnalysis { return &ActivityAnalysis{SpawnCount: c} },
			wantID:    "claxon_horn",
		},
		{
			name:      "infra has a lowered threshold of 2",
			threshold: 2,
			build:     func(c int) *ActivityAnalysis { return &ActivityAnalysis{InfraWork: c} },
			wantID:    "engineers_wrench",
		},
		{
			name:      "edit uses the default threshold of 5",
			threshold: 5,
			build:     func(c int) *ActivityAnalysis { return &ActivityAnalysis{EditCount: c} },
			wantID:    "smithing_hammer",
		},
		{
			name:      "search uses the default threshold of 5",
			threshold: 5,
			build:     func(c int) *ActivityAnalysis { return &ActivityAnalysis{SearchCount: c} },
			wantID:    "magnifying_glass",
		},
		{
			name:      "docs uses the default threshold of 5",
			threshold: 5,
			build:     func(c int) *ActivityAnalysis { return &ActivityAnalysis{DocWriting: c} },
			wantID:    "scribes_scroll",
		},
		{
			name:      "debug uses the default threshold of 5",
			threshold: 5,
			build:     func(c int) *ActivityAnalysis { return &ActivityAnalysis{DebugSessions: c} },
			wantID:    "debugging_probe",
		},
		{
			name:      "create uses the default threshold of 5",
			threshold: 5,
			build:     func(c int) *ActivityAnalysis { return &ActivityAnalysis{CreateCount: c} },
			wantID:    "builders_trowel",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Below the threshold: no item at all.
			below := getHeldItemsForAgent(tc.build(tc.threshold - 1))
			if len(below) != 0 {
				t.Errorf("count %d (threshold-1) yielded %d held items, want 0: %v",
					tc.threshold-1, len(below), idsOf(below))
			}

			// At the threshold: exactly this item. The comparison is >=, so the
			// threshold value itself qualifies.
			at := getHeldItemsForAgent(tc.build(tc.threshold))
			if len(at) != 1 {
				t.Fatalf("count %d (threshold) yielded %d held items, want exactly 1: %v",
					tc.threshold, len(at), idsOf(at))
			}
			if at[0].ID != tc.wantID {
				t.Errorf("count %d yielded item %q, want %q", tc.threshold, at[0].ID, tc.wantID)
			}
		})
	}
}

// The catalog's seven priority values, each read off the returned struct.
//
// Driving one activity at a time is what makes these observable: the function
// returns at most two items, so a single fixture could never show all seven.
// This pins the absolute value, not just the ranking that
// TestHeldItemsSortByPriorityDescending covers.
func TestHeldItemCatalogPrioritiesAndIdentities(t *testing.T) {
	// count is set to each activity's own threshold, so this table also states
	// which activities are the lowered ones.
	cases := []struct {
		analysis     *ActivityAnalysis
		wantID       string
		wantName     string
		wantActivity string
		wantPriority int
	}{
		{&ActivityAnalysis{SpawnCount: 2}, "claxon_horn", "Claxon Horn", "spawn", 15},
		{&ActivityAnalysis{InfraWork: 2}, "engineers_wrench", "Engineer's Wrench", "infra", 12},
		{&ActivityAnalysis{CreateCount: 5}, "builders_trowel", "Builder's Trowel", "create", 11},
		{&ActivityAnalysis{EditCount: 5}, "smithing_hammer", "Smithing Hammer", "edit", 10},
		{&ActivityAnalysis{DebugSessions: 5}, "debugging_probe", "Debugging Probe", "debug", 9},
		{&ActivityAnalysis{SearchCount: 5}, "magnifying_glass", "Investigator's Glass", "search", 8},
		{&ActivityAnalysis{DocWriting: 5}, "scribes_scroll", "Scribe's Scroll", "docs", 7},
	}

	for _, tc := range cases {
		t.Run(tc.wantID, func(t *testing.T) {
			items := getHeldItemsForAgent(tc.analysis)
			if len(items) != 1 {
				t.Fatalf("got %d held items, want exactly 1: %v", len(items), idsOf(items))
			}
			got := items[0]
			if got.ID != tc.wantID {
				t.Errorf("ID = %q, want %q", got.ID, tc.wantID)
			}
			if got.Name != tc.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tc.wantName)
			}
			if got.ActivityType != tc.wantActivity {
				t.Errorf("ActivityType = %q, want %q", got.ActivityType, tc.wantActivity)
			}
			if got.Priority != tc.wantPriority {
				t.Errorf("Priority = %d, want %d", got.Priority, tc.wantPriority)
			}
		})
	}
}

// maxItems is 2. Three activities qualify with distinct counts, so the ranking
// is deterministic and the two highest-scoring survive.
func TestHeldItemsAreCappedAtTwo(t *testing.T) {
	// Distinct counts, all at or above the default threshold of 5.
	items := getHeldItemsForAgent(&ActivityAnalysis{
		EditCount:     9,
		CreateCount:   8,
		DebugSessions: 7,
	})

	if len(items) != 2 {
		t.Fatalf("three qualifying activities yielded %d held items, want 2 (maxItems): %v",
			len(items), idsOf(items))
	}

	// The cap is applied to the score ranking, so the lowest-scoring of the
	// three is the one dropped — regardless of its priority.
	for _, item := range items {
		if item.ID == "debugging_probe" {
			t.Errorf("lowest-scoring activity (debug, count 7) survived the cap: %v", idsOf(items))
		}
	}
}

// The final sort is by Priority descending, and it is a separate step from the
// score ranking. The fixture makes the two orders disagree: edit scores higher
// but spawn outranks it on priority, so a correct sort returns spawn first.
func TestHeldItemsSortByPriorityDescending(t *testing.T) {
	items := getHeldItemsForAgent(&ActivityAnalysis{
		EditCount:  9, // score 9, priority 10
		SpawnCount: 2, // score 2, priority 15 — lower score, higher priority
	})

	if len(items) != 2 {
		t.Fatalf("got %d held items, want 2: %v", len(items), idsOf(items))
	}
	if items[0].ID != "claxon_horn" || items[1].ID != "smithing_hammer" {
		t.Errorf("order = %v, want [claxon_horn smithing_hammer]: the higher priority (15) "+
			"must come first even though it scored lower (2 vs 9)", idsOf(items))
	}
	if items[0].Priority <= items[1].Priority {
		t.Errorf("priorities not descending: %d then %d", items[0].Priority, items[1].Priority)
	}
}

// No activity reaching any threshold yields no items at all — the empty case is
// an empty slice, not a default item.
func TestHeldItemsEmptyWhenNothingQualifies(t *testing.T) {
	// Every count one below its own threshold.
	items := getHeldItemsForAgent(&ActivityAnalysis{
		SpawnCount:    1,
		InfraWork:     1,
		EditCount:     4,
		SearchCount:   4,
		DocWriting:    4,
		DebugSessions: 4,
		CreateCount:   4,
	})
	if len(items) != 0 {
		t.Errorf("got %d held items, want 0: %v", len(items), idsOf(items))
	}
}

func idsOf(items []RPGHeldItem) []string {
	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = item.ID
	}
	return ids
}

// ---------------------------------------------------------------------------
// estimateToolDuration — the float table
// ---------------------------------------------------------------------------

func TestEstimateToolDurationTable(t *testing.T) {
	cases := []struct {
		tool string
		want float64
	}{
		{"read", 1.0},
		{"write", 1.0},
		{"edit", 1.0},
		{"exec", 3.0},
		{"process", 3.0},
		{"web_search", 2.0},
		{"web_fetch", 2.0},
		{"image", 4.0},
		{"tts", 4.0},
		{"browser", 5.0},
		// The default arm. Pinned by three inputs that must all reach it: an
		// unknown tool, the empty string, and a known tool in the wrong case —
		// the switch is case-sensitive and nothing said so before.
		{"no_such_tool", 1.5},
		{"", 1.5},
		{"Read", 1.5},
	}

	for _, tc := range cases {
		t.Run(tc.tool, func(t *testing.T) {
			if got := estimateToolDuration(tc.tool); got != tc.want {
				t.Errorf("estimateToolDuration(%q) = %v, want %v", tc.tool, got, tc.want)
			}
		})
	}
}

// The arms must stay distinguishable from each other. A table where two arms
// collapse onto one value would still pass the table above for the merged
// value, so the separation is asserted on its own.
func TestEstimateToolDurationArmsAreDistinct(t *testing.T) {
	seen := map[float64]string{}
	for _, tool := range []string{"read", "exec", "web_search", "image", "browser", "no_such_tool"} {
		d := estimateToolDuration(tool)
		if other, dup := seen[d]; dup {
			t.Errorf("%q and %q both return %v; the arms are no longer separable", other, tool, d)
		}
		seen[d] = tool
	}
}

// ---------------------------------------------------------------------------
// generateActivityDescription — the 1000 / 500 token ladder
// ---------------------------------------------------------------------------

// neutralInput contains none of the keywords the function matches on, so the
// token ladder is what decides the output.
const neutralInput = "wander the quiet meadow"

func TestActivityDescriptionTokenLadderIsStraddled(t *testing.T) {
	s := &Store{}

	cases := []struct {
		tokens       int
		wantFragment string
		why          string
	}{
		{1001, "epic quest", "just above the 1000 boundary"},
		{1000, "challenging task", "1000 is NOT > 1000, so it falls to the next arm"},
		{501, "challenging task", "just above the 500 boundary"},
		{500, "routine task", "500 is NOT > 500, so it falls through"},
		{0, "routine task", "the floor"},
	}

	for _, tc := range cases {
		t.Run(tc.why, func(t *testing.T) {
			got := s.generateActivityDescription(neutralInput, "", tc.tokens, 0)
			if !strings.Contains(got, tc.wantFragment) {
				t.Errorf("tokens=%d: got %q, want it to contain %q (%s)",
					tc.tokens, got, tc.wantFragment, tc.why)
			}
		})
	}
}

// Every keyword arm, and the fact that they outrank the token ladder: each case
// passes a token count that would otherwise select "epic quest".
func TestActivityDescriptionKeywordsOutrankTheTokenLadder(t *testing.T) {
	s := &Store{}
	const epicTokens = 9999

	cases := []struct {
		input        string
		wantFragment string
	}{
		{"fix the gate", "Battled bugs"},
		{"debug the gate", "Battled bugs"},
		{"error in the gate", "Battled bugs"},
		{"test the gate", "test scrolls"},
		{"spec the gate", "test scrolls"},
		{"refactor the gate", "Organized the code realm"},
		{"clean the gate", "Organized the code realm"},
		{"organize the gate", "Organized the code realm"},
		{"document the gate", "sacred documentation scrolls"},
		{"readme for the gate", "sacred documentation scrolls"},
		{"feature for the gate", "magical features"},
		{"implement the gate", "magical features"},
		{"build the gate", "magical features"},
		{"research the gate", "mystical research"},
		{"analyze the gate", "mystical research"},
		{"investigate the gate", "mystical research"},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := s.generateActivityDescription(tc.input, "", epicTokens, 0)
			if !strings.Contains(got, tc.wantFragment) {
				t.Errorf("input %q: got %q, want it to contain %q", tc.input, got, tc.wantFragment)
			}
			if strings.Contains(got, "epic quest") {
				t.Errorf("input %q at %d tokens fell through to the token ladder: %q",
					tc.input, epicTokens, got)
			}
		})
	}
}

// The keyword match is case-insensitive: the function lowercases its input
// before testing. Nothing pinned that, so the ToLower could have been dropped.
func TestActivityDescriptionKeywordMatchIsCaseInsensitive(t *testing.T) {
	s := &Store{}
	got := s.generateActivityDescription("FIX THE GATE", "", 10, 0)
	if !strings.Contains(got, "Battled bugs") {
		t.Errorf("uppercase input did not match the keyword arm: %q", got)
	}
}

// ---------------------------------------------------------------------------
// generateNarrative — the 2000 / 1000 ladder and the list-join arms
// ---------------------------------------------------------------------------

func testAgent() *RPGAgent {
	return &RPGAgent{Name: "Brigid", Class: "Smith"}
}

// A fixed date, so the weekday in the opening line is stable.
func testDate() time.Time {
	return time.Date(2026, 8, 12, 9, 0, 0, 0, time.UTC) // a Wednesday
}

func TestNarrativeTokenLadderIsStraddled(t *testing.T) {
	s := &Store{}

	cases := []struct {
		tokens       int
		wantFragment string
		why          string
	}{
		{2001, "vast amounts of magical energy", "just above the 2000 boundary"},
		{2000, "exertions were considerable", "2000 is NOT > 2000"},
		{1001, "exertions were considerable", "just above the 1000 boundary"},
		{1000, "measured use of their powers", "1000 is NOT > 1000"},
		{1, "measured use of their powers", "the floor above zero"},
	}

	for _, tc := range cases {
		t.Run(tc.why, func(t *testing.T) {
			got := s.generateNarrative(testAgent(), []string{"forged a blade"},
				JournalStats{TokensUsed: tc.tokens}, testDate())
			if !strings.Contains(got, tc.wantFragment) {
				t.Errorf("tokens=%d: want %q in narrative (%s), got:\n%s",
					tc.tokens, tc.wantFragment, tc.why, got)
			}
		})
	}
}

// TokensUsed > 0 gates the whole sentence, and zero must omit it entirely
// rather than reporting "0 tokens".
func TestNarrativeOmitsTokenSentenceAtZero(t *testing.T) {
	s := &Store{}
	got := s.generateNarrative(testAgent(), []string{"forged a blade"},
		JournalStats{TokensUsed: 0}, testDate())
	if strings.Contains(got, "tokens of concentrated effort") {
		t.Errorf("zero tokens still produced the token sentence:\n%s", got)
	}
}

// The two joining arms are separate branches and each needs a length that
// reaches it: one activity reaches neither, two reach "and finally", three
// reach both.
func TestNarrativeListJoinArms(t *testing.T) {
	s := &Store{}
	stats := JournalStats{}

	t.Run("one activity uses neither joiner", func(t *testing.T) {
		got := s.generateNarrative(testAgent(), []string{"Forged a blade"}, stats, testDate())
		if strings.Contains(got, ", then ") {
			t.Errorf("single activity emitted \", then \":\n%s", got)
		}
		if strings.Contains(got, "and finally ") {
			t.Errorf("single activity emitted \"and finally \":\n%s", got)
		}
	})

	t.Run("two activities use only the final joiner", func(t *testing.T) {
		got := s.generateNarrative(testAgent(),
			[]string{"Forged a blade", "Banked the fire"}, stats, testDate())
		if strings.Contains(got, ", then ") {
			t.Errorf("two activities emitted \", then \":\n%s", got)
		}
		if strings.Count(got, "and finally ") != 1 {
			t.Errorf("want exactly one \"and finally \", got %d:\n%s",
				strings.Count(got, "and finally "), got)
		}
	})

	t.Run("three activities use both joiners once each", func(t *testing.T) {
		got := s.generateNarrative(testAgent(),
			[]string{"Forged a blade", "Banked the fire", "Swept the hall"}, stats, testDate())
		if strings.Count(got, ", then ") != 1 {
			t.Errorf("want exactly one \", then \", got %d:\n%s",
				strings.Count(got, ", then "), got)
		}
		if strings.Count(got, "and finally ") != 1 {
			t.Errorf("want exactly one \"and finally \", got %d:\n%s",
				strings.Count(got, "and finally "), got)
		}
	})

	// Activities are lowercased as they are joined.
	t.Run("activities are lowercased", func(t *testing.T) {
		got := s.generateNarrative(testAgent(), []string{"FORGED A BLADE"}, stats, testDate())
		if !strings.Contains(got, "forged a blade") {
			t.Errorf("activity was not lowercased into the narrative:\n%s", got)
		}
	})
}

// An empty activity list returns early with the contemplation line and none of
// the stats prose, however rich the stats are.
func TestNarrativeEmptyActivitiesReturnsEarly(t *testing.T) {
	s := &Store{}
	got := s.generateNarrative(testAgent(), nil,
		JournalStats{TokensUsed: 5000, QuestsCompleted: 9, XPGained: 900}, testDate())

	if !strings.Contains(got, "quiet contemplation") {
		t.Errorf("empty activity list did not produce the contemplation line:\n%s", got)
	}
	// The early return means no stats sentence can appear.
	for _, leaked := range []string{"tokens of concentrated effort", "experience points", "quests demanded"} {
		if strings.Contains(got, leaked) {
			t.Errorf("early return leaked stats prose (%q):\n%s", leaked, got)
		}
	}
}

// The singular/plural arms for the three counted stats. Each is a separate
// branch on == 1, and each has a zero case that must emit nothing at all.
func TestNarrativeSingularAndPluralArms(t *testing.T) {
	s := &Store{}
	activities := []string{"forged a blade"}

	cases := []struct {
		name    string
		stats   JournalStats
		want    string
		notWant string
	}{
		{"one quest completed", JournalStats{QuestsCompleted: 1},
			"A single quest called to them", "quests demanded their expertise"},
		{"two quests completed", JournalStats{QuestsCompleted: 2},
			"2 quests demanded their expertise", "A single quest"},
		{"one collaboration", JournalStats{Collaborations: 1},
			"a fellow adventurer", "allies to form"},
		{"two collaborations", JournalStats{Collaborations: 2},
			"summoned 2 allies", "a fellow adventurer"},
		{"one quest failed", JournalStats{QuestsFailed: 1},
			"one quest did not reach completion", "quests encountered obstacles"},
		{"two quests failed", JournalStats{QuestsFailed: 2},
			"While 2 quests encountered obstacles", "one quest did not reach completion"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := s.generateNarrative(testAgent(), activities, tc.stats, testDate())
			if !strings.Contains(got, tc.want) {
				t.Errorf("want %q in narrative, got:\n%s", tc.want, got)
			}
			if strings.Contains(got, tc.notWant) {
				t.Errorf("the other arm leaked (%q):\n%s", tc.notWant, got)
			}
		})
	}

	// Zero of each emits neither arm.
	t.Run("zeros emit nothing", func(t *testing.T) {
		got := s.generateNarrative(testAgent(), activities, JournalStats{}, testDate())
		for _, leaked := range []string{
			"A single quest", "quests demanded", "fellow adventurer", "allies to form",
			"did not reach completion", "encountered obstacles", "experience points",
		} {
			if strings.Contains(got, leaked) {
				t.Errorf("zero stats leaked %q:\n%s", leaked, got)
			}
		}
	})
}

// XPGained > 0 is its own gate, and the count is interpolated.
func TestNarrativeExperienceLine(t *testing.T) {
	s := &Store{}
	activities := []string{"forged a blade"}

	got := s.generateNarrative(testAgent(), activities, JournalStats{XPGained: 42}, testDate())
	if !strings.Contains(got, "gaining 42 experience points") {
		t.Errorf("want the XP line with the count, got:\n%s", got)
	}

	none := s.generateNarrative(testAgent(), activities, JournalStats{XPGained: 0}, testDate())
	if strings.Contains(none, "experience points") {
		t.Errorf("zero XP still produced the XP line:\n%s", none)
	}
}

// The opening reads the weekday off the date argument, so the date is load
// -bearing and not decoration.
func TestNarrativeOpeningNamesTheWeekday(t *testing.T) {
	s := &Store{}
	got := s.generateNarrative(testAgent(), []string{"forged a blade"},
		JournalStats{}, testDate())
	if !strings.Contains(got, "On this Wednesday") {
		t.Errorf("narrative did not name the date's weekday:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// generateJournalTitle — the >= 3 and >= 500 boundaries
// ---------------------------------------------------------------------------

// Both thresholds are inclusive (>=), so each straddle pins the value AND the
// direction of the comparison.
func TestJournalTitleThresholdsAreStraddled(t *testing.T) {
	s := &Store{}
	agent := testAgent()

	t.Run("quests completed >= 3", func(t *testing.T) {
		below := s.generateJournalTitle(agent, JournalStats{QuestsCompleted: 2}, 1)
		if strings.Contains(below, "Great Endeavors") {
			t.Errorf("2 completed quests reached the >= 3 arm: %q", below)
		}
		at := s.generateJournalTitle(agent, JournalStats{QuestsCompleted: 3}, 1)
		if !strings.Contains(at, "The Great Endeavors of Brigid") {
			t.Errorf("3 completed quests did not reach the >= 3 arm: %q", at)
		}
	})

	t.Run("XP gained >= 500", func(t *testing.T) {
		// QuestsCompleted stays below 3 so the earlier arm cannot absorb these.
		below := s.generateJournalTitle(agent, JournalStats{XPGained: 499}, 1)
		if strings.Contains(below, "Epic Trials") {
			t.Errorf("499 XP reached the >= 500 arm: %q", below)
		}
		at := s.generateJournalTitle(agent, JournalStats{XPGained: 500}, 1)
		if !strings.Contains(at, "Epic Trials of Brigid the Smith") {
			t.Errorf("500 XP did not reach the >= 500 arm: %q", at)
		}
	})
}

// The arms are ordered, and the order is itself behaviour: a stat that would
// match a later arm must lose to an earlier one.
func TestJournalTitleArmPrecedence(t *testing.T) {
	s := &Store{}
	agent := testAgent()

	cases := []struct {
		name       string
		questCount int
		stats      JournalStats
		want       string
	}{
		{
			// questCount == 0 short-circuits everything, however good the stats.
			name:       "no quests outranks every stat",
			questCount: 0,
			stats:      JournalStats{QuestsCompleted: 9, XPGained: 9000, Collaborations: 5},
			want:       "A Day of Rest for Brigid",
		},
		{
			name:       "completed quests outrank XP",
			questCount: 1,
			stats:      JournalStats{QuestsCompleted: 3, XPGained: 9000},
			want:       "The Great Endeavors of Brigid",
		},
		{
			name:       "XP outranks collaborations",
			questCount: 1,
			stats:      JournalStats{XPGained: 500, Collaborations: 5},
			want:       "Epic Trials of Brigid the Smith",
		},
		{
			name:       "collaborations outrank failures",
			questCount: 1,
			stats:      JournalStats{Collaborations: 1, QuestsFailed: 5},
			want:       "Brigid and the Fellowship",
		},
		{
			name:       "failures outrank the fallback",
			questCount: 1,
			stats:      JournalStats{QuestsFailed: 1},
			want:       "Trials and Tribulations: Brigid's Journey",
		},
		{
			name:       "the fallback",
			questCount: 1,
			stats:      JournalStats{},
			want:       "Adventures of Brigid the Smith",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := s.generateJournalTitle(agent, tc.stats, tc.questCount)
			if got != tc.want {
				t.Errorf("title = %q, want %q", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// truncateText — the caller's shipping value is NOT pinned here, and saying so
// is the point.
//
// truncateText has exactly one production call site, (*Store).GetAgentJournal,
// which passes 100. Reaching that call needs a fixture schema this package does
// not have: GetAgentJournal's query selects r.output_text, and the requests
// table in inber_test.go does not define that column. So the 100 that actually
// ships is unpinned, and the arithmetic below is all that is held.
//
// Do not read these as covering the caller. Pinning the shipping 100 is the
// fixture-schema child's job.
// ---------------------------------------------------------------------------

func TestTruncateTextArithmetic(t *testing.T) {
	// The guard is <=, so text exactly at maxLen is returned untouched.
	if got := truncateText("abcdefghij", 10); got != "abcdefghij" {
		t.Errorf("text at exactly maxLen was altered: %q", got)
	}
	if got := truncateText("abcde", 10); got != "abcde" {
		t.Errorf("text below maxLen was altered: %q", got)
	}

	// One byte over: cut to maxLen-3 and append the three-byte ellipsis, so the
	// result is exactly maxLen. Both halves of that arithmetic are asserted —
	// the cut length and the total — because a drift in either alone would
	// otherwise be invisible.
	got := truncateText("abcdefghijk", 10)
	if got != "abcdefg..." {
		t.Errorf("truncateText(11 bytes, 10) = %q, want %q", got, "abcdefg...")
	}
	if len(got) != 10 {
		t.Errorf("result is %d bytes, want exactly maxLen (10): %q", len(got), got)
	}
}
