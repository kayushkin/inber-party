package logstack

// Boundary VALUES for the conversation-title cut.
//
// logstack_test.go slides a rune across the cut and pins that the cut is
// rune-safe. It never varies the input length across the guard, so both numbers
// in `if len(title) > 60 { title = TruncateAtRuneBoundary(title, 57) + "..." }`
// are free to drift: the slide test's input is 93 bytes for every offset, which
// is over the guard whatever the guard says, and it asserts a 53-byte prefix
// survived, which stays true whatever the budget says.
//
// The guard and the budget are separate numbers even though they describe one
// intent ("titles are at most 60 bytes"). They are pinned separately here
// because a drift in either one alone is invisible to a test of the other.

import (
	"strings"
	"testing"
)

// titleOf drives the real entry point rather than the cut, so a mutation that
// moves the cut out of this branch is caught too.
func titleOf(content string) string {
	return generateConversationTitle([]ConversationMsg{{Role: "user", Content: content}})
}

// TestATitleOfExactlyTheGuardIsLeftWhole and its sibling below are a straddle
// pair on `len(title) > 60`. 60 must survive whole, 61 must be cut. Either test
// alone pins the guard from one side only, and a guard that drifts the other
// way stays green.
func TestATitleOfExactlyTheGuardIsLeftWhole(t *testing.T) {
	const guard = 60
	in := strings.Repeat("a", guard)

	got := titleOf(in)
	if got != in {
		t.Errorf("a title of exactly %d bytes was not returned whole:\n got  (%d) %q\n want (%d) %q",
			guard, len(got), got, len(in), in)
	}
}

// TestATitleOneByteOverTheGuardIsCutToTheBudget pins the guard's upper side and
// the budget at once, by asserting the whole returned string rather than a
// property of it. 57 + "..." lands back on 60, which is the convention this
// site uses: the ellipsis is *inside* the byte budget.
//
// ⛔ Do not "fix" that to agree with http_client.go and api.go, which put the
// ellipsis outside. textutil.go's doc comment records the disagreement as
// deliberate and unsettled; making the six sites agree is a product change.
func TestATitleOneByteOverTheGuardIsCutToTheBudget(t *testing.T) {
	const (
		guard  = 60
		budget = 57
	)
	in := strings.Repeat("a", guard+1)

	got := titleOf(in)
	want := in[:budget] + "..."
	if got != want {
		t.Errorf("a title of %d bytes was not cut to the budget:\n got  (%d) %q\n want (%d) %q",
			len(in), len(got), got, len(want), want)
	}
	// The ellipsis convention itself, stated as a number so a change to it is a
	// failure here rather than a silent widening of every title on the page.
	if len(got) != guard {
		t.Errorf("a cut title is %d bytes, want %d — this site counts the ellipsis inside the budget",
			len(got), guard)
	}
}
