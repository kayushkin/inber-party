// Package inber provides read-only access to inber's SQLite databases,
// mapping agent session data to RPG concepts for visualization.
package inber

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	_ "github.com/mattn/go-sqlite3"
)

// RPG model types derived from inber data

type RPGAgent struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Title       string         `json:"title"`
	Class       string         `json:"class"`
	Level       int            `json:"level"`
	XP          int            `json:"xp"`
	XPToNext    int            `json:"xp_to_next"`
	Energy      int            `json:"energy"`
	MaxEnergy   int            `json:"max_energy"`
	Status       string         `json:"status"` // idle, working, resting, stuck
	Orchestrator string         `json:"orchestrator"`
	AvatarEmoji  string         `json:"avatar_emoji"`
	TotalTokens int            `json:"total_tokens"`
	TotalCost   float64        `json:"total_cost"`
	SessionCount int           `json:"session_count"`
	QuestCount  int            `json:"quest_count"`
	ErrorCount  int            `json:"error_count"`
	Skills      []RPGSkill     `json:"skills"`
	LastActive  *time.Time     `json:"last_active,omitempty"`
	HeldItems   []RPGHeldItem  `json:"held_items,omitempty"`
}

type RPGHeldItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Icon         string `json:"icon"`
	ActivityType string `json:"activity_type"`
	Priority     int    `json:"priority"`
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

type RPGAchievement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	UnlockedAt  string `json:"unlocked_at"`
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
	TotalSessions     int     `json:"total_sessions"`
	Uptime            string  `json:"uptime,omitempty"`
}

// QuestHistoryEntry is a lightweight quest record for chart data.
type QuestHistoryEntry struct {
	ID         int     `json:"id"`
	Tokens     int     `json:"tokens"`
	Cost       float64 `json:"cost"`
	Status     string  `json:"status"`
	StartedAt  string  `json:"started_at"`
	CompletedAt string `json:"completed_at,omitempty"`
}

// RPGConversation represents a conversation thread between agents.
type RPGConversation struct {
	ID           string             `json:"id"`
	ParticipantIDs []string         `json:"participant_ids"`
	Participants []string           `json:"participants"`
	Title        string             `json:"title"`
	Messages     []RPGMessage       `json:"messages"`
	StartedAt    string             `json:"started_at"`
	LastActive   string             `json:"last_active"`
	Type         string             `json:"type"` // "spawn_chain", "session_chat", "inter_agent"
}

// RPGMessage represents a single message in an agent conversation.
type RPGMessage struct {
	ID        string `json:"id"`
	FromAgent string `json:"from_agent"`
	ToAgent   string `json:"to_agent,omitempty"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"` // "message", "spawn", "task", "result"
}

// Agent class assignments based on agent name patterns
var agentClasses = map[string]struct{ class, emoji, title string }{
	// Inber agents
	"claxon":    {"Overlord", "♚", "the All-Seeing"},
	"fionn":     {"Engineer", "⚙️", "the Builder"},
	"brigid":    {"Artificer", "🔨", "the Radiant"},
	"oisin":     {"Bard", "🎵", "the Storyteller"},
	"manannan":  {"Ranger", "🌊", "the Voyager"},
	"ogma":      {"Scribe", "📜", "the Chronicler"},
	"scathach":  {"Shadow", "🗡️", "the Unseen"},
	"goibniu":   {"Smith", "⚒️", "the Forgemaster"},
	"bench":     {"Gladiator", "🏛️", "the Proven"},
	"bran":      {"Scout", "🐕", "the Loyal"},
	// OpenClaw agents
	"main":           {"Sovereign", "👑", "the Wise"},
	"kayushkin":      {"Artificer", "🔨", "the Maker"},
	"si":             {"Bard", "🎵", "the Melodic"},
	"downloadstack":  {"Ranger", "🌊", "the Fetcher"},
	"claxon-android": {"Shadow", "📱", "the Mobile"},
	"inber":          {"Engineer", "⚙️", "the Orchestrator"},
	"forge":          {"Smith", "⚒️", "the Forgemaster"},
	"logstack":       {"Scribe", "📜", "the Recorder"},
	"healthcheck":    {"Cleric", "🏥", "the Healer"},
	"inber-party":    {"Jester", "🃏", "the Entertainer"},
	"agent-bench":    {"Gladiator", "🏛️", "the Tested"},
	"argraphments":   {"Sage", "📚", "the Learned"},
	"keyboard":       {"Tinker", "🎹", "the Melodic"},
	"claxon-watch":   {"Sentinel", "⌚", "the Watchful"},
	"run":            {"Ranger", "🏹", "the Swift"},
}

