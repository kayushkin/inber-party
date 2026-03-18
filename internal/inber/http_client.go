// Package inber — HTTP API client fallback when SQLite DBs are unavailable.
package inber

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient fetches data from inber's HTTP API.
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPClient creates a client for inber's HTTP API.
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// inberAPIAgent is the shape returned by GET /api/agents
type inberAPIAgent struct {
	Name         string  `json:"name"`
	Agent        string  `json:"agent"`
	Orchestrator string  `json:"orchestrator"`
	Enabled      bool    `json:"enabled"`
	SessionCount int     `json:"session_count"`
	TotalTokens  int     `json:"total_tokens"`
	TotalCost    float64 `json:"total_cost"`
	Status       string  `json:"status"`
	LastActive   string  `json:"last_active"`
}

// inberAPISession is the shape returned by GET /api/sessions
type inberAPISession struct {
	Key        string  `json:"key"`
	Agent      string  `json:"agent"`
	Status     string  `json:"status"`
	Turns      int     `json:"turns"`
	InTokens   int     `json:"in_tokens"`
	OutTokens  int     `json:"out_tokens"`
	Cost       float64 `json:"cost"`
	StartedAt  string  `json:"started_at"`
	FinishedAt string  `json:"finished_at"`
	InputText  string  `json:"input_text"`
	ToolCalls  int     `json:"tool_calls"`
	ErrorText  string  `json:"error_text"`
}

