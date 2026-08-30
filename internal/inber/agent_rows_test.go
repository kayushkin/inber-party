package inber

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

// What these tests pin.
//
// GetAgents skipped a row whose columns would not scan and returned the shorter slice with a
// nil error. GetStats then published len(agents) as total_agents AND divided the summed levels
// by it to produce average_agent_level, so a dropped row left both the numerator and the
// denominator and the average moved with nothing on screen changing shape. A shortened list is
// at least visible; a moved average is not.
//
// The repair records the drop and never corrects the count -- correcting it would invent an
// agent. So the arms below split into two families: the count must be REPORTED (and reported
// as a number, on every response, including zero), and the published totals must still be the
// honest totals of what was actually read.

// sessionsDBWithUnreadableRows writes a sessions fixture where exactly the named agents carry a
// REAL in-token value.
//
// This is the schema's own hazard, not an invented one. `in_tokens INTEGER DEFAULT 0` has
// INTEGER affinity and is not NOT NULL, and SQLite keeps a value it cannot narrow losslessly --
// so 1.5 stays REAL, SUM(in_tokens+out_tokens) comes back REAL, and the scan into a non-pointer
// int fails for that agent's group and no other. Measured: "converting driver.Value type
// float64 to a int". rows.Err stays nil throughout, which is exactly why the drop was silent.
func sessionsDBWithUnreadableRows(t *testing.T, readable []string, unreadable []string) string {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "sessions.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			agent TEXT NOT NULL,
			status TEXT DEFAULT 'completed',
			started_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE turns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL REFERENCES sessions(id),
			in_tokens INTEGER DEFAULT 0,
			out_tokens INTEGER DEFAULT 0,
			cost REAL DEFAULT 0,
			tool_calls INTEGER DEFAULT 0
		);
	`); err != nil {
		t.Fatal(err)
	}

	insert := func(agent string, inTokens interface{}) {
		sessionID := "sess-" + agent
		if _, err := db.Exec(`INSERT INTO sessions (id, agent, status) VALUES (?, ?, 'completed')`, sessionID, agent); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO turns (session_id, in_tokens, out_tokens, cost, tool_calls) VALUES (?, ?, 100, 0.01, 2)`, sessionID, inTokens); err != nil {
			t.Fatal(err)
		}
	}
	for _, agent := range readable {
		insert(agent, 500)
	}
	for _, agent := range unreadable {
		insert(agent, 1.5)
	}

	// The fixture is worthless unless the poisoning actually took, and SQLite's affinity rules
	// are subtle enough that asserting it here is cheaper than debugging a green run later.
	for _, agent := range unreadable {
		var columnType string
		if err := db.QueryRow(`SELECT typeof(in_tokens) FROM turns WHERE session_id = ?`, "sess-"+agent).Scan(&columnType); err != nil {
			t.Fatal(err)
		}
		if columnType != "real" {
			t.Fatalf("fixture did not poison %s: in_tokens stored as %q, want \"real\" -- this fixture proves nothing", agent, columnType)
		}
	}
	return dbPath
}

// TestAnUnreadableSessionsRowIsCountedNotErased is the headline arm.
func TestAnUnreadableSessionsRowIsCountedNotErased(t *testing.T) {
	dbPath := sessionsDBWithUnreadableRows(t, []string{"brigid", "claxon"}, []string{"morrigan"})
	store, err := NewStore(dbPath, "", "")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	agents, unreadableRows, err := store.GetAgents()
	if err != nil {
		t.Fatalf("GetAgents returned an error for a fixture where two of three rows read fine: %v", err)
	}

	if unreadableRows != 1 {
		t.Errorf("unreadable rows = %d, want 1 -- the row is dropped either way, and the count is the only thing that says so", unreadableRows)
	}
	if len(agents) != 2 {
		t.Errorf("got %d agents, want 2 -- the repair records the drop, it does not invent the missing agent back", len(agents))
	}
	for _, a := range agents {
		if a.ID == "morrigan" {
			t.Errorf("morrigan is present, so the fixture's unreadable row was read after all and this test pins nothing")
		}
	}
}

