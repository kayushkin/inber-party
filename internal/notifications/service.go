package notifications

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kayushkin/inber-party/internal/ws"
)

// Service handles notification creation and delivery
type Service struct {
	db  *sql.DB
	hub *ws.Hub
}

// NewService creates a new notification service
func NewService(database *sql.DB, hub *ws.Hub) *Service {
	return &Service{
		db:  database,
		hub: hub,
	}
}

// CreateNotification creates a new notification for the specified recipient
func (s *Service) CreateNotification(recipientID int, notificationType NotificationType, title, message string, data interface{}, priority NotificationPriority) (*Notification, error) {
	// Check if recipient has preferences that might suppress this notification
	if !s.shouldSendNotification(recipientID, notificationType) {
		log.Printf("Notification suppressed by user preferences: user=%d, type=%s", recipientID, notificationType)
		return nil, nil
	}

	// Convert data to JSON
	var dataJSON []byte
	if data != nil {
		var err error
		dataJSON, err = json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal notification data: %w", err)
		}
	}

	notification := &Notification{
		RecipientID: recipientID,
		Type:        notificationType,
		Title:       title,
		Message:     message,
		Priority:    priority,
		Status:      StatusUnread,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Insert notification into database
	query := `
		INSERT INTO notifications (recipient_id, type, title, message, data, priority, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	err := s.db.QueryRow(query, recipientID, notificationType, title, message, string(dataJSON), priority, StatusUnread, notification.CreatedAt, notification.UpdatedAt).Scan(&notification.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Parse data back into the struct
	if dataJSON != nil {
		var parsedData map[string]interface{}
		if err := json.Unmarshal(dataJSON, &parsedData); err == nil {
			notification.Data = parsedData
		}
	}

	// Send real-time notification via WebSocket
	s.sendWebSocketNotification(notification)

	log.Printf("📢 Created notification: %s for agent %d: %s", notificationType, recipientID, title)
	return notification, nil
}

// NotifyBountyClaimed notifies the bounty creator that their bounty was claimed
func (s *Service) NotifyBountyClaimed(bountyID, creatorID, claimerID int, bountyTitle string, payoutAmount int) error {
	// Get claimer name
	claimerName := s.getAgentName(claimerID)
	
	data := BountyEventData{
		BountyID:     bountyID,
		BountyTitle:  bountyTitle,
		PayoutAmount: payoutAmount,
		ActorID:      claimerID,
		ActorName:    claimerName,
	}

	title := "Bounty Claimed"
	message := fmt.Sprintf("%s has claimed your bounty \"%s\" (₹%d)", claimerName, bountyTitle, payoutAmount)
	
	_, err := s.CreateNotification(creatorID, NotificationTypeBountyClaimed, title, message, data, PriorityNormal)
	return err
}

// NotifyBountyCompleted notifies the bounty creator that work was submitted
func (s *Service) NotifyBountyCompleted(bountyID, creatorID, claimerID int, bountyTitle string, payoutAmount int) error {
	claimerName := s.getAgentName(claimerID)
	
	data := BountyEventData{
		BountyID:     bountyID,
		BountyTitle:  bountyTitle,
		PayoutAmount: payoutAmount,
		ActorID:      claimerID,
		ActorName:    claimerName,
	}

	title := "Work Submitted"
	message := fmt.Sprintf("%s has submitted work for your bounty \"%s\". Please review and verify.", claimerName, bountyTitle)
	
	_, err := s.CreateNotification(creatorID, NotificationTypeBountyCompleted, title, message, data, PriorityHigh)
	return err
}

// NotifyBountyDisputed notifies the bounty creator that their rejection was disputed
func (s *Service) NotifyBountyDisputed(bountyID, creatorID, claimerID int, bountyTitle string, reason string) error {
	claimerName := s.getAgentName(claimerID)
	
	data := BountyEventData{
		BountyID:     bountyID,
		BountyTitle:  bountyTitle,
		ActorID:      claimerID,
		ActorName:    claimerName,
	}

	title := "Bounty Disputed"
	message := fmt.Sprintf("%s has disputed your rejection of \"%s\". Reason: %s", claimerName, bountyTitle, reason)
	
	_, err := s.CreateNotification(creatorID, NotificationTypeBountyDisputed, title, message, data, PriorityUrgent)
	return err
}

// NotifyBountyRejected notifies the claimer that their work was rejected
func (s *Service) NotifyBountyRejected(bountyID, claimerID, creatorID int, bountyTitle string, notes string) error {
	creatorName := s.getAgentName(creatorID)
	
	data := BountyEventData{
		BountyID:    bountyID,
		BountyTitle: bountyTitle,
		ActorID:     creatorID,
		ActorName:   creatorName,
	}

	title := "Work Rejected"
	message := fmt.Sprintf("Your work on \"%s\" was rejected by %s", bountyTitle, creatorName)
	if notes != "" {
		message += fmt.Sprintf(": %s", notes)
	}
	
	_, err := s.CreateNotification(claimerID, NotificationTypeBountyRejected, title, message, data, PriorityHigh)
	return err
}

// NotifySkillMatch notifies agents when new bounties match their skills
func (s *Service) NotifySkillMatch(bountyID int, bountyTitle string, requiredSkills []string, payoutAmount int, tier string) error {
	// Get all agents who have skills matching this bounty
	matchingAgents := s.findAgentsWithMatchingSkills(requiredSkills, tier)
	
	for _, match := range matchingAgents {
		data := SkillMatchData{
			BountyID:       bountyID,
			BountyTitle:    bountyTitle,
			PayoutAmount:   payoutAmount,
			RequiredSkills: requiredSkills,
			MatchedSkills:  match.MatchedSkills,
			MatchScore:     match.MatchScore,
		}

		title := "New Bounty Matches Your Skills"
		skillsText := strings.Join(match.MatchedSkills, ", ")
		message := fmt.Sprintf("New %s bounty \"%s\" (₹%d) matches your skills: %s", strings.Title(tier), bountyTitle, payoutAmount, skillsText)
		
		_, err := s.CreateNotification(match.AgentID, NotificationTypeSkillMatch, title, message, data, PriorityNormal)
		if err != nil {
			log.Printf("Failed to create skill match notification for agent %d: %v", match.AgentID, err)
		}
	}
	
	return nil
}

// NotifyDisputeResolved notifies the claimer about dispute resolution
func (s *Service) NotifyDisputeResolved(disputeID, claimerID int, bountyTitle string, inFavorOfClaimer bool, resolution string) error {
	title := "Dispute Resolved"
	var message string
	var priority NotificationPriority
	
	if inFavorOfClaimer {
		message = fmt.Sprintf("Your dispute for \"%s\" was resolved in your favor! You will receive payment.", bountyTitle)
		priority = PriorityHigh
	} else {
		message = fmt.Sprintf("Your dispute for \"%s\" was resolved against you. %s", bountyTitle, resolution)
		priority = PriorityNormal
	}

	data := map[string]interface{}{
		"dispute_id":           disputeID,
		"bounty_title":         bountyTitle,
		"in_favor_of_claimer":  inFavorOfClaimer,
		"resolution":           resolution,
	}

	_, err := s.CreateNotification(claimerID, NotificationTypeDisputeResolved, title, message, data, priority)
	return err
}

// NotifyPayoutReceived notifies an agent when they receive payment
func (s *Service) NotifyPayoutReceived(agentID, amount int, source, description string) error {
	title := "Payment Received"
	message := fmt.Sprintf("You received ₹%d from %s: %s", amount, source, description)
	
	data := map[string]interface{}{
		"amount":      amount,
		"source":      source,
		"description": description,
	}

	_, err := s.CreateNotification(agentID, NotificationTypePayoutReceived, title, message, data, PriorityNormal)
	return err
}

// GetNotifications retrieves notifications for an agent with filtering
func (s *Service) GetNotifications(filter NotificationFilter) ([]Notification, error) {
	query := `
		SELECT id, recipient_id, type, title, message, COALESCE(data, '{}'), priority, status, read_at, created_at, updated_at
		FROM notifications WHERE 1=1
	`
	args := []interface{}{}
	argCount := 0

	if filter.RecipientID != nil {
		argCount++
		query += " AND recipient_id = $" + fmt.Sprintf("%d", argCount)
		args = append(args, *filter.RecipientID)
	}

	if len(filter.Types) > 0 {
		argCount++
		query += " AND type = ANY($" + fmt.Sprintf("%d", argCount) + ")"
		typeStrings := make([]string, len(filter.Types))
		for i, t := range filter.Types {
			typeStrings[i] = string(t)
		}
		args = append(args, typeStrings)
	}

	if len(filter.Statuses) > 0 {
		argCount++
		query += " AND status = ANY($" + fmt.Sprintf("%d", argCount) + ")"
		statusStrings := make([]string, len(filter.Statuses))
		for i, s := range filter.Statuses {
			statusStrings[i] = string(s)
		}
		args = append(args, statusStrings)
	}

	if filter.Unread != nil && *filter.Unread {
		query += " AND status = 'unread'"
	}

	query += " ORDER BY created_at DESC"
	
	if filter.Limit > 0 {
		argCount++
		query += " LIMIT $" + fmt.Sprintf("%d", argCount)
		args = append(args, filter.Limit)
	}
	
	if filter.Offset > 0 {
		argCount++
		query += " OFFSET $" + fmt.Sprintf("%d", argCount)
		args = append(args, filter.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	notifications := []Notification{}
	for rows.Next() {
		var n Notification
		var dataJSON string
		
		err := rows.Scan(&n.ID, &n.RecipientID, &n.Type, &n.Title, &n.Message, &dataJSON, &n.Priority, &n.Status, &n.ReadAt, &n.CreatedAt, &n.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning notification: %v", err)
			continue
		}

		// Parse JSON data
		if dataJSON != "" && dataJSON != "{}" {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(dataJSON), &data); err == nil {
				n.Data = data
			}
		}

		notifications = append(notifications, n)
	}

	return notifications, nil
}

// MarkAsRead marks a notification as read
func (s *Service) MarkAsRead(notificationID, agentID int) error {
	now := time.Now()
	query := `UPDATE notifications SET status = $1, read_at = $2, updated_at = $3 WHERE id = $4 AND recipient_id = $5`
	result, err := s.db.Exec(query, StatusRead, now, now, notificationID, agentID)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("notification not found or not owned by agent")
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for an agent
func (s *Service) MarkAllAsRead(agentID int) error {
	now := time.Now()
	query := `UPDATE notifications SET status = $1, read_at = $2, updated_at = $3 WHERE recipient_id = $4 AND status = 'unread'`
	_, err := s.db.Exec(query, StatusRead, now, now, agentID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}
	return nil
}

// GetUnreadCount returns the number of unread notifications for an agent
func (s *Service) GetUnreadCount(agentID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM notifications WHERE recipient_id = $1 AND status = 'unread'`
	err := s.db.QueryRow(query, agentID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}
	return count, nil
}

