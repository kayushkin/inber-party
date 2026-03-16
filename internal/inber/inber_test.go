package inber

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// createTestGatewayDB creates a temporary gateway.db with test data.
func createTestGatewayDB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "gateway.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Create schema matching inber's gateway.db
	_, err = db.Exec(`
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
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Insert test data
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = db.Exec(`
		INSERT INTO sessions (key, agent, last_active) VALUES
			('sess-1', 'claxon', ?),
			('sess-2', 'brigid', ?),
			('sess-3', 'run', ?);
	`, now, now, now)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		INSERT INTO requests (session_key, status, input_text, turns, input_tokens, output_tokens, cost, started_at, completed_at, parent_request_id) VALUES
			('sess-1', 'completed', 'Fix the bug in auth module', 5, 1000, 500, 0.02, ?, ?, NULL),
			('sess-1', 'completed', 'Write tests for auth', 3, 800, 400, 0.015, ?, ?, NULL),
			('sess-1', 'error', 'Deploy to prod', 1, 200, 100, 0.003, ?, NULL, NULL),
			('sess-2', 'completed', 'Review pull request', 2, 500, 300, 0.01, ?, ?, NULL),
			('sess-3', 'running', 'Build new feature', 10, 5000, 3000, 0.1, ?, NULL, NULL),
			('sess-1', 'completed', 'Sub-task of bug fix', 2, 300, 200, 0.005, ?, ?, 1);
	`, now, now, now, now, now, now, now, now, now, now)
	if err != nil {
		t.Fatal(err)
	}

	return dbPath
}