// TestACleanReadReportsZeroUnreadableRows is the negative control the headline arm needs.
// Without it, an implementation that returns a hardcoded 1 passes the arm above.
func TestACleanReadReportsZeroUnreadableRows(t *testing.T) {
	dbPath := sessionsDBWithUnreadableRows(t, []string{"brigid", "claxon", "morrigan"}, nil)
	store, err := NewStore(dbPath, "", "")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	agents, unreadableRows, err := store.GetAgents()
	if err != nil {
		t.Fatal(err)
	}
	if unreadableRows != 0 {
		t.Errorf("unreadable rows = %d on a fixture every row of which scans, want 0", unreadableRows)
	}
	if len(agents) != 3 {
		t.Errorf("got %d agents, want 3", len(agents))
	}
}

// TestTheCountDistinguishesEveryNumberOfDroppedRows catches a boolean wearing an int's clothes.
// "Some rows were lost" passes both arms above and loses the only number a reader can compare
// against a neighbouring count.
func TestTheCountDistinguishesEveryNumberOfDroppedRows(t *testing.T) {
	poisoned := []string{"morrigan", "lugh", "danu"}
	seen := map[int]int{}
	for dropped := 0; dropped <= 3; dropped++ {
		dbPath := sessionsDBWithUnreadableRows(t, []string{"brigid", "claxon"}, poisoned[:dropped])
		store, err := NewStore(dbPath, "", "")
		if err != nil {
			t.Fatal(err)
		}
		agents, unreadableRows, err := store.GetAgents()
		store.Close()
		if err != nil {
			t.Fatal(err)
		}
		if unreadableRows != dropped {
			t.Errorf("poisoned %d rows, reported %d", dropped, unreadableRows)
		}
		if len(agents) != 2 {
			t.Errorf("poisoned %d rows: got %d agents, want the 2 readable ones", dropped, len(agents))
		}
		seen[unreadableRows]++
	}
	if len(seen) != 4 {
		t.Errorf("four different drop counts produced %d distinct reported values (%v), want 4 -- a count that cannot separate them reports nothing", len(seen), seen)
	}
}

// TestTheAverageIsQualifiedByTheCountItWasTakenOver is the arm the card was filed for.
//
// total_agents and average_agent_level are both computed from the shortened slice, and the
// repair deliberately leaves them that way. What changes is that the object now says how many
// rows it could not read, so a reader can tell "three agents averaging level 4" from "four
// agents, one of which could not be read, averaging level 4 over the three that could".
func TestTheAverageIsQualifiedByTheCountItWasTakenOver(t *testing.T) {
	dbPath := sessionsDBWithUnreadableRows(t, []string{"brigid", "claxon"}, []string{"morrigan"})
	store, err := NewStore(dbPath, "", "")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	stats, err := store.GetStats()
	if err != nil {
		t.Fatal(err)
	}

	if stats.UnreadableAgentRows != 1 {
		t.Errorf("stats.UnreadableAgentRows = %d, want 1 -- without it the average is published as fact over a subset", stats.UnreadableAgentRows)
	}
	if stats.TotalAgents != 2 {
		t.Errorf("stats.TotalAgents = %d, want 2 -- the repair records the drop and never corrects the count", stats.TotalAgents)
	}
	if stats.TotalAgents+stats.UnreadableAgentRows != 3 {
		t.Errorf("total_agents(%d) + unreadable_agent_rows(%d) = %d, want the 3 rows the database held -- the two numbers together are what make the shortfall recoverable", stats.TotalAgents, stats.UnreadableAgentRows, stats.TotalAgents+stats.UnreadableAgentRows)
	}
}

