package questgiver

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/kayushkin/inber-party/internal/db"
	"github.com/kayushkin/inber-party/internal/inber"
)

// QuestGiver is an AI agent that reviews and assigns tasks to appropriate agents
type QuestGiver struct {
	db          *db.DB
	inberSource inber.DataSource
	nameGen     *QuestNameGenerator
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
	EpicQuestName   string   `json:"epic_quest_name"`  // Generated immersive quest name
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
		nameGen:     NewQuestNameGenerator(),
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
	
	// Generate epic quest name based on the analysis
	analysis.EpicQuestName = qg.nameGen.GenerateQuestName(analysis, taskName)
	
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

// ── Procedural Quest Name Generation ──────────────────────

// SeasonalEvent represents a seasonal event that can affect quest names
type SeasonalEvent struct {
	Name        string
	StartMonth  int
	StartDay    int
	EndMonth    int
	EndDay      int
	Prefixes    []string
	Suffixes    []string
	Themes      []string
	Decorations []string
}

// QuestNameGenerator generates immersive, contextual quest names based on task content
type QuestNameGenerator struct {
	// Name templates organized by task type and theme
	nameTemplates map[string][]QuestNameTemplate
	// Seasonal events that can modify quest names
	seasonalEvents []SeasonalEvent
}

// QuestNameTemplate represents a template for generating quest names
type QuestNameTemplate struct {
	Pattern     string   // Name pattern with placeholders
	Themes      []string // Thematic words to use in placeholders
	Descriptors []string // Descriptive adjectives
}

// NewQuestNameGenerator creates a new quest name generator with predefined templates
func NewQuestNameGenerator() *QuestNameGenerator {
	return &QuestNameGenerator{
		nameTemplates:  initQuestNameTemplates(),
		seasonalEvents: initSeasonalEvents(),
	}
}

// GenerateQuestName creates an immersive quest name based on task analysis
func (qng *QuestNameGenerator) GenerateQuestName(analysis *TaskAnalysis, originalName string) string {
	var questName string
	
	// If we have templates for this task type, use them
	if templates, exists := qng.nameTemplates[analysis.TaskType]; exists {
		questName = qng.generateFromTemplate(templates, analysis, originalName)
	} else {
		// Fall back to generic epic name generation
		questName = qng.generateGenericEpicName(analysis, originalName)
	}
	
	// Apply seasonal modifications if there's an active seasonal event
	return qng.applySeasonalModifications(questName, analysis)
}

// applySeasonalModifications enhances quest names with seasonal themes
func (qng *QuestNameGenerator) applySeasonalModifications(questName string, analysis *TaskAnalysis) string {
	activeEvent := qng.getActiveSeasonalEvent()
	if activeEvent == nil {
		return questName
	}
	
	// Apply seasonal prefix (25% chance)
	if len(activeEvent.Prefixes) > 0 && analysis.Complexity >= 2 {
		prefix := activeEvent.Prefixes[len(analysis.RequiredSkills)%len(activeEvent.Prefixes)]
		questName = prefix + " " + questName
	}
	
	// Apply seasonal suffix (20% chance)
	if len(activeEvent.Suffixes) > 0 && analysis.Complexity >= 3 {
		suffix := activeEvent.Suffixes[analysis.EstimatedXP%len(activeEvent.Suffixes)]
		questName = questName + " " + suffix
	}
	
	// Replace generic themes with seasonal themes (30% chance)
	if len(activeEvent.Themes) > 0 && len(analysis.RequiredSkills) > 0 {
		for _, theme := range activeEvent.Themes {
			// Replace common words with seasonal equivalents
			questName = strings.ReplaceAll(questName, "Great", theme)
			questName = strings.ReplaceAll(questName, "Epic", theme)
			questName = strings.ReplaceAll(questName, "Mighty", theme)
		}
	}
	
	return questName
}

// getActiveSeasonalEvent returns the currently active seasonal event, if any
func (qng *QuestNameGenerator) getActiveSeasonalEvent() *SeasonalEvent {
	now := time.Now()
	currentMonth := int(now.Month())
	currentDay := now.Day()
	
	for _, event := range qng.seasonalEvents {
		if qng.isEventActive(event, currentMonth, currentDay) {
			return &event
		}
	}
	
	return nil
}

// isEventActive checks if a seasonal event is currently active
func (qng *QuestNameGenerator) isEventActive(event SeasonalEvent, month, day int) bool {
	// Handle events that span across year boundary (e.g., Christmas to New Year)
	if event.StartMonth > event.EndMonth {
		return (month > event.StartMonth || month < event.EndMonth) ||
			(month == event.StartMonth && day >= event.StartDay) ||
			(month == event.EndMonth && day <= event.EndDay)
	}
	
	// Handle events within the same year
	if month < event.StartMonth || month > event.EndMonth {
		return false
	}
	
	if month == event.StartMonth && day < event.StartDay {
		return false
	}
	
	if month == event.EndMonth && day > event.EndDay {
		return false
	}
	
	return true
}

// generateFromTemplate generates a name using task-specific templates
func (qng *QuestNameGenerator) generateFromTemplate(templates []QuestNameTemplate, analysis *TaskAnalysis, originalName string) string {
	// Pick a random template
	template := templates[len(templates)%3] // Use modulo for deterministic selection
	
	// Extract key terms from the original name for context
	keyTerms := extractKeyTerms(originalName)
	
	// Replace placeholders in the pattern
	name := template.Pattern
	
	// Replace {object} with a key term from the original task
	if len(keyTerms) > 0 {
		name = strings.ReplaceAll(name, "{object}", keyTerms[0])
	} else {
		name = strings.ReplaceAll(name, "{object}", "Code")
	}
	
	// Replace {theme} with a thematic word
	if len(template.Themes) > 0 {
		theme := template.Themes[len(analysis.RequiredSkills)%len(template.Themes)]
		name = strings.ReplaceAll(name, "{theme}", theme)
	}
	
	// Replace {descriptor} with an adjective
	if len(template.Descriptors) > 0 {
		descriptor := template.Descriptors[analysis.Complexity%len(template.Descriptors)]
		name = strings.ReplaceAll(name, "{descriptor}", descriptor)
	}
	
	// Replace {difficulty} with complexity-based adjectives
	difficultyWords := []string{"Simple", "Challenging", "Complex", "Heroic", "Legendary"}
	if analysis.Complexity >= 1 && analysis.Complexity <= 5 {
		name = strings.ReplaceAll(name, "{difficulty}", difficultyWords[analysis.Complexity-1])
	}
	
	return name
}

// generateGenericEpicName creates a generic but epic-sounding quest name
func (qng *QuestNameGenerator) generateGenericEpicName(analysis *TaskAnalysis, originalName string) string {
	keyTerms := extractKeyTerms(originalName)
	
	// Epic prefixes based on complexity
	prefixes := map[int][]string{
		1: {"The Simple", "A Quick", "The Minor"},
		2: {"The", "A Noble", "The Modest"},
		3: {"The Great", "The Epic", "The Grand"},
		4: {"The Mighty", "The Legendary", "The Heroic"},
		5: {"The Ultimate", "The Mythical", "The Supreme"},
	}
	
	// Epic suffixes based on task type
	suffixes := map[string][]string{
		"debugging":     {"of Debugging", "of Bug Hunting", "of Error Slaying"},
		"development":   {"of Creation", "of Building", "of Crafting"},
		"deployment":    {"of Deployment", "of Release", "of Publishing"},
		"testing":       {"of Verification", "of Testing", "of Quality"},
		"documentation": {"of Writing", "of Knowledge", "of Wisdom"},
		"optimization":  {"of Optimization", "of Enhancement", "of Improvement"},
		"design":        {"of Architecture", "of Design", "of Planning"},
		"monitoring":    {"of Vigilance", "of Watching", "of Guarding"},
	}
	
	complexity := analysis.Complexity
	if complexity < 1 {
		complexity = 1
	}
	if complexity > 5 {
		complexity = 5
	}
	
	prefix := "The Great"
	if prefixList, exists := prefixes[complexity]; exists {
		prefix = prefixList[0] // Use first option for consistency
	}
	
	suffix := "of Adventure"
	if suffixList, exists := suffixes[analysis.TaskType]; exists {
		suffix = suffixList[len(analysis.RequiredSkills)%len(suffixList)]
	}
	
	// Use the main subject from the original name if available
	subject := "Quest"
	if len(keyTerms) > 0 {
		subject = strings.Title(strings.ToLower(keyTerms[0]))
	}
	
	return fmt.Sprintf("%s %s %s", prefix, subject, suffix)
}

// extractKeyTerms extracts meaningful terms from the original task name
func extractKeyTerms(originalName string) []string {
	words := strings.Fields(strings.ToLower(originalName))
	var keyTerms []string
	
	// Skip common words and extract meaningful terms
	skipWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true, "in": true, "on": true, "at": true, "to": true, "for": true, "of": true, "with": true, "by": true, "from": true, "up": true, "about": true, "into": true, "through": true, "during": true, "before": true, "after": true, "above": true, "below": true, "between": true, "among": true, "this": true, "that": true, "these": true, "those": true, "i": true, "me": true, "my": true, "myself": true, "we": true, "us": true, "our": true, "ours": true, "you": true, "your": true, "yours": true, "he": true, "him": true, "his": true, "she": true, "her": true, "hers": true, "it": true, "its": true, "they": true, "them": true, "their": true, "theirs": true,
	}
	
	for _, word := range words {
		if len(word) > 2 && !skipWords[word] {
			keyTerms = append(keyTerms, strings.Title(word))
			if len(keyTerms) >= 3 { // Limit to 3 key terms
				break
			}
		}
	}
	
	return keyTerms
}