// createTestSessionsDB creates a temporary sessions.db with test data.
func createTestSessionsDB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "sessions.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
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
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		INSERT INTO sessions (id, agent, status) VALUES
			('s1', 'claxon', 'completed'),
			('s2', 'brigid', 'completed');
		INSERT INTO turns (session_id, in_tokens, out_tokens, cost, tool_calls) VALUES
			('s1', 500, 300, 0.01, 5),
			('s1', 600, 400, 0.012, 3),
			('s2', 200, 100, 0.005, 2);
	`)
	if err != nil {
		t.Fatal(err)
	}

	return dbPath
}

func TestNewStore_NoDBs(t *testing.T) {
	_, err := NewStore("/nonexistent/sessions.db", "/nonexistent/gateway.db")
	if err == nil {
		t.Fatal("expected error when no DBs found")
	}
}

func TestGetAgents_GatewayOnly(t *testing.T) {
	gwPath := createTestGatewayDB(t)
	store, err := NewStore("", gwPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	agents, err := store.GetAgents()
	if err != nil {
		t.Fatal(err)
	}

	if len(agents) != 3 {
		t.Fatalf("expected 3 agents, got %d", len(agents))
	}

	// Find claxon
	var claxon *RPGAgent
	for i := range agents {
		if agents[i].ID == "claxon" {
			claxon = &agents[i]
			break
		}
	}
	if claxon == nil {
		t.Fatal("claxon agent not found")
	}

	if claxon.Class != "Wizard" {
		t.Errorf("expected claxon class Wizard, got %s", claxon.Class)
	}
	if claxon.TotalTokens != 3500 { // 1000+500 + 800+400 + 200+100 + 300+200
		t.Errorf("expected claxon tokens 3500, got %d", claxon.TotalTokens)
	}
	if claxon.ErrorCount != 1 {
		t.Errorf("expected claxon error count 1, got %d", claxon.ErrorCount)
	}
	if claxon.Name != "Claxon" {
		t.Errorf("expected claxon name 'Claxon', got %s", claxon.Name)
	}
}

func TestGetAgents_BothDBs(t *testing.T) {
	gwPath := createTestGatewayDB(t)
	sessPath := createTestSessionsDB(t)
	store, err := NewStore(sessPath, gwPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	agents, err := store.GetAgents()
	if err != nil {
		t.Fatal(err)
	}

	// Should have 3 agents (claxon, brigid, run — brigid appears in both DBs)
	if len(agents) != 3 {
		t.Fatalf("expected 3 agents, got %d", len(agents))
	}

	// Claxon should have Tool Mastery skill from sessions.db
	var claxon *RPGAgent
	for i := range agents {
		if agents[i].ID == "claxon" {
			claxon = &agents[i]
			break
		}
	}
	if claxon == nil {
		t.Fatal("claxon not found")
	}

	hasToolMastery := false
	for _, s := range claxon.Skills {
		if s.Name == "Tool Mastery" {
			hasToolMastery = true
			if s.TaskCount != 8 { // 5 + 3 tool calls
				t.Errorf("expected 8 tool calls for Tool Mastery, got %d", s.TaskCount)
			}
		}
	}
	if !hasToolMastery {
		t.Error("expected claxon to have Tool Mastery skill")
	}
}

func TestGetQuests(t *testing.T) {
	gwPath := createTestGatewayDB(t)
	store, err := NewStore("", gwPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	quests, err := store.GetQuests(50)
	if err != nil {
		t.Fatal(err)
	}

	if len(quests) != 6 {
		t.Fatalf("expected 6 quests, got %d", len(quests))
	}

	// Check statuses mapped correctly
	statusCounts := map[string]int{}
	for _, q := range quests {
		statusCounts[q.Status]++
	}
	if statusCounts["completed"] != 4 { // 3 top-level + 1 sub-quest
		t.Errorf("expected 4 completed quests, got %d", statusCounts["completed"])
	}
	if statusCounts["failed"] != 1 {
		t.Errorf("expected 1 failed quest, got %d", statusCounts["failed"])
	}
	if statusCounts["in_progress"] != 1 {
		t.Errorf("expected 1 in_progress quest, got %d", statusCounts["in_progress"])
	}

	// Check sub-quest (request 6 has parent_request_id=1)
	// Request 1 should have children=1
	var parentQuest *RPGQuest
	for i := range quests {
		if quests[i].ID == 1 {
			parentQuest = &quests[i]
			break
		}
	}
	if parentQuest != nil && parentQuest.Children != 1 {
		t.Errorf("expected parent quest to have 1 child, got %d", parentQuest.Children)
	}
}

func TestGetQuests_Limit(t *testing.T) {
	gwPath := createTestGatewayDB(t)
	store, err := NewStore("", gwPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	quests, err := store.GetQuests(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(quests) != 2 {
		t.Fatalf("expected 2 quests with limit, got %d", len(quests))
	}
}

func TestGetStats(t *testing.T) {
	gwPath := createTestGatewayDB(t)
	store, err := NewStore("", gwPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	stats, err := store.GetStats()
	if err != nil {
		t.Fatal(err)
	}

	if stats.TotalAgents != 3 {
		t.Errorf("expected 3 agents, got %d", stats.TotalAgents)
	}
	if stats.CompletedQuests != 4 {
		t.Errorf("expected 4 completed, got %d", stats.CompletedQuests)
	}
	if stats.FailedQuests != 1 {
		t.Errorf("expected 1 failed, got %d", stats.FailedQuests)
	}
	if stats.ActiveQuests != 1 {
		t.Errorf("expected 1 active, got %d", stats.ActiveQuests)
	}
	if stats.TotalTokens <= 0 {
		t.Error("expected positive total tokens")
	}
}

func TestLevelForXP(t *testing.T) {
	tests := []struct {
		xp    int
		level int
	}{
		{0, 1},
		{50, 1},
		{100, 2},
		{300, 3},
		{600, 4},
	}
	for _, tc := range tests {
		level, _ := levelForXP(tc.xp)
		if level != tc.level {
			t.Errorf("levelForXP(%d) = %d, want %d", tc.xp, level, tc.level)
		}
	}
}

func TestXPForTokens(t *testing.T) {
	if xpForTokens(0) != 0 {
		t.Error("0 tokens should give 0 XP")
	}
	if xpForTokens(100) != 1 {
		t.Error("100 tokens should give 1 XP")
	}
	if xpForTokens(1500) != 15 {
		t.Error("1500 tokens should give 15 XP")
	}
}

func TestEnergyFromActivity(t *testing.T) {
	// nil = full energy
	if energyFromActivity(nil) != 100 {
		t.Error("nil activity should give 100 energy")
	}

	// very recent = low energy
	now := time.Now()
	e := energyFromActivity(&now)
	if e < 20 || e > 30 {
		t.Errorf("very recent activity should give ~20 energy, got %d", e)
	}

	// 10 hours ago = full energy
	old := time.Now().Add(-10 * time.Hour)
	e = energyFromActivity(&old)
	if e != 100 {
		t.Errorf("10h ago should give 100 energy, got %d", e)
	}
}

func TestTitleCase(t *testing.T) {
	if titleCase("claxon") != "Claxon" {
		t.Error("expected Claxon")
	}
	if titleCase("") != "" {
		t.Error("empty string should stay empty")
	}
}

func TestClassFor(t *testing.T) {
	class, emoji, _ := classFor("claxon")
	if class != "Wizard" || emoji != "🧙" {
		t.Errorf("claxon should be Wizard 🧙, got %s %s", class, emoji)
	}

	class, emoji, _ = classFor("unknown_agent")
	if class != "Warrior" || emoji != "⚔️" {
		t.Error("unknown agent should default to Warrior")
	}
}

func TestStoreClose(t *testing.T) {
	gwPath := createTestGatewayDB(t)
	store, err := NewStore("", gwPath)
	if err != nil {
		t.Fatal(err)
	}
	// Should not panic
	store.Close()
}

func TestGenerateQuestName(t *testing.T) {
	tests := []struct {
		input  string
		status string
		prefix string
	}{
		{"Fix the bug", "completed", "🛡️"},
		{"Build new feature", "completed", "🔨"},
		{"Test the auth module", "running", "⚗️"},
		{"Deploy to production", "completed", "🚀"},
		{"", "error", ""},
		{"Something random", "completed", "📋"},
	}
	for _, tc := range tests {
		name := generateQuestName(tc.input, tc.status)
		if tc.prefix != "" && len(name) < len(tc.prefix) {
			t.Errorf("generateQuestName(%q, %q) = %q, expected prefix %q", tc.input, tc.status, name, tc.prefix)
		}
		if name == "" {
			t.Errorf("generateQuestName(%q, %q) returned empty", tc.input, tc.status)
		}
	}
}

func TestDefaultDBPaths(t *testing.T) {
	sessPath, gwPath := DefaultDBPaths()
	if sessPath == "" || gwPath == "" {
		t.Error("DefaultDBPaths returned empty paths")
	}
	// Should contain .inber
	home, _ := os.UserHomeDir()
	expectedSess := filepath.Join(home, ".inber", "sessions.db")
	if sessPath != expectedSess {
		t.Errorf("expected sessions path %s, got %s", expectedSess, sessPath)
	}
}
