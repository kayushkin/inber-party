package events

import (
	"log"
	"sync"
	"time"

	"github.com/kayushkin/inber-party/internal/inber"
	"github.com/kayushkin/inber-party/internal/ws"
)

// QuestEvent represents different types of quest events
type QuestEvent struct {
	Type     string      `json:"type"`     // "quest_started", "quest_completed", "quest_failed", "quest_progress", "agent_status_changed"
	QuestID  int         `json:"quest_id,omitempty"`
	AgentID  string      `json:"agent_id,omitempty"`
	Data     interface{} `json:"data"`     // The actual quest or agent data
	Message  string      `json:"message"`  // Human-readable event message
}

// QuestMonitor watches for changes in inber data and broadcasts events
type QuestMonitor struct {
	dataSource   inber.DataSource
	hub          *ws.Hub
	pollInterval time.Duration
	
	// Cache for change detection
	lastQuests    map[int]inber.RPGQuest
	lastAgents    map[string]inber.RPGAgent
	mutex         sync.RWMutex
	
	// Control
	stopCh chan bool
	running bool
}

// NewQuestMonitor creates a new quest monitor
func NewQuestMonitor(dataSource inber.DataSource, hub *ws.Hub) *QuestMonitor {
	return &QuestMonitor{
		dataSource:   dataSource,
		hub:          hub,
		pollInterval: 3 * time.Second, // Poll every 3 seconds for real-time feel
		lastQuests:   make(map[int]inber.RPGQuest),
		lastAgents:   make(map[string]inber.RPGAgent),
		stopCh:       make(chan bool),
		running:      false,
	}
}

// Start begins monitoring for quest events
func (m *QuestMonitor) Start() {
	m.mutex.Lock()
	if m.running {
		m.mutex.Unlock()
		return
	}
	m.running = true
	m.mutex.Unlock()
	
	log.Println("🔍 Quest Monitor started - watching for real-time quest events")
	
	// Initialize cache
	m.updateCache()
	
	// Start monitoring loop
	go m.monitorLoop()
}

// Stop stops the quest monitor
func (m *QuestMonitor) Stop() {
	m.mutex.Lock()
	if !m.running {
		m.mutex.Unlock()
		return
	}
	m.running = false
	m.mutex.Unlock()
	
	m.stopCh <- true
	log.Println("🔍 Quest Monitor stopped")
}

// monitorLoop is the main monitoring loop
func (m *QuestMonitor) monitorLoop() {
	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.checkForChanges()
		}
	}
}

// updateCache updates the internal cache with current data
func (m *QuestMonitor) updateCache() {
	// Update quests cache
	quests, err := m.dataSource.GetQuests(100)
	if err != nil {
		log.Printf("Quest Monitor: Error fetching quests: %v", err)
		return
	}
	
	m.mutex.Lock()
	m.lastQuests = make(map[int]inber.RPGQuest)
	for _, quest := range quests {
		m.lastQuests[quest.ID] = quest
	}
	m.mutex.Unlock()
	
	// Update agents cache
	agents, err := m.dataSource.GetAgents()
	if err != nil {
		log.Printf("Quest Monitor: Error fetching agents: %v", err)
		return
	}
	
	m.mutex.Lock()
	m.lastAgents = make(map[string]inber.RPGAgent)
	for _, agent := range agents {
		m.lastAgents[agent.ID] = agent
	}
	m.mutex.Unlock()
}

// checkForChanges compares current data with cache and broadcasts events
func (m *QuestMonitor) checkForChanges() {
	// Check quest changes
	m.checkQuestChanges()
	
	// Check agent changes
	m.checkAgentChanges()
}

// checkQuestChanges detects and broadcasts quest-related events
func (m *QuestMonitor) checkQuestChanges() {
	quests, err := m.dataSource.GetQuests(100)
	if err != nil {
		return
	}
	
	m.mutex.RLock()
	oldQuests := m.lastQuests
	m.mutex.RUnlock()
	
	newQuests := make(map[int]inber.RPGQuest)
	for _, quest := range quests {
		newQuests[quest.ID] = quest
		
		oldQuest, existed := oldQuests[quest.ID]
		
		if !existed {
			// New quest detected
			m.broadcastQuestEvent("quest_started", quest.ID, quest.AgentID, quest, 
				formatQuestStartMessage(quest))
		} else if oldQuest.Status != quest.Status {
			// Quest status changed
			switch quest.Status {
			case "completed":
				m.broadcastQuestEvent("quest_completed", quest.ID, quest.AgentID, quest,
					formatQuestCompletedMessage(quest))
			case "failed":
				m.broadcastQuestEvent("quest_failed", quest.ID, quest.AgentID, quest,
					formatQuestFailedMessage(quest))
			}
		} else if oldQuest.Progress != quest.Progress {
			// Quest progress changed
			m.broadcastQuestEvent("quest_progress", quest.ID, quest.AgentID, quest,
				formatQuestProgressMessage(quest, oldQuest.Progress))
		}
	}
	
	// Update cache
	m.mutex.Lock()
	m.lastQuests = newQuests
	m.mutex.Unlock()
}

