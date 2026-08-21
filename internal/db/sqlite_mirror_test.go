package db

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// What these tests hold is the mirror itself: that ApplyProductionMigrationsToSQLite
// really does build the schema db.go declares, and that it says so loudly when it cannot.
// The check they replace asserted a hand-written list of ten table names against a
// hand-written schema, so it passed by agreeing with itself — 4 of its 10 tables did not
// exist in production at all.

func TestTheSQLiteMirrorCreatesEveryTableTheProductionSchemaDeclares(t *testing.T) {
	db, cleanup := setupInMemoryDB(t)
	defer cleanup()

	if err := ApplyProductionMigrationsToSQLite(db.DB); err != nil {
		t.Fatalf("apply production migrations: %v", err)
	}

	declared := declaredProductionSchema(t)
	for _, table := range sortedTableNames(declared) {
		var count int
		err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&count)
		if err != nil {
			t.Errorf("look for table %s: %v", table, err)
			continue
		}
		if count != 1 {
			t.Errorf("db.go declares table %s and the mirror did not create it", table)
		}
	}
}

func TestTheSQLiteMirrorCreatesEveryColumnTheProductionSchemaDeclares(t *testing.T) {
	db, cleanup := setupInMemoryDB(t)
	defer cleanup()

	if err := ApplyProductionMigrationsToSQLite(db.DB); err != nil {
		t.Fatalf("apply production migrations: %v", err)
	}

	declared := declaredProductionSchema(t)
	mirrored := make(map[string][]string, len(declared))
	for _, table := range sortedTableNames(declared) {
		rows, err := db.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, table))
		if err != nil {
			t.Fatalf("read columns of %s: %v", table, err)
		}
		for rows.Next() {
			var (
				index            int
				name, columnType string
				notNull          int
				defaultValue     any
				partOfPrimaryKey int
			)
			if err := rows.Scan(&index, &name, &columnType, &notNull, &defaultValue, &partOfPrimaryKey); err != nil {
				rows.Close()
				t.Fatalf("scan columns of %s: %v", table, err)
			}
			mirrored[table] = append(mirrored[table], name)
		}
		rows.Close()
	}

	for _, missing := range columnsMissingFromMirror(declared, mirrored) {
		t.Errorf("db.go declares %s and the mirror does not have it", missing)
	}
}

// TestColumnComparisonNoticesAMissingColumn is the control for the check above: a
// comparison that can only ever return nothing is indistinguishable from a schema with
// nothing wrong with it.
func TestColumnComparisonNoticesAMissingColumn(t *testing.T) {
	declared := map[string][]string{"bounties": {"id", "payout_amount", "claimer_id"}}
	mirrored := map[string][]string{"bounties": {"id", "claimer_id"}}

	missing := columnsMissingFromMirror(declared, mirrored)
	if len(missing) != 1 || missing[0] != "bounties.payout_amount" {
		t.Fatalf("expected the comparison to report bounties.payout_amount, got %v", missing)
	}

	if leftovers := columnsMissingFromMirror(declared, map[string][]string{"bounties": declared["bounties"]}); len(leftovers) != 0 {
		t.Fatalf("expected no report when the mirror matches, got %v", leftovers)
	}
}

