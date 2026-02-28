package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/kayushkin/inber-party/internal/db"
	"github.com/kayushkin/inber-party/internal/ws"
)

type Server struct {
	DB  *db.DB
	Hub *ws.Hub
}

func NewServer(database *db.DB, hub *ws.Hub) *Server {
	return &Server{DB: database, Hub: hub}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/agents", s.handleAgents)
	mux.HandleFunc("/api/agents/", s.handleAgentDetail)
	mux.HandleFunc("/api/tasks", s.handleTasks)
	mux.HandleFunc("/api/tasks/", s.handleTaskDetail)
	mux.HandleFunc("/api/stats", s.handleStats)
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
	rows, err := s.DB.Query(`
		SELECT id, name, title, class, level, xp, energy, status, avatar_emoji, created_at, updated_at
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
		if err := rows.Scan(&a.ID, &a.Name, &a.Title, &a.Class, &a.Level, &a.XP, &a.Energy, &a.Status, &a.AvatarEmoji, &a.CreatedAt, &a.UpdatedAt); err != nil {
			log.Printf("Error scanning agent: %v", err)
			continue
		}
		agents = append(agents, a)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

func (s *Server) getAgentDetail(w http.ResponseWriter, r *http.Request, id int) {
	var agent db.Agent
	err := s.DB.QueryRow(`
		SELECT id, name, title, class, level, xp, energy, status, avatar_emoji, created_at, updated_at
		FROM agents WHERE id = $1
	`, id).Scan(&agent.ID, &agent.Name, &agent.Title, &agent.Class, &agent.Level, &agent.XP, &agent.Energy, &agent.Status, &agent.AvatarEmoji, &agent.CreatedAt, &agent.UpdatedAt)
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
		SELECT id, name, description, difficulty, xp_reward, status, assigned_agent_id, progress, created_at, started_at, completed_at
		FROM tasks WHERE assigned_agent_id = $1
		ORDER BY created_at DESC
	`, id)
	defer taskRows.Close()
	for taskRows.Next() {
		var task db.Task
		taskRows.Scan(&task.ID, &task.Name, &task.Description, &task.Difficulty, &task.XPReward, &task.Status, &task.AssignedAgentID, &task.Progress, &task.CreatedAt, &task.StartedAt, &task.CompletedAt)
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
	rows, err := s.DB.Query(`
		SELECT id, name, description, difficulty, xp_reward, status, assigned_agent_id, progress, created_at, started_at, completed_at
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
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Difficulty, &t.XPReward, &t.Status, &t.AssignedAgentID, &t.Progress, &t.CreatedAt, &t.StartedAt, &t.CompletedAt); err != nil {
			log.Printf("Error scanning task: %v", err)
			continue
		}
		tasks = append(tasks, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (s *Server) updateTask(w http.ResponseWriter, r *http.Request, id int) {
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

	s.Hub.Broadcast(ws.Message{Type: "task_updated", Data: map[string]interface{}{"id": id, "updates": updates}})

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	var stats db.Stats

	s.DB.QueryRow("SELECT COUNT(*) FROM agents").Scan(&stats.TotalAgents)
	s.DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE status = 'in_progress'").Scan(&stats.ActiveTasks)
	s.DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE status = 'completed'").Scan(&stats.CompletedTasks)
	s.DB.QueryRow("SELECT COALESCE(SUM(xp), 0) FROM agents").Scan(&stats.TotalXP)
	s.DB.QueryRow("SELECT COALESCE(AVG(level), 0) FROM agents").Scan(&stats.AverageAgentLevel)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
