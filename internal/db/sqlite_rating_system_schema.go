package db

import (
	"database/sql"
	"fmt"
)

// bounty_ratings is the one table the tests need that the SQLite mirror cannot give them,
// because the mirror translates postgresMigrations() and postgresMigrations() does not
// declare it. It is declared in schema/rating_system.sql instead, a file Migrate() never
// reads, and internal/db/ratings.go queries it — so either deployment applies schema/*.sql
// by some path outside Migrate(), or the table is absent in production and the ratings code
// has never run there.
//
// ⚠️ That question is open and belongs to whoever owns the deployment, not to this file.
// Until it is answered the honest arrangement is the one below: the tables db.go declares
// come from the mirror, and bounty_ratings is applied separately and visibly, so nothing
// here can be mistaken for a claim that Migrate() creates it. When the question is
// answered, the fix is to move the table into postgresMigrations() and delete this file —
// not to grow it.
//
// notifications.sql is the same shape and no test needs it yet.

// ratingSystemSchemaStatements is schema/rating_system.sql's bounty_ratings table in
// SQLite's dialect: SERIAL PRIMARY KEY to INTEGER PRIMARY KEY AUTOINCREMENT, JSONB to TEXT,
// TIMESTAMP to DATETIME, and the three FOREIGN KEY clauses written inline. The indexes,
// sample rows, trigger and views in that file are PostgreSQL-only or test data, and are not
// carried. The UNIQUE is on two columns, (bounty_id, rater_id) — one rating per bounty per
// rater — which is the constraint db.CreateBountyRating's hand-written guard checks.
var ratingSystemSchemaStatements = []string{
	`CREATE TABLE IF NOT EXISTS bounty_ratings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bounty_id INTEGER NOT NULL REFERENCES bounties(id) ON DELETE CASCADE,
		rater_id INTEGER NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
		rated_id INTEGER NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
		rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
		comment TEXT DEFAULT '',
		categories TEXT DEFAULT '{}',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(bounty_id, rater_id)
	)`,
}

// ApplyRatingSystemSchemaToSQLite creates schema/rating_system.sql's bounty_ratings table
// in a SQLite database that already carries the production mirror. It is separate from
// ApplyProductionMigrationsToSQLite on purpose: that function's contract is "everything
// db.go declares", and this table is the exception to it.
//
// It refuses a connection that is not enforcing foreign keys, for the same reason the
// mirror does — the table is three REFERENCES clauses deep and SQLite ignores all of them
// by default. Call it on a database opened through OpenSQLiteMirrorOfProductionSchema; the
// bounties and agents rows it points at have to exist first, so the mirror has to be
// applied before this.
func ApplyRatingSystemSchemaToSQLite(sqlDB *sql.DB) error {
	if err := confirmForeignKeyEnforcementIsOn(sqlDB); err != nil {
		return err
	}
	for index, statement := range ratingSystemSchemaStatements {
		if _, err := sqlDB.Exec(statement); err != nil {
			return fmt.Errorf("rating system statement %d failed: %w\nstatement: %s", index, err, statement)
		}
	}
	return nil
}
