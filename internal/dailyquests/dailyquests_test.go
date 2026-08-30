package dailyquests

import (
	"database/sql"
	"errors"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/kayushkin/inber-party/internal/db"

	_ "github.com/mattn/go-sqlite3"
)

// What these tests pin, and why the count is the subject rather than the slice.
//
// getActiveAgents skips an agent row it cannot scan. Six of the eleven columns it selects
// (level, xp, energy, status, created_at, updated_at) carry DEFAULTs but are not declared
// NOT NULL, so the schema permits a NULL that will not scan into db.Agent. Before the
// change these tests cover, such a row was logged and then dropped: GenerateDailyQuests
// printed len(agents) as the number of agents, so an agent that could not be read and an
// agent that did not exist produced exactly the same number. The agent got no daily quest,
// permanently, because nothing revisits it.
//
// So the assertion that matters is not "the right agents came back" — that was already true
// — but "the agents that did NOT come back are counted, and the count reaches the caller".

// agentsFixtureDDL creates the agents table the way db.go's production migration declares
// it: NOT NULL on name, title, class and avatar_emoji, and nullable everywhere else. The
// nullability is the whole point of the fixture, so it is copied deliberately rather than
// tightened.
//
// ⚠️ This is a hand-written mirror of production's schema, which is the thing that drifts.
// A derived mirror (db.OpenSQLiteMirrorOfProductionSchema, translating postgresMigrations()
// statement by statement) exists on the branch
// test/the-sqlite-test-mirror-is-a-translation-of-the-production-schema and is the right
// fixture for this test once that branch reaches main. Until then
// TestFixtureDeclaresTheColumnsTheShippedQuerySelects below is the guard against the drift
// that matters here: the scan is positional, so a fixture whose column list disagrees with
// the shipped query would exercise a different statement than the one that ships, and would
// do it without failing.
const agentsFixtureDDL = `
	CREATE TABLE agents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		title TEXT NOT NULL,
		class TEXT NOT NULL,
		level INTEGER DEFAULT 1,
		xp INTEGER DEFAULT 0,
		energy INTEGER DEFAULT 100,
		status TEXT DEFAULT 'idle',
		avatar_emoji TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

func newManagerWithAgentsTable(t *testing.T) *DailyQuestManager {
	t.Helper()
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory SQLite: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	if _, err := sqlDB.Exec(agentsFixtureDDL); err != nil {
		t.Fatalf("create agents fixture table: %v", err)
	}
	return NewDailyQuestManager(&db.DB{DB: sqlDB})
}

// insertReadableAgent writes an agent every column of which will scan.
func insertReadableAgent(t *testing.T, m *DailyQuestManager, name string, level int) {
	t.Helper()
	_, err := m.db.Exec(`
		INSERT INTO agents (name, title, class, level, xp, energy, status, avatar_emoji, created_at, updated_at)
		VALUES (?, 'Tester', 'mage', ?, 0, 100, 'idle', '@', '2026-08-30 00:00:00', '2026-08-30 00:00:00')`,
		name, level)
	if err != nil {
		t.Fatalf("insert readable agent %q: %v", name, err)
	}
}

// insertAgentWithNullLevel writes an agent whose level is NULL. The production schema
// permits this — level carries DEFAULT 1 but is not NOT NULL — and db.Agent.Level is an int,
// so the scan fails on this row and only this row.
func insertAgentWithNullLevel(t *testing.T, m *DailyQuestManager, name string) {
	t.Helper()
	_, err := m.db.Exec(`
		INSERT INTO agents (name, title, class, level, xp, energy, status, avatar_emoji, created_at, updated_at)
		VALUES (?, 'Tester', 'mage', NULL, 0, 100, 'idle', '@', '2026-08-30 00:00:00', '2026-08-30 00:00:00')`,
		name)
	if err != nil {
		t.Fatalf("insert agent with NULL level %q: %v", name, err)
	}
}

// TestEveryRowReadableReportsZeroUnreadable is the known-negative arm. Without it a
// getActiveAgents that returned a constant 1 would pass the positive arm below.
func TestEveryRowReadableReportsZeroUnreadable(t *testing.T) {
	m := newManagerWithAgentsTable(t)
	insertReadableAgent(t, m, "Brigid", 3)
	insertReadableAgent(t, m, "Lugh", 7)

	agents, unreadableRows, err := m.getActiveAgents()
	if err != nil {
		t.Fatalf("getActiveAgents: %v", err)
	}
	if len(agents) != 2 {
		t.Errorf("read %d agents, want 2", len(agents))
	}
	if unreadableRows != 0 {
		t.Errorf("reported %d unreadable rows, want 0 — every row in this fixture scans", unreadableRows)
	}
}

// TestEveryColumnLandsInItsOwnField pins the SELECT list against the Scan destinations,
// which the name comparison in TestFixtureDeclaresTheColumnsTheShippedQuerySelects cannot do.
//
// scanAgentRows binds by position. Swap two columns of the same type in activeAgentsQuery —
// xp and energy are both integers — and the query still runs, every row still scans, the
// column sets still agree, and every other test in this file still passes. The values simply
// land in each other's fields. Measured: that mutation was MISSED by all seven other tests
// here. The only thing that catches it is giving each column a value no other column has and
// checking where it arrives.
func TestEveryColumnLandsInItsOwnField(t *testing.T) {
	m := newManagerWithAgentsTable(t)
	_, err := m.db.Exec(`
		INSERT INTO agents (name, title, class, level, xp, energy, status, avatar_emoji, created_at, updated_at)
		VALUES ('Dagda', 'Chief', 'druid', 3, 17, 41, 'questing', '#', '2026-08-30 00:00:00', '2026-08-30 00:00:00')`)
	if err != nil {
		t.Fatalf("insert agent: %v", err)
	}

	agents, unreadableRows, err := m.getActiveAgents()
	if err != nil {
		t.Fatalf("getActiveAgents: %v", err)
	}
	if unreadableRows != 0 {
		t.Fatalf("unreadableRows = %d, want 0", unreadableRows)
	}
	if len(agents) != 1 {
		t.Fatalf("read %d agents, want 1", len(agents))
	}
	agent := agents[0]

	// Every value below is unique across the row, so a column landing in the wrong field
	// cannot be masked by another column happening to hold the same value.
	for _, check := range []struct {
		column string
		got    any
		want   any
	}{
		{"name", agent.Name, "Dagda"},
		{"title", agent.Title, "Chief"},
		{"class", agent.Class, "druid"},
		{"level", agent.Level, 3},
		{"xp", agent.XP, 17},
		{"energy", agent.Energy, 41},
		{"status", agent.Status, "questing"},
		{"avatar_emoji", agent.AvatarEmoji, "#"},
	} {
		if check.got != check.want {
			t.Errorf("column %s landed as %v, want %v — the SELECT list and the Scan destinations are out of step", check.column, check.got, check.want)
		}
	}
	if agent.ID == 0 {
		t.Error("id landed as 0; the first inserted row cannot have id 0")
	}
	if agent.CreatedAt.IsZero() || agent.UpdatedAt.IsZero() {
		t.Errorf("created_at (%v) / updated_at (%v) did not land; one of them is zero", agent.CreatedAt, agent.UpdatedAt)
	}
}

// TestAnUnreadableAgentRowIsCountedNotErased is the defect this change repairs. The row is
// still skipped; what changed is that the caller is told it was skipped.
func TestAnUnreadableAgentRowIsCountedNotErased(t *testing.T) {
	m := newManagerWithAgentsTable(t)
	insertReadableAgent(t, m, "Brigid", 3)
	insertAgentWithNullLevel(t, m, "Nuada")
	insertReadableAgent(t, m, "Lugh", 7)

	agents, unreadableRows, err := m.getActiveAgents()
	if err != nil {
		t.Fatalf("getActiveAgents: %v", err)
	}
	if len(agents) != 2 {
		t.Fatalf("read %d agents, want 2 — the NULL level row must still be skipped", len(agents))
	}
	if unreadableRows != 1 {
		t.Errorf("reported %d unreadable rows, want 1 — the skipped row is the whole finding, and a 0 here is the old behaviour where it vanished", unreadableRows)
	}

	// The distinguishing assertion. Two agents came back either way; only the count
	// separates "there are two agents" from "there are three and one could not be read".
	readable := 0
	for _, a := range agents {
		if a.Name == "Nuada" {
			t.Errorf("agent %q was returned, but its level is NULL and cannot scan", a.Name)
		}
		readable++
	}
	if readable+unreadableRows != 3 {
		t.Errorf("agents read (%d) + rows unreadable (%d) = %d, want 3 — the two numbers together must account for every row the database returned", readable, unreadableRows, readable+unreadableRows)
	}
}

// TestGenerateDailyQuestsReportsUnreadableRowsToItsCaller pins the call site rather than the
// helper. A repair that only fixed getActiveAgents and left GenerateDailyQuests reading
// len(agents) would pass both tests above.
//
// It asserts on the report's agent counts, not on QuestsCreated: quest insertion runs
// PostgreSQL-flavoured SQL that this SQLite fixture does not serve, so createDailyQuest
// fails here and GenerateDailyQuests logs and continues past it by design. That is stated
// rather than worked around, because a test that quietly asserted QuestsCreated == 0 would
// be passing for a reason it does not name.
func TestGenerateDailyQuestsReportsUnreadableRowsToItsCaller(t *testing.T) {
	m := newManagerWithAgentsTable(t)
	insertReadableAgent(t, m, "Brigid", 3)
	insertAgentWithNullLevel(t, m, "Nuada")

	report, err := m.GenerateDailyQuests()
	if err != nil {
		t.Fatalf("GenerateDailyQuests: %v", err)
	}
	if report.UnreadableAgentRows != 1 {
		t.Errorf("report.UnreadableAgentRows = %d, want 1", report.UnreadableAgentRows)
	}
	if report.AgentsQuested != 1 {
		t.Errorf("report.AgentsQuested = %d, want 1", report.AgentsQuested)
	}
	if report.ReadEveryAgent() {
		t.Error("report.ReadEveryAgent() = true, but one agent row was unreadable")
	}
}

// TestReadEveryAgentIsTrueOnlyWhenNothingWasLost is the known-negative for ReadEveryAgent.
func TestReadEveryAgentIsTrueOnlyWhenNothingWasLost(t *testing.T) {
	m := newManagerWithAgentsTable(t)
	insertReadableAgent(t, m, "Brigid", 3)

	report, err := m.GenerateDailyQuests()
	if err != nil {
		t.Fatalf("GenerateDailyQuests: %v", err)
	}
	if !report.ReadEveryAgent() {
		t.Errorf("report.ReadEveryAgent() = false with %d unreadable rows, want true", report.UnreadableAgentRows)
	}
}

// --- the iteration-failure arm ------------------------------------------------------------

// stubAgentRowCursor drives scanAgentRows through outcomes a real driver gives no reliable
// way to produce, in particular a failure of the iteration itself.
type stubAgentRowCursor struct {
	remaining int
	scanErr   error
	iterErr   error
}

func (c *stubAgentRowCursor) Next() bool {
	if c.remaining == 0 {
		return false
	}
	c.remaining--
	return true
}

func (c *stubAgentRowCursor) Scan(dest ...any) error { return c.scanErr }
func (c *stubAgentRowCursor) Err() error             { return c.iterErr }

// TestAnIterationFailureIsReturnedNotCounted pins the second silent drop at this site. If
// row iteration fails part-way, the loop simply ends: every remaining row is lost and how
// many were lost is not knowable, because the driver stopped wherever it failed. Counting
// them would be a guess and returning the short slice with a nil error would present it as
// complete, so the error is returned instead.
func TestAnIterationFailureIsReturnedNotCounted(t *testing.T) {
	iterationFailure := errors.New("connection reset mid-scan")
	cursor := &stubAgentRowCursor{remaining: 0, iterErr: iterationFailure}

	agents, unreadableRows, err := scanAgentRows(cursor)
	if err == nil {
		t.Fatal("scanAgentRows returned a nil error after rows.Err() failed — a truncated result presented as complete is the drop this repair exists to stop")
	}
	if !errors.Is(err, iterationFailure) {
		t.Errorf("returned error %v does not wrap the iteration failure", err)
	}
	if agents != nil {
		t.Errorf("returned %d agents alongside the error, want none — a partial list invites the caller to use it", len(agents))
	}
	if unreadableRows != 0 {
		t.Errorf("returned unreadableRows = %d alongside the error, want 0 — the number lost is not knowable here, so any count would be invented", unreadableRows)
	}
}

// TestScanFailuresAreCountedNotReturnedAsAnError is the discriminating negative for the test
// above: a scan failure and an iteration failure must not be handled the same way. A repair
// that returned an error on any failure would pass the iteration arm and break the service
// on a single NULL.
func TestScanFailuresAreCountedNotReturnedAsAnError(t *testing.T) {
	cursor := &stubAgentRowCursor{remaining: 3, scanErr: errors.New("converting NULL to int is unsupported")}

	agents, unreadableRows, err := scanAgentRows(cursor)
	if err != nil {
		t.Fatalf("scanAgentRows returned %v; a row that will not scan is counted, not fatal", err)
	}
	if len(agents) != 0 {
		t.Errorf("read %d agents, want 0", len(agents))
	}
	if unreadableRows != 3 {
		t.Errorf("unreadableRows = %d, want 3", unreadableRows)
	}
}

// --- the fixture-agreement control ---------------------------------------------------------

// TestFixtureDeclaresTheColumnsTheShippedQuerySelects is the guard on the hand-written
// fixture above. scanAgentRows binds by position, so if activeAgentsQuery gains, loses or
// reorders a column and the fixture does not, every test in this file would keep running —
// against a statement that is not the one that ships. Comparing the two column lists is what
// makes that a failure instead of a silence.
func TestFixtureDeclaresTheColumnsTheShippedQuerySelects(t *testing.T) {
	selected := selectedColumns(t, activeAgentsQuery)
	declared := declaredColumns(t, agentsFixtureDDL)

	for _, column := range selected {
		if !contains(declared, column) {
			t.Errorf("activeAgentsQuery selects %q, which the fixture table does not declare", column)
		}
	}

	// Also assert the query is actually runnable against the fixture, which catches a
	// mismatch the name comparison above cannot — a column of the wrong type, or a WHERE
	// clause naming something absent.
	m := newManagerWithAgentsTable(t)
	rows, err := m.db.Query(activeAgentsQuery)
	if err != nil {
		t.Fatalf("the shipped query does not run against the fixture table: %v", err)
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		t.Fatalf("read result columns: %v", err)
	}
	if len(columns) != len(selected) {
		t.Errorf("the query returned %d columns but its SELECT list parses to %d; the parser below is wrong, not the query", len(columns), len(selected))
	}
}

// selectedColumns pulls the column names out of a single-table SELECT. It is deliberately
// narrow: it handles the one statement in this package and fails loudly on anything else,
// rather than guessing and letting the control above pass on a shape it never read.
func selectedColumns(t *testing.T, query string) []string {
	t.Helper()
	match := regexp.MustCompile(`(?is)\bSELECT\s+(.*?)\s+\bFROM\b`).FindStringSubmatch(query)
	if match == nil {
		t.Fatalf("could not find a SELECT ... FROM in %q", query)
	}
	var columns []string
	for _, part := range strings.Split(match[1], ",") {
		name := strings.TrimSpace(part)
		if name == "" || strings.ContainsAny(name, "()* ") {
			t.Fatalf("column %q is not a bare name; this parser only reads simple column lists", name)
		}
		columns = append(columns, strings.ToLower(name))
	}
	return columns
}

// declaredColumns pulls the column names out of the fixture's CREATE TABLE.
func declaredColumns(t *testing.T, ddl string) []string {
	t.Helper()
	open := strings.Index(ddl, "(")
	closeAt := strings.LastIndex(ddl, ")")
	if open < 0 || closeAt < open {
		t.Fatalf("could not find a column list in %q", ddl)
	}
	var columns []string
	for _, line := range strings.Split(ddl[open+1:closeAt], ",") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) == 0 {
			continue
		}
		columns = append(columns, strings.ToLower(fields[0]))
	}
	sort.Strings(columns)
	return columns
}

func contains(haystack []string, needle string) bool {
	for _, candidate := range haystack {
		if candidate == needle {
			return true
		}
	}
	return false
}
