package db

import (
	"fmt"
	"strings"
	"time"
)

// UpdateReputation updates an agent's reputation in a specific domain based on task completion
func (db *DB) UpdateReputation(agentID int, domain string, success bool) error {
	// Ensure the reputation record exists
	_, err := db.Exec(`
		INSERT INTO reputation (agent_id, domain, score, task_count, success_rate, last_update)
		VALUES ($1, $2, 100, 0, 1.0, CURRENT_TIMESTAMP)
		ON CONFLICT (agent_id, domain) DO NOTHING
	`, agentID, domain)
	if err != nil {
		return fmt.Errorf("failed to create reputation record: %w", err)
	}

	// Get current stats
	var currentScore, taskCount int
	var successRate float64
	err = db.QueryRow(`
		SELECT score, task_count, success_rate 
		FROM reputation 
		WHERE agent_id = $1 AND domain = $2
	`, agentID, domain).Scan(&currentScore, &taskCount, &successRate)
	if err != nil {
		return fmt.Errorf("failed to get current reputation: %w", err)
	}

	// Calculate new stats
	newTaskCount := taskCount + 1
	var newSuccessRate float64
	if success {
		newSuccessRate = (successRate*float64(taskCount) + 1.0) / float64(newTaskCount)
	} else {
		newSuccessRate = (successRate * float64(taskCount)) / float64(newTaskCount)
	}

	// Calculate score change
	scoreChange := 0
	if success {
		// Successful tasks increase reputation (more for harder tasks)
		scoreChange = 5 + int(10*newSuccessRate) // 5-15 points for success
	} else {
		// Failed tasks decrease reputation
		scoreChange = -10 - int(5*(1.0-newSuccessRate)) // -10 to -15 points for failure
	}

	newScore := currentScore + scoreChange
	if newScore < 0 {
		newScore = 0
	}
	if newScore > 1000 {
		newScore = 1000
	}

	// Update the record
	_, err = db.Exec(`
		UPDATE reputation 
		SET score = $1, task_count = $2, success_rate = $3, last_update = CURRENT_TIMESTAMP
		WHERE agent_id = $4 AND domain = $5
	`, newScore, newTaskCount, newSuccessRate, agentID, domain)
	if err != nil {
		return fmt.Errorf("failed to update reputation: %w", err)
	}

	return nil
}

// GetAgentReputation returns all reputation records for an agent
func (db *DB) GetAgentReputation(agentID int) ([]Reputation, error) {
	rows, err := db.Query(`
		SELECT id, agent_id, domain, score, task_count, success_rate, last_update
		FROM reputation
		WHERE agent_id = $1
		ORDER BY score DESC
	`, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query reputation: %w", err)
	}
	defer rows.Close()

	var reputations []Reputation
	for rows.Next() {
		var r Reputation
		err := rows.Scan(&r.ID, &r.AgentID, &r.Domain, &r.Score, &r.TaskCount, &r.SuccessRate, &r.LastUpdate)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reputation: %w", err)
		}
		reputations = append(reputations, r)
	}

	return reputations, nil
}

// InferTaskDomain attempts to categorize a task into a domain based on its name/description
func InferTaskDomain(taskName, taskDescription string) string {
	text := strings.ToLower(taskName + " " + taskDescription)
	
	// Define domain keywords
	domains := map[string][]string{
		"coding": {"code", "implement", "develop", "program", "bug", "fix", "feature", "api", "function", "class", "method"},
		"testing": {"test", "spec", "unit", "integration", "coverage", "assert", "mock", "validate", "verify"},
		"documentation": {"doc", "readme", "comment", "manual", "guide", "tutorial", "explain", "document"},
		"devops": {"deploy", "infrastructure", "docker", "kubernetes", "ci", "cd", "pipeline", "server", "cloud", "aws", "build"},
		"design": {"ui", "ux", "interface", "design", "layout", "style", "css", "theme", "visual", "mockup"},
		"security": {"security", "auth", "permission", "vulnerability", "encrypt", "ssl", "https", "secure"},
		"database": {"database", "sql", "query", "migration", "schema", "table", "index", "postgres", "mysql"},
		"maintenance": {"refactor", "cleanup", "optimize", "performance", "memory", "speed", "improve", "update"},
	}

	// Score each domain based on keyword matches
	domainScores := make(map[string]int)
	for domain, keywords := range domains {
		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				domainScores[domain]++
			}
		}
	}

	// Return the domain with the highest score, default to "general"
	maxScore := 0
	bestDomain := "general"
	for domain, score := range domainScores {
		if score > maxScore {
			maxScore = score
			bestDomain = domain
		}
	}

	return bestDomain
}

// GetTopReputationByDomain returns the top agents in each domain
func (db *DB) GetTopReputationByDomain(limit int) (map[string][]struct {
	Agent      Agent `json:"agent"`
	Reputation Reputation `json:"reputation"`
}, error) {
	query := `
		WITH ranked_agents AS (
			SELECT r.*, a.name, a.title, a.class, a.level, a.avatar_emoji,
			       ROW_NUMBER() OVER (PARTITION BY r.domain ORDER BY r.score DESC, r.task_count DESC) as rank
			FROM reputation r
			JOIN agents a ON r.agent_id = a.id
			WHERE r.task_count > 0  -- Only include agents who have completed tasks in this domain
		)
		SELECT agent_id, name, title, class, level, avatar_emoji,
		       id, domain, score, task_count, success_rate, last_update
		FROM ranked_agents
		WHERE rank <= $1
		ORDER BY domain, rank
	`
	
	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top reputation: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]struct {
		Agent      Agent `json:"agent"`
		Reputation Reputation `json:"reputation"`
	})

	for rows.Next() {
		var agent Agent
		var rep Reputation
		var lastUpdate time.Time
		
		err := rows.Scan(
			&agent.ID, &agent.Name, &agent.Title, &agent.Class, &agent.Level, &agent.AvatarEmoji,
			&rep.ID, &rep.Domain, &rep.Score, &rep.TaskCount, &rep.SuccessRate, &lastUpdate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reputation ranking: %w", err)
		}
		
		rep.AgentID = agent.ID
		rep.LastUpdate = lastUpdate

		entry := struct {
			Agent      Agent `json:"agent"`
			Reputation Reputation `json:"reputation"`
		}{
			Agent:      agent,
			Reputation: rep,
		}

		result[rep.Domain] = append(result[rep.Domain], entry)
	}

	return result, nil
}