// initQuestNameTemplates creates the predefined quest name templates
func initQuestNameTemplates() map[string][]QuestNameTemplate {
	return map[string][]QuestNameTemplate{
		"debugging": {
			{
				Pattern:     "The {difficulty} Hunt for the {object} Bug",
				Themes:      []string{"Elusive", "Hidden", "Ancient", "Cursed", "Phantom"},
				Descriptors: []string{"Tricky", "Mysterious", "Complex", "Devious", "Legendary"},
			},
			{
				Pattern:     "Siege of the Broken {object}",
				Themes:      []string{"Fortress", "Tower", "System", "Module", "Service"},
				Descriptors: []string{"Crumbling", "Corrupted", "Failing", "Unstable", "Cursed"},
			},
			{
				Pattern:     "The {descriptor} Debugging of {object}",
				Themes:      []string{"Code", "Logic", "System", "Function", "Process"},
				Descriptors: []string{"Great", "Epic", "Noble", "Heroic", "Legendary"},
			},
		},
		"development": {
			{
				Pattern:     "The {difficulty} Forging of {object}",
				Themes:      []string{"Code", "System", "Module", "Feature", "Tool"},
				Descriptors: []string{"Grand", "Mighty", "Epic", "Noble", "Legendary"},
			},
			{
				Pattern:     "Quest for the Perfect {object}",
				Themes:      []string{"Implementation", "Solution", "Architecture", "Design", "System"},
				Descriptors: []string{"Elegant", "Robust", "Scalable", "Efficient", "Beautiful"},
			},
			{
				Pattern:     "The {descriptor} Creation of {object}",
				Themes:      []string{"Feature", "System", "Tool", "Module", "Service"},
				Descriptors: []string{"Great", "Magnificent", "Ambitious", "Revolutionary", "Masterful"},
			},
		},
		"deployment": {
			{
				Pattern:     "The {difficulty} Deployment of {object}",
				Themes:      []string{"Service", "System", "Application", "Platform", "Infrastructure"},
				Descriptors: []string{"Critical", "Major", "Strategic", "Essential", "Vital"},
			},
			{
				Pattern:     "Launch of the {descriptor} {object}",
				Themes:      []string{"Platform", "Service", "System", "Application", "Release"},
				Descriptors: []string{"Mighty", "Revolutionary", "Advanced", "Powerful", "Ultimate"},
			},
			{
				Pattern:     "The {theme} Release Quest",
				Themes:      []string{"Golden", "Silver", "Platinum", "Diamond", "Epic"},
				Descriptors: []string{"Smooth", "Flawless", "Perfect", "Triumphant", "Glorious"},
			},
		},
		"testing": {
			{
				Pattern:     "The {difficulty} Verification of {object}",
				Themes:      []string{"Code", "System", "Function", "Feature", "Logic"},
				Descriptors: []string{"Thorough", "Rigorous", "Comprehensive", "Meticulous", "Exhaustive"},
			},
			{
				Pattern:     "Trial by {theme} for {object}",
				Themes:      []string{"Fire", "Code", "Logic", "Data", "Users"},
				Descriptors: []string{"Intense", "Demanding", "Challenging", "Rigorous", "Ultimate"},
			},
			{
				Pattern:     "The Quality Assurance {theme}",
				Themes:      []string{"Crusade", "Mission", "Campaign", "Journey", "Adventure"},
				Descriptors: []string{"Noble", "Vital", "Critical", "Essential", "Sacred"},
			},
		},
		"documentation": {
			{
				Pattern:     "The {difficulty} Chronicle of {object}",
				Themes:      []string{"Knowledge", "Wisdom", "Code", "System", "Process"},
				Descriptors: []string{"Sacred", "Ancient", "Comprehensive", "Detailed", "Masterful"},
			},
			{
				Pattern:     "Scribing the {descriptor} {object} Tome",
				Themes:      []string{"Knowledge", "Wisdom", "Code", "API", "Guide"},
				Descriptors: []string{"Great", "Essential", "Comprehensive", "Ultimate", "Definitive"},
			},
			{
				Pattern:     "The {theme} Documentation Quest",
				Themes:      []string{"Sacred", "Holy", "Divine", "Essential", "Noble"},
				Descriptors: []string{"Important", "Vital", "Critical", "Fundamental", "Crucial"},
			},
		},
		"optimization": {
			{
				Pattern:     "The {difficulty} Refinement of {object}",
				Themes:      []string{"Code", "System", "Performance", "Logic", "Algorithm"},
				Descriptors: []string{"Great", "Major", "Significant", "Revolutionary", "Masterful"},
			},
			{
				Pattern:     "Enhancement of the {descriptor} {object}",
				Themes:      []string{"System", "Code", "Performance", "Function", "Service"},
				Descriptors: []string{"Ancient", "Legacy", "Critical", "Core", "Essential"},
			},
			{
				Pattern:     "The {theme} Optimization Saga",
				Themes:      []string{"Performance", "Speed", "Efficiency", "Resource", "Memory"},
				Descriptors: []string{"Epic", "Heroic", "Legendary", "Ambitious", "Monumental"},
			},
		},
		"design": {
			{
				Pattern:     "The {difficulty} Architectural Vision of {object}",
				Themes:      []string{"System", "Platform", "Framework", "Infrastructure", "Solution"},
				Descriptors: []string{"Grand", "Ambitious", "Visionary", "Revolutionary", "Masterful"},
			},
			{
				Pattern:     "Blueprint of the {descriptor} {object}",
				Themes:      []string{"System", "Architecture", "Framework", "Platform", "Solution"},
				Descriptors: []string{"Perfect", "Ideal", "Ultimate", "Legendary", "Divine"},
			},
			{
				Pattern:     "The {theme} Design Odyssey",
				Themes:      []string{"Architectural", "Structural", "Strategic", "Systematic", "Holistic"},
				Descriptors: []string{"Epic", "Grand", "Visionary", "Revolutionary", "Transcendent"},
			},
		},
		"monitoring": {
			{
				Pattern:     "The {difficulty} Vigil of {object}",
				Themes:      []string{"System", "Service", "Health", "Performance", "Status"},
				Descriptors: []string{"Eternal", "Watchful", "Vigilant", "Constant", "Unwavering"},
			},
			{
				Pattern:     "Guardian Watch over {object}",
				Themes:      []string{"System", "Service", "Infrastructure", "Platform", "Network"},
				Descriptors: []string{"Sacred", "Noble", "Essential", "Critical", "Vital"},
			},
			{
				Pattern:     "The {theme} Monitoring Mission",
				Themes:      []string{"Silent", "Vigilant", "Watchful", "Alert", "Observant"},
				Descriptors: []string{"Critical", "Essential", "Vital", "Important", "Sacred"},
			},
		},
		"general": {
			{
				Pattern:     "The {difficulty} Adventure of {object}",
				Themes:      []string{"Discovery", "Exploration", "Journey", "Quest", "Mission"},
				Descriptors: []string{"Grand", "Noble", "Epic", "Heroic", "Legendary"},
			},
			{
				Pattern:     "Quest for the {descriptor} {object}",
				Themes:      []string{"Solution", "Answer", "Truth", "Knowledge", "Wisdom"},
				Descriptors: []string{"Perfect", "Ultimate", "Ideal", "Supreme", "Divine"},
			},
			{
				Pattern:     "The {theme} Mission",
				Themes:      []string{"Noble", "Sacred", "Epic", "Heroic", "Legendary"},
				Descriptors: []string{"Important", "Critical", "Essential", "Vital", "Sacred"},
			},
		},
	}
}

