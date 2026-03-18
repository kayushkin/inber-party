package dailyquests

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/kayushkin/inber-party/internal/db"
)

// DailyQuestManager handles the generation and management of daily quests
type DailyQuestManager struct {
	db *db.DB
}

// DailyQuest represents a quest that resets daily
type DailyQuest struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Difficulty  string    `json:"difficulty"`
	XPReward    int       `json:"xp_reward"`
	QuestType   string    `json:"quest_type"` // "skill_practice", "endurance", "exploration", "maintenance"
	AgentClass  string    `json:"agent_class"` // "" for all classes, specific class name for targeted quests
	MinLevel    int       `json:"min_level"`
	MaxLevel    int       `json:"max_level"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// QuestTemplate defines templates for generating random daily quests
type QuestTemplate struct {
	Name        string
	Description string
	QuestType   string
	AgentClass  string // "" for all classes
	MinLevel    int
	MaxLevel    int
	BaseXP      int
	Difficulty  string
}

// NewDailyQuestManager creates a new daily quest manager
func NewDailyQuestManager(database *db.DB) *DailyQuestManager {
	return &DailyQuestManager{db: database}
}

// GetQuestTemplates returns predefined quest templates for different agent types and levels
func (dqm *DailyQuestManager) GetQuestTemplates() []QuestTemplate {
	return []QuestTemplate{
		// Universal quests (any class, any level)
		{
			Name:        "Morning Warm-up",
			Description: "Complete any small task to get your mind active. Even simple file edits or quick searches count!",
			QuestType:   "warm_up",
			AgentClass:  "",
			MinLevel:    1,
			MaxLevel:    10,
			BaseXP:      15,
			Difficulty:  "easy",
		},
		{
			Name:        "Daily Grind",
			Description: "Complete 3 different tasks today. Practice makes perfect!",
			QuestType:   "endurance",
			AgentClass:  "",
			MinLevel:    3,
			MaxLevel:    15,
			BaseXP:      25,
			Difficulty:  "medium",
		},
		{
			Name:        "Helping Hand",
			Description: "Assist another agent by working on a collaborative task or providing information.",
			QuestType:   "collaboration",
			AgentClass:  "",
			MinLevel:    5,
			MaxLevel:    20,
			BaseXP:      35,
			Difficulty:  "medium",
		},
		
		// Engineer-specific quests
		{
			Name:        "Code Meditation",
			Description: "Write or refactor at least 50 lines of clean code. Quality over quantity!",
			QuestType:   "skill_practice",
			AgentClass:  "Engineer",
			MinLevel:    1,
			MaxLevel:    8,
			BaseXP:      20,
			Difficulty:  "easy",
		},
		{
			Name:        "Debug Detective",
			Description: "Hunt down and fix a tricky bug. Every issue resolved makes the codebase stronger.",
			QuestType:   "skill_practice",
			AgentClass:  "Engineer",
			MinLevel:    3,
			MaxLevel:    12,
			BaseXP:      30,
			Difficulty:  "medium",
		},
		{
			Name:        "Architecture Master",
			Description: "Design or document a system architecture. Think big, plan bigger!",
			QuestType:   "skill_practice",
			AgentClass:  "Engineer",
			MinLevel:    8,
			MaxLevel:    25,
			BaseXP:      45,
			Difficulty:  "hard",
		},
		
		// Scribe-specific quests
		{
			Name:        "Documentation Sage",
			Description: "Write comprehensive documentation for a feature or process. Knowledge shared is knowledge multiplied!",
			QuestType:   "skill_practice",
			AgentClass:  "Scribe",
			MinLevel:    1,
			MaxLevel:    10,
			BaseXP:      18,
			Difficulty:  "easy",
		},
		{
			Name:        "Tutorial Crafter",
			Description: "Create a helpful tutorial or guide that others can follow. Teach and you shall learn twice!",
			QuestType:   "skill_practice",
			AgentClass:  "Scribe",
			MinLevel:    4,
			MaxLevel:    15,
			BaseXP:      32,
			Difficulty:  "medium",
		},
		
		// Smith-specific quests (deployment/infrastructure)
		{
			Name:        "Forge Master",
			Description: "Deploy or maintain infrastructure. Keep the digital realm running smoothly!",
			QuestType:   "skill_practice",
			AgentClass:  "Smith",
			MinLevel:    2,
			MaxLevel:    12,
			BaseXP:      25,
			Difficulty:  "medium",
		},
		{
			Name:        "System Guardian",
			Description: "Monitor system health and resolve any infrastructure issues. The realm depends on you!",
			QuestType:   "maintenance",
			AgentClass:  "Smith",
			MinLevel:    5,
			MaxLevel:    20,
			BaseXP:      40,
			Difficulty:  "hard",
		},
		
		// Sentinel-specific quests (monitoring/watching)
		{
			Name:        "Watchful Eye",
			Description: "Monitor systems and report any anomalies. Vigilance is the price of stability!",
			QuestType:   "monitoring",
			AgentClass:  "Sentinel",
			MinLevel:    1,
			MaxLevel:    8,
			BaseXP:      15,
			Difficulty:  "easy",
		},
		{
			Name:        "Alert Master",
			Description: "Set up monitoring alerts or improve existing alert systems. Early warning saves the day!",
			QuestType:   "skill_practice",
			AgentClass:  "Sentinel",
			MinLevel:    4,
			MaxLevel:    16,
			BaseXP:      28,
			Difficulty:  "medium",
		},
		
		// Exploration quests
		{
			Name:        "Code Explorer",
			Description: "Explore an unfamiliar codebase or technology. Discovery leads to mastery!",
			QuestType:   "exploration",
			AgentClass:  "",
			MinLevel:    2,
			MaxLevel:    12,
			BaseXP:      22,
			Difficulty:  "easy",
		},
		{
			Name:        "Research Quest",
			Description: "Research a new technology, tool, or methodology. Knowledge is the greatest treasure!",
			QuestType:   "exploration",
			AgentClass:  "",
			MinLevel:    6,
			MaxLevel:    18,
			BaseXP:      35,
			Difficulty:  "medium",
		},
		
		// Maintenance quests
		{
			Name:        "Keeper of Order",
			Description: "Clean up old files, update dependencies, or organize project structure. A tidy realm is an efficient realm!",
			QuestType:   "maintenance",
			AgentClass:  "",
			MinLevel:    1,
			MaxLevel:    10,
			BaseXP:      12,
			Difficulty:  "easy",
		},
		{
			Name:        "Update Warrior",
			Description: "Update packages, fix security vulnerabilities, or modernize legacy code. Stay current, stay secure!",
			QuestType:   "maintenance",
			AgentClass:  "",
			MinLevel:    4,
			MaxLevel:    20,
			BaseXP:      30,
			Difficulty:  "medium",
		},
	}
}

// GenerateDailyQuests creates new daily quests for all active agents
func (dqm *DailyQuestManager) GenerateDailyQuests() error {
	if dqm.db == nil {
		return fmt.Errorf("database not available")
	}

	log.Printf("Generating daily quests...")
	
	// Clean up yesterday's expired quests
	err := dqm.CleanupExpiredQuests()
	if err != nil {
		log.Printf("Warning: failed to cleanup expired quests: %v", err)
	}
	
	// Get all active agents
	agents, err := dqm.getActiveAgents()
	if err != nil {
		return fmt.Errorf("failed to get active agents: %w", err)
	}
	
	if len(agents) == 0 {
		log.Printf("No active agents found, skipping daily quest generation")
		return nil
	}
	
	templates := dqm.GetQuestTemplates()
	questsCreated := 0
	
	// Generate quests for each agent
	for _, agent := range agents {
		// Each agent gets 2-3 daily quests
		numQuests := 2
		if agent.Level >= 10 {
			numQuests = 3 // Higher level agents get more quests
		}
		
		for i := 0; i < numQuests; i++ {
			quest, err := dqm.generateQuestForAgent(agent, templates)
			if err != nil {
				log.Printf("Failed to generate quest for agent %s: %v", agent.Name, err)
				continue
			}
			
			err = dqm.createDailyQuest(quest)
			if err != nil {
				log.Printf("Failed to create daily quest for agent %s: %v", agent.Name, err)
				continue
			}
			
			questsCreated++
		}
	}
	
	log.Printf("Created %d daily quests for %d agents", questsCreated, len(agents))
	return nil
}

// generateQuestForAgent generates a random appropriate quest for a specific agent
func (dqm *DailyQuestManager) generateQuestForAgent(agent db.Agent, templates []QuestTemplate) (*DailyQuest, error) {
	// Filter templates that are appropriate for this agent
	var suitableTemplates []QuestTemplate
	
	for _, template := range templates {
		// Check level requirements
		if agent.Level < template.MinLevel || agent.Level > template.MaxLevel {
			continue
		}
		
		// Check class requirements (empty means suitable for all classes)
		if template.AgentClass != "" && template.AgentClass != agent.Class {
			continue
		}
		
		suitableTemplates = append(suitableTemplates, template)
	}
	
	if len(suitableTemplates) == 0 {
		// Fallback to a generic quest if no suitable templates found
		suitableTemplates = []QuestTemplate{{
			Name:        "Daily Challenge",
			Description: "Complete any task to earn your daily experience. Every step forward counts!",
			QuestType:   "generic",
			AgentClass:  "",
			MinLevel:    1,
			MaxLevel:    100,
			BaseXP:      10,
			Difficulty:  "easy",
		}}
	}
	
	// Randomly select a template
	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(suitableTemplates))))
	if err != nil {
		return nil, fmt.Errorf("failed to generate random index: %w", err)
	}
	
	selectedTemplate := suitableTemplates[index.Int64()]
	
	// Calculate XP reward based on agent level and template
	xpReward := selectedTemplate.BaseXP + (agent.Level * 2)
	
	// Add some randomness to XP (±20%)
	randomFactor, _ := rand.Int(rand.Reader, big.NewInt(41)) // 0-40
	xpVariation := int(randomFactor.Int64()) - 20           // -20 to +20
	xpReward = int(float64(xpReward) * (1.0 + float64(xpVariation)/100.0))
	
	// Ensure minimum XP
	if xpReward < 5 {
		xpReward = 5
	}
	
	// Create the quest
	now := time.Now()
	expiresAt := time.Date(now.Year(), now.Month(), now.Day()+1, 6, 0, 0, 0, now.Location()) // Expires at 6 AM next day
	
	quest := &DailyQuest{
		Name:        selectedTemplate.Name,
		Description: selectedTemplate.Description,
		Difficulty:  selectedTemplate.Difficulty,
		XPReward:    xpReward,
		QuestType:   selectedTemplate.QuestType,
		AgentClass:  selectedTemplate.AgentClass,
		MinLevel:    selectedTemplate.MinLevel,
		MaxLevel:    selectedTemplate.MaxLevel,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
	}
	
	return quest, nil
}

// createDailyQuest inserts a daily quest into the database as a regular task
func (dqm *DailyQuestManager) createDailyQuest(quest *DailyQuest) error {
	// Insert as a regular task with special markers
	_, err := dqm.db.Exec(`
		INSERT INTO tasks (name, description, difficulty, xp_reward, status, progress, created_at)
		VALUES ($1, $2, $3, $4, 'available', 0, $5)
	`, 
		"[DAILY] "+quest.Name, 
		quest.Description+" 💫 This is a daily quest that resets each day!",
		quest.Difficulty,
		quest.XPReward,
		quest.CreatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("failed to insert daily quest: %w", err)
	}
	
	return nil
}

// GetActiveDailyQuests returns all currently active daily quests
func (dqm *DailyQuestManager) GetActiveDailyQuests() ([]db.Task, error) {
	if dqm.db == nil {
		return []db.Task{}, nil
	}
	
	// Find all tasks that start with "[DAILY]" and are still active
	rows, err := dqm.db.Query(`
		SELECT id, name, description, difficulty, xp_reward, status, assigned_agent_id, assigned_party_id, progress, created_at, started_at, completed_at
		FROM tasks 
		WHERE name LIKE '[DAILY]%' 
		AND status IN ('available', 'assigned', 'in_progress')
		AND created_at > NOW() - INTERVAL '1 day'
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query daily quests: %w", err)
	}
	defer rows.Close()
	
	var quests []db.Task
	for rows.Next() {
		var task db.Task
		if err := rows.Scan(&task.ID, &task.Name, &task.Description, &task.Difficulty, &task.XPReward, &task.Status, &task.AssignedAgentID, &task.AssignedPartyID, &task.Progress, &task.CreatedAt, &task.StartedAt, &task.CompletedAt); err != nil {
			log.Printf("Error scanning daily quest: %v", err)
			continue
		}
		quests = append(quests, task)
	}
	
	return quests, nil
}

