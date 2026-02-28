package db

import "time"

type Agent struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Title       string    `json:"title"`
	Class       string    `json:"class"`
	Level       int       `json:"level"`
	XP          int       `json:"xp"`
	Energy      int       `json:"energy"`
	Status      string    `json:"status"`
	AvatarEmoji string    `json:"avatar_emoji"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AgentDetail struct {
	Agent
	Skills       []Skill       `json:"skills"`
	Achievements []Achievement `json:"achievements"`
	Tasks        []Task        `json:"tasks"`
}

type Task struct {
	ID              int        `json:"id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Difficulty      string     `json:"difficulty"`
	XPReward        int        `json:"xp_reward"`
	Status          string     `json:"status"`
	AssignedAgentID *int       `json:"assigned_agent_id,omitempty"`
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

type Stats struct {
	TotalAgents        int `json:"total_agents"`
	ActiveTasks        int `json:"active_tasks"`
	CompletedTasks     int `json:"completed_tasks"`
	TotalXP            int `json:"total_xp"`
	AverageAgentLevel  int `json:"average_agent_level"`
}
