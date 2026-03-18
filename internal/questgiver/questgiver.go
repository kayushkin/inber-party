package questgiver

import (
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/kayushkin/inber-party/internal/db"
	"github.com/kayushkin/inber-party/internal/inber"
)

// QuestGiver is an AI agent that reviews and assigns tasks to appropriate agents
type QuestGiver struct {
	db          *db.DB
	inberSource inber.DataSource
}

// TaskAnalysis contains the AI's analysis of a task
type TaskAnalysis struct {
	TaskType        string   `json:"task_type"`        // "coding", "debugging", "deployment", "documentation", etc.
	RequiredSkills  []string `json:"required_skills"`  // Skills needed for the task
	Complexity      int      `json:"complexity"`       // 1-5 scale
	EstimatedXP     int      `json:"estimated_xp"`     // Expected XP reward
	RequiredClass   string   `json:"required_class"`   // Best agent class for this task
	SuggestedAgent  string   `json:"suggested_agent"`  // Specific agent recommendation
	Reasoning       string   `json:"reasoning"`        // Why this assignment was chosen
	CanBeShared     bool     `json:"can_be_shared"`    // Whether multiple agents can work on this
	Priority        string   `json:"priority"`         // "low", "medium", "high", "urgent"
}

// AgentCapability represents an agent's suitability for a task
type AgentCapability struct {
	AgentID         string  `json:"agent_id"`
	AgentName       string  `json:"agent_name"`
	MatchScore      float64 `json:"match_score"`      // 0-100 compatibility score
	Availability    float64 `json:"availability"`     // 0-100 based on current workload
	SkillMatch      float64 `json:"skill_match"`      // How well skills align
	ClassBonus      float64 `json:"class_bonus"`      // Bonus for matching class
	ExperienceLevel float64 `json:"experience_level"` // Agent's experience level
	RecentPerform   float64 `json:"recent_performance"` // Recent success rate
}

// NewQuestGiver creates a new quest-giver instance
func NewQuestGiver(database *db.DB, inberSource inber.DataSource) *QuestGiver {
	return &QuestGiver{
		db:          database,
		inberSource: inberSource,
	}
}

// AnalyzeTask uses AI-like heuristics to analyze a task and determine requirements
func (qg *QuestGiver) AnalyzeTask(taskName, taskDescription string) *TaskAnalysis {
	analysis := &TaskAnalysis{
		RequiredSkills: []string{},
		Complexity:     1,
		EstimatedXP:    10,
		Priority:       "medium",
		CanBeShared:    false,
	}

	// Combine task name and description for analysis
	combined := strings.ToLower(taskName + " " + taskDescription)
	
	// Determine task type based on keywords
	switch {
	case containsAny(combined, []string{"bug", "fix", "error", "debug", "broken"}):
		analysis.TaskType = "debugging"
		analysis.RequiredClass = "Engineer"
		analysis.RequiredSkills = append(analysis.RequiredSkills, "Debugging", "Problem Solving")
		analysis.Complexity = 3
		analysis.Priority = "high"
		
	case containsAny(combined, []string{"build", "create", "implement", "develop", "code"}):
		analysis.TaskType = "development"
		analysis.RequiredClass = "Engineer"
		analysis.RequiredSkills = append(analysis.RequiredSkills, "Programming", "System Design")
		analysis.Complexity = 4
		analysis.CanBeShared = true
		
	case containsAny(combined, []string{"deploy", "release", "publish", "ship"}):
		analysis.TaskType = "deployment"
		analysis.RequiredClass = "Smith"
		analysis.RequiredSkills = append(analysis.RequiredSkills, "Deployment", "Infrastructure")
		analysis.Complexity = 3
		analysis.Priority = "high"
		
	case containsAny(combined, []string{"test", "verify", "validate", "check"}):
		analysis.TaskType = "testing"
		analysis.RequiredClass = "Gladiator"
		analysis.RequiredSkills = append(analysis.RequiredSkills, "Testing", "Quality Assurance")
		analysis.Complexity = 2
		
	case containsAny(combined, []string{"document", "write", "explain", "guide"}):
		analysis.TaskType = "documentation"
		analysis.RequiredClass = "Scribe"
		analysis.RequiredSkills = append(analysis.RequiredSkills, "Writing", "Documentation")
		analysis.Complexity = 2
		analysis.CanBeShared = true
		
	case containsAny(combined, []string{"refactor", "clean", "optimize", "improve"}):
		analysis.TaskType = "optimization"
		analysis.RequiredClass = "Artificer"
		analysis.RequiredSkills = append(analysis.RequiredSkills, "Code Quality", "Optimization")
		analysis.Complexity = 3
		
	case containsAny(combined, []string{"design", "architecture", "plan", "structure"}):
		analysis.TaskType = "design"
		analysis.RequiredClass = "Sage"
		analysis.RequiredSkills = append(analysis.RequiredSkills, "Architecture", "Design Patterns")
		analysis.Complexity = 4
		analysis.Priority = "medium"
		
	case containsAny(combined, []string{"monitor", "watch", "alert", "health"}):
		analysis.TaskType = "monitoring"
		analysis.RequiredClass = "Sentinel"
		analysis.RequiredSkills = append(analysis.RequiredSkills, "Monitoring", "System Health")
		analysis.Complexity = 2
		
	default:
		analysis.TaskType = "general"
		analysis.RequiredClass = "Adventurer"
		analysis.RequiredSkills = append(analysis.RequiredSkills, "General Skills")
		analysis.Complexity = 2
	}
	
	// Adjust complexity based on description length and keywords
	if len(taskDescription) > 500 {
		analysis.Complexity = min(5, analysis.Complexity+1)
	}
	
	if containsAny(combined, []string{"urgent", "critical", "emergency", "asap"}) {
		analysis.Priority = "urgent"
		analysis.Complexity = min(5, analysis.Complexity+1)
	}
	
	if containsAny(combined, []string{"simple", "quick", "easy", "small"}) {
		analysis.Complexity = max(1, analysis.Complexity-1)
	}
	
	if containsAny(combined, []string{"complex", "difficult", "large", "major"}) {
		analysis.Complexity = min(5, analysis.Complexity+2)
	}
	
	// Calculate estimated XP based on complexity
	analysis.EstimatedXP = analysis.Complexity * 20
	
	return analysis
}