// initSeasonalEvents defines seasonal events that can modify quest names
func initSeasonalEvents() []SeasonalEvent {
	return []SeasonalEvent{
		{
			Name:        "New Year's Resolutions",
			StartMonth:  1,  // January
			StartDay:    1,
			EndMonth:    1,
			EndDay:      3,
			Prefixes:    []string{"New Year's", "Resolution", "Fresh Start", "Renewed"},
			Suffixes:    []string{"of New Beginnings", "of Fresh Starts", "of Resolutions"},
			Themes:      []string{"Sparkling", "Golden", "Celebratory", "Ambitious"},
			Decorations: []string{"✨", "🎆", "🥳", "🎊"},
		},
		{
			Name:       "Valentine's Festival",
			StartMonth: 2,  // February
			StartDay:   13,
			EndMonth:   2,
			EndDay:     16,
			Prefixes:   []string{"Romantic", "Lovely", "Heartfelt", "Passionate"},
			Suffixes:   []string{"of Love", "of Hearts", "of Romance", "of Harmony"},
			Themes:     []string{"Pink", "Rosy", "Loving", "Harmonious"},
			Decorations: []string{"💖", "🌹", "💕", "💝"},
		},
		{
			Name:       "St. Patrick's Luck",
			StartMonth: 3,  // March
			StartDay:   17,
			EndMonth:   3,
			EndDay:     17,
			Prefixes:   []string{"Lucky", "Emerald", "Golden", "Fortunate"},
			Suffixes:   []string{"of Luck", "of Fortune", "of Prosperity", "of the Emerald Isle"},
			Themes:     []string{"Green", "Lucky", "Golden", "Magical"},
			Decorations: []string{"🍀", "🌈", "🪙", "🎩"},
		},
		{
			Name:       "Easter Renewal",
			StartMonth: 3,  // March/April (simplified to March)
			StartDay:   25,
			EndMonth:   4,
			EndDay:     10,
			Prefixes:   []string{"Easter", "Spring", "Renewal", "Blooming"},
			Suffixes:   []string{"of Renewal", "of Spring", "of Rebirth", "of Growth"},
			Themes:     []string{"Blooming", "Fresh", "Colorful", "Vibrant"},
			Decorations: []string{"🐰", "🥚", "🌷", "🌸"},
		},
		{
			Name:       "Halloween Haunt",
			StartMonth: 10, // October
			StartDay:   25,
			EndMonth:   10,
			EndDay:     31,
			Prefixes:   []string{"Spooky", "Haunted", "Eerie", "Phantom", "Ghostly"},
			Suffixes:   []string{"of Shadows", "of the Haunted", "of Dark Magic", "of Spirits"},
			Themes:     []string{"Spooky", "Haunted", "Mysterious", "Dark"},
			Decorations: []string{"🎃", "👻", "🦇", "🕷️", "💀"},
		},
		{
			Name:       "Thanksgiving Gratitude",
			StartMonth: 11, // November
			StartDay:   22,
			EndMonth:   11,
			EndDay:     28,
			Prefixes:   []string{"Thankful", "Grateful", "Harvest", "Abundant"},
			Suffixes:   []string{"of Gratitude", "of Abundance", "of Harvest", "of Blessing"},
			Themes:     []string{"Golden", "Abundant", "Grateful", "Warm"},
			Decorations: []string{"🦃", "🍂", "🌽", "🥧"},
		},
		{
			Name:       "Winter Solstice Magic",
			StartMonth: 12, // December
			StartDay:   20,
			EndMonth:   12,
			EndDay:     26,
			Prefixes:   []string{"Festive", "Magical", "Winter", "Holiday", "Enchanted"},
			Suffixes:   []string{"of Wonder", "of Magic", "of the Holidays", "of Winter Solstice"},
			Themes:     []string{"Magical", "Festive", "Sparkling", "Enchanted"},
			Decorations: []string{"🎄", "❄️", "🎁", "⭐", "🔔"},
		},
		{
			Name:       "Independence Day",
			StartMonth: 7, // July
			StartDay:   4,
			EndMonth:   7,
			EndDay:     4,
			Prefixes:   []string{"Patriotic", "Freedom", "Liberty", "Independence"},
			Suffixes:   []string{"of Freedom", "of Liberty", "of Independence", "of the Republic"},
			Themes:     []string{"Patriotic", "Free", "Bold", "Courageous"},
			Decorations: []string{"🇺🇸", "🎆", "🗽", "🎇"},
		},
	}
}