package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kayushkin/inber-party/internal/db"
)

// createRating (api.go:3489) decides its status code by comparing the error text the
// database layer returned:
//
//	err.Error() == "bounty already rated"   → 409 Conflict
//	else                                    → 500 Internal Server Error
//
// That is a join on a name across a package boundary, and it is the only thing standing
// between a client-correctable conflict and a reported server fault. The db-side wording
// is pinned in internal/db/bounty_rating_guard_test.go; this test pins the wiring, so a
// change on either side of the string shows up as a red rather than as a silent 500 in
// production.

// completedBountyForRating seeds an agent pair and a completed bounty straight through
// SQL, so the fixture cannot be satisfied by the handler under test.
func completedBountyForRating(t *testing.T, database *db.DB) (creatorID, claimerID, bountyID int) {
	t.Helper()
	insertAgent := func(name string) int {
		var id int
		err := database.QueryRow(
			`INSERT INTO agents (name, title, class, avatar_emoji) VALUES ($1, '', '', '') RETURNING id`,
			name).Scan(&id)
		if err != nil {
			t.Fatalf("insert agent %s: %v", name, err)
		}
		return id
	}
	creatorID = insertAgent("creator")
	claimerID = insertAgent("claimer")

	err := database.QueryRow(`
		INSERT INTO bounties (title, description, payout_amount, status, creator_id, claimer_id, tier)
		VALUES ('tidy the readme', 'documentation chore', 25, 'completed', $1, $2, 'bronze')
		RETURNING id`, creatorID, claimerID).Scan(&bountyID)
	if err != nil {
		t.Fatalf("insert bounty: %v", err)
	}
	return creatorID, claimerID, bountyID
}

func postRating(t *testing.T, server *Server, rating db.BountyRating) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(rating)
	if err != nil {
		t.Fatalf("marshal rating: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/ratings", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	server.handleRatings(recorder, req)
	return recorder
}

func TestCreateRatingAnswers409NotInternalErrorWhenTheBountyIsAlreadyRated(t *testing.T) {
	server, cleanup := createTestServer(t)
	defer cleanup()

	creatorID, claimerID, bountyID := completedBountyForRating(t, server.DB)
	rating := db.BountyRating{
		BountyID: bountyID,
		RaterID:  creatorID,
		RatedID:  claimerID,
		Rating:   5,
		Comment:  "ahead of schedule",
	}

	// The known-positive control: the first rating has to get through, or the conflict
	// below would be reported by a handler that refuses everything.
	if got := postRating(t, server, rating).Code; got != http.StatusCreated {
		t.Fatalf("first POST /api/ratings = %d, want %d", got, http.StatusCreated)
	}

	second := rating
	second.Rating = 1
	second.Comment = "changed my mind"
	recorder := postRating(t, server, second)

	if recorder.Code != http.StatusConflict {
		t.Errorf("second POST /api/ratings = %d, want %d — a duplicate rating is the client's to correct, not a server fault",
			recorder.Code, http.StatusConflict)
	}
	if body := strings.TrimSpace(recorder.Body.String()); body != "bounty already rated" {
		t.Errorf("conflict body = %q, want %q — the reason must reach the client", body, "bounty already rated")
	}

	// And the refused rating must not have landed.
	var stored int
	if err := server.DB.QueryRow(`SELECT rating FROM bounty_ratings WHERE bounty_id = $1`, bountyID).Scan(&stored); err != nil {
		t.Fatalf("read stored rating: %v", err)
	}
	if stored != 5 {
		t.Errorf("stored rating = %d, want 5 — the first rating must survive the refused second one", stored)
	}
}