// FindBestAgent evaluates all available agents and finds the best match for a task
func (qg *QuestGiver) FindBestAgent(analysis *TaskAnalysis) (*AgentCapability, error) {
	// Get agents from inber source
	agents, err := qg.inberSource.GetAgents()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents: %w", err)
	}
	
	if len(agents) == 0 {
		return nil, fmt.Errorf("no agents available")
	}
	
	capabilities := make([]AgentCapability, 0, len(agents))
	
	for _, agent := range agents {
		capability := qg.evaluateAgent(agent, analysis)
		capabilities = append(capabilities, capability)
	}
	
	// Find the best match
	var bestMatch *AgentCapability
	var bestScore float64
	
	for i := range capabilities {
		cap := &capabilities[i]
		// Overall score combines match score and availability
		overallScore := (cap.MatchScore * 0.7) + (cap.Availability * 0.3)
		
		if overallScore > bestScore {
			bestScore = overallScore
			bestMatch = cap
		}
	}
	
	if bestMatch != nil {
		analysis.SuggestedAgent = bestMatch.AgentID
		analysis.Reasoning = fmt.Sprintf(
			"Selected %s (match: %.1f%%, available: %.1f%%) - %s with strong %s skills",
			bestMatch.AgentName, bestMatch.MatchScore, bestMatch.Availability,
			classDescription(bestMatch.AgentID), skillDescription(analysis.RequiredSkills),
		)
	}
	
	return bestMatch, nil
}

// evaluateAgent calculates how suitable an agent is for a specific task
func (qg *QuestGiver) evaluateAgent(agent inber.RPGAgent, analysis *TaskAnalysis) AgentCapability {
	capability := AgentCapability{
		AgentID:      agent.ID,
		AgentName:    agent.Name,
		MatchScore:   0,
		Availability: 100, // Default to full availability
	}
	
	// Calculate skill match
	skillMatch := 0.0
	for _, skill := range agent.Skills {
		for _, requiredSkill := range analysis.RequiredSkills {
			if strings.Contains(strings.ToLower(skill.Name), strings.ToLower(requiredSkill)) {
				skillMatch += float64(skill.Level) * 10
			}
		}
	}
	capability.SkillMatch = math.Min(100, skillMatch)
	
	// Calculate class bonus
	if strings.EqualFold(agent.Class, analysis.RequiredClass) {
		capability.ClassBonus = 25.0
	} else if isCompatibleClass(agent.Class, analysis.RequiredClass) {
		capability.ClassBonus = 10.0
	}
	
	// Calculate experience level based on agent level and XP
	capability.ExperienceLevel = math.Min(100, float64(agent.Level*10))
	
	// Calculate recent performance based on quest success rate
	if agent.QuestCount > 0 && agent.ErrorCount >= 0 {
		successRate := float64(agent.QuestCount-agent.ErrorCount) / float64(agent.QuestCount)
		capability.RecentPerform = math.Max(0, successRate*100)
	} else {
		capability.RecentPerform = 80 // Default for new agents
	}
	
	// Reduce availability based on current status
	switch agent.Status {
	case "working":
		capability.Availability = 30
	case "stuck":
		capability.Availability = 10
	case "resting":
		capability.Availability = 50
	default: // idle
		capability.Availability = 100
	}
	
	// Reduce availability based on energy level
	if agent.Energy < agent.MaxEnergy {
		energyPenalty := 100 - float64(agent.Energy)/float64(agent.MaxEnergy)*100
		capability.Availability -= energyPenalty * 0.5
	}
	
	// Calculate overall match score
	capability.MatchScore = (capability.SkillMatch*0.4 + 
		capability.ClassBonus*0.3 + 
		capability.ExperienceLevel*0.2 + 
		capability.RecentPerform*0.1)
	
	capability.MatchScore = math.Max(0, math.Min(100, capability.MatchScore))
	capability.Availability = math.Max(0, math.Min(100, capability.Availability))
	
	return capability
}

