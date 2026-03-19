package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kayushkin/inber-party/internal/dailyquests"
	"github.com/kayushkin/inber-party/internal/db"
	"github.com/kayushkin/inber-party/internal/logstack"
	"github.com/kayushkin/inber-party/internal/mood"
	"github.com/kayushkin/inber-party/internal/questgiver"
	"github.com/kayushkin/inber-party/internal/sync"
	"github.com/kayushkin/inber-party/internal/verifiers"
	"github.com/kayushkin/inber-party/internal/ws"
)

type Server struct {
	DB               *db.DB
	Hub              *ws.Hub
	QuestGiver       *questgiver.QuestGiver
	DailyQuestMgr    *dailyquests.DailyQuestManager
	MoodCalc         *mood.MoodCalculator
	LogstackClient   *logstack.LogstackClient
	AgentSync        *sync.AgentRegistrySync
	VerifierRegistry *verifiers.VerifierRegistry
}

func NewServer(database *db.DB, hub *ws.Hub, qg *questgiver.QuestGiver, dqm *dailyquests.DailyQuestManager) *Server {
	moodCalc := mood.NewMoodCalculator(database)
	logstackClient := logstack.NewLogstackClient()
	verifierRegistry := verifiers.NewVerifierRegistry()
	return &Server{
		DB:               database, 
		Hub:              hub, 
		QuestGiver:       qg, 
		DailyQuestMgr:    dqm, 
		MoodCalc:         moodCalc,
		LogstackClient:   logstackClient,
		VerifierRegistry: verifierRegistry,
	}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/agents", s.handleAgents)
	mux.HandleFunc("/api/agents/", s.handleAgentDetail)
	mux.HandleFunc("/api/parties", s.handleParties)
	mux.HandleFunc("/api/parties/", s.handlePartyDetail)
	mux.HandleFunc("/api/tasks", s.handleTasks)
	mux.HandleFunc("/api/tasks/", s.handleTaskDetail)
	mux.HandleFunc("/api/stats", s.handleStats)
	mux.HandleFunc("/api/leaderboard", s.handleLeaderboard)
	mux.HandleFunc("/api/quest-giver/analyze", s.handleQuestAnalyze)
	mux.HandleFunc("/api/quest-giver/recommendations", s.handleQuestRecommendations)
	mux.HandleFunc("/api/agents/sync", s.handleAgentSync)
	mux.HandleFunc("/api/agents/sync/status", s.handleAgentSyncStatus)
	mux.HandleFunc("/api/agents/managed", s.handleManagedAgents)
	mux.HandleFunc("/api/quest-giver/assign", s.handleQuestAssign)
	mux.HandleFunc("/api/daily-quests", s.handleDailyQuests)
	mux.HandleFunc("/api/daily-quests/generate", s.handleGenerateDailyQuests)
	mux.HandleFunc("/api/daily-quests/stats", s.handleDailyQuestStats)
	mux.HandleFunc("/api/mood/update", s.handleMoodUpdate)
	mux.HandleFunc("/api/mood/levels", s.handleMoodLevels)
	mux.HandleFunc("/api/reputation/rankings", s.handleReputationRankings)
	mux.HandleFunc("/api/conversations", s.handleConversations)
	mux.HandleFunc("/api/conversations/", s.handleConversationDetail)
	mux.HandleFunc("/api/webhooks/spawn", s.handleSpawnWebhook)
	mux.HandleFunc("/api/health", s.handleHealthCheck)
	mux.HandleFunc("/api/costs", s.handleCosts)
	mux.HandleFunc("/api/costs/summary", s.handleCostsSummary)
	// Bounty marketplace endpoints
	mux.HandleFunc("/api/bounties", s.handleBounties)
	mux.HandleFunc("/api/bounties/", s.handleBountyDetail)
	mux.HandleFunc("/api/verifiers/types", s.handleVerifierTypes)
	mux.HandleFunc("/api/verifiers/run/", s.handleRunVerifiers)
	// Payout tracking endpoints
	mux.HandleFunc("/api/payouts", s.handlePayouts)
	mux.HandleFunc("/api/payouts/summary", s.handlePayoutSummary)
	mux.HandleFunc("/api/payouts/agent/", s.handleAgentPayouts)
	// Rating system endpoints
	mux.HandleFunc("/api/ratings", s.handleRatings)
	mux.HandleFunc("/api/ratings/agent/", s.handleAgentRatings)
	mux.HandleFunc("/api/ratings/stats/", s.handleRatingStats)
	mux.HandleFunc("/api/ratings/leaderboard", s.handleRatingLeaderboard)
}

