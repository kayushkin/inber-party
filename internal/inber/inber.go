// Package inber provides read-only access to inber's SQLite databases,
// mapping agent session data to RPG concepts for visualization.
package inber

import (
	"database/sql"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	_ "github.com/mattn/go-sqlite3"
)

// RPG model types derived from inber data

type RPGAgent struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Title       string     `json:"title"`
	Class       string     `json:"class"`
	Level       int        `json:"level"`
	XP          int        `json:"xp"`
	XPToNext    int        `json:"xp_to_next"`
	Energy      int        `json:"energy"`
	MaxEnergy   int        `json:"max_energy"`
	Status      string     `json:"status"` // idle, working, resting, stuck
	AvatarEmoji string     `json:"avatar_emoji"`
	TotalTokens int        `json:"total_tokens"`
	TotalCost   float64    `json:"total_cost"`
	SessionCount int       `json:"session_count"`
	QuestCount  int        `json:"quest_count"`
	ErrorCount  int        `json:"error_count"`
	Skills      []RPGSkill `json:"skills"`
	LastActive  *time.Time `json:"last_active,omitempty"`
}

type RPGSkill struct {
	Name      string `json:"skill_name"`
	Level     int    `json:"level"`
	TaskCount int    `json:"task_count"`
}

type RPGQuest struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Difficulty  int        `json:"difficulty"`
	XPReward    int        `json:"xp_reward"`
	Status      string     `json:"status"` // available, in_progress, completed, failed
	AgentID     string     `json:"assigned_agent_id,omitempty"`
	AgentName   string     `json:"assigned_agent_name,omitempty"`
	Progress    int        `json:"progress"`
	Turns       int        `json:"turns"`
	TokensUsed  int        `json:"tokens_used"`
	Cost        float64    `json:"cost"`
	CreatedAt   string     `json:"created_at"`
	StartedAt   string     `json:"started_at,omitempty"`
	CompletedAt string     `json:"completed_at,omitempty"`
	ErrorText   string     `json:"error_text,omitempty"`
	Children    int        `json:"children,omitempty"` // sub-quests spawned
}

type RPGStats struct {
	TotalAgents       int     `json:"total_agents"`
	ActiveQuests      int     `json:"active_quests"`
	CompletedQuests   int     `json:"completed_quests"`
	FailedQuests      int     `json:"failed_quests"`
	TotalXP           int     `json:"total_xp"`
	TotalTokens       int     `json:"total_tokens"`
	TotalCost         float64 `json:"total_cost"`
	AverageAgentLevel float64 `json:"average_agent_level"`
}

// Agent class assignments based on agent name patterns
var agentClasses = map[string]struct{ class, emoji, title string }{
	"claxon": {"Wizard", "🧙", "the All-Seeing"},
	"brigid": {"Healer", "✨", "the Radiant"},
	"run":    {"Ranger", "🏹", "the Swift"},
	"main":   {"Wizard", "🧙", "the Wise"},
}

func classFor(agent string) (string, string, string) {
	agent = strings.ToLower(agent)
	if c, ok := agentClasses[agent]; ok {
		return c.class, c.emoji, c.title
	}
	return "Warrior", "⚔️", "the Unknown"
}

// Store provides read-only access to inber's databases.
type Store struct {
	sessionsDB *sql.DB // ~/.inber/sessions.db
	gatewayDB  *sql.DB // ~/.inber/gateway/gateway.db
}

// DefaultDBPaths returns the default paths for inber databases.
func DefaultDBPaths() (sessionsDB, gatewayDB string) {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".inber", "sessions.db"),
		filepath.Join(home, ".inber", "gateway", "gateway.db")
}

// NewStore opens read-only connections to inber's databases.
func NewStore(sessionsDBPath, gatewayDBPath string) (*Store, error) {
	s := &Store{}
	var err error

	if sessionsDBPath != "" {
		if _, err := os.Stat(sessionsDBPath); err == nil {
			s.sessionsDB, err = sql.Open("sqlite3", sessionsDBPath+"?mode=ro&_journal_mode=WAL")
			if err != nil {
				return nil, fmt.Errorf("open sessions db: %w", err)
			}
		}
	}

	if gatewayDBPath != "" {
		if _, err = os.Stat(gatewayDBPath); err == nil {
			s.gatewayDB, err = sql.Open("sqlite3", gatewayDBPath+"?mode=ro&_journal_mode=WAL")
			if err != nil {
				if s.sessionsDB != nil {
					s.sessionsDB.Close()
				}
				return nil, fmt.Errorf("open gateway db: %w", err)
			}
		}
	}

	if s.sessionsDB == nil && s.gatewayDB == nil {
		return nil, fmt.Errorf("no inber databases found at %s or %s", sessionsDBPath, gatewayDBPath)
	}

	return s, nil
}

