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
		`CREATE TABLE IF NOT EXISTS parties (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			leader_id INTEGER REFERENCES agents(id) ON DELETE CASCADE,
			status VARCHAR(50) DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS party_members (
			id SERIAL PRIMARY KEY,
			party_id INTEGER REFERENCES parties(id) ON DELETE CASCADE,
			agent_id INTEGER REFERENCES agents(id) ON DELETE CASCADE,
			role VARCHAR(50) DEFAULT 'member',
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(party_id, agent_id)
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			difficulty INTEGER DEFAULT 1,
			xp_reward INTEGER DEFAULT 10,
			status VARCHAR(50) DEFAULT 'available',
			assigned_agent_id INTEGER REFERENCES agents(id),
			assigned_party_id INTEGER REFERENCES parties(id),
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
		// Add mood/morale system fields to agents table
		// Add missing gold column
		`ALTER TABLE agents ADD COLUMN IF NOT EXISTS gold INTEGER DEFAULT 0`,
		// Add mood/morale system fields to agents table
		`ALTER TABLE agents ADD COLUMN IF NOT EXISTS mood VARCHAR(20) DEFAULT 'neutral'`,
		`ALTER TABLE agents ADD COLUMN IF NOT EXISTS mood_score INTEGER DEFAULT 75`,
		`ALTER TABLE agents ADD COLUMN IF NOT EXISTS workload INTEGER DEFAULT 0`,
		`ALTER TABLE agents ADD COLUMN IF NOT EXISTS last_active TIMESTAMP`,
		// Add unique constraint on agent names for sync
		`CREATE UNIQUE INDEX IF NOT EXISTS agents_name_unique ON agents(name)`,
		// Add reputation system
		`CREATE TABLE IF NOT EXISTS reputation (
			id SERIAL PRIMARY KEY,
			agent_id INTEGER REFERENCES agents(id) ON DELETE CASCADE,
			domain VARCHAR(100) NOT NULL,
			score INTEGER DEFAULT 100,
			task_count INTEGER DEFAULT 0,
			success_rate DECIMAL(5,4) DEFAULT 1.0,
			last_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(agent_id, domain)
		)`,
		// Add gold/currency system
		`ALTER TABLE agents ADD COLUMN IF NOT EXISTS gold INTEGER DEFAULT 0`,
		`ALTER TABLE tasks ADD COLUMN IF NOT EXISTS gold_reward INTEGER DEFAULT 0`,
		// Add cost tracking system
		`CREATE TABLE IF NOT EXISTS cost_entries (
			id SERIAL PRIMARY KEY,
			agent_id INTEGER REFERENCES agents(id) ON DELETE CASCADE,
			task_id INTEGER REFERENCES tasks(id) ON DELETE CASCADE,
			session_id VARCHAR(255) NOT NULL,
			tokens_used INTEGER NOT NULL DEFAULT 0,
			cost_usd DECIMAL(10,6) NOT NULL DEFAULT 0,
			model_name VARCHAR(100) NOT NULL,
			operation_type VARCHAR(50) NOT NULL DEFAULT 'task',
			date DATE NOT NULL DEFAULT CURRENT_DATE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			metadata JSONB
		)`,
		`CREATE INDEX IF NOT EXISTS cost_entries_agent_date_idx ON cost_entries(agent_id, date)`,
		`CREATE INDEX IF NOT EXISTS cost_entries_date_idx ON cost_entries(date)`,
		`CREATE INDEX IF NOT EXISTS cost_entries_session_idx ON cost_entries(session_id)`,
		// Add bounty marketplace system
		`CREATE TABLE IF NOT EXISTS bounties (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			description TEXT NOT NULL,
			requirements TEXT NOT NULL DEFAULT '',
			payout_amount INTEGER NOT NULL DEFAULT 0,
			status VARCHAR(50) NOT NULL DEFAULT 'open',
			deadline TIMESTAMP,
			creator_id INTEGER NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
			claimer_id INTEGER REFERENCES agents(id) ON DELETE SET NULL,
			work_submission TEXT,
			verification_notes TEXT,
			required_skills JSONB DEFAULT '[]',
			tier VARCHAR(20) NOT NULL DEFAULT 'bronze',
			claimed_at TIMESTAMP,
			submitted_at TIMESTAMP,
			completed_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS bounties_status_idx ON bounties(status)`,
		`CREATE INDEX IF NOT EXISTS bounties_creator_id_idx ON bounties(creator_id)`,
		`CREATE INDEX IF NOT EXISTS bounties_claimer_id_idx ON bounties(claimer_id)`,
		`CREATE INDEX IF NOT EXISTS bounties_tier_idx ON bounties(tier)`,
		`CREATE INDEX IF NOT EXISTS bounties_created_at_idx ON bounties(created_at DESC)`,
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
		startingGold := agent.level * 50 // Give agents starting gold based on their level
		_, err := db.Exec(`
			INSERT INTO agents (name, title, class, level, xp, gold, energy, status, avatar_emoji)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, agent.name, agent.title, agent.class, agent.level, agent.level*100, startingGold, 100, "idle", agent.emoji)
		if err != nil {
			return fmt.Errorf("failed to seed agent %s: %w", agent.name, err)
		}
	}

	log.Println("✓ Seeded default agents")
	return nil
}

// NewAgentSyncService creates a new agent sync service for this database
func (db *DB) NewAgentSyncService() *AgentSyncService {
	return NewAgentSyncService(db)
}

// GetAgentByID returns an agent by ID
func (db *DB) GetAgentByID(id int) (*Agent, error) {
	var agent Agent
	err := db.QueryRow(`
		SELECT id, name, title, class, level, xp, gold, energy, status, avatar_emoji, 
		       mood, mood_score, workload, last_active, created_at, updated_at
		FROM agents WHERE id = $1
	`, id).Scan(&agent.ID, &agent.Name, &agent.Title, &agent.Class, &agent.Level, 
		&agent.XP, &agent.Gold, &agent.Energy, &agent.Status, &agent.AvatarEmoji,
		&agent.Mood, &agent.MoodScore, &agent.Workload, &agent.LastActive, 
		&agent.CreatedAt, &agent.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	
	return &agent, nil
}
