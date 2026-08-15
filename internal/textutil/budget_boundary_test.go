package textutil

// Boundary VALUES, not the rune-boundary mechanism.
//
// textutil_test.go pins what the cut does — it never splits a rune. This file
// pins the one number that decides whether there is a cut at all: the guard
// `maxBytes <= 0`. The two questions are independent, and the existing sweep
// answers only the first.

import "testing"

// TestABudgetOfOneKeepsASingleByteRune separates `maxBytes <= 0` from
// `maxBytes <= 1`.
//
// TestTruncateAtRuneBoundaryHandlesBudgetsSmallerThanOneRune already sweeps
// {-1, 0, 1, 2, 3}, so it looks like this boundary is held. It is not: that
// sweep runs every budget over a *four-byte* leading rune, and for a four-byte
// rune a budget of 1 correctly yields "" under both spellings of the guard. The
// sweep covers the value without separating it. Only a one-byte leading rune
// tells the two apart — at budget 1 the correct code keeps "h" and the widened
// guard returns "".
func TestABudgetOfOneKeepsASingleByteRune(t *testing.T) {
	if out := TruncateAtRuneBoundary("hello", 1); out != "h" {
		t.Errorf(`TruncateAtRuneBoundary("hello", 1) = %q, want "h"`, out)
	}
}

// TestTheBudgetIsAnUpperBoundNotAFixedLength pins the other end of the same
// guard: a budget one byte under the input's length must cut, and a budget
// exactly equal to it must not. `len(s) <= maxBytes` is the second comparison
// in the function and no existing test straddles it with a one-byte rune
// either — the sweep above stops at 3 and the slide test only ever passes 20.
func TestTheBudgetIsAnUpperBoundNotAFixedLength(t *testing.T) {
	const s = "hello" // 5 bytes, all single-byte runes

	if out := TruncateAtRuneBoundary(s, len(s)); out != s {
		t.Errorf("a budget of exactly len(s)=%d truncated: got %q, want %q", len(s), out, s)
	}
	if out := TruncateAtRuneBoundary(s, len(s)-1); out != "hell" {
		t.Errorf(`a budget of len(s)-1 = %d: got %q, want "hell"`, len(s)-1, out)
	}
	if out := TruncateAtRuneBoundary(s, len(s)+1); out != s {
		t.Errorf("a budget over len(s) truncated: got %q, want %q", out, s)
	}
}
