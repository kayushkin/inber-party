package db

import (
	"fmt"
	"log"
)

// AnalyzeAndUpdateRelationships analyzes collaboration patterns and updates relationships
func (db *DB) AnalyzeAndUpdateRelationships() error {
	// Get all pairs of agents that have worked together or competed
	agentPairs, err := db.getAgentInteractionPairs()
	if err != nil {
		return fmt.Errorf("failed to get agent pairs: %v", err)
	}
	
	for _, pair := range agentPairs {
		err := db.updateRelationshipForPair(pair.Agent1ID, pair.Agent2ID)
		if err != nil {
			log.Printf("Failed to update relationship for agents %d and %d: %v", pair.Agent1ID, pair.Agent2ID, err)
		}
	}
	
	return nil
}

// AgentPair represents a pair of agents who have interacted
type AgentPair struct {
	Agent1ID int
	Agent2ID int
}

// getAgentInteractionPairs finds all pairs of agents that have worked together or competed
func (db *DB) getAgentInteractionPairs() ([]AgentPair, error) {
	// Find pairs from party memberships (worked together)
	query := `
		SELECT DISTINCT 
			LEAST(pm1.agent_id, pm2.agent_id) as agent1_id,
			GREATEST(pm1.agent_id, pm2.agent_id) as agent2_id
		FROM party_members pm1
		JOIN party_members pm2 ON pm1.party_id = pm2.party_id 
		WHERE pm1.agent_id != pm2.agent_id
		
		UNION
		
		-- Find pairs from tasks assigned to same party (worked together)
		SELECT DISTINCT 
			LEAST(pm1.agent_id, pm2.agent_id) as agent1_id,
			GREATEST(pm1.agent_id, pm2.agent_id) as agent2_id
		FROM tasks t
		JOIN party_members pm1 ON t.assigned_party_id = pm1.party_id
		JOIN party_members pm2 ON t.assigned_party_id = pm2.party_id
		WHERE pm1.agent_id != pm2.agent_id
		AND t.status IN ('completed', 'failed')
	`
	
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var pairs []AgentPair
	for rows.Next() {
		var pair AgentPair
		err := rows.Scan(&pair.Agent1ID, &pair.Agent2ID)
		if err != nil {
			continue
		}
		pairs = append(pairs, pair)
	}
	
	return pairs, nil
}

// updateRelationshipForPair analyzes and updates the relationship between two agents
func (db *DB) updateRelationshipForPair(agent1ID, agent2ID int) error {
	// Ensure agent1ID is always smaller for consistent storage
	if agent1ID > agent2ID {
		agent1ID, agent2ID = agent2ID, agent1ID
	}
	
	// Calculate relationship metrics
	collabCount, successfulCollabs, err := db.getCollaborationStats(agent1ID, agent2ID)
	if err != nil {
		return err
	}
	
	competitionCount, err := db.getCompetitionStats(agent1ID, agent2ID)
	if err != nil {
		return err
	}
	
	// Determine relationship type and strength
	relationshipType, strength := db.calculateRelationshipType(collabCount, successfulCollabs, competitionCount)
	
	// Update or insert relationship
	return db.upsertRelationship(agent1ID, agent2ID, relationshipType, strength, collabCount, successfulCollabs, competitionCount)
}

// getCollaborationStats calculates how many times agents worked together and succeeded
func (db *DB) getCollaborationStats(agent1ID, agent2ID int) (collabCount, successfulCollabs int, err error) {
	// Count party collaborations
	query := `
		SELECT COUNT(DISTINCT t.id) as total_tasks,
			   COUNT(DISTINCT CASE WHEN t.status = 'completed' THEN t.id END) as successful_tasks
		FROM tasks t
		JOIN party_members pm1 ON t.assigned_party_id = pm1.party_id
		JOIN party_members pm2 ON t.assigned_party_id = pm2.party_id
		WHERE pm1.agent_id = $1 AND pm2.agent_id = $2
		AND t.status IN ('completed', 'failed')
	`
	
	err = db.QueryRow(query, agent1ID, agent2ID).Scan(&collabCount, &successfulCollabs)
	if err != nil {
		return 0, 0, err
	}
	
	return collabCount, successfulCollabs, nil
}