func classFor(agent string) (string, string, string) {
	agent = strings.ToLower(agent)
	if c, ok := agentClasses[agent]; ok {
		return c.class, c.emoji, c.title
	}
	// Sensible defaults for unknown agents
	return "Adventurer", "⚔️", "the Unknown"
}

// inberRegistryAgent is the shape returned by inber's GET /api/agents registry endpoint.
type inberRegistryAgent struct {
	Name         string `json:"name"`
	Orchestrator string `json:"orchestrator"`
	Enabled      bool   `json:"enabled"`
}

// Store provides read-only access to inber's databases.
type Store struct {
	sessionsDB *sql.DB // ~/.inber/sessions.db
	gatewayDB  *sql.DB // ~/.inber/gateway/gateway.db
	inberURL   string  // base URL for inber HTTP API (for registry)
}

// DefaultDBPaths returns the default paths for inber databases.
func DefaultDBPaths() (sessionsDB, gatewayDB string) {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".inber", "sessions.db"),
		filepath.Join(home, ".inber", "gateway", "gateway.db")
}

// NewStore opens read-only connections to inber's databases.
func NewStore(sessionsDBPath, gatewayDBPath, inberURL string) (*Store, error) {
	s := &Store{inberURL: inberURL}
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

	// It's OK if both are nil — caller can use HTTP fallback
	return s, nil
}

