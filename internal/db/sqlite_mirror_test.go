package db

import (
	"database/sql"
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

// The four tests below are the controls for foreign key enforcement. Without them, "the
// suite is green under enforcement" and "the DSN parameter never took" print the same
// result — which is the state this mirror was in until 2026-08-21, carrying all 25 of
// db.go's REFERENCES clauses and obeying none of them.

func TestAMirrorDatabaseEnforcesForeignKeys(t *testing.T) {
	db, cleanup := setupInMemoryDB(t)
	defer cleanup()

	var enforced int
	if err := db.QueryRow(`PRAGMA foreign_keys`).Scan(&enforced); err != nil {
		t.Fatalf("read PRAGMA foreign_keys: %v", err)
	}
	if enforced != 1 {
		t.Fatalf("a mirror database must enforce foreign keys, PRAGMA foreign_keys = %d", enforced)
	}
}

// TestTheMirrorRefusesAChildRowWhoseParentDoesNotExist is the positive control: it drives
// the enforcement itself rather than the pragma that switches it on, because a pragma
// reading 1 on one pooled connection is not proof that the connection running the INSERT
// has it too.
func TestTheMirrorRefusesAChildRowWhoseParentDoesNotExist(t *testing.T) {
	db, cleanup := setupInMemoryDB(t)
	defer cleanup()

	// party_members.agent_id REFERENCES agents(id), and agents is empty.
	_, err := db.Exec(`INSERT INTO party_members (party_id, agent_id, role) VALUES (NULL, ?, 'member')`, 999999)
	if err == nil {
		t.Fatal("expected the mirror to refuse a party_members row naming an agent that does not exist")
	}
	if !strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
		t.Errorf("expected a foreign key error, got: %v", err)
	}
}

// TestTheMirrorCascadesADeleteTheProductionSchemaDeclaresCascading covers the other half
// of what enforcement buys. ON DELETE CASCADE is inert on an unenforcing connection, so a
// test that deleted an agent and asserted its rows went with it was asserting against a
// database that never cascaded and would have passed either way.
func TestTheMirrorCascadesADeleteTheProductionSchemaDeclaresCascading(t *testing.T) {
	db, cleanup := setupInMemoryDB(t)
	defer cleanup()

	result, err := db.Exec(
		`INSERT INTO agents (name, title, class, avatar_emoji) VALUES ('cascade-subject', 'Tester', 'rogue', 'x')`)
	if err != nil {
		t.Fatalf("insert agent: %v", err)
	}
	agentID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read the inserted agent's id: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO skills (agent_id, skill_name) VALUES (?, 'lockpicking')`, agentID); err != nil {
		t.Fatalf("insert skill: %v", err)
	}

	if _, err := db.Exec(`DELETE FROM agents WHERE id = ?`, agentID); err != nil {
		t.Fatalf("delete agent: %v", err)
	}

	var remaining int
	if err := db.QueryRow(`SELECT COUNT(*) FROM skills WHERE agent_id = ?`, agentID).Scan(&remaining); err != nil {
		t.Fatalf("count skills after deleting their agent: %v", err)
	}
	if remaining != 0 {
		t.Errorf("skills.agent_id is declared ON DELETE CASCADE and %d row(s) survived the delete", remaining)
	}
}

// TestTheMirrorRefusesAConnectionThatIsNotEnforcingForeignKeys proves the guard in
// ApplyProductionMigrationsToSQLite is live rather than dead code. It is the only test
// that opens SQLite by hand, and it does so to produce exactly the connection every other
// caller used to get by accident.
func TestTheMirrorRefusesAConnectionThatIsNotEnforcingForeignKeys(t *testing.T) {
	unenforcing, err := sql.Open(sqliteDriverName, ":memory:")
	if err != nil {
		t.Fatalf("open an unenforcing SQLite database: %v", err)
	}
	defer unenforcing.Close()

	err = ApplyProductionMigrationsToSQLite(unenforcing)
	if err == nil {
		t.Fatal("expected the mirror to refuse a connection with foreign key enforcement off")
	}
	if !strings.Contains(err.Error(), "foreign key enforcement") {
		t.Errorf("expected the error to name the problem, got: %v", err)
	}

	var tables int
	if err := unenforcing.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table'`).Scan(&tables); err != nil {
		t.Fatalf("count tables after the refusal: %v", err)
	}
	if tables != 0 {
		t.Errorf("the mirror refused the connection and still created %d table(s) in it", tables)
	}
}

