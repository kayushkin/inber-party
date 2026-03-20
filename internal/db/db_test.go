package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Test helper to create a temporary SQLite database for testing
func setupTestDB(t *testing.T) (*DB, func()) {
	// Create a temporary file for SQLite
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	
	// Use SQLite for testing (easier to set up than PostgreSQL)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	testDB := &DB{db}
	
	// Cleanup function
	cleanup := func() {
		testDB.Close()
		os.Remove(dbPath)
	}
	
	return testDB, cleanup
}

// Test helper to create an in-memory SQLite database
func setupInMemoryDB(t *testing.T) (*DB, func()) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	testDB := &DB{db}
	
	cleanup := func() {
		testDB.Close()
	}
	
	return testDB, cleanup
}

func TestConnect(t *testing.T) {
	tests := []struct {
		name        string
		databaseURL string
		shouldFail  bool
	}{
		{
			name:        "invalid database URL",
			databaseURL: "invalid://url",
			shouldFail:  true,
		},
		{
			name:        "empty database URL",
			databaseURL: "",
			shouldFail:  true,
		},
		{
			name:        "malformed postgres URL",
			databaseURL: "postgres://invalid",
			shouldFail:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := Connect(tt.databaseURL)
			
			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tt.name)
					if db != nil {
						db.Close()
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for %s, but got: %v", tt.name, err)
				}
				if db != nil {
					db.Close()
				}
			}
		})
	}
}

func TestDB_Close(t *testing.T) {
	db, cleanup := setupInMemoryDB(t)
	defer cleanup()

	// Verify the connection works
	err := db.Ping()
	if err != nil {
		t.Fatalf("Database should be accessible before closing: %v", err)
	}

	// Close the database
	err = db.Close()
	if err != nil {
		t.Fatalf("Failed to close database: %v", err)
	}

	// Verify the connection no longer works
	err = db.Ping()
	if err == nil {
		t.Error("Database should not be accessible after closing")
	}
}

func TestDB_Migrate(t *testing.T) {
	db, cleanup := setupInMemoryDB(t)
	defer cleanup()

	// Convert PostgreSQL-specific migrations to SQLite for testing
	err := db.migrateSQLite()
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify that tables were created
	tables := []string{"agents", "parties", "bounties", "party_members", "applications", "skills", "agent_skills", "payouts", "ratings", "reputation"}
	
	for _, table := range tables {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='%s'", table)
		err := db.QueryRow(query).Scan(&count)
		if err != nil {
			t.Errorf("Failed to check for table %s: %v", table, err)
			continue
		}
		
		if count != 1 {
			t.Errorf("Expected table %s to exist after migration", table)
		}
	}
}

func TestDB_Migrate_Idempotent(t *testing.T) {
	db, cleanup := setupInMemoryDB(t)
	defer cleanup()

	// Run migration multiple times - should not fail
	for i := 0; i < 3; i++ {
		err := db.migrateSQLite()
		if err != nil {
			t.Fatalf("Migration attempt %d failed: %v", i+1, err)
		}
	}

	// Verify tables still exist and are accessible
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM agents").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query agents table after multiple migrations: %v", err)
	}
}

func TestDB_BasicOperations(t *testing.T) {
	db, cleanup := setupInMemoryDB(t)
	defer cleanup()

	// Run migrations first
	err := db.migrateSQLite()
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Test inserting an agent
	_, err = db.Exec(`
		INSERT INTO agents (name, title, class, level, xp, energy, avatar_emoji) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"TestAgent", "Test Engineer", "engineer", 1, 0, 100, "🤖")
	if err != nil {
		t.Fatalf("Failed to insert test agent: %v", err)
	}

	// Test reading the agent back
	var name, title, class, emoji string
	var level, xp, energy int
	err = db.QueryRow("SELECT name, title, class, level, xp, energy, avatar_emoji FROM agents WHERE name = ?", "TestAgent").
		Scan(&name, &title, &class, &level, &xp, &energy, &emoji)
	if err != nil {
		t.Fatalf("Failed to query test agent: %v", err)
	}

	// Verify the data
	if name != "TestAgent" {
		t.Errorf("Expected name 'TestAgent', got '%s'", name)
	}
	if class != "engineer" {
		t.Errorf("Expected class 'engineer', got '%s'", class)
	}
	if level != 1 {
		t.Errorf("Expected level 1, got %d", level)
	}
}

func TestDB_TransactionRollback(t *testing.T) {
	db, cleanup := setupInMemoryDB(t)
	defer cleanup()

	// Run migrations
	err := db.migrateSQLite()
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Insert test data in transaction
	_, err = tx.Exec(`
		INSERT INTO agents (name, title, class, level, xp, energy, avatar_emoji) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"TransactionTest", "Test Engineer", "engineer", 1, 0, 100, "🔧")
	if err != nil {
		t.Fatalf("Failed to insert in transaction: %v", err)
	}

	// Rollback the transaction
	err = tx.Rollback()
	if err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}

	// Verify the data was not persisted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM agents WHERE name = ?", "TransactionTest").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check rollback result: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 agents after rollback, got %d", count)
	}
}