// HasData returns true if at least one database is connected.
func (s *Store) HasData() bool {
	return s.sessionsDB != nil || s.gatewayDB != nil
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

// fetchRegistry fetches the full agent list from inber's HTTP API.
func (s *Store) fetchRegistry() []inberRegistryAgent {
	if s.inberURL == "" {
		return nil
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(s.inberURL + "/api/agents")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil
	}
	var agents []inberRegistryAgent
	if err := json.NewDecoder(resp.Body).Decode(&agents); err != nil {
		return nil
	}
	return agents
}

// GetAgents returns all known agents mapped to RPG characters.
func (s *Store) GetAgents() ([]RPGAgent, error) {
	agentMap := make(map[string]*RPGAgent)

	// Build orchestrator map from registry
	orchestratorMap := make(map[string]string)
	registry := s.fetchRegistry()
	for _, ra := range registry {
		if ra.Name != "" {
			orchestratorMap[ra.Name] = ra.Orchestrator
		}
	}

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
				Orchestrator: orchestratorMap[agentName],
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
						Orchestrator: orchestratorMap[agentName],
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

	// Add registry agents that have no DB data
	for _, ra := range registry {
		if ra.Name == "" {
			continue
		}
		if _, exists := agentMap[ra.Name]; exists {
			continue
		}
		class, emoji, title := classFor(ra.Name)
		a := &RPGAgent{
			ID:           ra.Name,
			Name:         titleCase(ra.Name),
			Title:        title,
			Class:        class,
			Level:        1,
			XP:           0,
			XPToNext:     100,
			Energy:       100,
			MaxEnergy:    100,
			Status:       "idle",
			Orchestrator: ra.Orchestrator,
			AvatarEmoji:  emoji,
			Skills:       []RPGSkill{{Name: "Brute Force", Level: 0, TaskCount: 0}},
		}
		agentMap[ra.Name] = a
	}

	result := make([]RPGAgent, 0, len(agentMap))
	for _, a := range agentMap {
		// Analyze recent activity and add held items
		analysis := s.analyzeRecentActivity(a.ID)
		a.HeldItems = getHeldItemsForAgent(analysis)
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
	var earliest *time.Time
	for _, a := range agents {
		stats.TotalXP += a.XP
		stats.TotalTokens += a.TotalTokens
		stats.TotalCost += a.TotalCost
		stats.TotalSessions += a.SessionCount
		totalLevel += a.Level
		if a.LastActive != nil && (earliest == nil || a.LastActive.Before(*earliest)) {
			t := *a.LastActive
			earliest = &t
		}
	}
	if len(agents) > 0 {
		stats.AverageAgentLevel = float64(totalLevel) / float64(len(agents))
	}
	if earliest != nil {
		dur := time.Since(*earliest)
		days := int(dur.Hours() / 24)
		if days > 0 {
			stats.Uptime = fmt.Sprintf("%dd", days)
		} else {
			hours := int(dur.Hours())
			stats.Uptime = fmt.Sprintf("%dh", hours)
		}
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

// GetAchievements computes achievements for an agent based on their data and quest history.
func (s *Store) GetAchievements(agentID string) ([]RPGAchievement, error) {
	var achievements []RPGAchievement

	agents, err := s.GetAgents()
	if err != nil {
		return nil, err
	}
	var agent *RPGAgent
	for i := range agents {
		if agents[i].ID == agentID {
			agent = &agents[i]
			break
		}
	}
	if agent == nil {
		return []RPGAchievement{}, nil
	}

	quests, err := s.GetQuests(1000)
	if err != nil {
		return nil, err
	}

	var agentQuests []RPGQuest
	for _, q := range quests {
		if q.AgentID == agentID {
			agentQuests = append(agentQuests, q)
		}
	}

	completedCount := 0
	hasError := false
	hasNightOwl := false
	hasMarathon := false
	var firstQuestTime string

	for _, q := range agentQuests {
		if q.Status == "completed" {
			completedCount++
		}
		if q.Status == "failed" {
			hasError = true
		}
		// Night owl: activity after midnight (00:00-05:00)
		if q.StartedAt != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", q.StartedAt); err == nil {
				h := t.Hour()
				if h >= 0 && h < 5 {
					hasNightOwl = true
				}
			}
			if firstQuestTime == "" || q.StartedAt < firstQuestTime {
				firstQuestTime = q.StartedAt
			}
		}
		// Marathon: >30 turns (proxy for long session)
		if q.Turns > 30 {
			hasMarathon = true
		}
	}

	ts := time.Now().Format("2006-01-02T15:04:05Z")

	if len(agentQuests) > 0 {
		achievements = append(achievements, RPGAchievement{
			ID: "first_quest", Name: "First Quest", Description: "Completed their first quest",
			Icon: "⚔️", UnlockedAt: firstQuestTime,
		})
	}

	if agent.TotalTokens >= 1000 {
		achievements = append(achievements, RPGAchievement{
			ID: "1k_tokens", Name: "Apprentice Scribe", Description: "Used 1,000 tokens",
			Icon: "📜", UnlockedAt: ts,
		})
	}
	if agent.TotalTokens >= 100000 {
		achievements = append(achievements, RPGAchievement{
			ID: "100k_tokens", Name: "Master Scribe", Description: "Used 100,000 tokens",
			Icon: "📚", UnlockedAt: ts,
		})
	}
	if agent.TotalTokens >= 1000000 {
		achievements = append(achievements, RPGAchievement{
			ID: "1m_tokens", Name: "Archmage of Words", Description: "Used 1,000,000 tokens",
			Icon: "🌟", UnlockedAt: ts,
		})
	}

	if hasError {
		achievements = append(achievements, RPGAchievement{
			ID: "first_error", Name: "Battle Scarred", Description: "Survived their first failed quest",
			Icon: "💀", UnlockedAt: ts,
		})
	}

	if completedCount >= 10 {
		achievements = append(achievements, RPGAchievement{
			ID: "10_quests", Name: "Veteran", Description: "Completed 10 quests",
			Icon: "🛡️", UnlockedAt: ts,
		})
	}
	if completedCount >= 50 {
		achievements = append(achievements, RPGAchievement{
			ID: "50_quests", Name: "Champion", Description: "Completed 50 quests",
			Icon: "👑", UnlockedAt: ts,
		})
	}

	if hasNightOwl {
		achievements = append(achievements, RPGAchievement{
			ID: "night_owl", Name: "Night Owl", Description: "Active after midnight",
			Icon: "🦉", UnlockedAt: ts,
		})
	}

	if hasMarathon {
		achievements = append(achievements, RPGAchievement{
			ID: "marathon", Name: "Marathon Runner", Description: "Completed a quest with 30+ turns",
			Icon: "🏃", UnlockedAt: ts,
		})
	}

	if agent.Level >= 5 {
		achievements = append(achievements, RPGAchievement{
			ID: "level5", Name: "Seasoned Adventurer", Description: "Reached level 5",
			Icon: "⭐", UnlockedAt: ts,
		})
	}
	if agent.Level >= 10 {
		achievements = append(achievements, RPGAchievement{
			ID: "level10", Name: "Elite", Description: "Reached level 10",
			Icon: "💎", UnlockedAt: ts,
		})
	}

	return achievements, nil
}

// GetQuestHistory returns recent quest data for an agent (for charts).
func (s *Store) GetQuestHistory(agentID string, limit int) ([]QuestHistoryEntry, error) {
	if limit <= 0 {
		limit = 20
	}
	if s.gatewayDB == nil {
		return []QuestHistoryEntry{}, nil
	}

	rows, err := s.gatewayDB.Query(`
		SELECT r.id, (r.input_tokens + r.output_tokens) as tokens, r.cost,
			r.status, r.started_at, r.completed_at
		FROM requests r
		JOIN sessions s ON s.key = r.session_key
		WHERE s.agent = ?
		ORDER BY r.id DESC
		LIMIT ?
	`, agentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []QuestHistoryEntry
	for rows.Next() {
		var e QuestHistoryEntry
		var startedAt, completedAt sql.NullString
		if err := rows.Scan(&e.ID, &e.Tokens, &e.Cost, &e.Status, &startedAt, &completedAt); err != nil {
			continue
		}
		if startedAt.Valid {
			e.StartedAt = startedAt.String
		}
		if completedAt.Valid {
			e.CompletedAt = completedAt.String
		}
		// Map status
		switch e.Status {
		case "completed", "success":
			e.Status = "completed"
		case "error", "timeout", "interrupted":
			e.Status = "failed"
		}
		entries = append(entries, e)
	}
	// Reverse to chronological order
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	if entries == nil {
		entries = []QuestHistoryEntry{}
	}
	return entries, nil
}

// GetConversations returns inter-agent conversations from session data.
func (s *Store) GetConversations(limit int) ([]RPGConversation, error) {
	if limit <= 0 {
		limit = 50
	}

	var conversations []RPGConversation

	if s.sessionsDB != nil {
		// Query for sessions that involve spawned agents or inter-agent communication
		rows, err := s.sessionsDB.Query(`
			SELECT DISTINCT 
				s1.id as session_id,
				s1.agent as main_agent,
				s2.agent as spawned_agent,
				s1.started_at,
				s1.last_message_at,
				s1.initial_message
			FROM sessions s1
			LEFT JOIN sessions s2 ON s2.parent_session_id = s1.id
			WHERE s2.agent IS NOT NULL OR s1.initial_message LIKE '%spawn%' OR s1.initial_message LIKE '%session%'
			ORDER BY s1.started_at DESC
			LIMIT ?
		`, limit)
		
		if err != nil {
			return nil, fmt.Errorf("query conversations: %w", err)
		}
		defer rows.Close()

		conversationMap := make(map[string]*RPGConversation)

		for rows.Next() {
			var (
				sessionID      string
				mainAgent      sql.NullString
				spawnedAgent   sql.NullString
				startedAt      sql.NullString
				lastMessageAt  sql.NullString
				initialMessage sql.NullString
			)
			if err := rows.Scan(&sessionID, &mainAgent, &spawnedAgent, &startedAt, &lastMessageAt, &initialMessage); err != nil {
				continue
			}

			conversationKey := sessionID
			if spawnedAgent.Valid && mainAgent.Valid {
				// Group by main agent
				conversationKey = mainAgent.String
			}

			conv, exists := conversationMap[conversationKey]
			if !exists {
				conv = &RPGConversation{
					ID:           conversationKey,
					ParticipantIDs: []string{},
					Participants: []string{},
					Messages:     []RPGMessage{},
					Type:         "spawn_chain",
				}
				if startedAt.Valid {
					conv.StartedAt = startedAt.String
				}
				if lastMessageAt.Valid {
					conv.LastActive = lastMessageAt.String
				}
				conversationMap[conversationKey] = conv
			}

			// Add participants
			if mainAgent.Valid && !contains(conv.ParticipantIDs, mainAgent.String) {
				conv.ParticipantIDs = append(conv.ParticipantIDs, mainAgent.String)
				conv.Participants = append(conv.Participants, titleCase(mainAgent.String))
			}
			if spawnedAgent.Valid && !contains(conv.ParticipantIDs, spawnedAgent.String) {
				conv.ParticipantIDs = append(conv.ParticipantIDs, spawnedAgent.String)
				conv.Participants = append(conv.Participants, titleCase(spawnedAgent.String))
			}

			// Create initial message from spawn action
			if initialMessage.Valid && spawnedAgent.Valid && mainAgent.Valid {
				msgContent := initialMessage.String
				if len(msgContent) > 100 {
					msgContent = msgContent[:100] + "..."
				}
				
				msg := RPGMessage{
					ID:        fmt.Sprintf("%s-spawn-%s", sessionID, spawnedAgent.String),
					FromAgent: titleCase(mainAgent.String),
					ToAgent:   titleCase(spawnedAgent.String),
					Content:   fmt.Sprintf("🎯 Spawned %s: %s", titleCase(spawnedAgent.String), msgContent),
					Type:      "spawn",
				}
				if startedAt.Valid {
					msg.Timestamp = startedAt.String
				}
				
				// Add if not already exists
				msgExists := false
				for _, existingMsg := range conv.Messages {
					if existingMsg.ID == msg.ID {
						msgExists = true
						break
					}
				}
				if !msgExists {
					conv.Messages = append(conv.Messages, msg)
				}
			}
		}

		// Now get turn messages for each conversation
		for _, conv := range conversationMap {
			turnRows, err := s.sessionsDB.Query(`
				SELECT t.id, s.agent, t.role, t.content, t.timestamp
				FROM turns t
				JOIN sessions s ON s.id = t.session_id
				WHERE s.agent IN (` + generateInClause(len(conv.ParticipantIDs)) + `)
				ORDER BY t.timestamp ASC
			`, stringSliceToInterfaceSlice(conv.ParticipantIDs)...)
			
			if err != nil {
				continue
			}
			defer turnRows.Close()

			for turnRows.Next() {
				var (
					turnID    string
					agent     string
					role      string
					content   sql.NullString
					timestamp sql.NullString
				)
				if err := turnRows.Scan(&turnID, &agent, &role, &content, &timestamp); err != nil {
					continue
				}

				// Only include assistant messages (agent responses) and system messages about inter-agent activity
				if role == "assistant" || (role == "system" && content.Valid && 
					(strings.Contains(content.String, "spawn") || strings.Contains(content.String, "session"))) {
					
					msgContent := "No content"
					if content.Valid && content.String != "" {
						msgContent = content.String
						if len(msgContent) > 200 {
							msgContent = msgContent[:200] + "..."
						}
					}

					msgType := "message"
					if role == "system" {
						msgType = "system"
					}

					msg := RPGMessage{
						ID:        fmt.Sprintf("%s-%s", turnID, agent),
						FromAgent: titleCase(agent),
						Content:   msgContent,
						Type:      msgType,
					}
					if timestamp.Valid {
						msg.Timestamp = timestamp.String
					}

					conv.Messages = append(conv.Messages, msg)
				}
			}
		}

		// Generate conversation titles and convert to slice
		for _, conv := range conversationMap {
			if len(conv.Participants) > 0 {
				if len(conv.Participants) == 1 {
					conv.Title = fmt.Sprintf("%s's Work", conv.Participants[0])
				} else {
					conv.Title = fmt.Sprintf("Collaboration: %s", strings.Join(conv.Participants, ", "))
				}
			} else {
				conv.Title = "Unknown Conversation"
			}
			conversations = append(conversations, *conv)
		}
	}

	if conversations == nil {
		conversations = []RPGConversation{}
	}
	return conversations, nil
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func generateInClause(count int) string {
	if count == 0 {
		return "''"
	}
	placeholders := make([]string, count)
	for i := range placeholders {
		placeholders[i] = "?"
	}
	return strings.Join(placeholders, ", ")
}

func stringSliceToInterfaceSlice(strs []string) []interface{} {
	result := make([]interface{}, len(strs))
	for i, s := range strs {
		result[i] = s
	}
	return result
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

// Activity analysis for held items
type ActivityAnalysis struct {
	EditCount     int
	SpawnCount    int
	SearchCount   int
	DocWriting    int
	InfraWork     int
	DebugSessions int
	CreateCount   int
}

// analyzeRecentActivity analyzes an agent's recent activity to determine held items
func (s *Store) analyzeRecentActivity(agentID string) *ActivityAnalysis {
	if s.sessionsDB == nil {
		return &ActivityAnalysis{}
	}
	
	analysis := &ActivityAnalysis{}
	
	// Look at the last 7 days of activity
	rows, err := s.sessionsDB.Query(`
		SELECT s.initial_message, t.tool_calls, t.content, s.status, s.started_at
		FROM sessions s
		LEFT JOIN turns t ON t.session_id = s.id
		WHERE s.agent = ? AND s.started_at > datetime('now', '-7 days')
		ORDER BY s.started_at DESC
		LIMIT 100
	`, agentID)
	if err != nil {
		return analysis
	}
	defer rows.Close()
	
	for rows.Next() {
		var initialMessage, content, status sql.NullString
		var toolCalls sql.NullInt64
		var startedAt sql.NullString
		
		if err := rows.Scan(&initialMessage, &toolCalls, &content, &status, &startedAt); err != nil {
			continue
		}
		
		// Analyze initial messages and content for activity patterns
		textToAnalyze := ""
		if initialMessage.Valid {
			textToAnalyze += initialMessage.String + " "
		}
		if content.Valid {
			textToAnalyze += content.String + " "
		}
		
		lowerText := strings.ToLower(textToAnalyze)
		
		// Count edit operations
		editKeywords := []string{"edit", "modify", "change", "update", "refactor", "rewrite"}
		for _, keyword := range editKeywords {
			analysis.EditCount += strings.Count(lowerText, keyword)
		}
		
		// Count spawn operations
		spawnKeywords := []string{"spawn", "delegate", "session", "sub-agent"}
		for _, keyword := range spawnKeywords {
			analysis.SpawnCount += strings.Count(lowerText, keyword)
		}
		
		// Count search operations
		searchKeywords := []string{"search", "find", "lookup", "browse", "web", "fetch"}
		for _, keyword := range searchKeywords {
			analysis.SearchCount += strings.Count(lowerText, keyword)
		}
		
		// Count documentation work
		docKeywords := []string{"document", "readme", "guide", "manual", "write", "explain"}
		for _, keyword := range docKeywords {
			analysis.DocWriting += strings.Count(lowerText, keyword)
		}
		
		// Count infrastructure work
		infraKeywords := []string{"deploy", "build", "install", "configure", "setup", "gateway", "system"}
		for _, keyword := range infraKeywords {
			analysis.InfraWork += strings.Count(lowerText, keyword)
		}
		
		// Count debugging sessions
		debugKeywords := []string{"debug", "error", "fix", "bug", "issue", "problem", "troubleshoot"}
		for _, keyword := range debugKeywords {
			analysis.DebugSessions += strings.Count(lowerText, keyword)
		}
		
		// Count creation work
		createKeywords := []string{"create", "build", "make", "develop", "implement", "add", "new"}
		for _, keyword := range createKeywords {
			analysis.CreateCount += strings.Count(lowerText, keyword)
		}
		
		// High tool usage indicates heavy work
		if toolCalls.Valid && toolCalls.Int64 > 10 {
			analysis.EditCount += int(toolCalls.Int64 / 5) // Convert tool calls to edit activity
		}
	}
	
	return analysis
}

// getHeldItemsForAgent determines held items based on recent activity analysis
func getHeldItemsForAgent(analysis *ActivityAnalysis) []RPGHeldItem {
	var heldItems []RPGHeldItem
	
	// Define held items catalog
	heldItemsCatalog := map[string]RPGHeldItem{
		"smithing_hammer": {
			ID:           "smithing_hammer",
			Name:         "Smithing Hammer",
			Description:  "Wielded by agents who have been forging code with heavy edits and refactoring.",
			Icon:         "🔨",
			ActivityType: "edit",
			Priority:     10,
		},
		"claxon_horn": {
			ID:           "claxon_horn",
			Name:         "Claxon Horn",
			Description:  "Blown by commanders who have been summoning and directing sub-agents.",
			Icon:         "📯",
			ActivityType: "spawn",
			Priority:     15,
		},
		"magnifying_glass": {
			ID:           "magnifying_glass",
			Name:         "Investigator's Glass",
			Description:  "Used by agents who have been searching and investigating across the web.",
			Icon:         "🔍",
			ActivityType: "search",
			Priority:     8,
		},
		"scribes_scroll": {
			ID:           "scribes_scroll",
			Name:         "Scribe's Scroll",
			Description:  "Carried by agents who have been documenting knowledge and writing guides.",
			Icon:         "📜",
			ActivityType: "docs",
			Priority:     7,
		},
		"engineers_wrench": {
			ID:           "engineers_wrench",
			Name:         "Engineer's Wrench",
			Description:  "Gripped by agents who have been maintaining infrastructure and deployments.",
			Icon:         "🔧",
			ActivityType: "infra",
			Priority:     12,
		},
		"debugging_probe": {
			ID:           "debugging_probe",
			Name:         "Debugging Probe",
			Description:  "Wielded by agents who have been hunting down bugs and solving problems.",
			Icon:         "🔬",
			ActivityType: "debug",
			Priority:     9,
		},
		"builders_trowel": {
			ID:           "builders_trowel",
			Name:         "Builder's Trowel",
			Description:  "Used by agents who have been constructing new features and components.",
			Icon:         "🧱",
			ActivityType: "create",
			Priority:     11,
		},
	}
	
	// Determine which items to show based on activity thresholds
	activityMap := map[string]int{
		"edit":   analysis.EditCount,
		"spawn":  analysis.SpawnCount,
		"search": analysis.SearchCount,
		"docs":   analysis.DocWriting,
		"infra":  analysis.InfraWork,
		"debug":  analysis.DebugSessions,
		"create": analysis.CreateCount,
	}
	
	// Get the top 2 most active categories above threshold
	type activityScore struct {
		activityType string
		score        int
	}
	
	var scores []activityScore
	for activityType, count := range activityMap {
		// Set thresholds for each activity type
		threshold := 0
		switch activityType {
		case "spawn":
			threshold = 2 // Lower threshold for spawning
		case "infra":
			threshold = 2 // Lower threshold for infra work
		default:
			threshold = 5 // Default threshold
		}
		
		if count >= threshold {
			scores = append(scores, activityScore{activityType, count})
		}
	}
	
	// Sort by score descending
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].score > scores[i].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}
	
	// Take top 2 activities and find matching held items
	maxItems := 2
	for i, score := range scores {
		if i >= maxItems {
			break
		}
		
		for _, item := range heldItemsCatalog {
			if item.ActivityType == score.activityType {
				heldItems = append(heldItems, item)
				break
			}
		}
	}
	
	// Sort held items by priority (highest first)
	for i := 0; i < len(heldItems)-1; i++ {
		for j := i + 1; j < len(heldItems); j++ {
			if heldItems[j].Priority > heldItems[i].Priority {
				heldItems[i], heldItems[j] = heldItems[j], heldItems[i]
			}
		}
	}
	
	return heldItems
}

// SessionReplay represents a complete session replay with all tool calls and turns
type SessionReplay struct {
	SessionID    string           `json:"session_id"`
	AgentID      string           `json:"agent_id"`
	Status       string           `json:"status"`
	StartedAt    string           `json:"started_at"`
	CompletedAt  string           `json:"completed_at,omitempty"`
	InitialTask  string           `json:"initial_task"`
	TotalTurns   int              `json:"total_turns"`
	TotalTokens  int              `json:"total_tokens"`
	TotalCost    float64          `json:"total_cost"`
	Turns        []ReplayTurn     `json:"turns"`
}

// ReplayTurn represents a single turn in the session replay
type ReplayTurn struct {
	TurnNumber   int              `json:"turn_number"`
	Timestamp    string           `json:"timestamp"`
	Input        string           `json:"input"`
	Output       string           `json:"output"`
	ToolCalls    []ReplayToolCall `json:"tool_calls"`
	InputTokens  int              `json:"input_tokens"`
	OutputTokens int              `json:"output_tokens"`
	Cost         float64          `json:"cost"`
	Duration     float64          `json:"duration"` // seconds
}

// ReplayToolCall represents a single tool call in the replay
type ReplayToolCall struct {
	Name        string      `json:"name"`
	Parameters  interface{} `json:"parameters"`
	Result      string      `json:"result"`
	Success     bool        `json:"success"`
	Duration    float64     `json:"duration"` // seconds
	Timestamp   string      `json:"timestamp"`
}

// GetSessionReplay returns detailed session data for replay visualization
func (s *Store) GetSessionReplay(sessionID string) (*SessionReplay, error) {
	if s.gatewayDB == nil {
		return nil, fmt.Errorf("gateway database not available")
	}

	// Get basic session information
	var replay SessionReplay
	var completedAt sql.NullString
	
	err := s.gatewayDB.QueryRow(`
		SELECT s.key, s.agent, s.status, s.started_at, s.completed_at, 
			   s.initial_message, s.total_turns, 
			   COALESCE((SELECT SUM(input_tokens + output_tokens) FROM requests WHERE session_key = s.key), 0) as total_tokens,
			   COALESCE((SELECT SUM(cost) FROM requests WHERE session_key = s.key), 0) as total_cost
		FROM sessions s
		WHERE s.key = ?
	`, sessionID).Scan(&replay.SessionID, &replay.AgentID, &replay.Status, &replay.StartedAt, 
		&completedAt, &replay.InitialTask, &replay.TotalTurns, &replay.TotalTokens, &replay.TotalCost)
	
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}
	
	if completedAt.Valid {
		replay.CompletedAt = completedAt.String
	}

	// Get all turns/requests for this session
	rows, err := s.gatewayDB.Query(`
		SELECT r.id, r.started_at, r.input_text, r.output_text, r.tool_calls,
			   r.input_tokens, r.output_tokens, r.cost,
			   COALESCE(julianday(r.completed_at) - julianday(r.started_at), 0) * 86400 as duration
		FROM requests r
		WHERE r.session_key = ?
		ORDER BY r.id ASC
	`, sessionID)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get session turns: %w", err)
	}
	defer rows.Close()

	turnNumber := 1
	for rows.Next() {
		var turn ReplayTurn
		var toolCallsJSON sql.NullString
		var duration sql.NullFloat64
		
		err := rows.Scan(&turn.TurnNumber, &turn.Timestamp, &turn.Input, &turn.Output,
			&toolCallsJSON, &turn.InputTokens, &turn.OutputTokens, &turn.Cost, &duration)
		if err != nil {
			continue
		}
		
		turn.TurnNumber = turnNumber
		turnNumber++
		
		if duration.Valid {
			turn.Duration = duration.Float64
		}

		// Parse tool calls if present
		if toolCallsJSON.Valid && toolCallsJSON.String != "" {
			var toolCalls []map[string]interface{}
			if err := json.Unmarshal([]byte(toolCallsJSON.String), &toolCalls); err == nil {
				for _, tc := range toolCalls {
					replayTool := ReplayToolCall{
						Timestamp: turn.Timestamp, // Use turn timestamp as default
					}
					
					if name, ok := tc["name"].(string); ok {
						replayTool.Name = name
					}
					if params, ok := tc["parameters"]; ok {
						replayTool.Parameters = params
					}
					if result, ok := tc["result"].(string); ok {
						replayTool.Result = result
						replayTool.Success = true
					}
					if errorStr, ok := tc["error"].(string); ok && errorStr != "" {
						replayTool.Result = errorStr
						replayTool.Success = false
					}
					// Duration estimation based on tool complexity
					replayTool.Duration = estimateToolDuration(replayTool.Name)
					
					turn.ToolCalls = append(turn.ToolCalls, replayTool)
				}
			}
		}
		
		replay.Turns = append(replay.Turns, turn)
	}

	return &replay, nil
}

// estimateToolDuration provides rough duration estimates for different tool types
func estimateToolDuration(toolName string) float64 {
	switch toolName {
	case "read", "write", "edit":
		return 1.0 // File operations are usually quick
	case "exec", "process":
		return 3.0 // Shell commands take longer
	case "web_search", "web_fetch":
		return 2.0 // Web operations have network latency
	case "image", "tts":
		return 4.0 // AI operations take time
	case "browser":
		return 5.0 // Browser automation is slow
	default:
		return 1.5 // Default duration
	}
}