// TestTheStatsFieldIsSerialisedEvenWhenNothingWasLost is the arm an `omitempty` fails.
//
// It is the whole design in one assertion: absent would mean both "nothing was lost" and "this
// build does not report", and a reader that cannot separate those is back where this started.
func TestTheStatsFieldIsSerialisedEvenWhenNothingWasLost(t *testing.T) {
	encoded, err := json.Marshal(&RPGStats{TotalAgents: 3})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"unreadable_agent_rows":0`) {
		t.Errorf("a clean read serialised to %s -- unreadable_agent_rows must be present and zero, because an absent field is a build that makes no claim", encoded)
	}
}

// TestTheGatewayArmCountsItsOwnUnreadableRows pins the second of the two loops.
//
// Both arms are the same shape and it would be easy to repair one. The gateway arm is latent on
// this box today -- ~/.inber/gateway/gateway.db does not exist -- which is exactly why it needs
// a test rather than a live check.
func TestTheGatewayArmCountsItsOwnUnreadableRows(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "gateway.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
		CREATE TABLE sessions (key TEXT PRIMARY KEY, agent TEXT, last_active DATETIME);
		CREATE TABLE requests (id INTEGER PRIMARY KEY AUTOINCREMENT, session_key TEXT, status TEXT, input_tokens INTEGER DEFAULT 0, output_tokens INTEGER DEFAULT 0, cost REAL DEFAULT 0);
		INSERT INTO sessions (key, agent) VALUES ('k1','brigid'),('k2','claxon'),('k3','morrigan');
		INSERT INTO requests (session_key, status, input_tokens, output_tokens, cost) VALUES
			('k1','completed',100,50,0.01),
			('k2','completed',200,60,0.02),
			('k3','completed',1.5,70,0.03);
	`); err != nil {
		t.Fatal(err)
	}
	db.Close()

	store, err := NewStore("", dbPath, "")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	agents, unreadableRows, err := store.GetAgents()
	if err != nil {
		t.Fatal(err)
	}
	if unreadableRows != 1 {
		t.Errorf("gateway unreadable rows = %d, want 1 -- the sessions arm can be repaired on its own and this is the arm that says whether it was", unreadableRows)
	}
	if len(agents) != 2 {
		t.Errorf("got %d agents from the gateway arm, want 2", len(agents))
	}
}

// --- the wire ---

// countingDataSource is a DataSource that reports exactly what it is told to report, so the
// handler arms below test the handler and not the store.
type countingDataSource struct {
	agents         []RPGAgent
	unreadableRows int
}

func (c *countingDataSource) GetAgents() ([]RPGAgent, int, error) {
	return c.agents, c.unreadableRows, nil
}
func (c *countingDataSource) GetQuests(int) ([]RPGQuest, error) { return []RPGQuest{}, nil }
func (c *countingDataSource) GetStats() (*RPGStats, error) {
	return &RPGStats{TotalAgents: len(c.agents), UnreadableAgentRows: c.unreadableRows}, nil
}
func (c *countingDataSource) GetAchievements(string) ([]RPGAchievement, error) {
	return []RPGAchievement{}, nil
}
func (c *countingDataSource) GetQuestHistory(string, int) ([]QuestHistoryEntry, error) {
	return []QuestHistoryEntry{}, nil
}
func (c *countingDataSource) GetConversations(int) ([]RPGConversation, error) {
	return []RPGConversation{}, nil
}
func (c *countingDataSource) GetSessionReplay(string) (*SessionReplay, error) { return nil, nil }
func (c *countingDataSource) GetAgentJournal(string, string) (*RPGJournal, error) {
	return nil, nil
}

func serveAgents(t *testing.T, source DataSource) *http.Response {
	t.Helper()
	mux := http.NewServeMux()
	NewHandler(source).RegisterRoutes(mux)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest("GET", "/api/inber/agents", nil))
	return recorder.Result()
}

// TestTheAgentListHeaderIsWrittenEvenWhenNothingWasLost is the header's version of the
// present-and-zero rule, and the arm an "only report problems" implementation fails.
func TestTheAgentListHeaderIsWrittenEvenWhenNothingWasLost(t *testing.T) {
	response := serveAgents(t, &countingDataSource{agents: []RPGAgent{{ID: "brigid"}}, unreadableRows: 0})

	values, ok := response.Header[unreadableRowsHeader]
	if !ok {
		t.Fatalf("%s is absent after a clean read; present-and-zero is a claim, absent is a build that makes no claim", unreadableRowsHeader)
	}
	if len(values) != 1 || values[0] != "0" {
		t.Errorf("%s = %v, want exactly [\"0\"]", unreadableRowsHeader, values)
	}
}

