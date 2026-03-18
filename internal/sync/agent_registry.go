// Package sync handles synchronization between external agent data sources and the local database
package sync

import (
	"fmt"
	"log"
	"time"

	"github.com/kayushkin/inber-party/internal/db"
	"github.com/kayushkin/inber-party/internal/inber"
)

// AgentRegistrySync coordinates agent discovery and sync between inber and the local database
type AgentRegistrySync struct {
	inberSource inber.DataSource
	syncService *db.AgentSyncService
	lastSyncAt  time.Time
}

// NewAgentRegistrySync creates a new agent registry sync coordinator
func NewAgentRegistrySync(inberSource inber.DataSource, database *db.DB) *AgentRegistrySync {
	return &AgentRegistrySync{
		inberSource: inberSource,
		syncService: database.NewAgentSyncService(),
		lastSyncAt:  time.Time{},
	}
}

// SyncAgents discovers agents from inber and syncs them to the local database
func (ars *AgentRegistrySync) SyncAgents() (*SyncResult, error) {
	if ars.inberSource == nil {
		return nil, fmt.Errorf("no inber source configured")
	}

	log.Println("🔄 Starting agent registry sync...")
	
	// Get agents from inber
	inberAgents, err := ars.inberSource.GetAgents()
	if err != nil {
		return nil, fmt.Errorf("failed to get agents from inber: %w", err)
	}

	// Convert inber agents to sync format
	syncAgents := make([]db.InberAgent, 0, len(inberAgents))
	activeAgentNames := make([]string, 0, len(inberAgents))
	
	for _, agent := range inberAgents {
		syncAgent := db.InberAgent{
			ID:           agent.ID,
			Name:         agent.Name,
			Title:        agent.Title,
			Class:        agent.Class,
			Level:        agent.Level,
			XP:           agent.XP,
			Energy:       agent.Energy,
			Status:       agent.Status,
			AvatarEmoji:  agent.AvatarEmoji,
			TotalTokens:  agent.TotalTokens,
			TotalCost:    agent.TotalCost,
			SessionCount: agent.SessionCount,
			QuestCount:   agent.QuestCount,
			ErrorCount:   agent.ErrorCount,
			LastActive:   agent.LastActive,
		}
		syncAgents = append(syncAgents, syncAgent)
		activeAgentNames = append(activeAgentNames, agent.Name)
	}

	// Sync to database
	created, updated, err := ars.syncService.SyncAgentsFromInber(syncAgents)
	if err != nil {
		return nil, fmt.Errorf("failed to sync agents to database: %w", err)
	}

	// Mark inactive agents
	if err := ars.syncService.MarkAgentsAsInactive(activeAgentNames); err != nil {
		log.Printf("Warning: failed to mark inactive agents: %v", err)
	}

	ars.lastSyncAt = time.Now()
	
	result := &SyncResult{
		Timestamp:       ars.lastSyncAt,
		TotalAgents:     len(inberAgents),
		CreatedAgents:   created,
		UpdatedAgents:   updated,
		ActiveAgents:    len(activeAgentNames),
		SyncDurationMs:  time.Since(ars.lastSyncAt).Milliseconds(),
	}

	log.Printf("✅ Agent sync complete: %d total, %d new, %d updated", 
		result.TotalAgents, result.CreatedAgents, result.UpdatedAgents)
	
	return result, nil
}

// GetSyncStatus returns the current sync status
func (ars *AgentRegistrySync) GetSyncStatus() SyncStatus {
	var minutesAgo int64 = 0
	if !ars.lastSyncAt.IsZero() {
		minutesAgo = int64(time.Since(ars.lastSyncAt).Minutes())
	}

	status := "unknown"
	if ars.inberSource == nil {
		status = "no_source"
	} else if ars.lastSyncAt.IsZero() {
		status = "never_synced"
	} else if minutesAgo < 5 {
		status = "healthy"
	} else if minutesAgo < 30 {
		status = "stale"
	} else {
		status = "outdated"
	}

	return SyncStatus{
		Status:       status,
		LastSyncAt:   ars.lastSyncAt,
		MinutesAgo:   minutesAgo,
		HasInberSource: ars.inberSource != nil,
	}
}

// GetManagedAgents returns all locally managed agents
func (ars *AgentRegistrySync) GetManagedAgents() ([]db.Agent, error) {
	return ars.syncService.GetManagedAgents()
}

// GetAgentByName returns a specific managed agent by name
func (ars *AgentRegistrySync) GetAgentByName(name string) (*db.Agent, error) {
	return ars.syncService.GetAgentByName(name)
}

// StartPeriodicSync starts a background goroutine that syncs agents periodically
func (ars *AgentRegistrySync) StartPeriodicSync(intervalMinutes int) {
	if ars.inberSource == nil {
		log.Println("⚠️ No inber source - skipping periodic agent sync")
		return
	}

	interval := time.Duration(intervalMinutes) * time.Minute
	log.Printf("🔄 Starting periodic agent sync every %v", interval)

	// Do initial sync
	if _, err := ars.SyncAgents(); err != nil {
		log.Printf("Warning: initial agent sync failed: %v", err)
	}

	// Start periodic sync
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			if _, err := ars.SyncAgents(); err != nil {
				log.Printf("Warning: periodic agent sync failed: %v", err)
			}
		}
	}()
}

// SyncResult contains the results of an agent sync operation
type SyncResult struct {
	Timestamp       time.Time `json:"timestamp"`
	TotalAgents     int       `json:"total_agents"`
	CreatedAgents   int       `json:"created_agents"`
	UpdatedAgents   int       `json:"updated_agents"`
	ActiveAgents    int       `json:"active_agents"`
	SyncDurationMs  int64     `json:"sync_duration_ms"`
}

// SyncStatus represents the current state of agent synchronization
type SyncStatus struct {
	Status         string    `json:"status"`
	LastSyncAt     time.Time `json:"last_sync_at"`
	MinutesAgo     int64     `json:"minutes_ago"`
	HasInberSource bool      `json:"has_inber_source"`
}