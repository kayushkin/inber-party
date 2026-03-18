package db

import (
	"fmt"
	"log"
	"time"
)

// AgentSyncService handles syncing agents from external sources (inber) to the local database
type AgentSyncService struct {
	db *DB
}

// NewAgentSyncService creates a new agent sync service
func NewAgentSyncService(db *DB) *AgentSyncService {
	return &AgentSyncService{db: db}
}

// InberAgent represents an agent from the inber system
type InberAgent struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Title        string     `json:"title"`
	Class        string     `json:"class"`
	Level        int        `json:"level"`
	XP           int        `json:"xp"`
	Energy       int        `json:"energy"`
	Status       string     `json:"status"`
	AvatarEmoji  string     `json:"avatar_emoji"`
	TotalTokens  int        `json:"total_tokens"`
	TotalCost    float64    `json:"total_cost"`
	SessionCount int        `json:"session_count"`
	QuestCount   int        `json:"quest_count"`
	ErrorCount   int        `json:"error_count"`
	LastActive   *time.Time `json:"last_active"`
}

// SyncAgentsFromInber syncs agents from inber data to the local database
func (s *AgentSyncService) SyncAgentsFromInber(inberAgents []InberAgent) (int, int, error) {
	var created, updated int

	for _, agent := range inberAgents {
		exists, err := s.agentExists(agent.ID)
		if err != nil {
			log.Printf("Error checking if agent %s exists: %v", agent.ID, err)
			continue
		}

		if exists {
			if err := s.updateAgent(agent); err != nil {
				log.Printf("Error updating agent %s: %v", agent.ID, err)
				continue
			}
			updated++
		} else {
			if err := s.createAgent(agent); err != nil {
				log.Printf("Error creating agent %s: %v", agent.ID, err)
				continue
			}
			created++
		}
	}

	log.Printf("Agent sync complete: %d created, %d updated", created, updated)
	return created, updated, nil
}

// agentExists checks if an agent with the given ID (name) already exists in the database
func (s *AgentSyncService) agentExists(agentName string) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM agents WHERE name = $1", agentName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check agent existence: %w", err)
	}
	return count > 0, nil
}

// createAgent creates a new agent in the database
func (s *AgentSyncService) createAgent(agent InberAgent) error {
	_, err := s.db.Exec(`
		INSERT INTO agents (
			name, title, class, level, xp, energy, status, avatar_emoji, 
			mood, mood_score, workload, last_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, 
		agent.Name, agent.Title, agent.Class, agent.Level, agent.XP, agent.Energy, 
		agent.Status, agent.AvatarEmoji, calculateMood(agent), calculateMoodScore(agent),
		calculateWorkload(agent), agent.LastActive,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create agent %s: %w", agent.Name, err)
	}

	log.Printf("✓ Created new agent: %s (%s)", agent.Name, agent.Class)
	return nil
}

// updateAgent updates an existing agent in the database
func (s *AgentSyncService) updateAgent(agent InberAgent) error {
	_, err := s.db.Exec(`
		UPDATE agents SET 
			title = $1, class = $2, level = $3, xp = $4, energy = $5, 
			status = $6, avatar_emoji = $7, mood = $8, mood_score = $9, 
			workload = $10, last_active = $11, updated_at = CURRENT_TIMESTAMP
		WHERE name = $12
	`,
		agent.Title, agent.Class, agent.Level, agent.XP, agent.Energy,
		agent.Status, agent.AvatarEmoji, calculateMood(agent), calculateMoodScore(agent),
		calculateWorkload(agent), agent.LastActive, agent.Name,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update agent %s: %w", agent.Name, err)
	}

	return nil
}

// GetManagedAgents returns all agents from the local database
func (s *AgentSyncService) GetManagedAgents() ([]Agent, error) {
	rows, err := s.db.Query(`
		SELECT id, name, title, class, level, xp, energy, status, avatar_emoji, 
		       mood, mood_score, workload, last_active, created_at, updated_at
		FROM agents 
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query managed agents: %w", err)
	}
	defer rows.Close()

	var agents []Agent
	for rows.Next() {
		var agent Agent
		err := rows.Scan(
			&agent.ID, &agent.Name, &agent.Title, &agent.Class, &agent.Level, &agent.XP,
			&agent.Energy, &agent.Status, &agent.AvatarEmoji, &agent.Mood, &agent.MoodScore,
			&agent.Workload, &agent.LastActive, &agent.CreatedAt, &agent.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

// GetAgentByName returns a specific agent by name
func (s *AgentSyncService) GetAgentByName(name string) (*Agent, error) {
	var agent Agent
	err := s.db.QueryRow(`
		SELECT id, name, title, class, level, xp, energy, status, avatar_emoji,
		       mood, mood_score, workload, last_active, created_at, updated_at
		FROM agents 
		WHERE name = $1
	`, name).Scan(
		&agent.ID, &agent.Name, &agent.Title, &agent.Class, &agent.Level, &agent.XP,
		&agent.Energy, &agent.Status, &agent.AvatarEmoji, &agent.Mood, &agent.MoodScore,
		&agent.Workload, &agent.LastActive, &agent.CreatedAt, &agent.UpdatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get agent %s: %w", name, err)
	}
	
	return &agent, nil
}

// MarkAgentsAsInactive marks agents that haven't been seen in recent syncs as inactive
func (s *AgentSyncService) MarkAgentsAsInactive(activeAgentNames []string) error {
	if len(activeAgentNames) == 0 {
		return nil
	}

	// Build placeholders for the IN clause
	placeholders := ""
	args := make([]interface{}, len(activeAgentNames))
	for i, name := range activeAgentNames {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += fmt.Sprintf("$%d", i+1)
		args[i] = name
	}

	query := fmt.Sprintf(`
		UPDATE agents 
		SET status = 'inactive', updated_at = CURRENT_TIMESTAMP 
		WHERE name NOT IN (%s) AND status != 'inactive'
	`, placeholders)

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to mark agents as inactive: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("Marked %d agents as inactive", rowsAffected)
	}

	return nil
}

// Helper functions to calculate mood and workload based on inber data

func calculateMood(agent InberAgent) string {
	moodScore := calculateMoodScore(agent)
	
	if moodScore >= 80 {
		return "happy"
	} else if moodScore >= 60 {
		return "content"
	} else if moodScore >= 40 {
		return "neutral"
	} else if moodScore >= 20 {
		return "stressed"
	}
	return "exhausted"
}

func calculateMoodScore(agent InberAgent) int {
	score := 75 // base mood
	
	// Factor in error rate
	if agent.QuestCount > 0 {
		errorRate := float64(agent.ErrorCount) / float64(agent.QuestCount)
		score -= int(errorRate * 30) // up to -30 for high error rate
	}
	
	// Factor in recent activity (being active is good, but too much is stressful)
	if agent.LastActive != nil {
		hoursAgo := time.Since(*agent.LastActive).Hours()
		if hoursAgo < 1 {
			score += 5 // recently active is good
		} else if hoursAgo > 24 {
			score -= 10 // been idle too long
		}
	}
	
	// Clamp between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	
	return score
}

func calculateWorkload(agent InberAgent) int {
	// Approximate current workload based on status
	switch agent.Status {
	case "working":
		return 5 // high workload
	case "idle":
		return 0 // no workload
	case "stuck":
		return 3 // some workload, but stuck
	default:
		return 1 // minimal workload
	}
}