package db

import (
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// CreateBountyRating refuses a second rating of the same bounty by the same rater with
// a hand-rolled SELECT-then-error check (ratings.go:64-72), and the UNIQUE index in
// schema/rating_system.sql refuses the same pair independently. Nothing executed either
// one. That matters more than an ordinary coverage hole because internal/api/api.go
// compares the returned error by STRING EQUALITY:
//
//	err.Error() == "bounty already rated"   → 409 Conflict
//
// so the Go check's exact wording is load-bearing at the HTTP edge. Delete the check and
// the database's own constraint message arrives instead, the comparison misses, and the
// endpoint reports a client-correctable conflict as a 500. The assertions below therefore
// name the wording rather than settling for "an error came back" — "an error came back"
// cannot tell the Go guard from the UNIQUE index, and the two produce different HTTP
// status codes.

// ⚠️ These tests print "Failed to update reputation from rating: ... no such function:
// GREATEST" on the way past. That is the SQLite mirror, not a defect: GREATEST/LEAST are
// real Postgres functions, and CreateBountyRating deliberately logs and continues when the
// reputation update fails (ratings.go, "Log error but don't fail the rating creation").
// The rating itself is unaffected, which is what these tests assert.

// ratedBounty drives a bounty all the way to paid through the real settlement functions,
// so the rating below is applied to a bounty in the only state CreateBountyRating accepts.
func ratedBounty(t *testing.T, database *DB, creatorID, claimerID int) int {
	t.Helper()
	b := claimAndSubmit(t, database, creatorID, claimerID, 25)
	if err := database.VerifyBounty(b.ID, true, "good work"); err != nil {
		t.Fatalf("VerifyBounty: %v", err)
	}
	return b.ID
}

func ratingRowCount(t *testing.T, database *DB, bountyID int) int {
	t.Helper()
	var n int
	if err := database.QueryRow(`SELECT COUNT(*) FROM bounty_ratings WHERE bounty_id = $1`, bountyID).Scan(&n); err != nil {
		t.Fatalf("count bounty_ratings for bounty %d: %v", bountyID, err)
	}
	return n
}

// The known-positive control. Without it a fixture that cannot store a rating at all
// would satisfy the duplicate test below for the wrong reason.
func TestCreateBountyRatingStoresTheFirstRatingInItsOwnColumns(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "claimer", 0)
	bountyID := ratedBounty(t, database, creatorID, claimerID)

	rating := &BountyRating{
		BountyID:   bountyID,
		RaterID:    creatorID,
		RatedID:    claimerID,
		Rating:     5,
		Comment:    "ahead of schedule",
		Categories: map[string]int{"quality": 5, "timeliness": 4},
	}
	if err := database.CreateBountyRating(rating); err != nil {
		t.Fatalf("CreateBountyRating: %v", err)
	}

	if rating.ID == 0 {
		t.Error("the id assigned by the database was not written back onto the rating")
	}
	if got := ratingRowCount(t, database, bountyID); got != 1 {
		t.Fatalf("bounty_ratings rows = %d, want 1", got)
	}

	// Read the columns back rather than trusting the struct the writer just filled in.
	var raterID, ratedID, score int
	var comment, categories string
	err := database.QueryRow(
		`SELECT rater_id, rated_id, rating, comment, categories FROM bounty_ratings WHERE bounty_id = $1`,
		bountyID).Scan(&raterID, &ratedID, &score, &comment, &categories)
	if err != nil {
		t.Fatalf("read stored rating: %v", err)
	}
	if raterID != creatorID {
		t.Errorf("rater_id = %d, want %d (the creator)", raterID, creatorID)
	}
	if ratedID != claimerID {
		t.Errorf("rated_id = %d, want %d (the claimer)", ratedID, claimerID)
	}
	if score != 5 {
		t.Errorf("rating = %d, want 5", score)
	}
	if comment != "ahead of schedule" {
		t.Errorf("comment = %q, want %q", comment, "ahead of schedule")
	}
	if !strings.Contains(categories, `"timeliness":4`) {
		t.Errorf("categories = %q, want the per-category breakdown stored as JSON", categories)
	}
}

func TestCreateBountyRatingRefusesASecondRatingOfTheSameBountyByItsOwnWording(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "claimer", 0)
	bountyID := ratedBounty(t, database, creatorID, claimerID)

	first := &BountyRating{BountyID: bountyID, RaterID: creatorID, RatedID: claimerID, Rating: 5, Comment: "first"}
	if err := database.CreateBountyRating(first); err != nil {
		t.Fatalf("first CreateBountyRating: %v", err)
	}

	second := &BountyRating{BountyID: bountyID, RaterID: creatorID, RatedID: claimerID, Rating: 1, Comment: "changed my mind"}
	err := database.CreateBountyRating(second)
	if err == nil {
		t.Fatal("expected CreateBountyRating to refuse a second rating of the same bounty")
	}
	// The wording, not merely the presence of an error: api.go:3513 switches on this exact
	// string to answer 409 rather than 500.
	if err.Error() != "bounty already rated" {
		t.Errorf("error = %q, want %q — api.go compares this string to decide 409 vs 500", err.Error(), "bounty already rated")
	}
	if got := ratingRowCount(t, database, bountyID); got != 1 {
		t.Errorf("bounty_ratings rows = %d, want 1 — a refused rating must not be stored", got)
	}

	// The refusal must not have overwritten the first rating either.
	var score int
	var comment string
	if err := database.QueryRow(
		`SELECT rating, comment FROM bounty_ratings WHERE bounty_id = $1`, bountyID).Scan(&score, &comment); err != nil {
		t.Fatalf("read stored rating: %v", err)
	}
	if score != 5 || comment != "first" {
		t.Errorf("stored rating = (%d, %q), want (5, %q) — the first rating must survive intact", score, comment, "first")
	}
}

// A different rater on the same bounty is a different pair, so neither the Go check nor
// the two-column UNIQUE may refuse it. This is what separates the real constraint from
// the three-column one in db_test.go's `ratings` mirror and from a hypothetical
// one-rating-per-bounty rule.
func TestCreateBountyRatingRefusesARaterWhoIsNotTheBountyCreator(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "claimer", 0)
	strangerID := makeAgent(t, database, "stranger", 0)
	bountyID := ratedBounty(t, database, creatorID, claimerID)

	err := database.CreateBountyRating(&BountyRating{
		BountyID: bountyID, RaterID: strangerID, RatedID: claimerID, Rating: 5,
	})
	if err == nil {
		t.Fatal("expected CreateBountyRating to refuse a rater who did not create the bounty")
	}
	if err.Error() != "only bounty creator can rate the work" {
		t.Errorf("error = %q, want %q — api.go compares this string to decide 400 vs 500", err.Error(), "only bounty creator can rate the work")
	}
	if got := ratingRowCount(t, database, bountyID); got != 0 {
		t.Errorf("bounty_ratings rows = %d, want 0", got)
	}
}
