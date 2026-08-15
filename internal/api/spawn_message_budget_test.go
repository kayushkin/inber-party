package api

// Boundary VALUES for the spawn-message task cut.
//
// spawn_message_test.go pins the rune-boundary mechanism and pins the guard
// from BELOW — TestSpawnEventMessageLeavesShortTasksWhole passes a task of
// exactly 100 bytes and asserts it is not truncated, which separates `> 100`
// from `> 99`. Nothing separates it from `> 101`, and nothing at all looks at
// the budget: `TruncateAtRuneBoundary(task, 100)` passes its number as a call
// argument, so it does not even appear in a grep for numeric comparisons in
// this file. It is the least visible of the two and the one that decides what
// the dashboard actually shows.

import (
	"strings"
	"testing"
)

// TestASpawnTaskOneByteOverTheGuardIsCutToTheBudget is the missing upper half
// of the straddle pair, and it pins the call-argument budget in the same
// assertion by comparing the whole message.
//
// This site puts the ellipsis OUTSIDE the budget, so a cut message carries 100
// bytes of task plus three more. logstack.go and generateQuestName put it
// inside. See the ⛔ note in internal/logstack/title_budget_test.go: the
// disagreement is deliberate and pinning it is not endorsing it.
func TestASpawnTaskOneByteOverTheGuardIsCutToTheBudget(t *testing.T) {
	const (
		guard  = 100
		budget = 100
	)
	task := strings.Repeat("a", guard+1)

	got := spawnEventMessage("spawn_started", "claxon", "", task)
	want := "🚀 claxon spawned a sub-agent: " + task[:budget] + "..."
	if got != want {
		t.Errorf("a task of %d bytes was not cut to the budget:\n got  (%d) %q\n want (%d) %q",
			len(task), len(got), got, len(want), want)
	}
}

// TestASpawnTaskOfExactlyTheGuardKeepsItsLastByte restates the lower half so
// the pair reads as one thing, and strengthens it: the existing test asserts
// the message does not END in "..." and that it ends with the task, which a
// budget drift cannot break. Comparing the whole message can.
func TestASpawnTaskOfExactlyTheGuardKeepsItsLastByte(t *testing.T) {
	const guard = 100
	task := strings.Repeat("a", guard-1) + "Z" // a distinct final byte

	got := spawnEventMessage("spawn_started", "claxon", "", task)
	want := "🚀 claxon spawned a sub-agent: " + task
	if got != want {
		t.Errorf("a task of exactly %d bytes was not passed through whole:\n got  (%d) %q\n want (%d) %q",
			guard, len(got), got, len(want), want)
	}
}