func (c *HTTPClient) get(path string, out interface{}) error {
	resp, err := c.httpClient.Get(c.baseURL + path)
	if err != nil {
		return fmt.Errorf("http get %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("http get %s: status %d: %s", path, resp.StatusCode, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// GetAgents fetches agents from the HTTP API and maps to RPG agents.
func (c *HTTPClient) GetAgents() ([]RPGAgent, error) {
	var raw []inberAPIAgent
	if err := c.get("/api/agents", &raw); err != nil {
		return nil, err
	}

	agents := make([]RPGAgent, 0, len(raw))
	for _, a := range raw {
		name := a.Name
		if name == "" {
			name = a.Agent
		}
		if name == "" {
			continue
		}
		class, emoji, title := classFor(name)
		xp := xpForTokens(a.TotalTokens)
		level, xpToNext := levelForXP(xp)

		var la *time.Time
		if a.LastActive != "" {
			for _, fmt := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05-07:00", "2006-01-02 15:04:05"} {
				if t, err := time.Parse(fmt, a.LastActive); err == nil {
					la = &t
					break
				}
			}
		}

		status := "idle"
		if a.Status == "running" || a.Status == "active" {
			status = "working"
		}

		agents = append(agents, RPGAgent{
			ID:           name,
			Name:         titleCase(name),
			Title:        title,
			Class:        class,
			Level:        level,
			XP:           xp,
			XPToNext:     xpToNext,
			Energy:       energyFromActivity(la),
			MaxEnergy:    100,
			Status:       status,
			Orchestrator: a.Orchestrator,
			AvatarEmoji:  emoji,
			TotalTokens:  a.TotalTokens,
			TotalCost:    a.TotalCost,
			SessionCount: a.SessionCount,
			LastActive:   la,
			Skills:       []RPGSkill{},
		})
	}

	// Add default skills
	for i := range agents {
		a := &agents[i]
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

	return agents, nil
}

// GetQuests fetches sessions from the HTTP API and maps to RPG quests.
func (c *HTTPClient) GetQuests(limit int) ([]RPGQuest, error) {
	var raw []inberAPISession
	if err := c.get("/api/sessions", &raw); err != nil {
		return nil, err
	}

	quests := make([]RPGQuest, 0, len(raw))
	for i, s := range raw {
		if limit > 0 && i >= limit {
			break
		}
		totalTokens := s.InTokens + s.OutTokens
		xpReward := xpForTokens(totalTokens)
		if xpReward < 1 {
			xpReward = 1
		}

		questStatus := "in_progress"
		switch s.Status {
		case "completed", "success":
			questStatus = "completed"
		case "error", "timeout", "interrupted":
			questStatus = "failed"
		case "pending":
			questStatus = "available"
		}

		questName := generateQuestName(s.InputText, s.Status)
		questDesc := s.InputText
		if len(questDesc) > 200 {
			questDesc = questDesc[:200] + "..."
		}

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

		progress := 50
		if questStatus == "completed" {
			progress = 100
		} else if questStatus == "failed" {
			progress = 30
		} else if questStatus == "available" {
			progress = 0
		}

		quests = append(quests, RPGQuest{
			ID:          i + 1,
			Name:        questName,
			Description: questDesc,
			Difficulty:  difficulty,
			XPReward:    xpReward,
			Status:      questStatus,
			AgentID:     s.Agent,
			AgentName:   titleCase(s.Agent),
			Progress:    progress,
			Turns:       s.Turns,
			TokensUsed:  totalTokens,
			Cost:        s.Cost,
			CreatedAt:   s.StartedAt,
			StartedAt:   s.StartedAt,
			CompletedAt: s.FinishedAt,
			ErrorText:   s.ErrorText,
		})
	}

	return quests, nil
}

// GetStats computes stats from HTTP API data.
func (c *HTTPClient) GetStats() (*RPGStats, error) {
	agents, err := c.GetAgents()
	if err != nil {
		return nil, err
	}
	quests, err := c.GetQuests(1000)
	if err != nil {
		return nil, err
	}

	stats := &RPGStats{TotalAgents: len(agents)}
	totalLevel := 0
	for _, a := range agents {
		stats.TotalXP += a.XP
		stats.TotalTokens += a.TotalTokens
		stats.TotalCost += a.TotalCost
		stats.TotalSessions += a.SessionCount
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

// GetAchievements computes achievements from HTTP API data.
func (c *HTTPClient) GetAchievements(agentID string) ([]RPGAchievement, error) {
	agents, err := c.GetAgents()
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

	quests, err := c.GetQuests(1000)
	if err != nil {
		return nil, err
	}

	return computeAchievements(agent, quests), nil
}

// GetQuestHistory returns recent quest data for an agent from HTTP API.
func (c *HTTPClient) GetQuestHistory(agentID string, limit int) ([]QuestHistoryEntry, error) {
	quests, err := c.GetQuests(1000)
	if err != nil {
		return nil, err
	}

	var entries []QuestHistoryEntry
	for _, q := range quests {
		if q.AgentID != agentID {
			continue
		}
		status := q.Status
		switch status {
		case "completed":
			// keep
		default:
			if status == "failed" {
				// keep
			} else {
				status = "in_progress"
			}
		}
		entries = append(entries, QuestHistoryEntry{
			ID:          q.ID,
			Tokens:      q.TokensUsed,
			Cost:        q.Cost,
			Status:      status,
			StartedAt:   q.StartedAt,
			CompletedAt: q.CompletedAt,
		})
		if limit > 0 && len(entries) >= limit {
			break
		}
	}
	if entries == nil {
		entries = []QuestHistoryEntry{}
	}
	return entries, nil
}

// computeAchievements is shared logic for computing achievements from agent+quests data.
func computeAchievements(agent *RPGAgent, allQuests []RPGQuest) []RPGAchievement {
	var achievements []RPGAchievement

	var agentQuests []RPGQuest
	for _, q := range allQuests {
		if q.AgentID == agent.ID {
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
		if q.StartedAt != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", q.StartedAt); err == nil {
				if h := t.Hour(); h >= 0 && h < 5 {
					hasNightOwl = true
				}
			}
			if firstQuestTime == "" || q.StartedAt < firstQuestTime {
				firstQuestTime = q.StartedAt
			}
		}
		if q.Turns > 30 {
			hasMarathon = true
		}
	}

	ts := time.Now().Format("2006-01-02T15:04:05Z")

	if len(agentQuests) > 0 {
		achievements = append(achievements, RPGAchievement{ID: "first_quest", Name: "First Quest", Description: "Completed their first quest", Icon: "⚔️", UnlockedAt: firstQuestTime})
	}
	if agent.TotalTokens >= 1000 {
		achievements = append(achievements, RPGAchievement{ID: "1k_tokens", Name: "Apprentice Scribe", Description: "Used 1,000 tokens", Icon: "📜", UnlockedAt: ts})
	}
	if agent.TotalTokens >= 100000 {
		achievements = append(achievements, RPGAchievement{ID: "100k_tokens", Name: "Master Scribe", Description: "Used 100,000 tokens", Icon: "📚", UnlockedAt: ts})
	}
	if agent.TotalTokens >= 1000000 {
		achievements = append(achievements, RPGAchievement{ID: "1m_tokens", Name: "Archmage of Words", Description: "Used 1,000,000 tokens", Icon: "🌟", UnlockedAt: ts})
	}
	if hasError {
		achievements = append(achievements, RPGAchievement{ID: "first_error", Name: "Battle Scarred", Description: "Survived their first failed quest", Icon: "💀", UnlockedAt: ts})
	}
	if completedCount >= 10 {
		achievements = append(achievements, RPGAchievement{ID: "10_quests", Name: "Veteran", Description: "Completed 10 quests", Icon: "🛡️", UnlockedAt: ts})
	}
	if completedCount >= 50 {
		achievements = append(achievements, RPGAchievement{ID: "50_quests", Name: "Champion", Description: "Completed 50 quests", Icon: "👑", UnlockedAt: ts})
	}
	if hasNightOwl {
		achievements = append(achievements, RPGAchievement{ID: "night_owl", Name: "Night Owl", Description: "Active after midnight", Icon: "🦉", UnlockedAt: ts})
	}
	if hasMarathon {
		achievements = append(achievements, RPGAchievement{ID: "marathon", Name: "Marathon Runner", Description: "Completed a quest with 30+ turns", Icon: "🏃", UnlockedAt: ts})
	}
	if agent.Level >= 5 {
		achievements = append(achievements, RPGAchievement{ID: "level5", Name: "Seasoned Adventurer", Description: "Reached level 5", Icon: "⭐", UnlockedAt: ts})
	}
	if agent.Level >= 10 {
		achievements = append(achievements, RPGAchievement{ID: "level10", Name: "Elite", Description: "Reached level 10", Icon: "💎", UnlockedAt: ts})
	}

	return achievements
}

// GetConversations returns empty conversations for HTTP client since this data is not available via API.
func (c *HTTPClient) GetConversations(limit int) ([]RPGConversation, error) {
	// HTTP API doesn't expose session conversation details
	return []RPGConversation{}, nil
}
