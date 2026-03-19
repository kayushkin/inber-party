package db

import "time"

type Agent struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Title       string    `json:"title"`
	Class       string    `json:"class"`
	Level       int       `json:"level"`
	XP          int       `json:"xp"`
	Gold        int       `json:"gold"`          // currency earned from completed quests
	Energy      int       `json:"energy"`
	Status      string    `json:"status"`
	AvatarEmoji string    `json:"avatar_emoji"`
	Mood        string    `json:"mood"`          // happy, content, neutral, stressed, exhausted
	MoodScore   int       `json:"mood_score"`    // 0-100, higher = better mood
	Workload    int       `json:"workload"`      // current active task count
	LastActive  *time.Time `json:"last_active"`  // last task completion time
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AgentDetail struct {
	Agent
	Skills       []Skill       `json:"skills"`
	Achievements []Achievement `json:"achievements"`
	Tasks        []Task        `json:"tasks"`
	Reputation   []Reputation  `json:"reputation"`
}

type Task struct {
	ID              int        `json:"id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Difficulty      string     `json:"difficulty"`
	XPReward        int        `json:"xp_reward"`
	GoldReward      int        `json:"gold_reward"`
	Status          string     `json:"status"`
	AssignedAgentID *int       `json:"assigned_agent_id,omitempty"`
	AssignedPartyID *int       `json:"assigned_party_id,omitempty"`
	Progress        int        `json:"progress"`
	CreatedAt       time.Time  `json:"created_at"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

type Skill struct {
	ID        int    `json:"id"`
	AgentID   int    `json:"agent_id"`
	SkillName string `json:"skill_name"`
	Level     int    `json:"level"`
	TaskCount int    `json:"task_count"`
}

type Achievement struct {
	ID              int       `json:"id"`
	AgentID         int       `json:"agent_id"`
	AchievementName string    `json:"achievement_name"`
	UnlockedAt      time.Time `json:"unlocked_at"`
}

type Party struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	LeaderID    int       `json:"leader_id"`
	Status      string    `json:"status"` // "active", "disbanded", "on_quest"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PartyDetail struct {
	Party
	Members []Agent `json:"members"`
	Tasks   []Task  `json:"tasks"`
	Leader  Agent   `json:"leader"`
}

type PartyMember struct {
	ID       int       `json:"id"`
	PartyID  int       `json:"party_id"`
	AgentID  int       `json:"agent_id"`
	Role     string    `json:"role"` // "leader", "member", "support"
	JoinedAt time.Time `json:"joined_at"`
}

type Reputation struct {
	ID         int       `json:"id"`
	AgentID    int       `json:"agent_id"`
	Domain     string    `json:"domain"`     // "coding", "testing", "documentation", "devops", "design", etc.
	Score      int       `json:"score"`      // 0-1000, starts at 100
	TaskCount  int       `json:"task_count"` // number of tasks completed in this domain
	SuccessRate float64  `json:"success_rate"` // percentage of successful tasks (0.0-1.0)
	LastUpdate time.Time `json:"last_update"`
}

// CostEntry tracks token costs for agents and tasks
type CostEntry struct {
	ID              int        `json:"id"`
	AgentID         *int       `json:"agent_id,omitempty"`         // Agent that incurred the cost
	TaskID          *int       `json:"task_id,omitempty"`          // Task that incurred the cost
	SessionID       string     `json:"session_id"`                 // Session/conversation ID
	TokensUsed      int        `json:"tokens_used"`                // Number of tokens consumed
	CostUSD         float64    `json:"cost_usd"`                   // Cost in USD
	ModelName       string     `json:"model_name"`                 // Model used (e.g., "claude-3-5-sonnet")
	OperationType   string     `json:"operation_type"`             // "task", "conversation", "spawn", etc.
	Date            string     `json:"date"`                       // Date in YYYY-MM-DD format for easy grouping
	CreatedAt       time.Time  `json:"created_at"`
	Metadata        *string    `json:"metadata,omitempty"`         // JSON metadata for additional context
}

// CostSummary provides aggregated cost data for dashboards
type CostSummary struct {
	Period        string  `json:"period"`        // "daily", "weekly", "monthly"
	AgentID       *int    `json:"agent_id,omitempty"`
	AgentName     string  `json:"agent_name,omitempty"`
	TotalCost     float64 `json:"total_cost"`
	TotalTokens   int     `json:"total_tokens"`
	TaskCount     int     `json:"task_count"`
	AvgCostPerTask float64 `json:"avg_cost_per_task"`
	DateRange     string  `json:"date_range"`   // e.g., "2026-03-18" or "2026-03-11 to 2026-03-18"
}

