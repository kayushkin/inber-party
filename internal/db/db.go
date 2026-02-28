package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

// Connect establishes a connection to PostgreSQL
func Connect(databaseURL string) (*DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✓ Connected to database")
	return &DB{db}, nil
}

// Migrate runs all database migrations
func (db *DB) Migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS agents (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			title VARCHAR(255) NOT NULL,
			class VARCHAR(100) NOT NULL,
			level INTEGER DEFAULT 1,
			xp INTEGER DEFAULT 0,
			energy INTEGER DEFAULT 100,
			status VARCHAR(50) DEFAULT 'idle',
			avatar_emoji VARCHAR(10) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			difficulty INTEGER DEFAULT 1,
			xp_reward INTEGER DEFAULT 10,
			status VARCHAR(50) DEFAULT 'available',
			assigned_agent_id INTEGER REFERENCES agents(id),
			progress INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			started_at TIMESTAMP,
			completed_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS skills (
			id SERIAL PRIMARY KEY,
			agent_id INTEGER REFERENCES agents(id) ON DELETE CASCADE,
			skill_name VARCHAR(255) NOT NULL,
			level INTEGER DEFAULT 1,
			task_count INTEGER DEFAULT 0,
			UNIQUE(agent_id, skill_name)
		)`,
		`CREATE TABLE IF NOT EXISTS achievements (
			id SERIAL PRIMARY KEY,
			agent_id INTEGER REFERENCES agents(id) ON DELETE CASCADE,
			achievement_name VARCHAR(255) NOT NULL,
			unlocked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(agent_id, achievement_name)
		)`,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i, err)
		}
	}

	log.Println("✓ Migrations complete")
	return nil
}

// Seed creates default agents if none exist
func (db *DB) Seed() error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM agents").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check agents: %w", err)
	}

	if count > 0 {
		log.Println("✓ Agents already exist, skipping seed")
		return nil
	}

	defaultAgents := []struct {
		name   string
		title  string
		class  string
		level  int
		emoji  string
	}{
		{"Bran", "the Methodical", "Wizard", 12, "🧙"},
		{"Scáthach", "the Swift", "Ranger", 15, "🏹"},
		{"Aoife", "the Bold", "Warrior", 9, "⚔️"},
	}

	for _, agent := range defaultAgents {
		_, err := db.Exec(`
			INSERT INTO agents (name, title, class, level, xp, energy, status, avatar_emoji)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, agent.name, agent.title, agent.class, agent.level, agent.level*100, 100, "idle", agent.emoji)
		if err != nil {
			return fmt.Errorf("failed to seed agent %s: %w", agent.name, err)
		}
	}

	log.Println("✓ Seeded default agents")
	return nil
}
