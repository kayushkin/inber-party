package inber

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// DataSource abstracts both SQLite Store and HTTP client.
type DataSource interface {
	GetAgents() ([]RPGAgent, error)
	GetQuests(limit int) ([]RPGQuest, error)
	GetStats() (*RPGStats, error)
	GetAchievements(agentID string) ([]RPGAchievement, error)
	GetQuestHistory(agentID string, limit int) ([]QuestHistoryEntry, error)
	GetConversations(limit int) ([]RPGConversation, error)
	GetSessionReplay(sessionID string) (*SessionReplay, error)
}

// Ensure both implement DataSource
var _ DataSource = (*Store)(nil)
var _ DataSource = (*HTTPClient)(nil)

// Handler serves the inber RPG API endpoints.
type Handler struct {
	source DataSource
}

// NewHandler creates a new API handler backed by any DataSource.
func NewHandler(source DataSource) *Handler {
	return &Handler{source: source}
}

// RegisterRoutes adds /api/inber/* routes to the mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/inber/agents", h.handleAgents)
	mux.HandleFunc("/api/inber/quests", h.handleQuests)
	mux.HandleFunc("/api/inber/stats", h.handleStats)
	mux.HandleFunc("/api/inber/achievements", h.handleAchievements)
	mux.HandleFunc("/api/inber/quest-history", h.handleQuestHistory)
	mux.HandleFunc("/api/inber/conversations", h.handleConversations)
	mux.HandleFunc("/api/inber/session-replay", h.handleSessionReplay)
}

func (h *Handler) handleAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agents, err := h.source.GetAgents()
	if err != nil {
		log.Printf("Error getting inber agents: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agents)
}

func (h *Handler) handleQuests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	agentFilter := r.URL.Query().Get("agent")

	quests, err := h.source.GetQuests(limit)
	if err != nil {
		log.Printf("Error getting inber quests: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if agentFilter != "" {
		filtered := make([]RPGQuest, 0)
		for _, q := range quests {
			if q.AgentID == agentFilter {
				filtered = append(filtered, q)
			}
		}
		quests = filtered
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quests)
}

func (h *Handler) handleAchievements(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	agentID := r.URL.Query().Get("agent")
	if agentID == "" {
		http.Error(w, "agent parameter required", http.StatusBadRequest)
		return
	}
	achievements, err := h.source.GetAchievements(agentID)
	if err != nil {
		log.Printf("Error getting achievements: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(achievements)
}

func (h *Handler) handleQuestHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	agentID := r.URL.Query().Get("agent")
	if agentID == "" {
		http.Error(w, "agent parameter required", http.StatusBadRequest)
		return
	}
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	history, err := h.source.GetQuestHistory(agentID, limit)
	if err != nil {
		log.Printf("Error getting quest history: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := h.source.GetStats()
	if err != nil {
		log.Printf("Error getting inber stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *Handler) handleConversations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	conversations, err := h.source.GetConversations(limit)
	if err != nil {
		log.Printf("Error getting conversations: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversations)
}

func (h *Handler) handleSessionReplay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		http.Error(w, "session parameter required", http.StatusBadRequest)
		return
	}

	replay, err := h.source.GetSessionReplay(sessionID)
	if err != nil {
		log.Printf("Error getting session replay: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(replay)
}
