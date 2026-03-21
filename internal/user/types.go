package user

import (
	"encoding/json"
	"time"
)

// User represents a user in the system
type User struct {
	ID                string    `json:"id" db:"id"`
	Username          string    `json:"username" db:"username"`
	Email             string    `json:"email" db:"email"`
	PasswordHash      string    `json:"-" db:"password_hash"` // Never expose password hash
	DisplayName       string    `json:"display_name" db:"display_name"`
	Bio               string    `json:"bio" db:"bio"`
	AvatarURL         string    `json:"avatar_url" db:"avatar_url"`
	Skills            []string  `json:"skills" db:"skills"`
	Reputation        float64   `json:"reputation" db:"reputation"`
	BountiesCompleted int       `json:"bounties_completed" db:"bounties_completed"`
	TotalEarned       float64   `json:"total_earned" db:"total_earned"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
	LastLogin         *time.Time `json:"last_login" db:"last_login"`
	IsActive          bool      `json:"is_active" db:"is_active"`
}

// PublicUser represents user data that can be shared publicly
type PublicUser struct {
	ID                string    `json:"id"`
	Username          string    `json:"username"`
	DisplayName       string    `json:"display_name"`
	Bio               string    `json:"bio"`
	AvatarURL         string    `json:"avatar_url"`
	Skills            []string  `json:"skills"`
	Reputation        float64   `json:"reputation"`
	BountiesCompleted int       `json:"bounties_completed"`
	TotalEarned       float64   `json:"total_earned"`
	CreatedAt         time.Time `json:"created_at"`
}

// ToPublic converts a User to PublicUser
func (u *User) ToPublic() PublicUser {
	return PublicUser{
		ID:                u.ID,
		Username:          u.Username,
		DisplayName:       u.DisplayName,
		Bio:               u.Bio,
		AvatarURL:         u.AvatarURL,
		Skills:            u.Skills,
		Reputation:        u.Reputation,
		BountiesCompleted: u.BountiesCompleted,
		TotalEarned:       u.TotalEarned,
		CreatedAt:         u.CreatedAt,
	}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username    string   `json:"username" binding:"required,min=3,max=50"`
	Email       string   `json:"email" binding:"required,email"`
	Password    string   `json:"password" binding:"required,min=8"`
	DisplayName string   `json:"display_name" binding:"required,min=1,max=100"`
	Bio         string   `json:"bio"`
	Skills      []string `json:"skills"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	User  PublicUser `json:"user"`
	Token string     `json:"token"`
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	DisplayName string   `json:"display_name"`
	Bio         string   `json:"bio"`
	AvatarURL   string   `json:"avatar_url"`
	Skills      []string `json:"skills"`
}

// UserFilter for listing/searching users
type UserFilter struct {
	Username   string   `json:"username"`
	Skills     []string `json:"skills"`
	MinRep     *float64 `json:"min_reputation"`
	IsActive   *bool    `json:"is_active"`
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
}

// MarshalSkills converts skills slice to JSON string for database storage
func MarshalSkills(skills []string) (string, error) {
	if skills == nil {
		skills = []string{}
	}
	data, err := json.Marshal(skills)
	return string(data), err
}

// UnmarshalSkills converts JSON string to skills slice from database
func UnmarshalSkills(data string) ([]string, error) {
	var skills []string
	if data == "" || data == "[]" {
		return skills, nil
	}
	err := json.Unmarshal([]byte(data), &skills)
	return skills, err
}