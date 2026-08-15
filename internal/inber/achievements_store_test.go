package inber

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sort"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// createAchievementGatewayDB builds a gateway database whose single agent lands
// exactly on three thresholds at once — 100,000 tokens, level 5, and a tenth
// completed quest — so a literal that drifts by one in either direction changes
// the answer this test asserts.
//
// The agent's ten completed requests carry 10,000 tokens each. The eleventh
// request fails and carries none, so adding it cannot move the token total off
// its threshold.
func createAchievementGatewayDB(t *testing.T) string {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "gateway.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE sessions (
			key TEXT PRIMARY KEY,
			agent TEXT NOT NULL,
			label TEXT,
			last_active DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_key TEXT NOT NULL REFERENCES sessions(key),
			status TEXT NOT NULL DEFAULT 'pending',
			input_text TEXT,
			turns INTEGER DEFAULT 0,
			input_tokens INTEGER DEFAULT 0,
			output_tokens INTEGER DEFAULT 0,
			cost REAL DEFAULT 0,
			started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			error_text TEXT,
			parent_request_id INTEGER
		);
	`); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Exec(`INSERT INTO sessions (key, agent, last_active) VALUES ('sess-1','claxon','2026-08-14 10:00:00')`); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		// One quest runs 31 turns, which is the first count above the marathon
		// line. One starts at 02:15, inside the night-owl window.
		turns := 1
		if i == 0 {
			turns = 31
		}
		startedAt := fmt.Sprintf("2026-08-14 10:%02d:00", i)
		if i == 1 {
			startedAt = "2026-08-14 02:15:00"
		}
		if _, err := db.Exec(`
			INSERT INTO requests (session_key, status, input_text, turns, input_tokens, output_tokens, cost, started_at, completed_at)
			VALUES ('sess-1','completed','Fix the bug in auth module',?,5000,5000,0.02,?,?)`,
			turns, startedAt, startedAt); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := db.Exec(`
		INSERT INTO requests (session_key, status, input_text, turns, input_tokens, output_tokens, cost, started_at)
		VALUES ('sess-1','error','Deploy to prod',1,0,0,0,'2026-08-14 10:30:00')`); err != nil {
		t.Fatal(err)
	}

	return dbPath
}

func achievementIDs(achievements []RPGAchievement) []string {
	ids := make([]string, 0, len(achievements))
	for _, a := range achievements {
		ids = append(ids, a.ID)
	}
	sort.Strings(ids)
	return ids
}

// TestStoreGetAchievementsUnlocksTheSameSetAsTheSharedComputation pins the Store
// path, which is the one that runs whenever a gateway database is present — that
// is, in the normal deployment. Until this test existed the Store computed
// achievements from its own hand copy of the rules, and moving any threshold in
// that copy by one left the whole suite green.
func TestStoreGetAchievementsUnlocksTheSameSetAsTheSharedComputation(t *testing.T) {
	store, err := NewStore("", createAchievementGatewayDB(t), "")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	agents, err := store.GetAgents()
	if err != nil {
		t.Fatal(err)
	}
	var agent *RPGAgent
	for i := range agents {
		if agents[i].ID == "claxon" {
			agent = &agents[i]
			break
		}
	}
	if agent == nil {
		t.Fatal("claxon agent not found")
	}

	// The fixture is only worth anything if it really lands on the thresholds it
	// claims to, so check that before asserting anything about achievements.
	if agent.TotalTokens != 100000 {
		t.Fatalf("fixture drifted: claxon has %d tokens, want exactly 100000", agent.TotalTokens)
	}
	if agent.Level != 5 {
		t.Fatalf("fixture drifted: claxon is level %d, want exactly 5", agent.Level)
	}

	got, err := store.GetAchievements("claxon")
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"100k_tokens", // exactly 100,000 tokens
		"10_quests",   // exactly 10 completed
		"1k_tokens",
		"first_error", // the one failed request
		"first_quest",
		"level5", // exactly level 5
		"marathon",
		"night_owl",
	}
	sort.Strings(want)
	gotIDs := achievementIDs(got)
	if len(gotIDs) != len(want) {
		t.Fatalf("achievements = %v, want %v", gotIDs, want)
	}
	for i := range want {
		if gotIDs[i] != want[i] {
			t.Fatalf("achievements = %v, want %v", gotIDs, want)
		}
	}

	// The Store must agree with the shared computation given the same inputs. A
	// second implementation of the same rules is free to drift, and this is the
	// assertion that notices when it does.
	quests, err := store.GetQuests(1000)
	if err != nil {
		t.Fatal(err)
	}
	shared := achievementIDs(computeAchievements(agent, quests))
	for i := range shared {
		if shared[i] != gotIDs[i] {
			t.Fatalf("Store.GetAchievements = %v, computeAchievements = %v", gotIDs, shared)
		}
	}
	if len(shared) != len(gotIDs) {
		t.Fatalf("Store.GetAchievements = %v, computeAchievements = %v", gotIDs, shared)
	}
}

// TestStoreGetAchievementsIsEmptyForAnUnknownAgent keeps the early return held.
func TestStoreGetAchievementsIsEmptyForAnUnknownAgent(t *testing.T) {
	store, err := NewStore("", createAchievementGatewayDB(t), "")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	got, err := store.GetAchievements("nobody")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("achievements for an unknown agent = %v, want none", achievementIDs(got))
	}
}