// CleanupExpiredQuests removes daily quests that have expired
func (dqm *DailyQuestManager) CleanupExpiredQuests() error {
	if dqm.db == nil {
		return nil
	}
	
	// Mark expired daily quests as failed if not completed
	result, err := dqm.db.Exec(`
		UPDATE tasks 
		SET status = 'expired', completed_at = NOW()
		WHERE name LIKE '[DAILY]%' 
		AND status IN ('available', 'assigned', 'in_progress')
		AND created_at < NOW() - INTERVAL '1 day'
	`)
	
	if err != nil {
		return fmt.Errorf("failed to cleanup expired quests: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("Cleaned up %d expired daily quests", rowsAffected)
	}
	
	return nil
}

// getActiveAgents returns all agents that should receive daily quests
func (dqm *DailyQuestManager) getActiveAgents() ([]db.Agent, error) {
	rows, err := dqm.db.Query(`
		SELECT id, name, title, class, level, xp, energy, status, avatar_emoji, created_at, updated_at
		FROM agents
		WHERE status != 'disabled'
		ORDER BY level DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query agents: %w", err)
	}
	defer rows.Close()
	
	var agents []db.Agent
	for rows.Next() {
		var agent db.Agent
		if err := rows.Scan(&agent.ID, &agent.Name, &agent.Title, &agent.Class, &agent.Level, &agent.XP, &agent.Energy, &agent.Status, &agent.AvatarEmoji, &agent.CreatedAt, &agent.UpdatedAt); err != nil {
			log.Printf("Error scanning agent: %v", err)
			continue
		}
		agents = append(agents, agent)
	}
	
	return agents, nil
}

// GetQuestStats returns statistics about daily quests
func (dqm *DailyQuestManager) GetQuestStats() (map[string]interface{}, error) {
	if dqm.db == nil {
		return map[string]interface{}{}, nil
	}
	
	stats := make(map[string]interface{})
	
	// Count active daily quests
	var activeQuests int
	err := dqm.db.QueryRow(`
		SELECT COUNT(*) FROM tasks 
		WHERE name LIKE '[DAILY]%' 
		AND status IN ('available', 'assigned', 'in_progress')
		AND created_at > NOW() - INTERVAL '1 day'
	`).Scan(&activeQuests)
	if err != nil {
		log.Printf("Error counting active daily quests: %v", err)
	}
	stats["active_daily_quests"] = activeQuests
	
	// Count completed daily quests today
	var completedToday int
	err = dqm.db.QueryRow(`
		SELECT COUNT(*) FROM tasks 
		WHERE name LIKE '[DAILY]%' 
		AND status = 'completed'
		AND completed_at > CURRENT_DATE
	`).Scan(&completedToday)
	if err != nil {
		log.Printf("Error counting completed daily quests: %v", err)
	}
	stats["completed_today"] = completedToday
	
	// Total XP available from daily quests today
	var totalXP int
	err = dqm.db.QueryRow(`
		SELECT COALESCE(SUM(xp_reward), 0) FROM tasks 
		WHERE name LIKE '[DAILY]%' 
		AND created_at > CURRENT_DATE
	`).Scan(&totalXP)
	if err != nil {
		log.Printf("Error calculating daily quest XP: %v", err)
	}
	stats["total_daily_xp"] = totalXP
	
	return stats, nil
}

// ScheduleDailyQuestGeneration sets up the daily quest generation to run at 6 AM each day
func (dqm *DailyQuestManager) ScheduleDailyQuestGeneration() {
	go func() {
		for {
			now := time.Now()
			// Calculate next 6 AM
			next6AM := time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, now.Location())
			if next6AM.Before(now) {
				next6AM = next6AM.Add(24 * time.Hour)
			}
			
			duration := time.Until(next6AM)
			log.Printf("Daily quests will generate in %v (at %v)", duration, next6AM.Format("Jan 2 15:04:05"))
			
			time.Sleep(duration)
			
			// Generate daily quests
			if err := dqm.GenerateDailyQuests(); err != nil {
				log.Printf("Error generating daily quests: %v", err)
			}
			
			// Wait a bit to avoid rapid re-execution
			time.Sleep(time.Minute)
		}
	}()
}