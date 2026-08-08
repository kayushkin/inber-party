package textutil

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// A four-byte rune. Three bytes is not enough to pin this: with a shorter rune a
// test that happens to pick one offset can land on the safe side of the cut and
// pass against the unfixed code, which is how the defect survived everywhere.
const fourByteRune = "\U0001D11E" // U+1D11E MUSICAL SYMBOL G CLEF, 4 bytes

// TestTruncateAtRuneBoundarySlidesARuneAcrossTheCut walks a four-byte rune
// through the cut one byte at a time. Offsets 1, 2 and 3 straddle it; offsets 0
// and 4 land exactly on a rune boundary and are known-negative controls — they
// pass against the unfixed byte-cut too, so a green run means the test detects a
// split rune rather than merely detecting non-ASCII input.
func TestTruncateAtRuneBoundarySlidesARuneAcrossTheCut(t *testing.T) {
	const maxBytes = 20

	for offset := 0; offset <= 4; offset++ {
		lead := strings.Repeat("a", maxBytes-4+offset)
		in := lead + fourByteRune + strings.Repeat("b", 10)

		kind := "straddling"
		if offset == 0 || offset == 4 {
			kind = "boundary-aligned control"
		}
		out := TruncateAtRuneBoundary(in, maxBytes)

		if !utf8.ValidString(out) {
			t.Errorf("offset %d (%s): result is not valid UTF-8: %q (% x)", offset, kind, out, out)
		}
		if len(out) > maxBytes {
			t.Errorf("offset %d (%s): result is %d bytes, over the %d budget: %q",
				offset, kind, len(out), maxBytes, out)
		}
		// Validity and budget alone are not falsifiable together: a function
		// returning "" for every input satisfies both. Pinning how much survives
		// is the assertion with teeth — at most one rune may be dropped.
		if want := maxBytes - 4; len(out) < want {
			t.Errorf("offset %d (%s): lost more than the straddling rune — kept %d bytes, want at least %d: %q",
				offset, kind, len(out), want, out)
		}
		if !strings.HasPrefix(in, out) {
			t.Errorf("offset %d (%s): result is not a prefix of the input: %q", offset, kind, out)
		}
	}
}

func TestTruncateAtRuneBoundaryLeavesShortStringsAlone(t *testing.T) {
	for _, s := range []string{"", "a", "héllo 🎯", strings.Repeat("x", 20)} {
		if out := TruncateAtRuneBoundary(s, 20); out != s {
			t.Errorf("TruncateAtRuneBoundary(%q, 20) = %q, want it returned unchanged", s, out)
		}
	}
}

func TestTruncateAtRuneBoundaryHandlesBudgetsSmallerThanOneRune(t *testing.T) {
	// A budget too small for the leading rune leaves nothing to keep. The result
	// must still be valid UTF-8 rather than a lone continuation byte.
	for _, maxBytes := range []int{-1, 0, 1, 2, 3} {
		out := TruncateAtRuneBoundary(fourByteRune+"tail", maxBytes)
		if out != "" {
			t.Errorf("TruncateAtRuneBoundary(4-byte rune, %d) = %q, want the empty string", maxBytes, out)
		}
	}
	if out := TruncateAtRuneBoundary(fourByteRune+"tail", 4); out != fourByteRune {
		t.Errorf("a budget of exactly one rune should keep it, got %q", out)
	}
}

func TestUpperFirstRuneKeepsMultiByteLeadingRunesIntact(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"claxon", "Claxon"},         // the ASCII path, unchanged by the fix
		{"Claxon", "Claxon"},         // already upper-cased
		{"émile", "Émile"},           // two-byte leading rune
		{"日本語", "日本語"},               // three-byte rune with no upper case
		{"🎯quest", "🎯quest"},         // four-byte rune with no upper case
		{"ñandú vive", "Ñandú vive"}, // the rest of the string is left alone
	}
	for _, c := range cases {
		got := UpperFirstRune(c.in)
		if got != c.want {
			t.Errorf("UpperFirstRune(%q) = %q, want %q", c.in, got, c.want)
		}
		if !utf8.ValidString(got) {
			t.Errorf("UpperFirstRune(%q) returned invalid UTF-8: % x", c.in, got)
		}
	}
}

func TestUpperFirstRuneLeavesAlreadyInvalidInputAlone(t *testing.T) {
	// A stray continuation byte cannot be upper-cased without changing the
	// string's length, and inventing a U+FFFD here would corrupt it further.
	in := "\xa9mile"
	if got := UpperFirstRune(in); got != in {
		t.Errorf("UpperFirstRune(%q) = %q, want it returned unchanged", in, got)
	}
}
