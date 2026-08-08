package logstack

import (
	"strings"
	"testing"
	"unicode/utf8"
)

const fourByteRune = "\U0001D11E" // U+1D11E MUSICAL SYMBOL G CLEF, 4 bytes

// TestGenerateConversationTitleNeverSplitsARune slides a four-byte rune through
// the cut at byte 57. Offsets 1, 2 and 3 straddle it; offsets 0 and 4 land on a
// rune boundary and are the known-negative controls, so a green run tells us the
// test detects a split rune rather than just detecting non-ASCII input.
func TestGenerateConversationTitleNeverSplitsARune(t *testing.T) {
	const cutAt = 57

	for offset := 0; offset <= 4; offset++ {
		content := strings.Repeat("a", cutAt-4+offset) + fourByteRune + strings.Repeat("b", 40)
		title := generateConversationTitle([]ConversationMsg{{Role: "user", Content: content}})

		kind := "straddling"
		if offset == 0 || offset == 4 {
			kind = "boundary-aligned control"
		}
		label := "generateConversationTitle rune at cut-4+" + string(rune('0'+offset)) + " (" + kind + ")"

		if !utf8.ValidString(title) {
			t.Errorf("%s: title is not valid UTF-8: %q (% x)", label, title, title)
		}
		// Validity alone is not falsifiable — a function returning "" passes it.
		if lead := strings.Repeat("a", cutAt-4); !strings.Contains(title, lead) {
			t.Errorf("%s: lost more than the straddling rune — %d leading bytes did not survive: %q",
				label, cutAt-4, title)
		}
	}
}
