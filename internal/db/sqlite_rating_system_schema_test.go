package db

import (
	"database/sql"
	"strings"
	"testing"
)

// What these tests hold is the one table the mirror does not create: bounty_ratings,
// applied by ApplyRatingSystemSchemaToSQLite from schema/rating_system.sql. The mirror is
// checked against db.go statement by statement, so it cannot drift on its own. This table
// is a translation nobody re-derives, so its constraints are checked here instead.
//
// Two of them are reachable only from SQL, and that is the finding rather than an
// awkwardness of the test: CreateBountyRating refuses a duplicate with its own Go check
// before the UNIQUE index is ever consulted, and it does not look at the rating value at
// all. So the constraint is the only thing standing between a 9-star rating and the
// database, and until now nothing exercised it. Measured by sabotage on 2026-08-21:
// widening the UNIQUE to three columns and deleting the CHECK both left the whole suite
// green.

// ratingSystemDatabase builds a mirror database with one rateable bounty in it and returns
// the database plus the ids the bounty_ratings rows below have to name.
func ratingSystemDatabase(t *testing.T) (database *DB, bountyID, raterID, ratedID int) {
	t.Helper()
	database = setupSettlementDB(t)
	raterID = makeAgent(t, database, "rater", 0)
	ratedID = makeAgent(t, database, "rated", 0)
	bounty := newBounty(t, database, raterID, 10, "a bounty", "to be rated")
	return database, bounty.ID, raterID, ratedID
}

func insertRating(database *DB, bountyID, raterID, ratedID, rating int) error {
	_, err := database.Exec(
		`INSERT INTO bounty_ratings (bounty_id, rater_id, rated_id, rating) VALUES ($1, $2, $3, $4)`,
		bountyID, raterID, ratedID, rating)
	return err
}

func TestTheRatingSystemSchemaKeepsARatingInsideOneToFive(t *testing.T) {
	database, bountyID, raterID, ratedID := ratingSystemDatabase(t)

	for _, rating := range []int{0, 6, -1, 9} {
		if err := insertRating(database, bountyID, raterID, ratedID, rating); err == nil {
			t.Errorf("a rating of %d was stored — the CHECK (rating >= 1 AND rating <= 5) is not there", rating)
			if _, err := database.Exec(`DELETE FROM bounty_ratings WHERE bounty_id = $1`, bountyID); err != nil {
				t.Fatalf("clear the stored rating: %v", err)
			}
		}
	}
	if err := insertRating(database, bountyID, raterID, ratedID, 3); err != nil {
		t.Fatalf("a rating of 3 is inside the range and must be stored: %v", err)
	}
}

// The UNIQUE is on (bounty_id, rater_id): one rating per bounty per rater, whoever is
// being rated. A three-column UNIQUE including rated_id would let the same rater rate the
// same bounty again by naming a different agent, and reads identically in the schema.
func TestTheRatingSystemSchemaAllowsOneRatingPerBountyPerRater(t *testing.T) {
	database, bountyID, raterID, ratedID := ratingSystemDatabase(t)
	otherRatedID := makeAgent(t, database, "another rated agent", 0)
	otherRaterID := makeAgent(t, database, "another rater", 0)

	if err := insertRating(database, bountyID, raterID, ratedID, 5); err != nil {
		t.Fatalf("store the first rating: %v", err)
	}

	err := insertRating(database, bountyID, raterID, otherRatedID, 5)
	if err == nil {
		t.Error("the same rater rated the same bounty twice — the UNIQUE is not on (bounty_id, rater_id)")
	} else if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
		t.Errorf("refused for the wrong reason: %v", err)
	}

	if err := insertRating(database, bountyID, otherRaterID, ratedID, 4); err != nil {
		t.Errorf("a different rater must be able to rate the same bounty: %v", err)
	}
}

// The three REFERENCES clauses, one per column. They are worth a test of their own
// because the table is applied outside the mirror, so the mirror's own foreign key test
// says nothing about it.
func TestTheRatingSystemSchemaEnforcesItsThreeForeignKeys(t *testing.T) {
	database, bountyID, raterID, ratedID := ratingSystemDatabase(t)
	const noSuchRow = 424242

	for _, argument := range []struct {
		column                     string
		bountyID, raterID, ratedID int
	}{
		{"bounty_id", noSuchRow, raterID, ratedID},
		{"rater_id", bountyID, noSuchRow, ratedID},
		{"rated_id", bountyID, raterID, noSuchRow},
	} {
		err := insertRating(database, argument.bountyID, argument.raterID, argument.ratedID, 5)
		if err == nil {
			t.Errorf("a rating whose %s names no row was stored — its REFERENCES clause is not enforced", argument.column)
			if _, err := database.Exec(`DELETE FROM bounty_ratings WHERE bounty_id = $1`, argument.bountyID); err != nil {
				t.Fatalf("clear the stored rating: %v", err)
			}
			continue
		}
		if !strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			t.Errorf("%s: refused for the wrong reason: %v", argument.column, err)
		}
	}
}

// ApplyRatingSystemSchemaToSQLite refuses a connection with foreign key enforcement off,
// for the same reason the mirror does: it would create three REFERENCES clauses and obey
// none of them, silently.
func TestApplyRatingSystemSchemaRefusesAConnectionThatIgnoresForeignKeys(t *testing.T) {
	unenforcing, err := sql.Open(sqliteDriverName, ":memory:")
	if err != nil {
		t.Fatalf("open a SQLite connection with foreign keys off: %v", err)
	}
	defer unenforcing.Close()

	err = ApplyRatingSystemSchemaToSQLite(unenforcing)
	if err == nil {
		t.Fatal("the rating system schema was applied to a connection that ignores foreign keys")
	}
	if !strings.Contains(err.Error(), "foreign key enforcement") {
		t.Errorf("refused for the wrong reason: %v", err)
	}
}