func TestDB_MultipleOperations(t *testing.T) {
	db, cleanup := setupInMemoryDB(t)
	defer cleanup()

	// Run migrations
	err := db.migrateSQLite()
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Test multiple sequential operations to verify database stability
	const numOperations = 10
	
	for i := 0; i < numOperations; i++ {
		agentName := fmt.Sprintf("SequentialAgent%d", i)
		_, err := db.Exec(`
			INSERT INTO agents (name, title, class, level, xp, energy, avatar_emoji) 
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			agentName, "Sequential Tester", "engineer", 1, 0, 100, "⚙️")
		if err != nil {
			t.Fatalf("Operation %d failed: %v", i, err)
		}
	}

	// Verify all agents were created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM agents WHERE name LIKE 'SequentialAgent%'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count sequential agents: %v", err)
	}

	if count != numOperations {
		t.Errorf("Expected %d sequential agents, got %d", numOperations, count)
	}

	// Test reading all agents
	rows, err := db.Query("SELECT name, class FROM agents WHERE name LIKE 'SequentialAgent%' ORDER BY name")
	if err != nil {
		t.Fatalf("Failed to query agents: %v", err)
	}
	defer rows.Close()

	readCount := 0
	for rows.Next() {
		var name, class string
		err := rows.Scan(&name, &class)
		if err != nil {
			t.Fatalf("Failed to scan row %d: %v", readCount, err)
		}
		
		expectedName := fmt.Sprintf("SequentialAgent%d", readCount)
		if name != expectedName {
			t.Errorf("Expected name %s, got %s", expectedName, name)
		}
		
		if class != "engineer" {
			t.Errorf("Expected class 'engineer', got '%s'", class)
		}
		
		readCount++
	}

	if readCount != numOperations {
		t.Errorf("Expected to read %d agents, got %d", numOperations, readCount)
	}
}

func TestDB_PreparedStatements(t *testing.T) {
	db, cleanup := setupInMemoryDB(t)
	defer cleanup()

	// Run migrations
	err := db.migrateSQLite()
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Test prepared statements
	stmt, err := db.Prepare("INSERT INTO agents (name, title, class, level, xp, energy, avatar_emoji) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		t.Fatalf("Failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	// Execute the prepared statement multiple times
	agents := []string{"PreparedAgent1", "PreparedAgent2", "PreparedAgent3"}
	for _, agentName := range agents {
		_, err = stmt.Exec(agentName, "Prepared Tester", "engineer", 1, 0, 100, "📝")
		if err != nil {
			t.Errorf("Failed to execute prepared statement for %s: %v", agentName, err)
		}
	}

	// Verify all agents were created
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM agents WHERE name LIKE 'PreparedAgent%'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count prepared agents: %v", err)
	}

	if count != len(agents) {
		t.Errorf("Expected %d prepared agents, got %d", len(agents), count)
	}
}

// migrateSQLite runs SQLite-compatible migrations for testing
func (db *DB) migrateSQLite() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS agents (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			title TEXT NOT NULL,
			class TEXT NOT NULL,
			level INTEGER DEFAULT 1,
			xp INTEGER DEFAULT 0,
			energy INTEGER DEFAULT 100,
			status TEXT DEFAULT 'idle',
			avatar_emoji TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS parties (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			leader_id INTEGER REFERENCES agents(id) ON DELETE CASCADE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS bounties (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			payout REAL NOT NULL,
			status TEXT DEFAULT 'open',
			created_by INTEGER REFERENCES agents(id) ON DELETE CASCADE,
			assigned_to INTEGER REFERENCES agents(id) ON DELETE SET NULL,
			tier TEXT NOT NULL,
			difficulty TEXT NOT NULL,
			type TEXT NOT NULL,
			tags TEXT,
			deadline DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS party_members (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			party_id INTEGER REFERENCES parties(id) ON DELETE CASCADE,
			agent_id INTEGER REFERENCES agents(id) ON DELETE CASCADE,
			role TEXT DEFAULT 'member',
			joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(party_id, agent_id)
		)`,
		`CREATE TABLE IF NOT EXISTS applications (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bounty_id INTEGER REFERENCES bounties(id) ON DELETE CASCADE,
			agent_id INTEGER REFERENCES agents(id) ON DELETE CASCADE,
			message TEXT,
			status TEXT DEFAULT 'pending',
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			reviewed_at DATETIME,
			UNIQUE(bounty_id, agent_id)
		)`,
		`CREATE TABLE IF NOT EXISTS skills (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			description TEXT,
			category TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS agent_skills (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			agent_id INTEGER REFERENCES agents(id) ON DELETE CASCADE,
			skill_id INTEGER REFERENCES skills(id) ON DELETE CASCADE,
			level INTEGER DEFAULT 1,
			acquired_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(agent_id, skill_id)
		)`,
		`CREATE TABLE IF NOT EXISTS payouts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bounty_id INTEGER REFERENCES bounties(id) ON DELETE CASCADE,
			recipient_id INTEGER REFERENCES agents(id) ON DELETE CASCADE,
			amount REAL NOT NULL,
			reason TEXT,
			status TEXT DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			processed_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS ratings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bounty_id INTEGER REFERENCES bounties(id) ON DELETE CASCADE,
			rater_id INTEGER REFERENCES agents(id) ON DELETE CASCADE,
			rated_id INTEGER REFERENCES agents(id) ON DELETE CASCADE,
			score INTEGER NOT NULL CHECK (score >= 1 AND score <= 5),
			comment TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(bounty_id, rater_id, rated_id)
		)`,
		`CREATE TABLE IF NOT EXISTS reputation (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			agent_id INTEGER UNIQUE REFERENCES agents(id) ON DELETE CASCADE,
			total_score REAL DEFAULT 0,
			total_ratings INTEGER DEFAULT 0,
			average_rating REAL DEFAULT 0,
			last_updated DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, migration := range migrations {
		_, err := db.Exec(migration)
		if err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}