func (s *Store) Close() {
	if s.sessionsDB != nil {
		s.sessionsDB.Close()
	}
	if s.gatewayDB != nil {
		s.gatewayDB.Close()
	}
}

// xpForTokens converts token usage to XP (1 XP per 100 tokens).
func xpForTokens(tokens int) int {
	return tokens / 100
}

// levelForXP calculates level from XP (each level requires more XP).
func levelForXP(xp int) (level int, xpToNext int) {
	// Level thresholds: level N requires N*100 XP total
	// So level 1 = 0 XP, level 2 = 100 XP, level 3 = 300 XP, etc.
	totalNeeded := 0
	level = 1
	for {
		next := level * 100
		if xp < totalNeeded+next {
			xpToNext = (totalNeeded + next) - xp
			return level, xpToNext
		}
		totalNeeded += next
		level++
		if level > 99 {
			return 99, 0
		}
	}
}

// energyFromActivity calculates energy based on recent activity.
// More recent activity = lower energy (agent is tired).
func energyFromActivity(lastActive *time.Time) int {
	if lastActive == nil {
		return 100
	}
	hoursSince := time.Since(*lastActive).Hours()
	// Recovers 10 energy per hour of rest, min 20 if recently active
	energy := int(math.Min(100, 20+hoursSince*10))
	return energy
}