func TestTheMirrorCanBeAppliedTwice(t *testing.T) {
	db, cleanup := setupInMemoryDB(t)
	defer cleanup()

	for attempt := 1; attempt <= 3; attempt++ {
		if err := ApplyProductionMigrationsToSQLite(db.DB); err != nil {
			t.Fatalf("apply production migrations, attempt %d: %v", attempt, err)
		}
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM agents`).Scan(&count); err != nil {
		t.Fatalf("query agents after three applications: %v", err)
	}
}

func TestTranslationRefusesAStatementItCannotClassify(t *testing.T) {
	_, runsOnSQLite, err := translatePostgresMigrationToSQLite(`GRANT ALL ON agents TO party_app`)
	if err == nil {
		t.Fatalf("expected an unclassified statement to be refused, got runsOnSQLite=%v", runsOnSQLite)
	}
	if !strings.Contains(err.Error(), "unclassified statement") {
		t.Errorf("expected the error to name the problem, got: %v", err)
	}
}

func TestTranslationRefusesAPostgresSpellingItCannotRewrite(t *testing.T) {
	_, _, err := translatePostgresMigrationToSQLite("CREATE TABLE IF NOT EXISTS scores (\n\tid SERIAL PRIMARY KEY,\n\tvalues INTEGER[]\n)")
	if err == nil {
		t.Fatal("expected an array column to be refused")
	}
	if !strings.Contains(err.Error(), "no SQLite rewrite") {
		t.Errorf("expected the error to name the problem, got: %v", err)
	}
}

func TestTranslationSkipsThePostgresOnlyTriggerStatements(t *testing.T) {
	for _, statement := range []string{
		"CREATE OR REPLACE FUNCTION update_updated_at_column()\nRETURNS TRIGGER AS $$ BEGIN END; $$ language 'plpgsql'",
		"DROP TRIGGER IF EXISTS update_notifications_updated_at ON notifications",
		"CREATE TRIGGER update_notifications_updated_at BEFORE UPDATE ON notifications\n\tFOR EACH ROW EXECUTE FUNCTION update_updated_at_column()",
	} {
		translated, runsOnSQLite, err := translatePostgresMigrationToSQLite(statement)
		if err != nil {
			t.Errorf("expected a skip, got an error for %q: %v", firstLine(statement), err)
		}
		if runsOnSQLite {
			t.Errorf("expected %q to be skipped, got: %s", firstLine(statement), translated)
		}
	}
}

func TestTranslationRewritesTheProductionSpellings(t *testing.T) {
	translated, runsOnSQLite, err := translatePostgresMigrationToSQLite(
		"CREATE TABLE IF NOT EXISTS sample (\n" +
			"\tid SERIAL PRIMARY KEY,\n" +
			"\tname VARCHAR(255) NOT NULL,\n" +
			"\trate DECIMAL(5,4) DEFAULT 1.0,\n" +
			"\tpayload JSONB DEFAULT '{}',\n" +
			"\tkeywords TEXT[],\n" +
			"\tseen_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP\n)")
	if err != nil || !runsOnSQLite {
		t.Fatalf("expected a translation, got runsOnSQLite=%v err=%v", runsOnSQLite, err)
	}

	for _, want := range []string{
		"id INTEGER PRIMARY KEY AUTOINCREMENT",
		"name TEXT NOT NULL",
		"rate REAL DEFAULT 1.0",
		"payload TEXT DEFAULT '{}'",
		"keywords TEXT",
		"seen_at DATETIME DEFAULT CURRENT_TIMESTAMP",
	} {
		if !strings.Contains(translated, want) {
			t.Errorf("expected the translation to contain %q, got:\n%s", want, translated)
		}
	}
}

// -----------------------------------------------------------------------------
// Reading the production schema back out of db.go
// -----------------------------------------------------------------------------

var (
	createTableHeader = regexp.MustCompile(`(?is)^\s*CREATE TABLE IF NOT EXISTS\s+(\w+)\s*\((.*)\)\s*$`)
	addColumnHeader   = regexp.MustCompile(`(?is)^\s*ALTER TABLE\s+(\w+)\s+ADD COLUMN IF NOT EXISTS\s+(\w+)\b`)
	tableConstraint   = regexp.MustCompile(`(?i)^(UNIQUE|PRIMARY KEY|FOREIGN KEY|CHECK|CONSTRAINT)\b`)
)

// declaredProductionSchema reads postgresMigrations() and returns the table and column
// names it declares. It parses db.go's own statements rather than restating them, so a
// table added there is checked here without anyone remembering to add it.
func declaredProductionSchema(t *testing.T) map[string][]string {
	t.Helper()

	schema := map[string][]string{}
	declared := map[string]bool{}
	appendColumn := func(table, column string) {
		if declared[table+"."+column] {
			// db.go adds agents.gold twice; the schema declares it once.
			return
		}
		declared[table+"."+column] = true
		schema[table] = append(schema[table], column)
	}

	for _, statement := range postgresMigrations() {
		if match := createTableHeader.FindStringSubmatch(statement); match != nil {
			table, body := match[1], match[2]
			for _, line := range strings.Split(body, "\n") {
				line = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(line), ","))
				if line == "" || tableConstraint.MatchString(line) {
					continue
				}
				appendColumn(table, strings.Fields(line)[0])
			}
			continue
		}
		if match := addColumnHeader.FindStringSubmatch(statement); match != nil {
			appendColumn(match[1], match[2])
		}
	}

	if len(schema) == 0 {
		t.Fatal("parsed no tables out of postgresMigrations() — the parser, not the schema, is what broke")
	}
	return schema
}

// columnsMissingFromMirror returns "table.column" for every declared column the mirror
// does not have, so the report names the gap rather than a count.
func columnsMissingFromMirror(declared, mirrored map[string][]string) []string {
	var missing []string
	for _, table := range sortedTableNames(declared) {
		present := map[string]bool{}
		for _, column := range mirrored[table] {
			present[column] = true
		}
		for _, column := range declared[table] {
			if !present[column] {
				missing = append(missing, table+"."+column)
			}
		}
	}
	return missing
}

func sortedTableNames(schema map[string][]string) []string {
	names := make([]string, 0, len(schema))
	for name := range schema {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func firstLine(statement string) string {
	return strings.SplitN(strings.TrimSpace(statement), "\n", 2)[0]
}
