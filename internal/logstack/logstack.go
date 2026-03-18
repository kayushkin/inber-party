package logstack

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// LogstackClient provides access to OpenClaw session logs
type LogstackClient struct {
	AgentsDir string
}

// NewLogstackClient creates a new client for reading OpenClaw session logs
func NewLogstackClient() *LogstackClient {
	home, _ := os.UserHomeDir()
	agentsDir := filepath.Join(home, ".openclaw", "agents")
	return &LogstackClient{
		AgentsDir: agentsDir,
	}
}

// SessionEvent represents a single event in an OpenClaw session log
type SessionEvent struct {
	Type      string                 `json:"type"`
	ID        string                 `json:"id"`
	ParentID  *string                `json:"parentId,omitempty"`
	Timestamp string                 `json:"timestamp"`
	Message   *MessageContent        `json:"message,omitempty"`
	Extra     map[string]interface{} `json:"-"` // For other fields we don't need
}

// MessageContent represents a message in the conversation
type MessageContent struct {
	Role      string          `json:"role"`
	Content   []ContentBlock  `json:"content"`
	Timestamp int64           `json:"timestamp"`
}

// ContentBlock represents a block of content in a message
type ContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Thinking string `json:"thinking,omitempty"`
}

// Conversation represents a conversation between user and agent
type Conversation struct {
	ID           string              `json:"id"`
	AgentID      string              `json:"agent_id"`
	AgentName    string              `json:"agent_name"`
	Title        string              `json:"title"`
	Messages     []ConversationMsg   `json:"messages"`
	StartTime    string              `json:"start_time"`
	EndTime      string              `json:"end_time"`
	MessageCount int                 `json:"message_count"`
}

// ConversationMsg represents a single message in a conversation
type ConversationMsg struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	HasTools  bool   `json:"has_tools"`
}

// GetAgentList returns list of available agents with sessions
func (lc *LogstackClient) GetAgentList() ([]string, error) {
	var agents []string
	
	entries, err := os.ReadDir(lc.AgentsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read agents directory: %w", err)
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			sessionsDir := filepath.Join(lc.AgentsDir, entry.Name(), "sessions")
			if _, err := os.Stat(sessionsDir); err == nil {
				agents = append(agents, entry.Name())
			}
		}
	}
	
	return agents, nil
}

// GetAgentConversations returns conversations for a specific agent
func (lc *LogstackClient) GetAgentConversations(agentID string, limit int) ([]Conversation, error) {
	if limit <= 0 {
		limit = 20
	}
	
	sessionsDir := filepath.Join(lc.AgentsDir, agentID, "sessions")
	
	// Get session files
	sessionFiles, err := lc.getSessionFiles(sessionsDir, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get session files for agent %s: %w", agentID, err)
	}
	
	var conversations []Conversation
	
	for _, sessionFile := range sessionFiles {
		conv, err := lc.parseSessionFile(sessionFile, agentID)
		if err != nil {
			continue // Skip files that can't be parsed
		}
		conversations = append(conversations, *conv)
	}
	
	// Sort by start time (newest first)
	sort.Slice(conversations, func(i, j int) bool {
		return conversations[i].StartTime > conversations[j].StartTime
	})
	
	return conversations, nil
}

// GetAllConversations returns recent conversations from all agents
func (lc *LogstackClient) GetAllConversations(limit int) ([]Conversation, error) {
	agents, err := lc.GetAgentList()
	if err != nil {
		return nil, fmt.Errorf("failed to get agent list: %w", err)
	}
	
	var allConversations []Conversation
	
	// Get conversations from each agent (limiting per agent to keep it reasonable)
	perAgentLimit := min(limit/len(agents)+1, 10)
	
	for _, agentID := range agents {
		conversations, err := lc.GetAgentConversations(agentID, perAgentLimit)
		if err != nil {
			continue // Skip agents we can't read
		}
		allConversations = append(allConversations, conversations...)
	}
	
	// Sort by start time (newest first) and limit
	sort.Slice(allConversations, func(i, j int) bool {
		return allConversations[i].StartTime > allConversations[j].StartTime
	})
	
	if len(allConversations) > limit {
		allConversations = allConversations[:limit]
	}
	
	return allConversations, nil
}