// GetAgents returns all known agents mapped to RPG characters.
func (s *Store) GetAgents() ([]RPGAgent, error) {
	agentMap := make(map[string]*RPGAgent)

	// Pull from gateway DB (has richer request-level data)
	if s.gatewayDB != nil {
		rows, err := s.gatewayDB.Query(`
			SELECT s.agent,
				COUNT(DISTINCT s.key) as session_count,
				COUNT(r.id) as request_count,
				COALESCE(SUM(r.input_tokens + r.output_tokens), 0) as total_tokens,
				COALESCE(SUM(r.cost), 0) as total_cost,
				COUNT(CASE WHEN r.status = 'completed' OR r.status = 'success' THEN 1 END) as completed,
				COUNT(CASE WHEN r.status = 'error' THEN 1 END) as errors,
				COUNT(CASE WHEN r.status = 'running' THEN 1 END) as running,
				MAX(s.last_active) as last_active
			FROM sessions s
			LEFT JOIN requests r ON r.session_key = s.key
			GROUP BY s.agent
		`)
		if err != nil {
			return nil, fmt.Errorf("query gateway agents: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				agentName    string
				sessionCount int
				requestCount int
				totalTokens  int
				totalCost    float64
				completed    int
				errors       int
				running      int
				lastActive   sql.NullString
			)
			if err := rows.Scan(&agentName, &sessionCount, &requestCount, &totalTokens, &totalCost,
				&completed, &errors, &running, &lastActive); err != nil {
				continue
			}

			class, emoji, title := classFor(agentName)
			xp := xpForTokens(totalTokens)
			level, xpToNext := levelForXP(xp)

			var la *time.Time
			if lastActive.Valid {
				if t, err := time.Parse("2006-01-02 15:04:05", lastActive.String); err == nil {
					la = &t
				}
			}

			status := "idle"
			if running > 0 {
				status = "working"
			} else if errors > 0 && float64(errors)/float64(requestCount) > 0.5 {
				status = "stuck"
			}

			agentMap[agentName] = &RPGAgent{
				ID:           agentName,
				Name:         titleCase(agentName),
				Title:        title,
				Class:        class,
				Level:        level,
				XP:           xp,
				XPToNext:     xpToNext,
				Energy:       energyFromActivity(la),
				MaxEnergy:    100,
				Status:       status,
				AvatarEmoji:  emoji,
				TotalTokens:  totalTokens,
				TotalCost:    totalCost,
				SessionCount: sessionCount,
				QuestCount:   completed,
				ErrorCount:   errors,
				LastActive:   la,
			}
		}
	}

	// Supplement from sessions DB (has turn-level detail and model info)
	if s.sessionsDB != nil {
		rows, err := s.sessionsDB.Query(`
			SELECT s.agent,
				COUNT(DISTINCT s.id) as session_count,
				COALESCE(SUM(t.in_tokens + t.out_tokens), 0) as total_tokens,
				COALESCE(SUM(t.cost), 0) as total_cost,
				COALESCE(SUM(t.tool_calls), 0) as total_tool_calls,
				COUNT(CASE WHEN s.status = 'completed' THEN 1 END) as completed,
				COUNT(CASE WHEN s.status = 'interrupted' THEN 1 END) as interrupted,
				MAX(s.started_at) as last_active
			FROM sessions s
			LEFT JOIN turns t ON t.session_id = s.id
			GROUP BY s.agent
		`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var (
					agentName  string
					sessions   int
					tokens     int
					cost       float64
					toolCalls  int
					completed  int
					interrupted int
					lastActive sql.NullString
				)
				if err := rows.Scan(&agentName, &sessions, &tokens, &cost, &toolCalls,
					&completed, &interrupted, &lastActive); err != nil {
					continue
				}

				// Add skills based on tool usage
				if a, ok := agentMap[agentName]; ok {
					if toolCalls > 0 {
						toolLevel, _ := levelForXP(toolCalls * 10)
						a.Skills = append(a.Skills, RPGSkill{
							Name:      "Tool Mastery",
							Level:     toolLevel,
							TaskCount: toolCalls,
						})
					}
					// Merge tokens if sessions DB has more
					if tokens > a.TotalTokens {
						a.TotalTokens = tokens
						xp := xpForTokens(tokens)
						a.XP = xp
						a.Level, a.XPToNext = levelForXP(xp)
					}
				} else {
					// Agent only in sessions DB
					class, emoji, title := classFor(agentName)
					xp := xpForTokens(tokens)
					level, xpToNext := levelForXP(xp)

					var la *time.Time
					if lastActive.Valid {
						if t, err := time.Parse("2006-01-02T15:04:05-07:00", lastActive.String); err == nil {
							la = &t
						} else if t, err := time.Parse(time.RFC3339Nano, lastActive.String); err == nil {
							la = &t
						}
					}

					a := &RPGAgent{
						ID:           agentName,
						Name:         titleCase(agentName),
						Title:        title,
						Class:        class,
						Level:        level,
						XP:           xp,
						XPToNext:     xpToNext,
						Energy:       energyFromActivity(la),
						MaxEnergy:    100,
						Status:       "idle",
						AvatarEmoji:  emoji,
						TotalTokens:  tokens,
						TotalCost:    cost,
						SessionCount: sessions,
						QuestCount:   completed,
						ErrorCount:   interrupted,
						LastActive:   la,
					}
					if toolCalls > 0 {
						toolLevel, _ := levelForXP(toolCalls * 10)
						a.Skills = append(a.Skills, RPGSkill{
							Name:      "Tool Mastery",
							Level:     toolLevel,
							TaskCount: toolCalls,
						})
					}
					agentMap[agentName] = a
				}
			}
		}
	}

	// Add default skills based on class for all agents
	for _, a := range agentMap {
		if a.Skills == nil {
			a.Skills = []RPGSkill{}
		}
		// Add class-based skills
		switch a.Class {
		case "Wizard":
			a.Skills = append(a.Skills, RPGSkill{Name: "Code Weaving", Level: a.Level / 2, TaskCount: a.QuestCount})
			a.Skills = append(a.Skills, RPGSkill{Name: "Arcane Debugging", Level: a.Level / 3, TaskCount: a.ErrorCount})
		case "Healer":
			a.Skills = append(a.Skills, RPGSkill{Name: "Restoration", Level: a.Level / 2, TaskCount: a.QuestCount})
			a.Skills = append(a.Skills, RPGSkill{Name: "Purification", Level: a.Level / 3, TaskCount: a.ErrorCount})
		case "Ranger":
			a.Skills = append(a.Skills, RPGSkill{Name: "Swift Execution", Level: a.Level / 2, TaskCount: a.QuestCount})
			a.Skills = append(a.Skills, RPGSkill{Name: "Path Finding", Level: a.Level / 3, TaskCount: a.SessionCount})
		default:
			a.Skills = append(a.Skills, RPGSkill{Name: "Brute Force", Level: a.Level / 2, TaskCount: a.QuestCount})
		}
	}

	result := make([]RPGAgent, 0, len(agentMap))
	for _, a := range agentMap {
		result = append(result, *a)
	}
	return result, nil
}

