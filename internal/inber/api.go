package inber

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

// DataSource abstracts both SQLite Store and HTTP client.
type DataSource interface {
	// GetAgents returns the agents and the number of source rows it could not read.
	//
	// The count is part of the interface, not of one implementation, because BOTH
	// implementations drop rows and both publish the shortened slice's length as a fact.
	// Store skips a row whose columns will not scan; HTTPClient skips a remote agent with no
	// name. Each counts its own drops, so the number is always a claim the implementation is
	// entitled to make -- there is no arm here that has to return a zero it cannot stand behind.
	GetAgents() ([]RPGAgent, int, error)
	GetQuests(limit int) ([]RPGQuest, error)
	GetStats() (*RPGStats, error)
	GetAchievements(agentID string) ([]RPGAchievement, error)
	GetQuestHistory(agentID string, limit int) ([]QuestHistoryEntry, error)
	GetConversations(limit int) ([]RPGConversation, error)
	GetSessionReplay(sessionID string) (*SessionReplay, error)
	GetAgentJournal(agentID string, date string) (*RPGJournal, error)
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
	mux.HandleFunc("/api/inber/agent-journal", h.handleAgentJournal)
}

func (h *Handler) handleAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agents, unreadableRows, err := h.source.GetAgents()
	if err != nil {
		log.Printf("Error getting inber agents: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	setUnreadableRowsHeader(w, unreadableRows)
	json.NewEncoder(w).Encode(agents)
}

// unreadableRowsHeader names the number of source rows a list response could not read.
//
// It is a header rather than a member of the body because /api/inber/agents answers a bare
// JSON array. Turning that into an object would be a wire change on a route the frontend
// already reads, which is a decision rather than a repair; a header is additive, so a caller
// that wants the array is unaffected and a caller that wants to know whether the array is the
// whole answer now has somewhere to look. /api/inber/stats answers an object already, so it
// carries the same number as the unreadable_agent_rows member instead.
const unreadableRowsHeader = "X-Unreadable-Rows"

// setUnreadableRowsHeader reports how many rows a list response lost.
//
// It is written on every response INCLUDING zero, deliberately. An absent header would mean two
// different things -- "nothing was lost" and "this build does not report" -- and a reader that
// cannot tell those apart is back where this repair started. Present-and-zero is a claim;
// absent is a build that makes no claim.
func setUnreadableRowsHeader(w http.ResponseWriter, unreadableRows int) {
	w.Header().Set(unreadableRowsHeader, strconv.Itoa(unreadableRows))
	if unreadableRows > 0 {
		log.Printf("inber agents: %d row(s) could not be read and are missing from this response", unreadableRows)
	}
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

func (h *Handler) handleAgentJournal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentID := r.URL.Query().Get("agent")
	if agentID == "" {
		http.Error(w, "agent parameter required", http.StatusBadRequest)
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		// Default to today
		date = time.Now().Format("2006-01-02")
	}

	journal, err := h.source.GetAgentJournal(agentID, date)
	if err != nil {
		log.Printf("Error getting agent journal: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(journal)
}