// Helper methods

// shouldSendNotification checks user preferences to determine if notification should be sent
func (s *Service) shouldSendNotification(agentID int, notificationType NotificationType) bool {
	// For now, send all notifications. In the future, check user preferences.
	// This could query a notification_preferences table
	return true
}

// sendWebSocketNotification sends a real-time notification via WebSocket
func (s *Service) sendWebSocketNotification(notification *Notification) {
	if s.hub == nil {
		return
	}

	// Send to specific agent and broadcast to all connected clients
	s.hub.Broadcast(ws.Message{
		Type: "notification",
		Data: map[string]interface{}{
			"notification": notification,
		},
	})
}

// getAgentName retrieves the name of an agent by ID
func (s *Service) getAgentName(agentID int) string {
	var name string
	query := `SELECT name FROM agents WHERE id = $1`
	err := s.db.QueryRow(query, agentID).Scan(&name)
	if err != nil {
		return fmt.Sprintf("Agent #%d", agentID)
	}
	return name
}

// SkillMatch represents an agent with matching skills
type SkillMatch struct {
	AgentID       int      `json:"agent_id"`
	MatchedSkills []string `json:"matched_skills"`
	MatchScore    float64  `json:"match_score"`
}

// findAgentsWithMatchingSkills finds agents whose skills match the bounty requirements
func (s *Service) findAgentsWithMatchingSkills(requiredSkills []string, tier string) []SkillMatch {
	if len(requiredSkills) == 0 {
		return nil
	}

	// Get agents who have at least one matching skill
	query := `
		SELECT DISTINCT a.id, a.name,
			array_agg(DISTINCT s.skill_name) as agent_skills
		FROM agents a
		JOIN skills s ON a.id = s.agent_id
		WHERE s.skill_name = ANY($1)
		GROUP BY a.id, a.name
		HAVING COUNT(DISTINCT s.skill_name) > 0
	`

	rows, err := s.db.Query(query, requiredSkills)
	if err != nil {
		log.Printf("Error finding agents with matching skills: %v", err)
		return nil
	}
	defer rows.Close()

	matches := []SkillMatch{}
	for rows.Next() {
		var agentID int
		var agentName string
		var agentSkills []string

		err := rows.Scan(&agentID, &agentName, &agentSkills)
		if err != nil {
			log.Printf("Error scanning agent skills: %v", err)
			continue
		}

		// Find which skills actually match
		matchedSkills := []string{}
		for _, required := range requiredSkills {
			for _, agentSkill := range agentSkills {
				if strings.EqualFold(required, agentSkill) {
					matchedSkills = append(matchedSkills, required)
					break
				}
			}
		}

		if len(matchedSkills) > 0 {
			// Calculate match score (percentage of requirements met)
			matchScore := float64(len(matchedSkills)) / float64(len(requiredSkills))
			
			matches = append(matches, SkillMatch{
				AgentID:       agentID,
				MatchedSkills: matchedSkills,
				MatchScore:    matchScore,
			})
		}
	}

	return matches
}