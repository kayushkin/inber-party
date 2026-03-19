package db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// IntMap handles PostgreSQL JSONB serialization for category ratings
type IntMap map[string]int

func (im *IntMap) Scan(value interface{}) error {
	if value == nil {
		*im = nil
		return nil
	}
	
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into IntMap", value)
	}
	
	return json.Unmarshal(data, im)
}

func (im IntMap) Value() (driver.Value, error) {
	if im == nil {
		return nil, nil
	}
	return json.Marshal(im)
}

// CreateBountyRating creates a new rating for a completed bounty
func (db *DB) CreateBountyRating(rating *BountyRating) error {
	// Ensure the bounty is completed and the rater is the creator
	var bounty Bounty
	err := db.QueryRow(`
		SELECT creator_id, claimer_id, status 
		FROM bounties 
		WHERE id = $1
	`, rating.BountyID).Scan(&bounty.CreatorID, &bounty.ClaimerID, &bounty.Status)
	
	if err != nil {
		return fmt.Errorf("bounty not found: %w", err)
	}
	
	if bounty.Status != "completed" && bounty.Status != "paid" {
		return fmt.Errorf("cannot rate incomplete bounty")
	}
	
	if bounty.CreatorID != rating.RaterID {
		return fmt.Errorf("only bounty creator can rate the work")
	}
	
	if bounty.ClaimerID == nil || *bounty.ClaimerID != rating.RatedID {
		return fmt.Errorf("rated agent must be the bounty claimer")
	}
	
	// Check if rating already exists
	var existingID int
	err = db.QueryRow(`
		SELECT id FROM bounty_ratings 
		WHERE bounty_id = $1 AND rater_id = $2
	`, rating.BountyID, rating.RaterID).Scan(&existingID)
	
	if err == nil {
		return fmt.Errorf("bounty already rated")
	}
	
	// Create the rating
	categories := IntMap(rating.Categories)
	err = db.QueryRow(`
		INSERT INTO bounty_ratings (bounty_id, rater_id, rated_id, rating, comment, categories)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`, rating.BountyID, rating.RaterID, rating.RatedID, rating.Rating, rating.Comment, categories).Scan(
		&rating.ID, &rating.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create rating: %w", err)
	}
	
	// Update reputation based on rating
	err = db.updateReputationFromRating(rating.RatedID, rating.BountyID, rating.Rating, rating.Categories)
	if err != nil {
		// Log error but don't fail the rating creation
		fmt.Printf("Failed to update reputation from rating: %v\n", err)
	}
	
	return nil
}

// GetBountyRating returns the rating for a specific bounty
func (db *DB) GetBountyRating(bountyID int) (*BountyRating, error) {
	var rating BountyRating
	var categories IntMap
	
	err := db.QueryRow(`
		SELECT id, bounty_id, rater_id, rated_id, rating, comment, categories, created_at
		FROM bounty_ratings 
		WHERE bounty_id = $1
	`, bountyID).Scan(
		&rating.ID, &rating.BountyID, &rating.RaterID, &rating.RatedID,
		&rating.Rating, &rating.Comment, &categories, &rating.CreatedAt)
	
	if err != nil {
		return nil, err
	}
	
	rating.Categories = map[string]int(categories)
	return &rating, nil
}