// getSessionFiles returns session files sorted by modification time (newest first)
func (lc *LogstackClient) getSessionFiles(sessionsDir string, limit int) ([]string, error) {
	var files []fs.FileInfo
	
	err := filepath.WalkDir(sessionsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Continue on errors
		}
		
		if !d.IsDir() && strings.HasSuffix(path, ".jsonl") && !strings.Contains(path, ".deleted.") {
			info, err := d.Info()
			if err == nil {
				files = append(files, &fileInfoWithPath{info, path})
			}
		}
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	// Sort by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})
	
	// Limit results
	if len(files) > limit {
		files = files[:limit]
	}
	
	var sessionFiles []string
	for _, file := range files {
		if pathInfo, ok := file.(*fileInfoWithPath); ok {
			sessionFiles = append(sessionFiles, pathInfo.path)
		}
	}
	
	return sessionFiles, nil
}

// parseSessionFile parses a JSONL session file into a conversation
func (lc *LogstackClient) parseSessionFile(sessionFile, agentID string) (*Conversation, error) {
	file, err := os.Open(sessionFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open session file: %w", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	
	var events []SessionEvent
	var sessionID string
	var startTime string
	
	// Parse all events from the JSONL file
	for scanner.Scan() {
		var event SessionEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue // Skip malformed lines
		}
		events = append(events, event)
		
		// Extract session ID and start time from session event
		if event.Type == "session" {
			sessionID = event.ID
			startTime = event.Timestamp
		}
	}
	
	if sessionID == "" {
		return nil, fmt.Errorf("no session event found in file")
	}
	
	// Extract messages from events
	var messages []ConversationMsg
	var lastTime string
	
	for _, event := range events {
		if event.Type == "message" && event.Message != nil {
			content := extractTextContent(event.Message.Content)
			if content == "" {
				continue // Skip empty messages
			}
			
			hasTools := hasToolCalls(event.Message.Content)
			
			msg := ConversationMsg{
				ID:        event.ID,
				Role:      event.Message.Role,
				Content:   content,
				Timestamp: event.Timestamp,
				HasTools:  hasTools,
			}
			messages = append(messages, msg)
			lastTime = event.Timestamp
		}
	}
	
	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages found in session")
	}
	
	// Generate a title from the first user message
	title := generateConversationTitle(messages)
	
	conv := &Conversation{
		ID:           sessionID,
		AgentID:      agentID,
		AgentName:    titleCase(agentID),
		Title:        title,
		Messages:     messages,
		StartTime:    startTime,
		EndTime:      lastTime,
		MessageCount: len(messages),
	}
	
	return conv, nil
}

// extractTextContent extracts readable text from message content blocks
func extractTextContent(content []ContentBlock) string {
	var textParts []string
	
	for _, block := range content {
		switch block.Type {
		case "text":
			if block.Text != "" {
				textParts = append(textParts, block.Text)
			}
		case "toolCall":
			// For tool calls, just indicate they happened
			textParts = append(textParts, "[Used tools]")
		}
	}
	
	return strings.Join(textParts, " ")
}

// hasToolCalls checks if the message content includes tool calls
func hasToolCalls(content []ContentBlock) bool {
	for _, block := range content {
		if block.Type == "toolCall" {
			return true
		}
	}
	return false
}

// generateConversationTitle creates a title from the first user message
func generateConversationTitle(messages []ConversationMsg) string {
	for _, msg := range messages {
		if msg.Role == "user" && msg.Content != "" {
			// Clean up and truncate the content
			title := strings.TrimSpace(msg.Content)
			
			// Remove cron job prefixes
			if strings.Contains(title, "[cron:") {
				if idx := strings.Index(title, "]"); idx != -1 && idx < len(title)-1 {
					title = strings.TrimSpace(title[idx+1:])
				}
			}
			
			// Take first line and truncate
			lines := strings.SplitN(title, "\n", 2)
			title = lines[0]
			
			if len(title) > 60 {
				title = title[:57] + "..."
			}
			
			if title != "" {
				return title
			}
		}
	}
	
	return "Untitled Conversation"
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// fileInfoWithPath wraps fs.FileInfo to include the full path
type fileInfoWithPath struct {
	fs.FileInfo
	path string
}