// getCompetitionStats calculates how many times agents competed for similar tasks
func (db *DB) getCompetitionStats(agent1ID, agent2ID int) (competitionCount int, err error) {
	// For now, we'll approximate competition as agents working on tasks with similar names/descriptions
	// within a similar time period (could indicate they were considered for the same work)
	// This is a simplified approach - in reality you might track bidding/assignment patterns
	
	query := `
		SELECT COUNT(*)
		FROM tasks t1
		JOIN tasks t2 ON (
			t1.name = t2.name OR 
			(t1.description IS NOT NULL AND t2.description IS NOT NULL AND 
			 SIMILARITY(t1.description, t2.description) > 0.5)
		)
		WHERE t1.assigned_agent_id = $1 
		AND t2.assigned_agent_id = $2
		AND ABS(EXTRACT(EPOCH FROM t1.created_at) - EXTRACT(EPOCH FROM t2.created_at)) < 86400 -- within 24 hours
		AND t1.id != t2.id
	`
	
	err = db.QueryRow(query, agent1ID, agent2ID).Scan(&competitionCount)
	if err != nil {
		// If similarity function doesn't exist, fall back to simpler approach
		if err.Error() == "function similarity(text, text) does not exist" {
			query = `
				SELECT COUNT(*)
				FROM tasks t1
				JOIN tasks t2 ON t1.name = t2.name
				WHERE t1.assigned_agent_id = $1 
				AND t2.assigned_agent_id = $2
				AND ABS(EXTRACT(EPOCH FROM t1.created_at) - EXTRACT(EPOCH FROM t2.created_at)) < 86400
				AND t1.id != t2.id
			`
			err = db.QueryRow(query, agent1ID, agent2ID).Scan(&competitionCount)
		}
	}
	
	return competitionCount, err
}