// GetAgentRatings returns all ratings received by an agent
func (db *DB) GetAgentRatings(agentID int, limit, offset int) ([]BountyRating, error) {
	query := `
		SELECT r.id, r.bounty_id, r.rater_id, r.rated_id, r.rating, r.comment, r.categories, r.created_at,
		       b.title as bounty_title, rater.name as rater_name
		FROM bounty_ratings r
		JOIN bounties b ON r.bounty_id = b.id
		JOIN agents rater ON r.rater_id = rater.id
		WHERE r.rated_id = $1
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := db.Query(query, agentID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query agent ratings: %w", err)
	}
	defer rows.Close()
	
	var ratings []BountyRating
	for rows.Next() {
		var rating BountyRating
		var categories IntMap
		var bountyTitle, raterName string
		
		err := rows.Scan(
			&rating.ID, &rating.BountyID, &rating.RaterID, &rating.RatedID,
			&rating.Rating, &rating.Comment, &categories, &rating.CreatedAt,
			&bountyTitle, &raterName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rating: %w", err)
		}
		
		rating.Categories = map[string]int(categories)
		ratings = append(ratings, rating)
	}
	
	return ratings, nil
}

// GetAgentRatingStats returns aggregated rating statistics for an agent
func (db *DB) GetAgentRatingStats(agentID int) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Overall stats
	var ratingCount int
	var avgRating, minRating, maxRating float64
	err := db.QueryRow(`
		SELECT 
			COUNT(*) as rating_count,
			AVG(rating) as average_rating,
			MIN(rating) as min_rating,
			MAX(rating) as max_rating
		FROM bounty_ratings 
		WHERE rated_id = $1
	`, agentID).Scan(&ratingCount, &avgRating, &minRating, &maxRating)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get rating stats: %w", err)
	}
	
	stats["rating_count"] = ratingCount
	stats["average_rating"] = avgRating
	stats["min_rating"] = minRating
	stats["max_rating"] = maxRating
	
	if err != nil {
		return nil, fmt.Errorf("failed to get rating stats: %w", err)
	}
	
	// Rating distribution (1-5 stars)
	distribution := make(map[string]int)
	rows, err := db.Query(`
		SELECT rating, COUNT(*) 
		FROM bounty_ratings 
		WHERE rated_id = $1 
		GROUP BY rating 
		ORDER BY rating
	`, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating distribution: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var rating int
		var count int
		rows.Scan(&rating, &count)
		distribution[fmt.Sprintf("%d_star", rating)] = count
	}
	stats["distribution"] = distribution
	
	// Recent ratings (last 30 days)
	var recentCount int
	var recentAverage float64
	err = db.QueryRow(`
		SELECT 
			COUNT(*) as recent_count,
			COALESCE(AVG(rating), 0) as recent_average
		FROM bounty_ratings 
		WHERE rated_id = $1 AND created_at >= CURRENT_DATE - INTERVAL '30 days'
	`, agentID).Scan(&recentCount, &recentAverage)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get recent rating stats: %w", err)
	}
	
	stats["recent_count"] = recentCount
	stats["recent_average"] = recentAverage
	
	return stats, nil
}

// updateReputationFromRating updates agent's reputation based on a new rating
func (db *DB) updateReputationFromRating(agentID, bountyID int, rating int, categories map[string]int) error {
	// Get bounty details to determine domain
	var title, description string
	err := db.QueryRow(`
		SELECT title, description FROM bounties WHERE id = $1
	`, bountyID).Scan(&title, &description)
	if err != nil {
		return fmt.Errorf("failed to get bounty details: %w", err)
	}
	
	domain := InferTaskDomain(title, description)
	
	// Convert 1-5 rating to success/failure for reputation system
	// Ratings 4-5 are considered success, 1-3 are considered failure
	success := rating >= 4
	
	// Update basic reputation
	err = db.UpdateReputation(agentID, domain, success)
	if err != nil {
		return fmt.Errorf("failed to update reputation: %w", err)
	}
	
	// Apply rating-based reputation bonus/penalty
	// Higher ratings give additional reputation points
	scoreAdjustment := 0
	switch rating {
	case 5:
		scoreAdjustment = 15 // Excellent work
	case 4:
		scoreAdjustment = 5  // Good work
	case 3:
		scoreAdjustment = -2 // Average work
	case 2:
		scoreAdjustment = -8 // Poor work
	case 1:
		scoreAdjustment = -15 // Very poor work
	}
	
	if scoreAdjustment != 0 {
		_, err = db.Exec(`
			UPDATE reputation 
			SET score = GREATEST(0, LEAST(1000, score + $1)), last_update = CURRENT_TIMESTAMP
			WHERE agent_id = $2 AND domain = $3
		`, scoreAdjustment, agentID, domain)
		if err != nil {
			return fmt.Errorf("failed to apply rating adjustment: %w", err)
		}
	}
	
	return nil
}

// GetHighestRatedAgents returns agents with the best ratings, optionally filtered by domain
func (db *DB) GetHighestRatedAgents(domain string, limit int) ([]struct {
	Agent Agent `json:"agent"`
	Stats map[string]interface{} `json:"rating_stats"`
}, error) {
	var whereClause string
	var args []interface{}
	argCount := 1
	
	// If domain is specified, only consider agents who have reputation in that domain
	if domain != "" {
		whereClause = `
			AND EXISTS (
				SELECT 1 FROM reputation r 
				WHERE r.agent_id = a.id AND r.domain = $` + fmt.Sprintf("%d", argCount) + `
			)
		`
		args = append(args, domain)
		argCount++
	}
	
	query := `
		SELECT 
			a.id, a.name, a.title, a.class, a.level, a.xp, a.gold, a.energy, 
			a.status, a.avatar_emoji, a.mood, a.mood_score, a.workload, 
			a.last_active, a.created_at, a.updated_at,
			COUNT(br.id) as rating_count,
			AVG(br.rating) as avg_rating
		FROM agents a
		LEFT JOIN bounty_ratings br ON a.id = br.rated_id
		WHERE 1=1 ` + whereClause + `
		GROUP BY a.id, a.name, a.title, a.class, a.level, a.xp, a.gold, a.energy,
		         a.status, a.avatar_emoji, a.mood, a.mood_score, a.workload,
		         a.last_active, a.created_at, a.updated_at
		HAVING COUNT(br.id) > 0  -- Only agents with ratings
		ORDER BY AVG(br.rating) DESC, COUNT(br.id) DESC
		LIMIT $` + fmt.Sprintf("%d", argCount)
	
	args = append(args, limit)
	
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query highest rated agents: %w", err)
	}
	defer rows.Close()
	
	var results []struct {
		Agent Agent `json:"agent"`
		Stats map[string]interface{} `json:"rating_stats"`
	}
	
	for rows.Next() {
		var agent Agent
		var ratingCount int
		var avgRating float64
		
		err := rows.Scan(
			&agent.ID, &agent.Name, &agent.Title, &agent.Class, &agent.Level,
			&agent.XP, &agent.Gold, &agent.Energy, &agent.Status, &agent.AvatarEmoji,
			&agent.Mood, &agent.MoodScore, &agent.Workload, &agent.LastActive,
			&agent.CreatedAt, &agent.UpdatedAt, &ratingCount, &avgRating,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}
		
		stats := map[string]interface{}{
			"rating_count":   ratingCount,
			"average_rating": avgRating,
		}
		
		results = append(results, struct {
			Agent Agent `json:"agent"`
			Stats map[string]interface{} `json:"rating_stats"`
		}{
			Agent: agent,
			Stats: stats,
		})
	}
	
	return results, nil
}