// checkAgentChanges detects and broadcasts agent status changes
func (m *QuestMonitor) checkAgentChanges() {
	agents, err := m.dataSource.GetAgents()
	if err != nil {
		return
	}
	
	m.mutex.RLock()
	oldAgents := m.lastAgents
	m.mutex.RUnlock()
	
	newAgents := make(map[string]inber.RPGAgent)
	for _, agent := range agents {
		newAgents[agent.ID] = agent
		
		oldAgent, existed := oldAgents[agent.ID]
		
		if existed {
			// Check for status changes
			if oldAgent.Status != agent.Status {
				m.broadcastQuestEvent("agent_status_changed", 0, agent.ID, agent,
					formatAgentStatusMessage(agent, oldAgent.Status))
			}
			
			// Check for level up
			if oldAgent.Level < agent.Level {
				m.broadcastQuestEvent("agent_level_up", 0, agent.ID, agent,
					formatAgentLevelUpMessage(agent, oldAgent.Level))
			}
		}
	}
	
	// Update cache
	m.mutex.Lock()
	m.lastAgents = newAgents
	m.mutex.Unlock()
}

// broadcastQuestEvent sends a quest event via WebSocket
func (m *QuestMonitor) broadcastQuestEvent(eventType string, questID int, agentID string, data interface{}, message string) {
	event := QuestEvent{
		Type:    eventType,
		QuestID: questID,
		AgentID: agentID,
		Data:    data,
		Message: message,
	}
	
	m.hub.Broadcast(ws.Message{
		Type: "quest_event",
		Data: event,
	})
	
	// Log interesting events
	if eventType != "quest_progress" { // Don't spam logs with progress updates
		log.Printf("🎯 Quest Event: %s", message)
	}
}

// Message formatting functions
func formatQuestStartMessage(quest inber.RPGQuest) string {
	return quest.AgentName + " has begun: " + quest.Name
}

func formatQuestCompletedMessage(quest inber.RPGQuest) string {
	return "🎉 " + quest.AgentName + " completed: " + quest.Name + " (+" + 
		   formatXP(quest.XPReward) + " XP)"
}

func formatQuestFailedMessage(quest inber.RPGQuest) string {
	return "💀 " + quest.AgentName + " failed: " + quest.Name
}

func formatQuestProgressMessage(quest inber.RPGQuest, oldProgress int) string {
	return quest.AgentName + " making progress on " + quest.Name + 
		   " (" + formatProgress(oldProgress) + " → " + formatProgress(quest.Progress) + ")"
}

func formatAgentStatusMessage(agent inber.RPGAgent, oldStatus string) string {
	statusEmoji := map[string]string{
		"idle":    "😴",
		"working": "⚔️", 
		"stuck":   "🚫",
		"resting": "💤",
	}
	
	emoji, ok := statusEmoji[agent.Status]
	if !ok {
		emoji = "📍"
	}
	
	return emoji + " " + agent.Name + " is now " + agent.Status
}

func formatAgentLevelUpMessage(agent inber.RPGAgent, oldLevel int) string {
	return "✨ " + agent.Name + " leveled up! " + formatLevel(oldLevel) + " → " + formatLevel(agent.Level)
}

// Helper formatting functions
func formatXP(xp int) string {
	if xp >= 1000 {
		return formatNumber(xp/1000) + "k"
	}
	return formatNumber(xp)
}

func formatProgress(progress int) string {
	return formatNumber(progress) + "%"
}

func formatLevel(level int) string {
	return "Lvl " + formatNumber(level)
}

func formatNumber(n int) string {
	if n >= 1000000 {
		return formatFloat(float64(n)/1000000.0, 1) + "M"
	} else if n >= 1000 {
		return formatFloat(float64(n)/1000.0, 1) + "k"
	}
	return itoa(n)
}

func formatFloat(f float64, decimals int) string {
	// Simple float formatting - could use fmt.Sprintf but avoiding imports
	if decimals == 1 {
		rounded := int(f*10 + 0.5)
		return itoa(rounded/10) + "." + itoa(rounded%10)
	}
	return itoa(int(f + 0.5))
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	
	negative := n < 0
	if negative {
		n = -n
	}
	
	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	
	// Reverse
	for i := 0; i < len(digits)/2; i++ {
		digits[i], digits[len(digits)-1-i] = digits[len(digits)-1-i], digits[i]
	}
	
	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}