// calculateRelationshipType determines friendship/rivalry based on collaboration patterns
func (db *DB) calculateRelationshipType(collabCount, successfulCollabs, competitionCount int) (relationshipType string, strength int) {
	if collabCount == 0 && competitionCount == 0 {
		return "neutral", 0
	}
	
	// Calculate success rate for collaborations
	successRate := 0.0
	if collabCount > 0 {
		successRate = float64(successfulCollabs) / float64(collabCount)
	}
	
	// Decision logic for relationship type
	if collabCount > competitionCount {
		// More collaboration than competition
		if successRate >= 0.7 {
			// High success rate = strong friendship
			strength = min(100, 30 + collabCount*10 + int(successRate*20))
			return "friendship", strength
		} else if successRate >= 0.4 {
			// Moderate success rate = weak friendship
			strength = min(100, 10 + collabCount*5 + int(successRate*10))
			return "friendship", strength
		} else {
			// Low success rate despite collaboration = developing rivalry
			strength = min(100, 20 + competitionCount*8)
			return "rivalry", strength
		}
	} else if competitionCount > collabCount {
		// More competition than collaboration = rivalry
		strength = min(100, 25 + competitionCount*10)
		return "rivalry", strength
	} else {
		// Equal collaboration and competition = neutral with some history
		strength = min(50, 10 + (collabCount+competitionCount)*5)
		return "neutral", strength
	}
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// upsertRelationship inserts or updates a relationship record
func (db *DB) upsertRelationship(agent1ID, agent2ID int, relationshipType string, strength, collabCount, successfulCollabs, competitionCount int) error {
	query := `
		INSERT INTO agent_relationships (
			agent1_id, agent2_id, relationship_type, strength, 
			collaboration_count, successful_collabs, competition_count,
			last_interaction, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW(), NOW())
		ON CONFLICT (agent1_id, agent2_id)
		DO UPDATE SET
			relationship_type = $3,
			strength = $4,
			collaboration_count = $5,
			successful_collabs = $6,
			competition_count = $7,
			last_interaction = NOW(),
			updated_at = NOW()
	`
	
	_, err := db.Exec(query, agent1ID, agent2ID, relationshipType, strength, collabCount, successfulCollabs, competitionCount)
	return err
}

// GetAgentRelationships retrieves all relationships for a specific agent
func (db *DB) GetAgentRelationships(agentID int) ([]AgentRelationshipDetail, error) {
	query := `
		SELECT 
			ar.id, ar.agent1_id, ar.agent2_id, ar.relationship_type, ar.strength,
			ar.collaboration_count, ar.successful_collabs, ar.competition_count,
			ar.last_interaction, ar.created_at, ar.updated_at,
			a1.id, a1.name, a1.title, a1.class, a1.level, a1.avatar_emoji,
			a2.id, a2.name, a2.title, a2.class, a2.level, a2.avatar_emoji
		FROM agent_relationships ar
		JOIN agents a1 ON ar.agent1_id = a1.id
		JOIN agents a2 ON ar.agent2_id = a2.id
		WHERE ar.agent1_id = $1 OR ar.agent2_id = $1
		ORDER BY ar.strength DESC
	`
	
	rows, err := db.Query(query, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var relationships []AgentRelationshipDetail
	for rows.Next() {
		var rel AgentRelationshipDetail
		var agent1, agent2 Agent
		
		err := rows.Scan(
			&rel.ID, &rel.Agent1ID, &rel.Agent2ID, &rel.RelationshipType, &rel.Strength,
			&rel.CollaborationCount, &rel.SuccessfulCollabs, &rel.CompetitionCount,
			&rel.LastInteraction, &rel.CreatedAt, &rel.UpdatedAt,
			&agent1.ID, &agent1.Name, &agent1.Title, &agent1.Class, &agent1.Level, &agent1.AvatarEmoji,
			&agent2.ID, &agent2.Name, &agent2.Title, &agent2.Class, &agent2.Level, &agent2.AvatarEmoji,
		)
		if err != nil {
			continue
		}
		
		// Ensure the target agent is always represented as the "other" agent
		if agent1.ID == agentID {
			rel.Agent1 = agent1
			rel.Agent2 = agent2
		} else {
			rel.Agent1 = agent2
			rel.Agent2 = agent1
		}
		
		relationships = append(relationships, rel)
	}
	
	return relationships, nil
}

// GetRelationshipStats provides aggregated relationship data for an agent
func (db *DB) GetRelationshipStats(agentID int) (*RelationshipStats, error) {
	relationships, err := db.GetAgentRelationships(agentID)
	if err != nil {
		return nil, err
	}
	
	stats := &RelationshipStats{
		AgentID:       agentID,
		Relationships: relationships,
	}
	
	var bestFriend, biggestRival *AgentRelationshipDetail
	maxFriendship, maxRivalry := 0, 0
	
	for i := range relationships {
		rel := &relationships[i]
		switch rel.RelationshipType {
		case "friendship":
			stats.TotalFriends++
			if rel.Strength > maxFriendship {
				maxFriendship = rel.Strength
				bestFriend = rel
			}
		case "rivalry":
			stats.TotalRivals++
			if rel.Strength > maxRivalry {
				maxRivalry = rel.Strength
				biggestRival = rel
			}
		}
	}
	
	stats.BestFriend = bestFriend
	stats.BiggestRival = biggestRival
	
	return stats, nil
}

// GetAllRelationships retrieves all relationships in the system
func (db *DB) GetAllRelationships() ([]AgentRelationshipDetail, error) {
	query := `
		SELECT 
			ar.id, ar.agent1_id, ar.agent2_id, ar.relationship_type, ar.strength,
			ar.collaboration_count, ar.successful_collabs, ar.competition_count,
			ar.last_interaction, ar.created_at, ar.updated_at,
			a1.id, a1.name, a1.title, a1.class, a1.level, a1.avatar_emoji,
			a2.id, a2.name, a2.title, a2.class, a2.level, a2.avatar_emoji
		FROM agent_relationships ar
		JOIN agents a1 ON ar.agent1_id = a1.id
		JOIN agents a2 ON ar.agent2_id = a2.id
		WHERE ar.strength > 0
		ORDER BY ar.strength DESC
	`
	
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var relationships []AgentRelationshipDetail
	for rows.Next() {
		var rel AgentRelationshipDetail
		
		err := rows.Scan(
			&rel.ID, &rel.Agent1ID, &rel.Agent2ID, &rel.RelationshipType, &rel.Strength,
			&rel.CollaborationCount, &rel.SuccessfulCollabs, &rel.CompetitionCount,
			&rel.LastInteraction, &rel.CreatedAt, &rel.UpdatedAt,
			&rel.Agent1.ID, &rel.Agent1.Name, &rel.Agent1.Title, &rel.Agent1.Class, &rel.Agent1.Level, &rel.Agent1.AvatarEmoji,
			&rel.Agent2.ID, &rel.Agent2.Name, &rel.Agent2.Title, &rel.Agent2.Class, &rel.Agent2.Level, &rel.Agent2.AvatarEmoji,
		)
		if err != nil {
			continue
		}
		
		relationships = append(relationships, rel)
	}
	
	return relationships, nil
}