// TestTheAgentListHeaderCarriesTheCountNotAFlag is the discriminating positive.
func TestTheAgentListHeaderCarriesTheCountNotAFlag(t *testing.T) {
	seen := map[string]int{}
	for _, unreadableRows := range []int{0, 1, 2, 17} {
		response := serveAgents(t, &countingDataSource{unreadableRows: unreadableRows})
		got := response.Header.Get(unreadableRowsHeader)
		if got != fmt.Sprint(unreadableRows) {
			t.Errorf("%s = %q after losing %d rows, want %q", unreadableRowsHeader, got, unreadableRows, fmt.Sprint(unreadableRows))
		}
		seen[got]++
	}
	if len(seen) != 4 {
		t.Errorf("four different row counts produced %d distinct header values (%v), want 4", len(seen), seen)
	}
}

// TestTheAgentListEndpointStillAnswersAnArray is the wire-shape control. The repair deliberately
// did NOT turn the body into an object -- that is a decision on a route the frontend reads, not
// a repair -- and this is the test that says so out loud if somebody later makes that change.
func TestTheAgentListEndpointStillAnswersAnArray(t *testing.T) {
	response := serveAgents(t, &countingDataSource{agents: []RPGAgent{{ID: "brigid"}}})

	var decoded []RPGAgent
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		t.Fatalf("/api/inber/agents no longer answers a bare JSON array: %v -- that is a wire change on a published route, and it needs a decision rather than a repair", err)
	}
	if len(decoded) != 1 {
		t.Errorf("decoded %d agents, want 1", len(decoded))
	}
}

// TestTheStatsEndpointPublishesTheCountInTheBody pins the other half of the channel choice:
// /api/inber/stats answers an object already, so naming one more member is additive and the
// count goes there rather than into a header nobody would look for on an object response.
func TestTheStatsEndpointPublishesTheCountInTheBody(t *testing.T) {
	mux := http.NewServeMux()
	NewHandler(&countingDataSource{agents: []RPGAgent{{ID: "brigid"}, {ID: "claxon"}}, unreadableRows: 3}).RegisterRoutes(mux)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, httptest.NewRequest("GET", "/api/inber/stats", nil))

	var decoded map[string]interface{}
	if err := json.NewDecoder(recorder.Body).Decode(&decoded); err != nil {
		t.Fatal(err)
	}
	got, present := decoded["unreadable_agent_rows"]
	if !present {
		t.Fatalf("/api/inber/stats published %v with no unreadable_agent_rows member -- average_agent_level is then a fact about an unnamed subset", decoded)
	}
	if got != float64(3) {
		t.Errorf("unreadable_agent_rows = %v, want 3", got)
	}
}

// TestTheHTTPClientCountsANamelessRemoteAgent pins the OTHER implementation of DataSource.
//
// It matters because the count is on the interface, and an implementation that returns a zero
// it cannot stand behind would make present-and-zero a lie rather than a claim. HTTPClient does
// drop rows -- a remote agent named in neither field -- so it has a real number to report.
func TestTheHTTPClientCountsANamelessRemoteAgent(t *testing.T) {
	remote := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/agents" {
			w.Write([]byte(`[]`))
			return
		}
		w.Write([]byte(`[{"name":"brigid"},{"name":"","agent":""},{"agent":"claxon"}]`))
	}))
	defer remote.Close()

	agents, unreadableRows, err := NewHTTPClient(remote.URL).GetAgents()
	if err != nil {
		t.Fatal(err)
	}
	if unreadableRows != 1 {
		t.Errorf("unreadable rows = %d, want 1 -- the nameless remote agent is dropped either way", unreadableRows)
	}
	if len(agents) != 2 {
		t.Errorf("got %d agents, want 2", len(agents))
	}

	stats, err := NewHTTPClient(remote.URL).GetStats()
	if err != nil {
		t.Fatal(err)
	}
	if stats.UnreadableAgentRows != 1 {
		t.Errorf("HTTPClient stats.UnreadableAgentRows = %d, want 1 -- this arm computes the same average over the same shortened slice", stats.UnreadableAgentRows)
	}
}
