package inber

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestGetQuestsDescriptionNeverSplitsARune exercises the whole path the defect
// actually travels: a session prompt arrives as JSON from inber, gets cut to fit
// the quest description, and is served on to the dashboard as JSON again. The
// helper tests pin the cut; this one pins that the cut is the one being used.
//
// encoding/json is why the defect is silent — it substitutes U+FFFD for invalid
// bytes rather than returning an error, so a corrupted description reaches the
// browser with nothing logged anywhere along the way.
func TestGetQuestsDescriptionNeverSplitsARune(t *testing.T) {
	const cutAt = 200 // http_client.go cuts InputText at 200 bytes

	for offset := 0; offset <= 4; offset++ {
		label := labelFor("HTTPClient.GetQuests description", offset)
		inputText := strings.Repeat("a", cutAt-4+offset) + fourByteRune + strings.Repeat("b", 40)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/sessions" {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode([]map[string]any{{
				"key":        "session-1",
				"agent":      "claxon",
				"status":     "running",
				"input_text": inputText,
			}}); err != nil {
				t.Errorf("%s: encoding the fake session failed: %v", label, err)
			}
		}))

		quests, err := NewHTTPClient(server.URL).GetQuests(10)
		server.Close()
		if err != nil {
			t.Fatalf("%s: GetQuests failed: %v", label, err)
		}
		if len(quests) != 1 {
			t.Fatalf("%s: expected 1 quest, got %d", label, len(quests))
		}

		assertCutIsRuneSafe(t, label, quests[0].Description, cutAt-4)
	}
}
