package inber

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// A four-byte rune. Three bytes is not enough: with a two- or three-byte rune a
// test that happens to pick one offset can land on the safe side of the cut and
// pass against the unfixed code, which is how this defect survived everywhere.
const fourByteRune = "\U0001D11E" // U+1D11E MUSICAL SYMBOL G CLEF, 4 bytes

// slideRuneAcrossCut builds inputs that walk a four-byte rune through a cut at
// byte offset cutAt. Offsets 1, 2 and 3 straddle the cut; offsets 0 and 4 land
// exactly on a rune boundary and are the known-negative controls — they must
// pass against the unfixed code too, so a green result on them tells us the test
// detects a split rune rather than merely detecting non-ASCII input.
func slideRuneAcrossCut(cutAt int, tail string) map[int]string {
	inputs := make(map[int]string, 5)
	for offset := 0; offset <= 4; offset++ {
		lead := strings.Repeat("a", cutAt-4+offset)
		inputs[offset] = lead + fourByteRune + tail
	}
	return inputs
}

// assertCutIsRuneSafe checks the two things that together pin the defect.
// Validity alone is not falsifiable: a function returning "" for every input
// produces valid UTF-8 within any budget. Pinning the surviving lead as well
// means a mutation that trims to nothing fails here.
func assertCutIsRuneSafe(t *testing.T, label, out string, minLeadingFiller int) {
	t.Helper()
	if !utf8.ValidString(out) {
		t.Errorf("%s: result is not valid UTF-8: %q (% x)", label, out, out)
	}
	if lead := strings.Repeat("a", minLeadingFiller); !strings.Contains(out, lead) {
		t.Errorf("%s: lost more than the straddling rune — %d leading bytes did not survive: %q",
			label, minLeadingFiller, out)
	}
}

func TestTruncateTextNeverSplitsARune(t *testing.T) {
	const maxLen = 40
	// truncateText cuts at maxLen-3 to make room for its ellipsis.
	for offset, in := range slideRuneAcrossCut(maxLen-3, strings.Repeat("b", 60)) {
		out := truncateText(in, maxLen)
		assertCutIsRuneSafe(t, labelFor("truncateText", offset), out, maxLen-3-4)
	}
}

func TestGenerateQuestNameNeverSplitsARune(t *testing.T) {
	// The trailing word is the quest subject (key terms are picked from the end,
	// and the leading token is rejected because a rune makes it non-alphabetic).
	// The naming pattern is chosen by a hash of the input, not at random, so
	// which inputs reach the cutting branch is fixed — but it is fixed per input,
	// and several patterns embed no subject and so never overrun the 80-byte
	// limit. This trailing length is one that drives all five offsets down the
	// fallback branch; the guard below fails loudly if that stops being true,
	// rather than letting the test pass without exercising the cut at all.
	for offset, in := range slideRuneAcrossCut(57, " "+strings.Repeat("z", 69)) {
		label := labelFor("generateQuestName", offset)
		out := generateQuestName(in, "running")
		if !strings.Contains(out, "aaaa") {
			t.Fatalf("%s: input no longer reaches the truncating branch, so this "+
				"test proves nothing; got the procedural name %q", label, out)
		}
		assertCutIsRuneSafe(t, label, out, 57-4)
	}
}

func labelFor(fn string, offset int) string {
	kind := "straddling"
	if offset == 0 || offset == 4 {
		kind = "boundary-aligned control"
	}
	return fn + " rune at cut-4+" + string(rune('0'+offset)) + " (" + kind + ")"
}
