package notifications

import (
	"time"
)

// Notification represents a notification to be delivered to an agent
type Notification struct {
	ID          int                    `json:"id" db:"id"`
	RecipientID int                    `json:"recipient_id" db:"recipient_id"`
	Type        NotificationType       `json:"type" db:"type"`
	Title       string                 `json:"title" db:"title"`
	Message     string                 `json:"message" db:"message"`
	Data        map[string]interface{} `json:"data" db:"data"`
	Priority    NotificationPriority   `json:"priority" db:"priority"`
	Status      NotificationStatus     `json:"status" db:"status"`
	ReadAt      *time.Time             `json:"read_at" db:"read_at"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeBountyClaimed    NotificationType = "bounty_claimed"
	NotificationTypeBountyCompleted  NotificationType = "bounty_completed"
	NotificationTypeBountyDisputed   NotificationType = "bounty_disputed"
	NotificationTypeBountyRejected   NotificationType = "bounty_rejected"
	NotificationTypeSkillMatch       NotificationType = "skill_match"
	NotificationTypeNewBounty        NotificationType = "new_bounty"
	NotificationTypeDisputeResolved  NotificationType = "dispute_resolved"
	NotificationTypePayoutReceived   NotificationType = "payout_received"
	NotificationTypeRatingReceived   NotificationType = "rating_received"
)

// NotificationPriority represents the priority level of a notification
type NotificationPriority string

const (
	PriorityLow    NotificationPriority = "low"
	PriorityNormal NotificationPriority = "normal"
	PriorityHigh   NotificationPriority = "high"
	PriorityUrgent NotificationPriority = "urgent"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	StatusUnread NotificationStatus = "unread"
	StatusRead   NotificationStatus = "read"
	StatusDismissed NotificationStatus = "dismissed"
)

// NotificationPreferences represents user preferences for notifications
type NotificationPreferences struct {
	ID                 int                         `json:"id" db:"id"`
	AgentID            int                         `json:"agent_id" db:"agent_id"`
	EmailEnabled       bool                        `json:"email_enabled" db:"email_enabled"`
	WebSocketEnabled   bool                        `json:"websocket_enabled" db:"websocket_enabled"`
	PushEnabled        bool                        `json:"push_enabled" db:"push_enabled"`
	TypePreferences    map[NotificationType]bool   `json:"type_preferences" db:"type_preferences"`
	SkillKeywords      []string                    `json:"skill_keywords" db:"skill_keywords"`
	MinimumBountyTier  string                      `json:"minimum_bounty_tier" db:"minimum_bounty_tier"`
	QuietHoursStart    *int                        `json:"quiet_hours_start" db:"quiet_hours_start"`
	QuietHoursEnd      *int                        `json:"quiet_hours_end" db:"quiet_hours_end"`
	CreatedAt          time.Time                   `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time                   `json:"updated_at" db:"updated_at"`
}

// BountyEventData represents data for bounty-related notifications
type BountyEventData struct {
	BountyID     int    `json:"bounty_id"`
	BountyTitle  string `json:"bounty_title"`
	PayoutAmount int    `json:"payout_amount"`
	ActorID      int    `json:"actor_id"`
	ActorName    string `json:"actor_name"`
}

// SkillMatchData represents data for skill-matching notifications
type SkillMatchData struct {
	BountyID        int      `json:"bounty_id"`
	BountyTitle     string   `json:"bounty_title"`
	PayoutAmount    int      `json:"payout_amount"`
	RequiredSkills  []string `json:"required_skills"`
	MatchedSkills   []string `json:"matched_skills"`
	MatchScore      float64  `json:"match_score"`
}

// NotificationFilter represents filters for querying notifications
type NotificationFilter struct {
	RecipientID *int                   `json:"recipient_id"`
	Types       []NotificationType     `json:"types"`
	Statuses    []NotificationStatus   `json:"statuses"`
	Priorities  []NotificationPriority `json:"priorities"`
	Unread      *bool                  `json:"unread"`
	Limit       int                    `json:"limit"`
	Offset      int                    `json:"offset"`
}