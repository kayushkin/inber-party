package inber

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// Handler serves the inber RPG API endpoints.
type Handler struct {
	store *Store
}

// NewHandler creates a new API handler backed by the inber store.
func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

// RegisterRoutes adds /api/inber/* routes to the mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/inber/agents", h.handleAgents)
	mux.HandleFunc("/api/inber/quests", h.handleQuests)
	mux.HandleFunc("/api/inber/stats", h.handleStats)
}

func (h *Handler) handleAgents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agents, err := h.store.GetAgents()
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

	quests, err := h.store.GetQuests(limit)
	if err != nil {
		log.Printf("Error getting inber quests: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Filter by agent if requested
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

func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := h.store.GetStats()
	if err != nil {
		log.Printf("Error getting inber stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
