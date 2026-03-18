package mood

import (
	"database/sql"
	"log"
	"math"
	"time"

	"github.com/kayushkin/inber-party/internal/db"
)

// MoodLevel represents different emotional states
type MoodLevel struct {
	Name        string `json:"name"`
	ScoreRange  [2]int `json:"score_range"`
	Description string `json:"description"`
	Emoji       string `json:"emoji"`
}

var MoodLevels = []MoodLevel{
	{"exhausted", [2]int{0, 20}, "Completely overwhelmed and burnt out", "😫"},
	{"stressed", [2]int{21, 40}, "Under pressure, making errors", "😰"},
	{"neutral", [2]int{41, 60}, "Stable but unremarkable", "😐"},
	{"content", [2]int{61, 80}, "Working well, good performance", "😊"},
	{"happy", [2]int{81, 100}, "Excellent performance, high energy", "😄"},
}

// MoodCalculator handles mood calculations for agents
type MoodCalculator struct {
	database *db.DB
}

func NewMoodCalculator(database *db.DB) *MoodCalculator {
	return &MoodCalculator{database: database}
}

// CalculateAgentMood computes mood based on recent performance
func (mc *MoodCalculator) CalculateAgentMood(agentID int) (string, int, error) {
	if mc.database == nil {
		return "neutral", 75, nil // Default mood when no DB
	}

	// Get recent task performance (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	
	var completedTasks, failedTasks, activeTasks int
	var lastActive sql.NullTime
	var avgTaskDuration sql.NullFloat64

	// Count task outcomes
	err := mc.database.QueryRow(`
		SELECT 
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed,
			COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as active,
			MAX(completed_at) as last_active,
			AVG(EXTRACT(EPOCH FROM (COALESCE(completed_at, NOW()) - started_at))/3600) as avg_hours
		FROM tasks 
		WHERE assigned_agent_id = $1 AND created_at > $2
	`, agentID, thirtyDaysAgo).Scan(&completedTasks, &failedTasks, &activeTasks, &lastActive, &avgTaskDuration)
	
	if err != nil {
		log.Printf("Error calculating mood for agent %d: %v", agentID, err)
		return "neutral", 75, err
	}

	// Calculate mood components
	
	// 1. Error Rate (0-40 points)
	totalTasks := completedTasks + failedTasks
	var errorRateScore int = 40 // Default high score
	if totalTasks > 0 {
		successRate := float64(completedTasks) / float64(totalTasks)
		errorRateScore = int(successRate * 40)
	}

	// 2. Workload (0-30 points) - inverse relationship
	workloadScore := 30
	if activeTasks > 5 {
		workloadScore = int(math.Max(0, 30-float64(activeTasks-5)*3))
	} else if activeTasks == 0 {
		workloadScore = 25 // Slightly lower for being idle
	}

	// 3. Rest Time (0-30 points)
	restScore := 30
	if lastActive.Valid {
		hoursSinceLastTask := time.Since(lastActive.Time).Hours()
		if hoursSinceLastTask < 1 {
			restScore = 10 // Very recent work, might be tired
		} else if hoursSinceLastTask < 6 {
			restScore = 20 // Recent work
		} else if hoursSinceLastTask < 24 {
			restScore = 30 // Good rest period
		} else if hoursSinceLastTask > 168 { // More than a week
			restScore = 20 // Might be rusty
		}
	} else {
		restScore = 25 // No recent activity, neutral
	}

	// Calculate final mood score
	moodScore := errorRateScore + workloadScore + restScore
	moodScore = int(math.Max(0, math.Min(100, float64(moodScore))))

	// Determine mood level
	moodName := "neutral"
	for _, level := range MoodLevels {
		if moodScore >= level.ScoreRange[0] && moodScore <= level.ScoreRange[1] {
			moodName = level.Name
			break
		}
	}

	log.Printf("Agent %d mood: %s (%d) - Error: %d, Workload: %d, Rest: %d", 
		agentID, moodName, moodScore, errorRateScore, workloadScore, restScore)

	return moodName, moodScore, nil
}

// UpdateAllAgentMoods recalculates mood for all agents
func (mc *MoodCalculator) UpdateAllAgentMoods() error {
	if mc.database == nil {
		return nil
	}

	rows, err := mc.database.Query("SELECT id FROM agents")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var agentID int
		if err := rows.Scan(&agentID); err != nil {
			log.Printf("Error scanning agent ID: %v", err)
			continue
		}

		mood, score, err := mc.CalculateAgentMood(agentID)
		if err != nil {
			log.Printf("Error calculating mood for agent %d: %v", agentID, err)
			continue
		}

		// Count current active tasks for workload
		var workload int
		mc.database.QueryRow("SELECT COUNT(*) FROM tasks WHERE assigned_agent_id = $1 AND status = 'in_progress'", agentID).Scan(&workload)

		// Update agent mood in database
		_, err = mc.database.Exec(`
			UPDATE agents 
			SET mood = $1, mood_score = $2, workload = $3, updated_at = NOW()
			WHERE id = $4
		`, mood, score, workload, agentID)
		
		if err != nil {
			log.Printf("Error updating mood for agent %d: %v", agentID, err)
		}
	}

	log.Println("✓ Updated all agent moods")
	return nil
}

// GetMoodLevel returns the mood level info for a given score
func GetMoodLevel(score int) MoodLevel {
	for _, level := range MoodLevels {
		if score >= level.ScoreRange[0] && score <= level.ScoreRange[1] {
			return level
		}
	}
	return MoodLevels[2] // Return neutral as fallback
}

// UpdateAgentLastActive updates the last_active timestamp for an agent
func (mc *MoodCalculator) UpdateAgentLastActive(agentID int) error {
	if mc.database == nil {
		return nil
	}

	_, err := mc.database.Exec(`
		UPDATE agents 
		SET last_active = NOW(), updated_at = NOW()
		WHERE id = $1
	`, agentID)
	
	return err
}