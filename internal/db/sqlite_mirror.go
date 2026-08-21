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
// One thing the product's schema does that this mirror deliberately does not carry,
// because SQLite has no equivalent of what db.go writes:
//
//   - the plpgsql updated_at touch triggers on notifications and notification_preferences.
//     No test can observe an automatic updated_at through the mirror.
//
// Foreign keys used to be a second entry on that list. They are not any more.
// OpenSQLiteMirrorOfProductionSchema turns enforcement on in the DSN and
// ApplyProductionMigrationsToSQLite refuses a connection that has it off, so a REFERENCES
// clause in db.go is a REFERENCES clause the tests obey. That matters because the failure
// it replaces was silent in both directions: a test could insert a bounty whose creator_id
// named no agent, and a test could delete an agent and assert its rows went with it while
// ON DELETE CASCADE did nothing at all.

// sqliteDriverName is the driver mattn/go-sqlite3 registers with database/sql. The
// package that imports it for its side effect is the one that owns the import; this file
// only names the driver.
const sqliteDriverName = "sqlite3"

// foreignKeyEnforcementParameter is go-sqlite3's DSN parameter for PRAGMA foreign_keys.
// It has to be in the DSN rather than executed as a pragma afterwards: enforcement is a
// property of a connection, database/sql hands out connections from a pool, and a pragma
// run through the pool reaches whichever single connection served it. A DSN parameter is
// applied by the driver every time it opens one, so it holds for the whole pool.
const foreignKeyEnforcementParameter = "_foreign_keys=on"

// OpenSQLiteMirrorOfProductionSchema opens a SQLite database with foreign key enforcement
// on and creates the production schema in it. It is how a test gets a mirror database:
// sql.Open("sqlite3", path) by hand returns a connection that ignores every REFERENCES
// clause in the schema it is about to be given, and ignores them without saying so.
//
// dataSourceName is a plain SQLite path — a file, or ":memory:". Any parameters already on
// it are kept.
func OpenSQLiteMirrorOfProductionSchema(dataSourceName string) (*sql.DB, error) {
	sqlDB, err := sql.Open(sqliteDriverName, withForeignKeyEnforcement(dataSourceName))
	if err != nil {
		return nil, fmt.Errorf("open SQLite mirror %q: %w", dataSourceName, err)
	}
	if err := ApplyProductionMigrationsToSQLite(sqlDB); err != nil {
		sqlDB.Close()
		return nil, err
	}
	return sqlDB, nil
}

// withForeignKeyEnforcement adds the foreign key parameter to a SQLite data source name,
// keeping whatever parameters it already carries.
func withForeignKeyEnforcement(dataSourceName string) string {
	if strings.Contains(dataSourceName, foreignKeyEnforcementParameter) {
		return dataSourceName
	}
	separator := "?"
	if strings.Contains(dataSourceName, "?") {
		separator = "&"
	}
	return dataSourceName + separator + foreignKeyEnforcementParameter
}

// confirmForeignKeyEnforcementIsOn reads PRAGMA foreign_keys back off the connection. The
// DSN parameter is the thing that turns enforcement on, and a misspelled or unsupported
// parameter is accepted silently by go-sqlite3 — so "the DSN said so" and "the connection
// does it" are two different claims, and this checks the second one.
func confirmForeignKeyEnforcementIsOn(sqlDB *sql.DB) error {
	var enforced int
	if err := sqlDB.QueryRow(`PRAGMA foreign_keys`).Scan(&enforced); err != nil {
		return fmt.Errorf("read PRAGMA foreign_keys: %w", err)
	}
	if enforced != 1 {
		return fmt.Errorf(
			"the SQLite mirror needs foreign key enforcement and this connection has it off "+
				"(PRAGMA foreign_keys = %d) — open it with %s in the data source name, or through "+
				"OpenSQLiteMirrorOfProductionSchema, which does that for you. Without it SQLite "+
				"creates every REFERENCES clause in the schema and obeys none of them",
			enforced, foreignKeyEnforcementParameter)
	}
	return nil
}

// ApplyProductionMigrationsToSQLite creates the production schema in a SQLite database by
// translating postgresMigrations() statement by statement. It is the single mirror: no
// test should write its own CREATE TABLE for a table db.go already declares.
//
// It refuses a connection that is not enforcing foreign keys, because the schema it
// applies is 25 REFERENCES clauses deep and a SQLite connection ignores all of them by
// default. A mirror that carries the clauses and does not obey them is the drift this file
// exists to stop, one level down.
//
// It is safe to call more than once on the same database, for the same reason Migrate is:
// every statement it applies is written IF NOT EXISTS.
func ApplyProductionMigrationsToSQLite(sqlDB *sql.DB) error {
	if err := confirmForeignKeyEnforcementIsOn(sqlDB); err != nil {
		return err
	}
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