// AutoAssignTask automatically analyzes a task and assigns it to the best agent
func (qg *QuestGiver) AutoAssignTask(taskID int, taskName, taskDescription string) error {
	// Analyze the task
	analysis := qg.AnalyzeTask(taskName, taskDescription)
	
	// Find the best agent
	bestAgent, err := qg.FindBestAgent(analysis)
	if err != nil {
		return fmt.Errorf("failed to find suitable agent: %w", err)
	}
	
	if bestAgent == nil {
		return fmt.Errorf("no suitable agent found for task")
	}
	
	// Assign the task in the database
	if qg.db != nil {
		err = qg.assignTaskToAgent(taskID, bestAgent.AgentID, analysis)
		if err != nil {
			return fmt.Errorf("failed to assign task to agent: %w", err)
		}
	}
	
	log.Printf("Quest Giver: Assigned task '%s' to agent '%s' (match: %.1f%%)", 
		taskName, bestAgent.AgentName, bestAgent.MatchScore)
	
	return nil
}

// assignTaskToAgent updates the database to assign a task to an agent
func (qg *QuestGiver) assignTaskToAgent(taskID int, agentID string, analysis *TaskAnalysis) error {
	// For now, we'll update the task with a comment about the assignment
	// In a full implementation, we'd need to map inber agent IDs to database agent IDs
	
	// Update task with quest-giver analysis
	_, err := qg.db.Exec(`
		UPDATE tasks 
		SET status = 'assigned', 
		    progress = 5,
		    xp_reward = ?
		WHERE id = ?
	`, analysis.EstimatedXP, taskID)
	
	if err != nil {
		return err
	}
	
	// Log the assignment decision
	log.Printf("Quest Giver Analysis - Task: %s, Agent: %s, Reasoning: %s", 
		analysis.TaskType, agentID, analysis.Reasoning)
	
	return nil
}

// GetTaskRecommendations returns a list of recommended task assignments
func (qg *QuestGiver) GetTaskRecommendations() ([]TaskAssignment, error) {
	if qg.db == nil {
		return []TaskAssignment{}, nil
	}
	
	// Get unassigned tasks
	rows, err := qg.db.Query(`
		SELECT id, name, description
		FROM tasks 
		WHERE status = 'available' AND assigned_agent_id IS NULL
		ORDER BY created_at ASC
		LIMIT 10
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var recommendations []TaskAssignment
	
	for rows.Next() {
		var taskID int
		var name, description string
		if err := rows.Scan(&taskID, &name, &description); err != nil {
			continue
		}
		
		// Analyze the task
		analysis := qg.AnalyzeTask(name, description)
		
		// Find the best agent
		bestAgent, err := qg.FindBestAgent(analysis)
		if err != nil {
			continue
		}
		
		assignment := TaskAssignment{
			TaskID:       taskID,
			TaskName:     name,
			TaskDesc:     description,
			Analysis:     *analysis,
			BestAgent:    bestAgent,
		}
		
		recommendations = append(recommendations, assignment)
	}
	
	return recommendations, nil
}

// TaskAssignment represents a recommended task assignment
type TaskAssignment struct {
	TaskID    int              `json:"task_id"`
	TaskName  string           `json:"task_name"`
	TaskDesc  string           `json:"task_description"`
	Analysis  TaskAnalysis     `json:"analysis"`
	BestAgent *AgentCapability `json:"best_agent"`
}

// Helper functions
func containsAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func isCompatibleClass(agentClass, requiredClass string) bool {
	compatibility := map[string][]string{
		"Engineer":   {"Artificer", "Smith"},
		"Artificer":  {"Engineer", "Smith"},
		"Smith":      {"Engineer", "Artificer"},
		"Scribe":     {"Sage", "Bard"},
		"Sage":       {"Scribe", "Bard"},
		"Bard":       {"Scribe", "Sage"},
		"Ranger":     {"Scout", "Sentinel"},
		"Scout":      {"Ranger", "Sentinel"},
		"Sentinel":   {"Ranger", "Scout"},
	}
	
	compatible, exists := compatibility[requiredClass]
	if !exists {
		return false
	}
	
	for _, compat := range compatible {
		if strings.EqualFold(agentClass, compat) {
			return true
		}
	}
	return false
}

func classDescription(agentID string) string {
	// Default description - could be enhanced to use agent class later
	return "skilled adventurer"
}

func skillDescription(skills []string) string {
	if len(skills) == 0 {
		return "general"
	}
	if len(skills) == 1 {
		return strings.ToLower(skills[0])
	}
	return fmt.Sprintf("%s and %s", strings.ToLower(skills[0]), strings.ToLower(skills[1]))
}