// Bounty represents a task posted to the marketplace that can be claimed and completed
type Bounty struct {
	ID                int        `json:"id"`
	Title             string     `json:"title"`
	Description       string     `json:"description"`
	Requirements      string     `json:"requirements"`       // acceptance criteria
	PayoutAmount      int        `json:"payout_amount"`      // gold reward
	Status            string     `json:"status"`             // "open", "claimed", "completed", "rejected", "paid"
	Deadline          *time.Time `json:"deadline,omitempty"` // optional deadline
	CreatorID         int        `json:"creator_id"`         // agent/user who posted the bounty
	ClaimerID         *int       `json:"claimer_id,omitempty"` // agent who claimed it
	WorkSubmission    *string    `json:"work_submission,omitempty"` // submitted work/evidence
	VerificationNotes *string    `json:"verification_notes,omitempty"` // reviewer feedback
	RequiredSkills    []string   `json:"required_skills"`    // array of skill tags
	Tier             string     `json:"tier"`               // bronze, silver, gold, legendary based on payout
	ClaimedAt        *time.Time `json:"claimed_at,omitempty"`
	SubmittedAt      *time.Time `json:"submitted_at,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// BountyDetail includes related data for full bounty view
type BountyDetail struct {
	Bounty
	Creator Agent  `json:"creator"`
	Claimer *Agent `json:"claimer,omitempty"`
}

// PayoutEntry tracks all gold/credit transactions for auditing and ledger purposes
type PayoutEntry struct {
	ID              int        `json:"id"`
	AgentID         int        `json:"agent_id"`             // Recipient of the payout
	Amount          int        `json:"amount"`               // Amount of gold/credits paid
	Source          string     `json:"source"`               // "bounty", "quest", "daily_quest", "manual"
	SourceID        *int       `json:"source_id,omitempty"`  // ID of the bounty/quest/etc that generated this payout
	Description     string     `json:"description"`          // Human readable description
	TransactionType string     `json:"transaction_type"`     // "credit", "debit", "adjustment"
	BalanceBefore   int        `json:"balance_before"`       // Agent's gold balance before this transaction
	BalanceAfter    int        `json:"balance_after"`        // Agent's gold balance after this transaction
	ProcessedBy     *int       `json:"processed_by,omitempty"` // Agent/user who processed this payout (for manual adjustments)
	CreatedAt       time.Time  `json:"created_at"`
}

// PayoutSummary provides aggregated payout data for participants
type PayoutSummary struct {
	AgentID        int     `json:"agent_id"`
	AgentName      string  `json:"agent_name"`
	TotalEarned    int     `json:"total_earned"`     // Total credits earned
	TotalSpent     int     `json:"total_spent"`      // Total credits spent/debited
	CurrentBalance int     `json:"current_balance"`  // Current gold balance
	PayoutCount    int     `json:"payout_count"`     // Number of payout transactions
	LastPayout     *time.Time `json:"last_payout,omitempty"` // Most recent payout timestamp
	Sources        map[string]int `json:"sources"`   // Breakdown by source (bounty: 100, quest: 50, etc.)
}

// BountyVerifier represents a verification requirement for a bounty
type BountyVerifier struct {
	ID           int                    `json:"id"`
	BountyID     int                    `json:"bounty_id"`
	VerifierType string                 `json:"verifier_type"` // 'manual', 'file_exists', 'test_suite', 'pr_approved', 'benchmark'
	Config       map[string]interface{} `json:"config"`        // configuration specific to verifier type
	Required     bool                   `json:"required"`      // if false, verifier is optional
	Weight       float64                `json:"weight"`        // weight for scoring (future use)
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// VerifierResult represents the result of running a verifier
type VerifierResult struct {
	ID           int                    `json:"id"`
	BountyID     int                    `json:"bounty_id"`
	VerifierID   int                    `json:"verifier_id"`
	Status       string                 `json:"status"`        // 'pending', 'passed', 'failed', 'error'
	ResultData   map[string]interface{} `json:"result_data"`   // detailed results from verifier
	ErrorMessage *string                `json:"error_message"` // error message if status = 'error'
	CheckedAt    time.Time              `json:"checked_at"`
	CheckedBy    *string                `json:"checked_by"` // agent/user/system that ran the check
}

// BountyWithVerifiers extends BountyDetail with verification info
type BountyWithVerifiers struct {
	BountyDetail
	Verifiers []BountyVerifier  `json:"verifiers"`
	Results   []VerifierResult  `json:"results"`
}

// BountyRating represents a rating given by a bounty creator to a claimer
type BountyRating struct {
	ID           int       `json:"id"`
	BountyID     int       `json:"bounty_id"`
	RaterID      int       `json:"rater_id"`      // Creator who is rating
	RatedID      int       `json:"rated_id"`      // Claimer who is being rated
	Rating       int       `json:"rating"`        // 1-5 stars
	Comment      string    `json:"comment"`       // Text feedback
	Categories   map[string]int `json:"categories"` // Breakdown: quality, timeliness, communication
	CreatedAt    time.Time `json:"created_at"`
}

type Stats struct {
	TotalAgents        int     `json:"total_agents"`
	ActiveTasks        int     `json:"active_tasks"`
	CompletedTasks     int     `json:"completed_tasks"`
	TotalXP            int     `json:"total_xp"`
	TotalGold          int     `json:"total_gold"`
	AverageAgentLevel  int     `json:"average_agent_level"`
	TotalParties       int     `json:"total_parties"`
	ActiveParties      int     `json:"active_parties"`
	TotalCostUSD       float64 `json:"total_cost_usd"`             // Total cost across all agents
	TotalTokensUsed    int     `json:"total_tokens_used"`          // Total tokens across all agents
}
