package api

import (
	"strings"
	"testing"
	"unicode/utf8"
)

const fourByteRune = "\U0001D11E" // U+1D11E MUSICAL SYMBOL G CLEF, 4 bytes

// TestSpawnEventMessageNeverSplitsARune slides a four-byte rune through the cut
// at byte 100 of the task text. Offsets 1, 2 and 3 straddle the cut; offsets 0
// and 4 land on a rune boundary and are the known-negative controls — they pass
// against the unfixed byte-cut too, so a green run means the test detects a
// split rune rather than merely detecting non-ASCII input.
func TestSpawnEventMessageNeverSplitsARune(t *testing.T) {
	const cutAt = 100

	for offset := 0; offset <= 4; offset++ {
		task := strings.Repeat("a", cutAt-4+offset) + fourByteRune + strings.Repeat("b", 40)
		got := spawnEventMessage("spawn_started", "claxon", "", task)

		kind := "straddling"
		if offset == 0 || offset == 4 {
			kind = "boundary-aligned control"
		}

		if !utf8.ValidString(got) {
			t.Errorf("offset %d (%s): message is not valid UTF-8: %q (% x)", offset, kind, got, got)
		}
		// Validity alone is not falsifiable: a formatter that dropped the task
		// entirely would still produce valid UTF-8. Pin what survives too.
		if lead := strings.Repeat("a", cutAt-4); !strings.Contains(got, lead) {
			t.Errorf("offset %d (%s): lost more than the straddling rune — %d leading bytes of the task did not survive: %q",
				offset, kind, cutAt-4, got)
		}
	}
}

// The other three branches carry no cut. They are here because the cut lives in
// a switch, and a mutation that routes spawn_started down the wrong arm would
// otherwise show up only as a missing truncation.
func TestSpawnEventMessageRendersEachEventType(t *testing.T) {
	cases := []struct {
		eventType string
		label     string
		task      string
		want      string
	}{
		{"spawn_started", "", "audit the logs", "🚀 claxon spawned a sub-agent: audit the logs"},
		{"spawn_started", "nightly", "audit the logs", "🚀 claxon spawned a sub-agent (nightly): audit the logs"},
		{"spawn_completed", "", "", "✅ Sub-agent completed task for claxon"},
		{"spawn_completed", "nightly", "", "✅ Sub-agent completed task for claxon (nightly)"},
		{"spawn_failed", "", "", "❌ Sub-agent failed task for claxon"},
		{"spawn_failed", "nightly", "", "❌ Sub-agent failed task for claxon (nightly)"},
		{"something_else", "", "", "📍 Spawn event something_else for claxon"},
	}
	for _, c := range cases {
		if got := spawnEventMessage(c.eventType, "claxon", c.label, c.task); got != c.want {
			t.Errorf("spawnEventMessage(%q, claxon, %q, %q) = %q, want %q",
				c.eventType, c.label, c.task, got, c.want)
		}
	}
}

// A task shorter than the budget must not be truncated, and must not gain an
// ellipsis — the cut and the ellipsis travel together and both are conditional.
func TestSpawnEventMessageLeavesShortTasksWhole(t *testing.T) {
	task := strings.Repeat("a", 100)
	got := spawnEventMessage("spawn_started", "claxon", "", task)
	if strings.HasSuffix(got, "...") {
		t.Errorf("a task of exactly the budget was truncated: %q", got)
	}
	if !strings.HasSuffix(got, task) {
		t.Errorf("a task within budget did not survive whole: %q", got)
	}
}
