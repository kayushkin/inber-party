package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/kayushkin/inber-party/internal/dailyquests"
	"github.com/kayushkin/inber-party/internal/db"
	"github.com/kayushkin/inber-party/internal/mood"
	"github.com/kayushkin/inber-party/internal/questgiver"
	"github.com/kayushkin/inber-party/internal/ws"
)

type Server struct {
	DB               *db.DB
	Hub              *ws.Hub
	QuestGiver       *questgiver.QuestGiver
	DailyQuestMgr    *dailyquests.DailyQuestManager
	MoodCalc         *mood.MoodCalculator
}

func NewServer(database *db.DB, hub *ws.Hub, qg *questgiver.QuestGiver, dqm *dailyquests.DailyQuestManager) *Server {
	moodCalc := mood.NewMoodCalculator(database)
	return &Server{DB: database, Hub: hub, QuestGiver: qg, DailyQuestMgr: dqm, MoodCalc: moodCalc}
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
	mux.HandleFunc("/api/quest-giver/assign", s.handleQuestAssign)
	mux.HandleFunc("/api/daily-quests", s.handleDailyQuests)
	mux.HandleFunc("/api/daily-quests/generate", s.handleGenerateDailyQuests)
	mux.HandleFunc("/api/daily-quests/stats", s.handleDailyQuestStats)
	mux.HandleFunc("/api/mood/update", s.handleMoodUpdate)
	mux.HandleFunc("/api/mood/levels", s.handleMoodLevels)
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
		SELECT id, name, title, class, level, xp, energy, status, avatar_emoji, mood, mood_score, workload, last_active, created_at, updated_at
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
		if err := rows.Scan(&a.ID, &a.Name, &a.Title, &a.Class, &a.Level, &a.XP, &a.Energy, &a.Status, &a.AvatarEmoji, &a.Mood, &a.MoodScore, &a.Workload, &a.LastActive, &a.CreatedAt, &a.UpdatedAt); err != nil {
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
		SELECT id, name, title, class, level, xp, energy, status, avatar_emoji, mood, mood_score, workload, last_active, created_at, updated_at
		FROM agents WHERE id = $1
	`, id).Scan(&agent.ID, &agent.Name, &agent.Title, &agent.Class, &agent.Level, &agent.XP, &agent.Energy, &agent.Status, &agent.AvatarEmoji, &agent.Mood, &agent.MoodScore, &agent.Workload, &agent.LastActive, &agent.CreatedAt, &agent.UpdatedAt)
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
		SELECT id, name, description, difficulty, xp_reward, status, assigned_agent_id, assigned_party_id, progress, created_at, started_at, completed_at
		FROM tasks WHERE assigned_agent_id = $1
		ORDER BY created_at DESC
	`, id)
	defer taskRows.Close()
	for taskRows.Next() {
		var task db.Task
		taskRows.Scan(&task.ID, &task.Name, &task.Description, &task.Difficulty, &task.XPReward, &task.Status, &task.AssignedAgentID, &task.AssignedPartyID, &task.Progress, &task.CreatedAt, &task.StartedAt, &task.CompletedAt)
		tasks = append(tasks, task)
	}

	detail := db.AgentDetail{
		Agent:        agent,
		Skills:       skills,
		Achievements: achievements,
		Tasks:        tasks,
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

	err := s.DB.QueryRow(`
		INSERT INTO agents (name, title, class, level, xp, avatar_emoji)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`, agent.Name, agent.Title, agent.Class, agent.Level, agent.XP, agent.AvatarEmoji).Scan(&agent.ID, &agent.CreatedAt, &agent.UpdatedAt)
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
		SELECT id, name, description, difficulty, xp_reward, status, assigned_agent_id, assigned_party_id, progress, created_at, started_at, completed_at
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
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Difficulty, &t.XPReward, &t.Status, &t.AssignedAgentID, &t.AssignedPartyID, &t.Progress, &t.CreatedAt, &t.StartedAt, &t.CompletedAt); err != nil {
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

	err := s.DB.QueryRow(`
		INSERT INTO tasks (name, description, difficulty, xp_reward)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`, task.Name, task.Description, task.Difficulty, task.XPReward).Scan(&task.ID, &task.CreatedAt)
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
			// Update last active time for completed tasks
			if status == "completed" {
				s.MoodCalc.UpdateAgentLastActive(int(agentID))
			}
			
			// Recalculate mood for this agent
			mood, moodScore, err := s.MoodCalc.CalculateAgentMood(int(agentID))
			if err == nil {
				var workload int
				s.DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE assigned_agent_id = $1 AND status = 'in_progress'", int(agentID)).Scan(&workload)
				
				s.DB.Exec(`
					UPDATE agents 
					SET mood = $1, mood_score = $2, workload = $3, updated_at = NOW()
					WHERE id = $4
				`, mood, moodScore, workload, int(agentID))
				
				s.Hub.Broadcast(ws.Message{Type: "agent_mood_updated", Data: map[string]interface{}{
					"agent_id": int(agentID),
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
	s.DB.QueryRow("SELECT COALESCE(AVG(level), 0) FROM agents").Scan(&stats.AverageAgentLevel)
	s.DB.QueryRow("SELECT COUNT(*) FROM parties").Scan(&stats.TotalParties)
	s.DB.QueryRow("SELECT COUNT(*) FROM parties WHERE status = 'active' OR status = 'on_quest'").Scan(&stats.ActiveParties)

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
		SELECT id, name, title, class, level, xp, energy, status, avatar_emoji, created_at, updated_at
		FROM agents WHERE id = $1
	`, party.LeaderID).Scan(&leader.ID, &leader.Name, &leader.Title, &leader.Class, &leader.Level, &leader.XP, &leader.Energy, &leader.Status, &leader.AvatarEmoji, &leader.CreatedAt, &leader.UpdatedAt)

	// Get members
	members := []db.Agent{}
	memberRows, _ := s.DB.Query(`
		SELECT a.id, a.name, a.title, a.class, a.level, a.xp, a.energy, a.status, a.avatar_emoji, a.created_at, a.updated_at
		FROM agents a
		JOIN party_members pm ON a.id = pm.agent_id
		WHERE pm.party_id = $1
		ORDER BY pm.joined_at
	`, id)
	defer memberRows.Close()
	for memberRows.Next() {
		var member db.Agent
		memberRows.Scan(&member.ID, &member.Name, &member.Title, &member.Class, &member.Level, &member.XP, &member.Energy, &member.Status, &member.AvatarEmoji, &member.CreatedAt, &member.UpdatedAt)
		members = append(members, member)
	}

	// Get tasks
	tasks := []db.Task{}
	taskRows, _ := s.DB.Query(`
		SELECT id, name, description, difficulty, xp_reward, status, assigned_agent_id, assigned_party_id, progress, created_at, started_at, completed_at
		FROM tasks WHERE assigned_party_id = $1
		ORDER BY created_at DESC
	`, id)
	defer taskRows.Close()
	for taskRows.Next() {
		var task db.Task
		taskRows.Scan(&task.ID, &task.Name, &task.Description, &task.Difficulty, &task.XPReward, &task.Status, &task.AssignedAgentID, &task.AssignedPartyID, &task.Progress, &task.CreatedAt, &task.StartedAt, &task.CompletedAt)
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