func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listAgents(w, r)
	case http.MethodPost:
		s.createAgent(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleAgentDetail(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/agents/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getAgentDetail(w, r, id)
	case http.MethodPatch:
		s.updateAgent(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listAgents(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]struct{}{})
		return
	}
	rows, err := s.DB.Query(`
		SELECT id, name, title, class, level, xp, gold, energy, status, avatar_emoji, mood, mood_score, workload, last_active, created_at, updated_at
		FROM agents
		ORDER BY id
	`)
	if err != nil {
		log.Printf("Error listing agents: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	agents := []db.Agent{}
	for rows.Next() {
		var a db.Agent
		if err := rows.Scan(&a.ID, &a.Name, &a.Title, &a.Class, &a.Level, &a.XP, &a.Gold, &a.Energy, &a.Status, &a.AvatarEmoji, &a.Mood, &a.MoodScore, &a.Workload, &a.LastActive, &a.CreatedAt, &a.UpdatedAt); err != nil {
			log.Printf("Error scanning agent: %v", err)
			continue
		}
		agents = append(agents, a)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

func (s *Server) getAgentDetail(w http.ResponseWriter, r *http.Request, id int) {
	if s.DB == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	var agent db.Agent
	err := s.DB.QueryRow(`
		SELECT id, name, title, class, level, xp, gold, energy, status, avatar_emoji, mood, mood_score, workload, last_active, created_at, updated_at
		FROM agents WHERE id = $1
	`, id).Scan(&agent.ID, &agent.Name, &agent.Title, &agent.Class, &agent.Level, &agent.XP, &agent.Gold, &agent.Energy, &agent.Status, &agent.AvatarEmoji, &agent.Mood, &agent.MoodScore, &agent.Workload, &agent.LastActive, &agent.CreatedAt, &agent.UpdatedAt)
	if err != nil {
		http.Error(w, "Agent not found", http.StatusNotFound)
		return
	}

	// Get skills
	skills := []db.Skill{}
	skillRows, _ := s.DB.Query("SELECT id, agent_id, skill_name, level, task_count FROM skills WHERE agent_id = $1", id)
	defer skillRows.Close()
	for skillRows.Next() {
		var skill db.Skill
		skillRows.Scan(&skill.ID, &skill.AgentID, &skill.SkillName, &skill.Level, &skill.TaskCount)
		skills = append(skills, skill)
	}

	// Get achievements
	achievements := []db.Achievement{}
	achRows, _ := s.DB.Query("SELECT id, agent_id, achievement_name, unlocked_at FROM achievements WHERE agent_id = $1", id)
	defer achRows.Close()
	for achRows.Next() {
		var ach db.Achievement
		achRows.Scan(&ach.ID, &ach.AgentID, &ach.AchievementName, &ach.UnlockedAt)
		achievements = append(achievements, ach)
	}

	// Get tasks
	tasks := []db.Task{}
	taskRows, _ := s.DB.Query(`
		SELECT id, name, description, difficulty, xp_reward, gold_reward, status, assigned_agent_id, assigned_party_id, progress, created_at, started_at, completed_at
		FROM tasks WHERE assigned_agent_id = $1
		ORDER BY created_at DESC
	`, id)
	defer taskRows.Close()
	for taskRows.Next() {
		var task db.Task
		taskRows.Scan(&task.ID, &task.Name, &task.Description, &task.Difficulty, &task.XPReward, &task.GoldReward, &task.Status, &task.AssignedAgentID, &task.AssignedPartyID, &task.Progress, &task.CreatedAt, &task.StartedAt, &task.CompletedAt)
		tasks = append(tasks, task)
	}

	// Get reputation
	reputation, _ := s.DB.GetAgentReputation(id)

	detail := db.AgentDetail{
		Agent:        agent,
		Skills:       skills,
		Achievements: achievements,
		Tasks:        tasks,
		Reputation:   reputation,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

func (s *Server) createAgent(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		http.Error(w, "PostgreSQL not configured", http.StatusServiceUnavailable)
		return
	}
	var agent db.Agent
	if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start new agents with basic gold allowance
	if agent.Gold == 0 {
		agent.Gold = 100 // Starting gold
	}
	
	err := s.DB.QueryRow(`
		INSERT INTO agents (name, title, class, level, xp, gold, avatar_emoji)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`, agent.Name, agent.Title, agent.Class, agent.Level, agent.XP, agent.Gold, agent.AvatarEmoji).Scan(&agent.ID, &agent.CreatedAt, &agent.UpdatedAt)
	if err != nil {
		log.Printf("Error creating agent: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	s.Hub.Broadcast(ws.Message{Type: "agent_created", Data: agent})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(agent)
}

func (s *Server) updateAgent(w http.ResponseWriter, r *http.Request, id int) {
	if s.DB == nil {
		http.Error(w, "PostgreSQL not configured", http.StatusServiceUnavailable)
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Build dynamic update query
	query := "UPDATE agents SET updated_at = NOW()"
	args := []interface{}{}
	argCount := 1

	if name, ok := updates["name"].(string); ok {
		query += ", name = $" + strconv.Itoa(argCount)
		args = append(args, name)
		argCount++
	}
	if status, ok := updates["status"].(string); ok {
		query += ", status = $" + strconv.Itoa(argCount)
		args = append(args, status)
		argCount++
	}
	if energy, ok := updates["energy"].(float64); ok {
		query += ", energy = $" + strconv.Itoa(argCount)
		args = append(args, int(energy))
		argCount++
	}
	if level, ok := updates["level"].(float64); ok {
		query += ", level = $" + strconv.Itoa(argCount)
		args = append(args, int(level))
		argCount++
	}
	if xp, ok := updates["xp"].(float64); ok {
		query += ", xp = $" + strconv.Itoa(argCount)
		args = append(args, int(xp))
		argCount++
	}
	if gold, ok := updates["gold"].(float64); ok {
		query += ", gold = $" + strconv.Itoa(argCount)
		args = append(args, int(gold))
		argCount++
	}

	query += " WHERE id = $" + strconv.Itoa(argCount)
	args = append(args, id)

	if _, err := s.DB.Exec(query, args...); err != nil {
		log.Printf("Error updating agent: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	s.Hub.Broadcast(ws.Message{Type: "agent_updated", Data: map[string]interface{}{"id": id, "updates": updates}})

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listTasks(w, r)
	case http.MethodPost:
		s.createTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleTaskDetail(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPatch:
		s.updateTask(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]struct{}{})
		return
	}
	rows, err := s.DB.Query(`
		SELECT id, name, description, difficulty, xp_reward, gold_reward, status, assigned_agent_id, assigned_party_id, progress, created_at, started_at, completed_at
		FROM tasks
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Printf("Error listing tasks: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	tasks := []db.Task{}
	for rows.Next() {
		var t db.Task
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Difficulty, &t.XPReward, &t.GoldReward, &t.Status, &t.AssignedAgentID, &t.AssignedPartyID, &t.Progress, &t.CreatedAt, &t.StartedAt, &t.CompletedAt); err != nil {
			log.Printf("Error scanning task: %v", err)
			continue
		}
		tasks = append(tasks, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		http.Error(w, "PostgreSQL not configured", http.StatusServiceUnavailable)
		return
	}
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Calculate gold reward based on difficulty and XP reward
	if task.GoldReward == 0 {
		task.GoldReward = calculateGoldReward(task.XPReward, task.Difficulty)
	}

	err := s.DB.QueryRow(`
		INSERT INTO tasks (name, description, difficulty, xp_reward, gold_reward)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, task.Name, task.Description, task.Difficulty, task.XPReward, task.GoldReward).Scan(&task.ID, &task.CreatedAt)
	if err != nil {
		log.Printf("Error creating task: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	task.Status = "available"
	task.Progress = 0

	s.Hub.Broadcast(ws.Message{Type: "task_created", Data: task})

	// Auto-assign using quest giver if available
	if s.QuestGiver != nil {
		go func() {
			if err := s.QuestGiver.AutoAssignTask(task.ID, task.Name, task.Description); err != nil {
				log.Printf("Quest Giver auto-assignment failed: %v", err)
			}
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (s *Server) updateTask(w http.ResponseWriter, r *http.Request, id int) {
	if s.DB == nil {
		http.Error(w, "PostgreSQL not configured", http.StatusServiceUnavailable)
		return
	}
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := "UPDATE tasks SET "
	args := []interface{}{}
	argCount := 1

	if status, ok := updates["status"].(string); ok {
		query += "status = $" + strconv.Itoa(argCount) + ", "
		args = append(args, status)
		argCount++
		
		if status == "in_progress" {
			query += "started_at = NOW(), "
		} else if status == "completed" || status == "failed" {
			query += "completed_at = NOW(), "
		}
	}
	if agentID, ok := updates["assigned_agent_id"].(float64); ok {
		query += "assigned_agent_id = $" + strconv.Itoa(argCount) + ", "
		args = append(args, int(agentID))
		argCount++
	}
	if partyID, ok := updates["assigned_party_id"].(float64); ok {
		query += "assigned_party_id = $" + strconv.Itoa(argCount) + ", "
		args = append(args, int(partyID))
		argCount++
	}
	if progress, ok := updates["progress"].(float64); ok {
		query += "progress = $" + strconv.Itoa(argCount) + ", "
		args = append(args, int(progress))
		argCount++
	}

	// Remove trailing comma and space
	query = strings.TrimSuffix(query, ", ")
	query += " WHERE id = $" + strconv.Itoa(argCount)
	args = append(args, id)

	if _, err := s.DB.Exec(query, args...); err != nil {
		log.Printf("Error updating task: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update agent mood if task is completed or failed and assigned to an agent
	if status, hasStatus := updates["status"].(string); hasStatus && (status == "completed" || status == "failed") {
		if agentID, hasAgent := updates["assigned_agent_id"].(float64); hasAgent && s.MoodCalc != nil {
			agentIDInt := int(agentID)
			
			// Award XP and gold for completed tasks
			if status == "completed" {
				// Get task rewards
				var xpReward, goldReward int
				err := s.DB.QueryRow("SELECT xp_reward, gold_reward FROM tasks WHERE id = $1", id).Scan(&xpReward, &goldReward)
				if err == nil {
					// Award XP directly (XP is not tracked in payout system, only gold)
					_, err = s.DB.Exec(`
						UPDATE agents 
						SET xp = xp + $1, updated_at = NOW()
						WHERE id = $2
					`, xpReward, agentIDInt)
					
					// Award gold using the payout tracking system
					if goldReward > 0 {
						err = s.DB.RecordQuestPayout(id, agentIDInt, goldReward)
					}
					
					if err == nil {
						s.Hub.Broadcast(ws.Message{Type: "rewards_awarded", Data: map[string]interface{}{
							"agent_id": agentIDInt,
							"task_id": id,
							"xp_awarded": xpReward,
							"gold_awarded": goldReward,
						}})
					}
				}
				
				// Update last active time for completed tasks
				s.MoodCalc.UpdateAgentLastActive(agentIDInt)
			}
			
			// Recalculate mood for this agent
			mood, moodScore, err := s.MoodCalc.CalculateAgentMood(agentIDInt)
			if err == nil {
				var workload int
				s.DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE assigned_agent_id = $1 AND status = 'in_progress'", agentIDInt).Scan(&workload)
				
				s.DB.Exec(`
					UPDATE agents 
					SET mood = $1, mood_score = $2, workload = $3, updated_at = NOW()
					WHERE id = $4
				`, mood, moodScore, workload, agentIDInt)
				
				s.Hub.Broadcast(ws.Message{Type: "agent_mood_updated", Data: map[string]interface{}{
					"agent_id": agentIDInt,
					"mood": mood,
					"mood_score": moodScore,
					"workload": workload,
				}})
			}
		}
	}

	s.Hub.Broadcast(ws.Message{Type: "task_updated", Data: map[string]interface{}{"id": id, "updates": updates}})

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(db.Stats{})
		return
	}
	var stats db.Stats

	s.DB.QueryRow("SELECT COUNT(*) FROM agents").Scan(&stats.TotalAgents)
	s.DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE status = 'in_progress'").Scan(&stats.ActiveTasks)
	s.DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE status = 'completed'").Scan(&stats.CompletedTasks)
	s.DB.QueryRow("SELECT COALESCE(SUM(xp), 0) FROM agents").Scan(&stats.TotalXP)
	s.DB.QueryRow("SELECT COALESCE(SUM(gold), 0) FROM agents").Scan(&stats.TotalGold)
	s.DB.QueryRow("SELECT COALESCE(AVG(level), 0) FROM agents").Scan(&stats.AverageAgentLevel)
	s.DB.QueryRow("SELECT COUNT(*) FROM parties").Scan(&stats.TotalParties)
	s.DB.QueryRow("SELECT COUNT(*) FROM parties WHERE status = 'active' OR status = 'on_quest'").Scan(&stats.ActiveParties)
	
	// Add cost tracking to stats
	s.DB.QueryRow("SELECT COALESCE(SUM(cost_usd), 0) FROM cost_entries").Scan(&stats.TotalCostUSD)
	s.DB.QueryRow("SELECT COALESCE(SUM(tokens_used), 0) FROM cost_entries").Scan(&stats.TotalTokensUsed)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.DB == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]struct{}{})
		return
	}

	// Get comprehensive agent stats including quest completion and efficiency
	query := `
		WITH agent_quest_stats AS (
			SELECT 
				a.id,
				COUNT(CASE WHEN t.status = 'completed' THEN 1 END) as quests_completed,
				COUNT(CASE WHEN t.status IN ('completed', 'failed') THEN 1 END) as quests_attempted,
				AVG(CASE WHEN t.status = 'completed' THEN t.xp_reward END) as avg_quest_xp,
				SUM(CASE WHEN t.status = 'completed' THEN t.xp_reward ELSE 0 END) as total_quest_xp
			FROM agents a
			LEFT JOIN tasks t ON a.id = t.assigned_agent_id
			GROUP BY a.id
		)
		SELECT 
			a.id, a.name, a.title, a.class, a.level, a.xp, a.energy, 
			a.status, a.avatar_emoji, a.created_at,
			COALESCE(aqs.quests_completed, 0) as quests_completed,
			COALESCE(aqs.quests_attempted, 0) as quests_attempted,
			COALESCE(aqs.avg_quest_xp, 0) as avg_quest_xp,
			COALESCE(aqs.total_quest_xp, 0) as total_quest_xp,
			-- Efficiency score: (quests_completed / max(quests_attempted, 1)) * 100
			CASE 
				WHEN aqs.quests_attempted > 0 
				THEN ROUND((aqs.quests_completed::float / aqs.quests_attempted * 100), 2)
				ELSE 0 
			END as efficiency_score
		FROM agents a
		LEFT JOIN agent_quest_stats aqs ON a.id = aqs.id
		ORDER BY 
			a.xp DESC, 
			aqs.quests_completed DESC, 
			CASE 
				WHEN aqs.quests_attempted > 0 
				THEN aqs.quests_completed::float / aqs.quests_attempted
				ELSE 0 
			END DESC,
			a.level DESC
	`

	rows, err := s.DB.Query(query)
	if err != nil {
		log.Printf("Error fetching leaderboard: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type LeaderboardEntry struct {
		Agent           db.Agent `json:"agent"`
		QuestsCompleted int      `json:"quests_completed"`
		QuestsAttempted int      `json:"quests_attempted"`
		AvgQuestXP      float64  `json:"avg_quest_xp"`
		TotalQuestXP    int      `json:"total_quest_xp"`
		EfficiencyScore float64  `json:"efficiency_score"`
		Rank            int      `json:"rank"`
	}

	leaderboard := []LeaderboardEntry{}
	rank := 1
	for rows.Next() {
		var entry LeaderboardEntry
		var a = &entry.Agent
		
		err := rows.Scan(
			&a.ID, &a.Name, &a.Title, &a.Class, &a.Level, &a.XP, &a.Energy,
			&a.Status, &a.AvatarEmoji, &a.CreatedAt,
			&entry.QuestsCompleted, &entry.QuestsAttempted, &entry.AvgQuestXP,
			&entry.TotalQuestXP, &entry.EfficiencyScore,
		)
		if err != nil {
			log.Printf("Error scanning leaderboard entry: %v", err)
			continue
		}
		
		entry.Rank = rank
		leaderboard = append(leaderboard, entry)
		rank++
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaderboard)
}

func (s *Server) handleParties(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listParties(w, r)
	case http.MethodPost:
		s.createParty(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePartyDetail(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/parties/")
	parts := strings.Split(idStr, "/")
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "Invalid party ID", http.StatusBadRequest)
		return
	}

	if len(parts) > 1 && parts[1] == "members" {
		switch r.Method {
		case http.MethodPost:
			s.addPartyMember(w, r, id)
		case http.MethodDelete:
			s.removePartyMember(w, r, id)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.getPartyDetail(w, r, id)
	case http.MethodPatch:
		s.updateParty(w, r, id)
	case http.MethodDelete:
		s.disbandParty(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listParties(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]struct{}{})
		return
	}

	rows, err := s.DB.Query(`
		SELECT p.id, p.name, p.description, p.leader_id, p.status, p.created_at, p.updated_at,
			   a.name as leader_name, a.title as leader_title, a.class as leader_class
		FROM parties p
		JOIN agents a ON p.leader_id = a.id
		ORDER BY p.created_at DESC
	`)
	if err != nil {
		log.Printf("Error listing parties: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	parties := []map[string]interface{}{}
	for rows.Next() {
		var p db.Party
		var leaderName, leaderTitle, leaderClass string
		
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.LeaderID, &p.Status, &p.CreatedAt, &p.UpdatedAt,
			&leaderName, &leaderTitle, &leaderClass); err != nil {
			log.Printf("Error scanning party: %v", err)
			continue
		}
		
		// Get member count
		var memberCount int
		s.DB.QueryRow("SELECT COUNT(*) FROM party_members WHERE party_id = $1", p.ID).Scan(&memberCount)
		
		party := map[string]interface{}{
			"id":          p.ID,
			"name":        p.Name,
			"description": p.Description,
			"leader_id":   p.LeaderID,
			"status":      p.Status,
			"created_at":  p.CreatedAt,
			"updated_at":  p.UpdatedAt,
			"leader": map[string]interface{}{
				"name":  leaderName,
				"title": leaderTitle,
				"class": leaderClass,
			},
			"member_count": memberCount,
		}
		parties = append(parties, party)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parties)
}

func (s *Server) createParty(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		http.Error(w, "PostgreSQL not configured", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		LeaderID    int    `json:"leader_id"`
		MemberIDs   []int  `json:"member_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create party
	var party db.Party
	err := s.DB.QueryRow(`
		INSERT INTO parties (name, description, leader_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`, req.Name, req.Description, req.LeaderID).Scan(&party.ID, &party.CreatedAt, &party.UpdatedAt)
	if err != nil {
		log.Printf("Error creating party: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	party.Name = req.Name
	party.Description = req.Description
	party.LeaderID = req.LeaderID
	party.Status = "active"

	// Add leader as member
	_, err = s.DB.Exec(`
		INSERT INTO party_members (party_id, agent_id, role)
		VALUES ($1, $2, $3)
	`, party.ID, req.LeaderID, "leader")
	if err != nil {
		log.Printf("Error adding party leader: %v", err)
	}

	// Add other members
	for _, memberID := range req.MemberIDs {
		if memberID != req.LeaderID {
			_, err = s.DB.Exec(`
				INSERT INTO party_members (party_id, agent_id, role)
				VALUES ($1, $2, $3)
			`, party.ID, memberID, "member")
			if err != nil {
				log.Printf("Error adding party member %d: %v", memberID, err)
			}
		}
	}

	s.Hub.Broadcast(ws.Message{Type: "party_created", Data: party})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(party)
}

func (s *Server) getPartyDetail(w http.ResponseWriter, r *http.Request, id int) {
	if s.DB == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	var party db.Party
	err := s.DB.QueryRow(`
		SELECT id, name, description, leader_id, status, created_at, updated_at
		FROM parties WHERE id = $1
	`, id).Scan(&party.ID, &party.Name, &party.Description, &party.LeaderID, &party.Status, &party.CreatedAt, &party.UpdatedAt)
	if err != nil {
		http.Error(w, "Party not found", http.StatusNotFound)
		return
	}

	// Get leader
	var leader db.Agent
	s.DB.QueryRow(`
		SELECT id, name, title, class, level, xp, gold, energy, status, avatar_emoji, created_at, updated_at
		FROM agents WHERE id = $1
	`, party.LeaderID).Scan(&leader.ID, &leader.Name, &leader.Title, &leader.Class, &leader.Level, &leader.XP, &leader.Gold, &leader.Energy, &leader.Status, &leader.AvatarEmoji, &leader.CreatedAt, &leader.UpdatedAt)

	// Get members
	members := []db.Agent{}
	memberRows, _ := s.DB.Query(`
		SELECT a.id, a.name, a.title, a.class, a.level, a.xp, a.gold, a.energy, a.status, a.avatar_emoji, a.created_at, a.updated_at
		FROM agents a
		JOIN party_members pm ON a.id = pm.agent_id
		WHERE pm.party_id = $1
		ORDER BY pm.joined_at
	`, id)
	defer memberRows.Close()
	for memberRows.Next() {
		var member db.Agent
		memberRows.Scan(&member.ID, &member.Name, &member.Title, &member.Class, &member.Level, &member.XP, &member.Gold, &member.Energy, &member.Status, &member.AvatarEmoji, &member.CreatedAt, &member.UpdatedAt)
		members = append(members, member)
	}

	// Get tasks
	tasks := []db.Task{}
	taskRows, _ := s.DB.Query(`
		SELECT id, name, description, difficulty, xp_reward, gold_reward, status, assigned_agent_id, assigned_party_id, progress, created_at, started_at, completed_at
		FROM tasks WHERE assigned_party_id = $1
		ORDER BY created_at DESC
	`, id)
	defer taskRows.Close()
	for taskRows.Next() {
		var task db.Task
		taskRows.Scan(&task.ID, &task.Name, &task.Description, &task.Difficulty, &task.XPReward, &task.GoldReward, &task.Status, &task.AssignedAgentID, &task.AssignedPartyID, &task.Progress, &task.CreatedAt, &task.StartedAt, &task.CompletedAt)
		tasks = append(tasks, task)
	}

	detail := db.PartyDetail{
		Party:   party,
		Members: members,
		Tasks:   tasks,
		Leader:  leader,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

func (s *Server) addPartyMember(w http.ResponseWriter, r *http.Request, partyID int) {
	if s.DB == nil {
		http.Error(w, "PostgreSQL not configured", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		AgentID int    `json:"agent_id"`
		Role    string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		req.Role = "member"
	}

	_, err := s.DB.Exec(`
		INSERT INTO party_members (party_id, agent_id, role)
		VALUES ($1, $2, $3)
	`, partyID, req.AgentID, req.Role)
	if err != nil {
		log.Printf("Error adding party member: %v", err)
		http.Error(w, "Failed to add member", http.StatusInternalServerError)
		return
	}

	s.Hub.Broadcast(ws.Message{Type: "party_member_added", Data: map[string]interface{}{
		"party_id": partyID,
		"agent_id": req.AgentID,
		"role":     req.Role,
	}})

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) removePartyMember(w http.ResponseWriter, r *http.Request, partyID int) {
	if s.DB == nil {
		http.Error(w, "PostgreSQL not configured", http.StatusServiceUnavailable)
		return
	}

	agentIDStr := r.URL.Query().Get("agent_id")
	agentID, err := strconv.Atoi(agentIDStr)
	if err != nil {
		http.Error(w, "Invalid agent_id", http.StatusBadRequest)
		return
	}

	_, err = s.DB.Exec(`
		DELETE FROM party_members 
		WHERE party_id = $1 AND agent_id = $2
	`, partyID, agentID)
	if err != nil {
		log.Printf("Error removing party member: %v", err)
		http.Error(w, "Failed to remove member", http.StatusInternalServerError)
		return
	}

	s.Hub.Broadcast(ws.Message{Type: "party_member_removed", Data: map[string]interface{}{
		"party_id": partyID,
		"agent_id": agentID,
	}})

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) updateParty(w http.ResponseWriter, r *http.Request, id int) {
	if s.DB == nil {
		http.Error(w, "PostgreSQL not configured", http.StatusServiceUnavailable)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Build dynamic update query
	query := "UPDATE parties SET updated_at = NOW()"
	args := []interface{}{}
	argCount := 1

	if name, ok := updates["name"].(string); ok {
		query += ", name = $" + strconv.Itoa(argCount)
		args = append(args, name)
		argCount++
	}
	if desc, ok := updates["description"].(string); ok {
		query += ", description = $" + strconv.Itoa(argCount)
		args = append(args, desc)
		argCount++
	}
	if status, ok := updates["status"].(string); ok {
		query += ", status = $" + strconv.Itoa(argCount)
		args = append(args, status)
		argCount++
	}

	query += " WHERE id = $" + strconv.Itoa(argCount)
	args = append(args, id)

	if _, err := s.DB.Exec(query, args...); err != nil {
		log.Printf("Error updating party: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	s.Hub.Broadcast(ws.Message{Type: "party_updated", Data: map[string]interface{}{"id": id, "updates": updates}})

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) disbandParty(w http.ResponseWriter, r *http.Request, id int) {
	if s.DB == nil {
		http.Error(w, "PostgreSQL not configured", http.StatusServiceUnavailable)
		return
	}

	// First remove all party members
	_, err := s.DB.Exec("DELETE FROM party_members WHERE party_id = $1", id)
	if err != nil {
		log.Printf("Error removing party members: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Then delete the party
	_, err = s.DB.Exec("DELETE FROM parties WHERE id = $1", id)
	if err != nil {
		log.Printf("Error disbanding party: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	s.Hub.Broadcast(ws.Message{Type: "party_disbanded", Data: map[string]interface{}{"id": id}})

	w.WriteHeader(http.StatusNoContent)
}

// Quest Giver handlers
func (s *Server) handleQuestAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	if s.QuestGiver == nil {
		http.Error(w, "Quest Giver not available", http.StatusServiceUnavailable)
		return
	}
	
	var req struct {
		TaskName        string `json:"task_name"`
		TaskDescription string `json:"task_description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	analysis := s.QuestGiver.AnalyzeTask(req.TaskName, req.TaskDescription)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysis)
}

func (s *Server) handleQuestRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	if s.QuestGiver == nil {
		http.Error(w, "Quest Giver not available", http.StatusServiceUnavailable)
		return
	}
	
	recommendations, err := s.QuestGiver.GetTaskRecommendations()
	if err != nil {
		log.Printf("Error getting quest recommendations: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recommendations)
}

func (s *Server) handleQuestAssign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	if s.QuestGiver == nil {
		http.Error(w, "Quest Giver not available", http.StatusServiceUnavailable)
		return
	}
	
	var req struct {
		TaskID          int    `json:"task_id"`
		TaskName        string `json:"task_name"`
		TaskDescription string `json:"task_description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	err := s.QuestGiver.AutoAssignTask(req.TaskID, req.TaskName, req.TaskDescription)
	if err != nil {
		log.Printf("Error assigning task: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	s.Hub.Broadcast(ws.Message{Type: "task_auto_assigned", Data: map[string]interface{}{
		"task_id": req.TaskID,
		"message": "Task automatically assigned by Quest Giver",
	}})
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Task assigned successfully"})
}

// Daily Quest handlers
func (s *Server) handleDailyQuests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	if s.DailyQuestMgr == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]struct{}{})
		return
	}
	
	quests, err := s.DailyQuestMgr.GetActiveDailyQuests()
	if err != nil {
		log.Printf("Error getting daily quests: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quests)
}

func (s *Server) handleGenerateDailyQuests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	if s.DailyQuestMgr == nil {
		http.Error(w, "Daily Quest Manager not available", http.StatusServiceUnavailable)
		return
	}
	
	err := s.DailyQuestMgr.GenerateDailyQuests()
	if err != nil {
		log.Printf("Error generating daily quests: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Broadcast that new daily quests were generated
	s.Hub.Broadcast(ws.Message{Type: "daily_quests_generated", Data: map[string]string{
		"message": "New daily quests have been generated!",
	}})
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Daily quests generated successfully"})
}

func (s *Server) handleDailyQuestStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	if s.DailyQuestMgr == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{})
		return
	}
	
	stats, err := s.DailyQuestMgr.GetQuestStats()
	if err != nil {
		log.Printf("Error getting daily quest stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// Mood endpoints
func (s *Server) handleMoodUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.MoodCalc == nil {
		http.Error(w, "Mood calculator not available", http.StatusServiceUnavailable)
		return
	}

	err := s.MoodCalc.UpdateAllAgentMoods()
	if err != nil {
		log.Printf("Error updating agent moods: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	s.Hub.Broadcast(ws.Message{Type: "moods_updated", Data: map[string]string{
		"message": "All agent moods have been recalculated",
	}})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Moods updated successfully"})
}

func (s *Server) handleMoodLevels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mood.MoodLevels)
}

func (s *Server) handleReputationRankings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.DB == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]struct{}{})
		return
	}

	// Get top 5 agents per domain by default
	rankings, err := s.DB.GetTopReputationByDomain(5)
	if err != nil {
		log.Printf("Error getting reputation rankings: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rankings)
}

// Conversation handlers
func (s *Server) handleConversations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Get query parameters
	agentID := r.URL.Query().Get("agent")
	limitStr := r.URL.Query().Get("limit")
	
	limit := 20
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 { // Cap at 100
				limit = 100
			}
		}
	}
	
	var conversations []logstack.Conversation
	var err error
	
	if agentID != "" {
		// Get conversations for specific agent
		conversations, err = s.LogstackClient.GetAgentConversations(agentID, limit)
	} else {
		// Get conversations from all agents
		conversations, err = s.LogstackClient.GetAllConversations(limit)
	}
	
	if err != nil {
		log.Printf("Error getting conversations: %v", err)
		// Return empty array instead of error - logstack may not be available
		conversations = []logstack.Conversation{}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversations)
}

func (s *Server) handleConversationDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Extract conversation ID from path
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/conversations/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Conversation ID required", http.StatusBadRequest)
		return
	}
	
	conversationID := pathParts[0]
	
	// For now, we'll need to search through all agents to find the conversation
	// This could be optimized with an index in the future
	agents, err := s.LogstackClient.GetAgentList()
	if err != nil {
		log.Printf("Error getting agent list: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	var foundConversation *logstack.Conversation
	
	for _, agentID := range agents {
		conversations, err := s.LogstackClient.GetAgentConversations(agentID, 100) // Check more conversations
		if err != nil {
			continue
		}
		
		for _, conv := range conversations {
			if conv.ID == conversationID {
				foundConversation = &conv
				break
			}
		}
		
		if foundConversation != nil {
			break
		}
	}
	
	if foundConversation == nil {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(foundConversation)
}

// Webhook handler for spawn events
func (s *Server) handleSpawnWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the spawn event payload
	var spawnEvent struct {
		Type      string `json:"type"`       // "spawn_started", "spawn_completed", "spawn_failed"
		SessionID string `json:"session_id"` // Session ID of the spawn
		AgentID   string `json:"agent_id"`   // Agent that spawned the task
		Task      string `json:"task"`       // Task description
		Label     string `json:"label"`      // Optional label for the spawn
		Data      interface{} `json:"data"`  // Additional spawn data
		Timestamp string `json:"timestamp"`  // ISO timestamp
	}

	if err := json.NewDecoder(r.Body).Decode(&spawnEvent); err != nil {
		log.Printf("Invalid webhook payload: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("📡 Received spawn webhook: %s for agent %s (session: %s)", 
		spawnEvent.Type, spawnEvent.AgentID, spawnEvent.SessionID)

	// Generate a human-readable message based on the event type
	var message string
	var eventIcon string
	switch spawnEvent.Type {
	case "spawn_started":
		eventIcon = "🚀"
		message = eventIcon + " " + spawnEvent.AgentID + " spawned a sub-agent"
		if spawnEvent.Label != "" {
			message += " (" + spawnEvent.Label + ")"
		}
		if len(spawnEvent.Task) > 100 {
			message += ": " + spawnEvent.Task[:100] + "..."
		} else {
			message += ": " + spawnEvent.Task
		}
	case "spawn_completed":
		eventIcon = "✅"
		message = eventIcon + " Sub-agent completed task for " + spawnEvent.AgentID
		if spawnEvent.Label != "" {
			message += " (" + spawnEvent.Label + ")"
		}
	case "spawn_failed":
		eventIcon = "❌"
		message = eventIcon + " Sub-agent failed task for " + spawnEvent.AgentID
		if spawnEvent.Label != "" {
			message += " (" + spawnEvent.Label + ")"
		}
	default:
		eventIcon = "📍"
		message = eventIcon + " Spawn event " + spawnEvent.Type + " for " + spawnEvent.AgentID
	}

	// Broadcast the event via WebSocket
	s.Hub.Broadcast(ws.Message{
		Type: "spawn_event",
		Data: map[string]interface{}{
			"type":       spawnEvent.Type,
			"session_id": spawnEvent.SessionID,
			"agent_id":   spawnEvent.AgentID,
			"task":       spawnEvent.Task,
			"label":      spawnEvent.Label,
			"message":    message,
			"timestamp":  spawnEvent.Timestamp,
			"data":       spawnEvent.Data,
		},
	})

	// Log the event for debugging
	log.Printf("🎯 Spawn Event: %s", message)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Spawn event processed successfully",
	})
}

// handleHealthCheck returns the health status of various services
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type ServiceHealth struct {
		Name         string  `json:"name"`
		Status       string  `json:"status"`       // "running", "warning", "down"
		Health       int     `json:"health"`       // 0-100
		ResponseTime *int    `json:"response_time,omitempty"` // milliseconds
		Error        *string `json:"error,omitempty"`
		Port         *int    `json:"port,omitempty"`
		Version      *string `json:"version,omitempty"`
		Uptime       *string `json:"uptime,omitempty"`
		LastCheck    string  `json:"last_check"`
	}

	type HealthResponse struct {
		Timestamp string          `json:"timestamp"`
		Services  []ServiceHealth `json:"services"`
		Overall   struct {
			Status  string  `json:"status"`
			Healthy int     `json:"healthy"`
			Total   int     `json:"total"`
			Score   float64 `json:"score"`
		} `json:"overall"`
	}

	var response HealthResponse
	response.Timestamp = time.Now().UTC().Format(time.RFC3339)

	// Check Database
	dbHealth := ServiceHealth{
		Name:      "database",
		LastCheck: response.Timestamp,
	}
	if s.DB != nil {
		start := time.Now()
		err := s.DB.Ping()
		responseTime := int(time.Since(start).Milliseconds())
		dbHealth.ResponseTime = &responseTime
		
		if err != nil {
			dbHealth.Status = "down"
			dbHealth.Health = 0
			errStr := err.Error()
			dbHealth.Error = &errStr
		} else {
			dbHealth.Status = "running"
			dbHealth.Health = 100
			port := 5432
			dbHealth.Port = &port
			version := "PostgreSQL"
			dbHealth.Version = &version
		}
	} else {
		dbHealth.Status = "down"
		dbHealth.Health = 0
		errStr := "Database connection not configured"
		dbHealth.Error = &errStr
	}

	// Check inber-party backend (self)
	backendHealth := ServiceHealth{
		Name:      "inber-party-backend",
		Status:    "running",
		Health:    100,
		LastCheck: response.Timestamp,
	}
	port := 8080
	backendHealth.Port = &port
	version := "v1.0.0" // You could read this from a version file
	backendHealth.Version = &version

	// Check OpenClaw Gateway (attempt connection to typical port)
	gatewayHealth := ServiceHealth{
		Name:      "openclaw-gateway",
		LastCheck: response.Timestamp,
	}
	gatewayUrl := "http://localhost:8080/status"  // Adjust based on actual OpenClaw port
	start := time.Now()
	resp, err := http.Get(gatewayUrl)
	responseTime := int(time.Since(start).Milliseconds())
	gatewayHealth.ResponseTime = &responseTime
	
	if err != nil {
		gatewayHealth.Status = "warning"
		gatewayHealth.Health = 50
		errStr := "Connection failed: " + err.Error()
		gatewayHealth.Error = &errStr
	} else {
		resp.Body.Close()
		if resp.StatusCode == 200 {
			gatewayHealth.Status = "running"
			gatewayHealth.Health = 100
		} else {
			gatewayHealth.Status = "warning"
			gatewayHealth.Health = 75
		}
		port := 8080
		gatewayHealth.Port = &port
	}

	// Add services to response
	response.Services = []ServiceHealth{
		backendHealth,
		dbHealth,
		gatewayHealth,
	}

	// Calculate overall health
	totalServices := len(response.Services)
	healthyServices := 0
	totalScore := 0

	for _, service := range response.Services {
		totalScore += service.Health
		if service.Status == "running" {
			healthyServices++
		}
	}

	response.Overall.Total = totalServices
	response.Overall.Healthy = healthyServices
	response.Overall.Score = float64(totalScore) / float64(totalServices)

	if healthyServices == totalServices {
		response.Overall.Status = "healthy"
	} else if healthyServices > 0 {
		response.Overall.Status = "degraded"
	} else {
		response.Overall.Status = "unhealthy"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// calculateGoldReward computes gold reward based on XP and difficulty
func calculateGoldReward(xpReward int, difficulty string) int {
	// Base conversion: 1 XP = 2 gold
	baseGold := xpReward * 2
	
	// Apply difficulty multiplier
	multiplier := 1.0
	switch difficulty {
	case "easy", "1":
		multiplier = 0.8
	case "medium", "2":
		multiplier = 1.0
	case "hard", "3":
		multiplier = 1.5
	case "expert", "4":
		multiplier = 2.0
	case "legendary", "5":
		multiplier = 3.0
	default:
		// Try to parse as integer
		if diff, err := strconv.Atoi(difficulty); err == nil {
			switch diff {
			case 1:
				multiplier = 0.8
			case 2:
				multiplier = 1.0
			case 3:
				multiplier = 1.5
			case 4:
				multiplier = 2.0
			case 5:
				multiplier = 3.0
			}
		}
	}
	
	goldReward := int(float64(baseGold) * multiplier)
	
	// Minimum of 5 gold for any task
	if goldReward < 5 {
		goldReward = 5
	}
	
	return goldReward
}

// handleAgentSync triggers manual agent sync from inber
func (s *Server) handleAgentSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.AgentSync == nil {
		http.Error(w, "Agent sync not available", http.StatusServiceUnavailable)
		return
	}

	result, err := s.AgentSync.SyncAgents()
	if err != nil {
		http.Error(w, "Sync failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleAgentSyncStatus returns the current agent sync status
func (s *Server) handleAgentSyncStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.AgentSync == nil {
		http.Error(w, "Agent sync not available", http.StatusServiceUnavailable)
		return
	}

	status := s.AgentSync.GetSyncStatus()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleManagedAgents returns agents managed in the local database
func (s *Server) handleManagedAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.AgentSync == nil {
		http.Error(w, "Agent sync not available", http.StatusServiceUnavailable)
		return
	}

	agents, err := s.AgentSync.GetManagedAgents()
	if err != nil {
		http.Error(w, "Failed to get managed agents: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

// handleCosts manages cost entries (GET for retrieving, POST for adding)
func (s *Server) handleCosts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listCosts(w, r)
	case http.MethodPost:
		s.createCostEntry(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listCosts retrieves cost entries with optional filters
func (s *Server) listCosts(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]struct{}{})
		return
	}

	// Parse query parameters for filtering
	agentID := r.URL.Query().Get("agent_id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	limit := r.URL.Query().Get("limit")

	if limit == "" {
		limit = "100" // Default limit
	}

	query := `
		SELECT ce.id, ce.agent_id, ce.task_id, ce.session_id, ce.tokens_used, 
			   ce.cost_usd, ce.model_name, ce.operation_type, ce.date, ce.created_at, ce.metadata
		FROM cost_entries ce
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 0

	if agentID != "" {
		argCount++
		query += " AND ce.agent_id = $" + strconv.Itoa(argCount)
		agentIDInt, _ := strconv.Atoi(agentID)
		args = append(args, agentIDInt)
	}

	if startDate != "" {
		argCount++
		query += " AND ce.date >= $" + strconv.Itoa(argCount)
		args = append(args, startDate)
	}

	if endDate != "" {
		argCount++
		query += " AND ce.date <= $" + strconv.Itoa(argCount)
		args = append(args, endDate)
	}

	argCount++
	query += " ORDER BY ce.created_at DESC LIMIT $" + strconv.Itoa(argCount)
	limitInt, _ := strconv.Atoi(limit)
	args = append(args, limitInt)

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		log.Printf("Error listing costs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	costs := []db.CostEntry{}
	for rows.Next() {
		var c db.CostEntry
		if err := rows.Scan(&c.ID, &c.AgentID, &c.TaskID, &c.SessionID, &c.TokensUsed, 
			&c.CostUSD, &c.ModelName, &c.OperationType, &c.Date, &c.CreatedAt, &c.Metadata); err != nil {
			log.Printf("Error scanning cost entry: %v", err)
			continue
		}
		costs = append(costs, c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(costs)
}

// createCostEntry adds a new cost entry
func (s *Server) createCostEntry(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		http.Error(w, "PostgreSQL not configured", http.StatusServiceUnavailable)
		return
	}

	var cost db.CostEntry
	if err := json.NewDecoder(r.Body).Decode(&cost); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set current date if not provided
	if cost.Date == "" {
		cost.Date = time.Now().Format("2006-01-02")
	}

	err := s.DB.QueryRow(`
		INSERT INTO cost_entries (agent_id, task_id, session_id, tokens_used, cost_usd, model_name, operation_type, date, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`, cost.AgentID, cost.TaskID, cost.SessionID, cost.TokensUsed, cost.CostUSD, 
	   cost.ModelName, cost.OperationType, cost.Date, cost.Metadata).Scan(&cost.ID, &cost.CreatedAt)
	
	if err != nil {
		log.Printf("Error creating cost entry: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cost)
}

// handleCostsSummary provides aggregated cost data for dashboards
func (s *Server) handleCostsSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.DB == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]struct{}{})
		return
	}

	// Parse query parameters
	period := r.URL.Query().Get("period") // "daily", "weekly", "monthly"
	if period == "" {
		period = "daily"
	}

	agentID := r.URL.Query().Get("agent_id")
	days := r.URL.Query().Get("days") // Number of periods to include
	if days == "" {
		days = "7" // Default to last 7 periods
	}

	var dateGrouping string
	var dateRange string
	switch period {
	case "weekly":
		dateGrouping = "DATE_TRUNC('week', ce.date)"
		dateRange = "ce.date >= CURRENT_DATE - INTERVAL '" + days + " weeks'"
	case "monthly":
		dateGrouping = "DATE_TRUNC('month', ce.date)"
		dateRange = "ce.date >= CURRENT_DATE - INTERVAL '" + days + " months'"
	default: // daily
		dateGrouping = "ce.date"
		dateRange = "ce.date >= CURRENT_DATE - INTERVAL '" + days + " days'"
	}

	query := `
		SELECT 
			$1 as period,
			ce.agent_id,
			COALESCE(a.name, 'Unknown') as agent_name,
			SUM(ce.cost_usd) as total_cost,
			SUM(ce.tokens_used) as total_tokens,
			COUNT(DISTINCT ce.task_id) as task_count,
			CASE 
				WHEN COUNT(DISTINCT ce.task_id) > 0 
				THEN SUM(ce.cost_usd) / COUNT(DISTINCT ce.task_id) 
				ELSE 0 
			END as avg_cost_per_task,
			TO_CHAR(` + dateGrouping + `, 'YYYY-MM-DD') as date_range
		FROM cost_entries ce
		LEFT JOIN agents a ON ce.agent_id = a.id
		WHERE ` + dateRange
	
	args := []interface{}{period}
	argCount := 1

	if agentID != "" {
		argCount++
		query += " AND ce.agent_id = $" + strconv.Itoa(argCount)
		agentIDInt, _ := strconv.Atoi(agentID)
		args = append(args, agentIDInt)
	}

	query += `
		GROUP BY ce.agent_id, a.name, ` + dateGrouping + `
		ORDER BY date_range DESC, total_cost DESC
	`

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		log.Printf("Error getting cost summary: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	summaries := []db.CostSummary{}
	for rows.Next() {
		var s db.CostSummary
		if err := rows.Scan(&s.Period, &s.AgentID, &s.AgentName, &s.TotalCost, 
			&s.TotalTokens, &s.TaskCount, &s.AvgCostPerTask, &s.DateRange); err != nil {
			log.Printf("Error scanning cost summary: %v", err)
			continue
		}
		summaries = append(summaries, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}

// Bounty marketplace handlers

func (s *Server) handleBounties(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listBounties(w, r)
	case http.MethodPost:
		s.createBounty(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleBountyDetail(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/bounties/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Bounty ID required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(pathParts[0])
	if err != nil {
		http.Error(w, "Invalid bounty ID", http.StatusBadRequest)
		return
	}

	// Handle sub-routes
	if len(pathParts) > 1 {
		action := pathParts[1]
		switch action {
		case "claim":
			if r.Method == http.MethodPost {
				s.claimBounty(w, r, id)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		case "submit":
			if r.Method == http.MethodPost {
				s.submitWork(w, r, id)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		case "verify":
			if r.Method == http.MethodPost {
				s.verifyBounty(w, r, id)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}
	}

	switch r.Method {
	case http.MethodGet:
		s.getBountyDetail(w, r, id)
	case http.MethodPatch:
		s.updateBounty(w, r, id)
	case http.MethodDelete:
		s.deleteBounty(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listBounties(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]struct{}{})
		return
	}

	// Parse query parameters
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	bounties, err := s.DB.GetBounties(status, limit, offset)
	if err != nil {
		log.Printf("Error listing bounties: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bounties)
}

func (s *Server) createBounty(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		http.Error(w, "PostgreSQL not configured", http.StatusServiceUnavailable)
		return
	}

	var bounty db.Bounty
	if err := json.NewDecoder(r.Body).Decode(&bounty); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if bounty.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}
	if bounty.Description == "" {
		http.Error(w, "Description is required", http.StatusBadRequest)
		return
	}
	if bounty.CreatorID == 0 {
		http.Error(w, "Creator ID is required", http.StatusBadRequest)
		return
	}
	if bounty.PayoutAmount <= 0 {
		http.Error(w, "Payout amount must be positive", http.StatusBadRequest)
		return
	}

	err := s.DB.CreateBounty(&bounty)
	if err != nil {
		log.Printf("Error creating bounty: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Broadcast bounty creation
	s.Hub.Broadcast(ws.Message{Type: "bounty_created", Data: bounty})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bounty)
}

func (s *Server) getBountyDetail(w http.ResponseWriter, r *http.Request, id int) {
	if s.DB == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	detail, err := s.DB.GetBountyDetail(id)
	if err != nil {
		log.Printf("Error getting bounty detail: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if detail == nil {
		http.Error(w, "Bounty not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detail)
}

func (s *Server) claimBounty(w http.ResponseWriter, r *http.Request, bountyID int) {
	if s.DB == nil {
		http.Error(w, "PostgreSQL not configured", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		ClaimerID int `json:"claimer_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ClaimerID <= 0 {
		http.Error(w, "Claimer ID is required", http.StatusBadRequest)
		return
	}

	err := s.DB.ClaimBounty(bountyID, req.ClaimerID)
	if err != nil {
		log.Printf("Error claiming bounty: %v", err)
		if err.Error() == "bounty not found or already claimed" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Broadcast bounty claimed
	s.Hub.Broadcast(ws.Message{Type: "bounty_claimed", Data: map[string]interface{}{
		"bounty_id":  bountyID,
		"claimer_id": req.ClaimerID,
	}})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "claimed"})
}

func (s *Server) submitWork(w http.ResponseWriter, r *http.Request, bountyID int) {
	if s.DB == nil {
		http.Error(w, "PostgreSQL not configured", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		WorkSubmission string `json:"work_submission"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.WorkSubmission == "" {
		http.Error(w, "Work submission is required", http.StatusBadRequest)
		return
	}

	err := s.DB.SubmitWork(bountyID, req.WorkSubmission)
	if err != nil {
		log.Printf("Error submitting work: %v", err)
		if err.Error() == "bounty not found or not in claimed status" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Broadcast work submitted
	s.Hub.Broadcast(ws.Message{Type: "bounty_work_submitted", Data: map[string]interface{}{
		"bounty_id": bountyID,
	}})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "submitted"})
}

func (s *Server) verifyBounty(w http.ResponseWriter, r *http.Request, bountyID int) {
	if s.DB == nil {
		http.Error(w, "PostgreSQL not configured", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Approved bool   `json:"approved"`
		Notes    string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := s.DB.VerifyBounty(bountyID, req.Approved, req.Notes)
	if err != nil {
		log.Printf("Error verifying bounty: %v", err)
		if err.Error() == "bounty not found or not in submitted status" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Broadcast bounty verified
	s.Hub.Broadcast(ws.Message{Type: "bounty_verified", Data: map[string]interface{}{
		"bounty_id": bountyID,
		"approved":  req.Approved,
	}})

	status := "rejected"
	if req.Approved {
		status = "completed"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": status})
}

func (s *Server) updateBounty(w http.ResponseWriter, r *http.Request, id int) {
	if s.DB == nil {
		http.Error(w, "PostgreSQL not configured", http.StatusServiceUnavailable)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// For now, implement basic updates (title, description, requirements)
	// More complex status changes should use specific endpoints (claim, submit, verify)
	allowedUpdates := []string{"title", "description", "requirements", "deadline", "payout_amount"}
	query := "UPDATE bounties SET updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{}
	argCount := 1

	for field, value := range updates {
		found := false
		for _, allowed := range allowedUpdates {
			if field == allowed {
				found = true
				break
			}
		}
		if !found {
			continue
		}

		query += ", " + field + " = $" + strconv.Itoa(argCount)
		args = append(args, value)
		argCount++
	}

	query += " WHERE id = $" + strconv.Itoa(argCount) + " AND status = 'open'"
	args = append(args, id)

	result, err := s.DB.Exec(query, args...)
	if err != nil {
		log.Printf("Error updating bounty: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Bounty not found or cannot be updated", http.StatusNotFound)
		return
	}

	s.Hub.Broadcast(ws.Message{Type: "bounty_updated", Data: map[string]interface{}{"id": id, "updates": updates}})

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) deleteBounty(w http.ResponseWriter, r *http.Request, id int) {
	if s.DB == nil {
		http.Error(w, "PostgreSQL not configured", http.StatusServiceUnavailable)
		return
	}

	// Only allow deletion of open bounties
	result, err := s.DB.Exec("DELETE FROM bounties WHERE id = $1 AND status = 'open'", id)
	if err != nil {
		log.Printf("Error deleting bounty: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Bounty not found or cannot be deleted", http.StatusNotFound)
		return
	}

	s.Hub.Broadcast(ws.Message{Type: "bounty_deleted", Data: map[string]interface{}{"id": id}})

	w.WriteHeader(http.StatusNoContent)
}

// Payout tracking handlers

func (s *Server) handlePayouts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.listPayouts(w, r)
	case http.MethodPost:
		s.createManualPayout(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) listPayouts(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]struct{}{})
		return
	}

	// Parse query parameters
	agentIDStr := r.URL.Query().Get("agent_id")
	source := r.URL.Query().Get("source")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	var agentID *int
	if agentIDStr != "" {
		if parsed, err := strconv.Atoi(agentIDStr); err == nil {
			agentID = &parsed
		}
	}

	limit := 50 // default
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	entries, err := s.DB.GetPayoutEntries(agentID, source, limit, offset)
	if err != nil {
		log.Printf("Error getting payout entries: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func (s *Server) createManualPayout(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		http.Error(w, "PostgreSQL not configured", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		AgentID     int    `json:"agent_id"`
		Amount      int    `json:"amount"`
		Description string `json:"description"`
		ProcessedBy *int   `json:"processed_by,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AgentID <= 0 {
		http.Error(w, "Agent ID is required", http.StatusBadRequest)
		return
	}
	if req.Amount == 0 {
		http.Error(w, "Amount cannot be zero", http.StatusBadRequest)
		return
	}

	err := s.DB.RecordManualAdjustment(req.AgentID, req.Amount, req.Description, req.ProcessedBy)
	if err != nil {
		log.Printf("Error creating manual payout: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Broadcast the adjustment
	s.Hub.Broadcast(ws.Message{Type: "manual_payout", Data: map[string]interface{}{
		"agent_id":    req.AgentID,
		"amount":      req.Amount,
		"description": req.Description,
	}})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Manual payout recorded"})
}

func (s *Server) handlePayoutSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.DB == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]struct{}{})
		return
	}

	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 25 // default
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	summaries, err := s.DB.GetPayoutSummary(limit)
	if err != nil {
		log.Printf("Error getting payout summary: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}

func (s *Server) handleAgentPayouts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.DB == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Extract agent ID from path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/payouts/agent/")
	agentID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	entries, err := s.DB.GetAgentPayoutHistory(agentID, limit)
	if err != nil {
		log.Printf("Error getting agent payout history: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

// Verifier system handlers

func (s *Server) handleVerifierTypes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.VerifierRegistry == nil {
		http.Error(w, "Verifier registry not available", http.StatusServiceUnavailable)
		return
	}

	response := map[string]interface{}{
		"types":   s.VerifierRegistry.ListTypes(),
		"schemas": s.VerifierRegistry.GetConfigSchemas(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleRunVerifiers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.DB == nil {
		http.Error(w, "Database not available", http.StatusServiceUnavailable)
		return
	}

	if s.VerifierRegistry == nil {
		http.Error(w, "Verifier registry not available", http.StatusServiceUnavailable)
		return
	}

	// Extract bounty ID from path
	bountyIDStr := strings.TrimPrefix(r.URL.Path, "/api/verifiers/run/")
	bountyID, err := strconv.Atoi(bountyIDStr)
	if err != nil {
		http.Error(w, "Invalid bounty ID", http.StatusBadRequest)
		return
	}

	// Get bounty with verifiers
	bountyWithVerifiers, err := s.DB.GetBountyWithVerifiers(bountyID)
	if err != nil {
		log.Printf("Error getting bounty with verifiers: %v", err)
		http.Error(w, "Failed to get bounty", http.StatusInternalServerError)
		return
	}
	if bountyWithVerifiers == nil {
		http.Error(w, "Bounty not found", http.StatusNotFound)
		return
	}

	// Check if bounty is in submitted status
	if bountyWithVerifiers.Status != "submitted" {
		http.Error(w, "Bounty must be in submitted status to run verifiers", http.StatusBadRequest)
		return
	}

	var runResults []map[string]interface{}
	allPassed := true

	// Run each verifier
	for _, verifierDef := range bountyWithVerifiers.Verifiers {
		verifier, err := s.VerifierRegistry.GetVerifier(verifierDef.VerifierType)
		if err != nil {
			log.Printf("Error getting verifier %s: %v", verifierDef.VerifierType, err)
			continue
		}

		// Create verification context
		workSubmission := ""
		if bountyWithVerifiers.WorkSubmission != nil {
			workSubmission = *bountyWithVerifiers.WorkSubmission
		}
		
		ctx := verifiers.VerificationContext{
			BountyID:       bountyID,
			WorkSubmission: workSubmission,
			BountyTitle:    bountyWithVerifiers.Title,
			Requirements:   bountyWithVerifiers.Requirements,
			ClaimerName:    "",
			Config:         verifierDef.Config,
		}

		if bountyWithVerifiers.Claimer != nil {
			ctx.ClaimerName = bountyWithVerifiers.Claimer.Name
		}

		// Run verifier
		result := verifier.Verify(ctx)

		// Save result to database
		dbResult := &db.VerifierResult{
			BountyID:     bountyID,
			VerifierID:   verifierDef.ID,
			Status:       result.Status,
			ResultData:   result.Details,
			ErrorMessage: result.ErrorMessage,
			CheckedBy:    stringPtr("system"),
		}

		if err := s.DB.CreateVerifierResult(dbResult); err != nil {
			log.Printf("Error saving verifier result: %v", err)
		}

		// Check if this verifier passed (or is manual and pending)
		if verifierDef.Required && result.Status != "passed" && result.Status != "pending" {
			allPassed = false
		}

		runResults = append(runResults, map[string]interface{}{
			"verifier_id":   verifierDef.ID,
			"verifier_type": verifierDef.VerifierType,
			"status":        result.Status,
			"message":       result.Message,
			"required":      verifierDef.Required,
		})
	}

	// Determine overall verification status
	overallStatus := "verification_pending"
	if allPassed {
		// Check if all verifiers are complete (no pending ones)
		allComplete := true
		for _, result := range runResults {
			if result["status"] == "pending" {
				allComplete = false
				break
			}
		}
		
		if allComplete {
			overallStatus = "verification_passed"
			
			// Automatically approve the bounty if all verifiers passed
			err := s.DB.VerifyBounty(bountyID, true, "All automated verifications passed")
			if err != nil {
				log.Printf("Error auto-approving bounty: %v", err)
			} else {
				s.Hub.Broadcast(ws.Message{Type: "bounty_auto_verified", Data: map[string]interface{}{
					"bounty_id": bountyID,
					"approved":  true,
					"message":   "All verifications passed automatically",
				}})
			}
		}
	} else {
		overallStatus = "verification_failed"
		
		// Automatically reject the bounty if required verifiers failed
		err := s.DB.VerifyBounty(bountyID, false, "Required verifications failed")
		if err != nil {
			log.Printf("Error auto-rejecting bounty: %v", err)
		} else {
			s.Hub.Broadcast(ws.Message{Type: "bounty_auto_verified", Data: map[string]interface{}{
				"bounty_id": bountyID,
				"approved":  false,
				"message":   "Required verifications failed",
			}})
		}
	}

	response := map[string]interface{}{
		"bounty_id":        bountyID,
		"overall_status":   overallStatus,
		"verification_results": runResults,
		"all_passed":       allPassed,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Rating system handlers

func (s *Server) handleRatings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.createRating(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) createRating(w http.ResponseWriter, r *http.Request) {
	if s.DB == nil {
		http.Error(w, "PostgreSQL not configured", http.StatusServiceUnavailable)
		return
	}

	var rating db.BountyRating
	if err := json.NewDecoder(r.Body).Decode(&rating); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate rating
	if rating.Rating < 1 || rating.Rating > 5 {
		http.Error(w, "Rating must be between 1 and 5", http.StatusBadRequest)
		return
	}
	if rating.BountyID <= 0 || rating.RaterID <= 0 || rating.RatedID <= 0 {
		http.Error(w, "Valid bounty_id, rater_id, and rated_id are required", http.StatusBadRequest)
		return
	}

	err := s.DB.CreateBountyRating(&rating)
	if err != nil {
		log.Printf("Error creating rating: %v", err)
		if err.Error() == "bounty already rated" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else if err.Error() == "cannot rate incomplete bounty" || err.Error() == "only bounty creator can rate the work" {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Broadcast rating created
	s.Hub.Broadcast(ws.Message{Type: "rating_created", Data: map[string]interface{}{
		"bounty_id": rating.BountyID,
		"rated_id":  rating.RatedID,
		"rating":    rating.Rating,
	}})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rating)
}

func (s *Server) handleAgentRatings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.DB == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Extract agent ID from path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/ratings/agent/")
	agentID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // default
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	ratings, err := s.DB.GetAgentRatings(agentID, limit, offset)
	if err != nil {
		log.Printf("Error getting agent ratings: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ratings)
}

func (s *Server) handleRatingStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.DB == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Extract agent ID from path
	idStr := strings.TrimPrefix(r.URL.Path, "/api/ratings/stats/")
	agentID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid agent ID", http.StatusBadRequest)
		return
	}

	stats, err := s.DB.GetAgentRatingStats(agentID)
	if err != nil {
		log.Printf("Error getting rating stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleRatingLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.DB == nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	// Parse query parameters
	domain := r.URL.Query().Get("domain")
	limitStr := r.URL.Query().Get("limit")

	limit := 10 // default
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	agents, err := s.DB.GetHighestRatedAgents(domain, limit)
	if err != nil {
		log.Printf("Error getting highest rated agents: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

func stringPtr(s string) *string {
	return &s
}