func TestForeignKeyEnforcementIsAddedToADataSourceNameWithoutLosingItsParameters(t *testing.T) {
	for _, testCase := range []struct {
		dataSourceName string
		want           string
	}{
		{":memory:", ":memory:?_foreign_keys=on"},
		{"/tmp/test.db", "/tmp/test.db?_foreign_keys=on"},
		{"/tmp/test.db?_journal_mode=WAL", "/tmp/test.db?_journal_mode=WAL&_foreign_keys=on"},
		{"/tmp/test.db?_foreign_keys=on", "/tmp/test.db?_foreign_keys=on"},
	} {
		if got := withForeignKeyEnforcement(testCase.dataSourceName); got != testCase.want {
			t.Errorf("withForeignKeyEnforcement(%q) = %q, want %q", testCase.dataSourceName, got, testCase.want)
		}
	}
}

// TestAPrivateInMemoryMirrorIsOneDatabaseAcrossThePool holds the reason
// OpenSQLiteMirrorOfProductionSchema caps such a pool at one connection. SQLite gives
// ":memory:" a per-connection meaning, and database/sql grows its pool on demand, so
// without the cap the schema lands in the connection that served the first call and the
// second connection opens an empty database — no error, no missing row, just every table
// gone. Measured before the cap: the second connection reported "no such table" for
// bounties.
func TestAPrivateInMemoryMirrorIsOneDatabaseAcrossThePool(t *testing.T) {
	mirror, err := OpenSQLiteMirrorOfProductionSchema(":memory:")
	if err != nil {
		t.Fatalf("open the mirror: %v", err)
	}
	defer mirror.Close()

	if capped := mirror.Stats().MaxOpenConnections; capped != 1 {
		t.Fatalf("a private in-memory pool has to be capped at one connection, got %d", capped)
	}

	for attempt := 0; attempt < 3; attempt++ {
		connection, err := mirror.Conn(t.Context())
		if err != nil {
			t.Fatalf("take connection %d from the pool: %v", attempt, err)
		}
		var name string
		err = connection.QueryRowContext(t.Context(),
			`SELECT name FROM sqlite_master WHERE type='table' AND name='bounties'`).Scan(&name)
		connection.Close()
		if err != nil {
			t.Fatalf("pool connection %d cannot see the schema the mirror just applied: %v", attempt, err)
		}
	}
}

// TestOnlyAPrivateMemoryDataSourceIsCapped keeps the cap off the data source names that do
// not need it. A file database and a shared-cache memory database are both one database
// however many connections read them, and capping either would serialise a pool for
// nothing.
func TestOnlyAPrivateMemoryDataSourceIsCapped(t *testing.T) {
	for _, dataSource := range []struct {
		name      string
		isPrivate bool
	}{
		{":memory:", true},
		{"file::memory:", true},
		{"file:test.db?mode=memory", true},
		{"file::memory:?cache=shared", false},
		{"file:named.db?mode=memory&cache=shared", false},
		{"/tmp/party.db", false},
		{"file:/tmp/party.db?_foreign_keys=on", false},
	} {
		if got := namesAPrivateInMemoryDatabase(dataSource.name); got != dataSource.isPrivate {
			t.Errorf("namesAPrivateInMemoryDatabase(%q) = %v, want %v", dataSource.name, got, dataSource.isPrivate)
		}
	}
}