// GetQuests returns requests mapped to RPG quests.
func (s *Store) GetQuests(limit int) ([]RPGQuest, error) {
	if limit <= 0 {
		limit = 50
	}

	var quests []RPGQuest

	if s.gatewayDB != nil {
		rows, err := s.gatewayDB.Query(`
			SELECT r.id, s.agent, r.status, r.input_text, r.turns,
				r.input_tokens, r.output_tokens, r.cost,
				r.started_at, r.completed_at, r.error_text,
				r.parent_request_id,
				(SELECT COUNT(*) FROM requests c WHERE c.parent_request_id = r.id) as children
			FROM requests r
			JOIN sessions s ON s.key = r.session_key
			ORDER BY r.id DESC
			LIMIT ?
		`, limit)
		if err != nil {
			return nil, fmt.Errorf("query quests: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				id            int
				agent         string
				status        string
				inputText     sql.NullString
				turns         int
				inTokens      int
				outTokens     int
				cost          float64
				startedAt     sql.NullString
				completedAt   sql.NullString
				errorText     sql.NullString
				parentID      sql.NullInt64
				children      int
			)
			if err := rows.Scan(&id, &agent, &status, &inputText, &turns,
				&inTokens, &outTokens, &cost, &startedAt, &completedAt,
				&errorText, &parentID, &children); err != nil {
				continue
			}

			totalTokens := inTokens + outTokens
			xpReward := xpForTokens(totalTokens)
			if xpReward < 1 {
				xpReward = 1
			}

			// Map inber status to RPG quest status
			questStatus := "in_progress"
			switch status {
			case "completed", "success":
				questStatus = "completed"
			case "error", "timeout":
				questStatus = "failed"
			case "pending":
				questStatus = "available"
			case "interrupted":
				questStatus = "failed"
			}

			// Generate quest name from input text
			questName := generateQuestName(inputText.String, status)
			questDesc := ""
			if inputText.Valid && inputText.String != "" {
				questDesc = inputText.String
				if len(questDesc) > 200 {
					questDesc = questDesc[:200] + "..."
				}
			}

			// Difficulty based on tokens used
			difficulty := 1
			if totalTokens > 5000 {
				difficulty = 5
			} else if totalTokens > 2000 {
				difficulty = 4
			} else if totalTokens > 1000 {
				difficulty = 3
			} else if totalTokens > 500 {
				difficulty = 2
			}

			// Progress: completed=100, running=50, error=progress at failure
			progress := 50
			if questStatus == "completed" {
				progress = 100
			} else if questStatus == "failed" {
				progress = 30 // died trying
			} else if questStatus == "available" {
				progress = 0
			}

			q := RPGQuest{
				ID:          id,
				Name:        questName,
				Description: questDesc,
				Difficulty:  difficulty,
				XPReward:    xpReward,
				Status:      questStatus,
				AgentID:     agent,
				AgentName:   titleCase(agent),
				Progress:    progress,
				Turns:       turns,
				TokensUsed:  totalTokens,
				Cost:        cost,
				CreatedAt:   startedAt.String,
				ErrorText:   errorText.String,
				Children:    children,
			}
			if startedAt.Valid {
				q.StartedAt = startedAt.String
			}
			if completedAt.Valid {
				q.CompletedAt = completedAt.String
			}

			// If it's a sub-quest, prefix the name
			if parentID.Valid {
				q.Name = "⚔️ " + q.Name
			}

			quests = append(quests, q)
		}
	}

	if quests == nil {
		quests = []RPGQuest{}
	}
	return quests, nil
}

// GetStats returns aggregate RPG stats.
func (s *Store) GetStats() (*RPGStats, error) {
	agents, err := s.GetAgents()
	if err != nil {
		return nil, err
	}

	quests, err := s.GetQuests(1000)
	if err != nil {
		return nil, err
	}

	stats := &RPGStats{
		TotalAgents: len(agents),
	}

	totalLevel := 0
	for _, a := range agents {
		stats.TotalXP += a.XP
		stats.TotalTokens += a.TotalTokens
		stats.TotalCost += a.TotalCost
		totalLevel += a.Level
	}
	if len(agents) > 0 {
		stats.AverageAgentLevel = float64(totalLevel) / float64(len(agents))
	}

	for _, q := range quests {
		switch q.Status {
		case "in_progress":
			stats.ActiveQuests++
		case "completed":
			stats.CompletedQuests++
		case "failed":
			stats.FailedQuests++
		}
	}

	return stats, nil
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// generateQuestName creates an RPG-flavored name from the request text.
func generateQuestName(input, status string) string {
	if input == "" {
		switch status {
		case "error":
			return "The Failed Incantation"
		case "running":
			return "An Ongoing Adventure"
		default:
			return "A Mysterious Quest"
		}
	}

	// Trim and take first line
	lines := strings.SplitN(input, "\n", 2)
	text := strings.TrimSpace(lines[0])
	if len(text) > 60 {
		text = text[:57] + "..."
	}

	// Add RPG flair based on keywords
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "fix") || strings.Contains(lower, "bug"):
		return "🛡️ " + text
	case strings.Contains(lower, "build") || strings.Contains(lower, "create"):
		return "🔨 " + text
	case strings.Contains(lower, "test"):
		return "⚗️ " + text
	case strings.Contains(lower, "deploy"):
		return "🚀 " + text
	case strings.Contains(lower, "refactor"):
		return "📜 " + text
	default:
		return "📋 " + text
	}
}
