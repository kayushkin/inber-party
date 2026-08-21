package db

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

// A SQLite mirror of the production schema, for tests.
//
// The product runs on PostgreSQL; its schema is postgresMigrations() in db.go. Tests run
// on SQLite, so they need that schema in SQLite's dialect. Writing the SQLite schema out
// by hand is what this file replaces: a hand-written mirror drifts from the product
// silently, and the tests that use it then pin tables the product does not have. Measured
// on 2026-08-21, the hand-written mirror in db_test.go had 4 tables production has never
// had and 4 more with the wrong columns, and TestDB_Migrate passed anyway because it
// checked the mirror against itself.
//
// So the mirror is derived instead: every statement comes from postgresMigrations(), and
// the only thing this file does is rewrite PostgreSQL spellings into SQLite ones. A
// statement it does not recognise is an error, not a silent skip — that is what keeps the
// mirror honest as db.go grows.
//
// Two things the product's schema does that this mirror deliberately does not carry, both
// because SQLite has no equivalent of what db.go writes:
//
//   - the plpgsql updated_at touch triggers on notifications and notification_preferences.
//     No test can observe an automatic updated_at through the mirror.
//   - foreign key enforcement. The mirror creates the REFERENCES clauses, but SQLite
//     ignores them unless the connection asks for them (`_foreign_keys=on` in the DSN), so
//     a test can insert a row PostgreSQL would reject.

// ApplyProductionMigrationsToSQLite creates the production schema in a SQLite database by
// translating postgresMigrations() statement by statement. It is the single mirror: no
// test should write its own CREATE TABLE for a table db.go already declares.
//
// It is safe to call more than once on the same database, for the same reason Migrate is:
// every statement it applies is written IF NOT EXISTS.
func ApplyProductionMigrationsToSQLite(sqlDB *sql.DB) error {
	for index, statement := range postgresMigrations() {
		translated, runsOnSQLite, err := translatePostgresMigrationToSQLite(statement)
		if err != nil {
			return fmt.Errorf("migration %d: %w", index, err)
		}
		if !runsOnSQLite {
			continue
		}
		if _, err := sqlDB.Exec(translated); err != nil {
			// SQLite has no ADD COLUMN IF NOT EXISTS, so a repeated ADD COLUMN — and
			// postgresMigrations() adds agents.gold twice — reports the duplicate rather
			// than ignoring it. Honouring the IF NOT EXISTS the source statement already
			// carries is the whole of this exception; every other error is returned.
			if addsAColumnIfNotExists(statement) && strings.Contains(err.Error(), "duplicate column name") {
				continue
			}
			return fmt.Errorf("migration %d failed: %w\nstatement: %s", index, err, translated)
		}
	}
	return nil
}

// postgresOnlyStatement matches the statements that have no SQLite form at all. Each is
// skipped, and the reason it can be skipped is in the file comment above.
var postgresOnlyStatement = regexp.MustCompile(`(?is)^\s*(CREATE OR REPLACE FUNCTION|DROP TRIGGER|CREATE TRIGGER)\b`)

// translatableStatement matches the statements this file knows how to run on SQLite.
// Anything matching neither this nor postgresOnlyStatement is an error: a new kind of
// statement in db.go must be classified deliberately, not guessed at.
var translatableStatement = regexp.MustCompile(`(?is)^\s*(CREATE TABLE|ALTER TABLE|CREATE INDEX|CREATE UNIQUE INDEX|INSERT INTO)\b`)

// addsColumnIfNotExists matches the ALTER TABLE form whose IF NOT EXISTS SQLite drops.
var addsColumnIfNotExists = regexp.MustCompile(`(?is)^\s*ALTER TABLE\s+\S+\s+ADD COLUMN IF NOT EXISTS\b`)

func addsAColumnIfNotExists(statement string) bool {
	return addsColumnIfNotExists.MatchString(statement)
}

// postgresTypeRewrites are the spelling differences between the two dialects, applied in
// order. \b keeps TIMESTAMP from matching inside CURRENT_TIMESTAMP: _ is a word character,
// so there is no boundary between them.
var postgresTypeRewrites = []struct {
	postgres *regexp.Regexp
	sqlite   string
}{
	{regexp.MustCompile(`\bSERIAL PRIMARY KEY\b`), "INTEGER PRIMARY KEY AUTOINCREMENT"},
	{regexp.MustCompile(`\bVARCHAR\(\d+\)`), "TEXT"},
	{regexp.MustCompile(`\bDECIMAL\(\d+,\s*\d+\)`), "REAL"},
	{regexp.MustCompile(`\bJSONB\b`), "TEXT"},
	{regexp.MustCompile(`\bTEXT\[\]`), "TEXT"},
	{regexp.MustCompile(`\bTIMESTAMP\b`), "DATETIME"},
	{regexp.MustCompile(`\bADD COLUMN IF NOT EXISTS\b`), "ADD COLUMN"},
}

// postgresLeftovers are spellings SQLite does not understand. Finding one after the
// rewrites means postgresMigrations() has grown a construct this file has never seen, and
// applying the statement anyway would either fail with SQLite's own error or, worse,
// create a column with an affinity nobody chose.
var postgresLeftovers = []*regexp.Regexp{
	regexp.MustCompile(`\bSERIAL\b`),
	regexp.MustCompile(`\bVARCHAR\b`),
	regexp.MustCompile(`\bJSONB\b`),
	regexp.MustCompile(`\bDECIMAL\b`),
	regexp.MustCompile(`\bTIMESTAMP\b`),
	regexp.MustCompile(`\[\]`),
	regexp.MustCompile(`::`),
}

// translatePostgresMigrationToSQLite rewrites one production migration into SQLite. The
// second return value is false for a statement that has no SQLite form and is skipped.
func translatePostgresMigrationToSQLite(statement string) (string, bool, error) {
	if postgresOnlyStatement.MatchString(statement) {
		return "", false, nil
	}
	if !translatableStatement.MatchString(statement) {
		return "", false, fmt.Errorf("unclassified statement — add it to translatePostgresMigrationToSQLite rather than letting the mirror guess:\n%s", statement)
	}

	translated := statement
	for _, rewrite := range postgresTypeRewrites {
		translated = rewrite.postgres.ReplaceAllString(translated, rewrite.sqlite)
	}

	// Scan for leftovers with the string literals blanked out. A DEFAULT '[]' is an
	// empty JSON array as a value, not an array type, and reads exactly like one to a
	// regexp that does not know where the quotes are.
	scannable := stringLiteral.ReplaceAllString(translated, "''")
	for _, leftover := range postgresLeftovers {
		if match := leftover.FindString(scannable); match != "" {
			return "", false, fmt.Errorf("no SQLite rewrite for %q in:\n%s", match, translated)
		}
	}
	return translated, true, nil
}

// stringLiteral matches a single-quoted SQL literal, so leftover detection can look at
// the DDL without looking at the values in it.
var stringLiteral = regexp.MustCompile(`'[^']*'`)
