package user

import (
	"database/sql"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Repository handles user data operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new user repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateUser creates a new user account
func (r *Repository) CreateUser(req RegisterRequest) (*User, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Marshal skills to JSON
	skillsJSON, err := MarshalSkills(req.Skills)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal skills: %w", err)
	}

	// Insert user into database
	query := `
		INSERT INTO users (username, email, password_hash, display_name, bio, skills)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at
	`
	var user User
	user.Username = req.Username
	user.Email = req.Email
	user.PasswordHash = string(hashedPassword)
	user.DisplayName = req.DisplayName
	user.Bio = req.Bio
	user.Skills = req.Skills
	user.Reputation = 0.0
	user.BountiesCompleted = 0
	user.TotalEarned = 0.0
	user.IsActive = true

	err = r.db.QueryRow(query, req.Username, req.Email, string(hashedPassword), req.DisplayName, req.Bio, skillsJSON).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		// Check for unique constraint violations
		if strings.Contains(err.Error(), "username") {
			return nil, fmt.Errorf("username already taken")
		}
		if strings.Contains(err.Error(), "email") {
			return nil, fmt.Errorf("email already registered")
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// AuthenticateUser validates credentials and returns user if successful
func (r *Repository) AuthenticateUser(username, password string) (*User, error) {
	user, err := r.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("account is deactivated")
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last login
	r.UpdateLastLogin(user.ID)

	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *Repository) GetUserByID(id string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, bio, avatar_url, skills,
		       reputation, bounties_completed, total_earned, created_at, updated_at, last_login, is_active
		FROM users WHERE id = ?
	`
	return r.scanUser(r.db.QueryRow(query, id))
}

// GetUserByUsername retrieves a user by username
func (r *Repository) GetUserByUsername(username string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, bio, avatar_url, skills,
		       reputation, bounties_completed, total_earned, created_at, updated_at, last_login, is_active
		FROM users WHERE username = ?
	`
	return r.scanUser(r.db.QueryRow(query, username))
}

// GetUserByEmail retrieves a user by email
func (r *Repository) GetUserByEmail(email string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, bio, avatar_url, skills,
		       reputation, bounties_completed, total_earned, created_at, updated_at, last_login, is_active
		FROM users WHERE email = ?
	`
	return r.scanUser(r.db.QueryRow(query, email))
}

// UpdateProfile updates user profile information
func (r *Repository) UpdateProfile(userID string, req UpdateProfileRequest) error {
	skillsJSON, err := MarshalSkills(req.Skills)
	if err != nil {
		return fmt.Errorf("failed to marshal skills: %w", err)
	}

	query := `
		UPDATE users 
		SET display_name = ?, bio = ?, avatar_url = ?, skills = ?
		WHERE id = ?
	`
	result, err := r.db.Exec(query, req.DisplayName, req.Bio, req.AvatarURL, skillsJSON, userID)
	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check update result: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdateLastLogin updates the user's last login timestamp
func (r *Repository) UpdateLastLogin(userID string) error {
	query := `UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.Exec(query, userID)
	return err
}

// UpdateUserStats updates user reputation and statistics
func (r *Repository) UpdateUserStats(userID string, reputationDelta float64, earnedAmount float64) error {
	query := `
		UPDATE users 
		SET reputation = reputation + ?, 
		    total_earned = total_earned + ?, 
		    bounties_completed = bounties_completed + 1
		WHERE id = ?
	`
	_, err := r.db.Exec(query, reputationDelta, earnedAmount, userID)
	return err
}

// ListUsers retrieves users based on filter criteria
func (r *Repository) ListUsers(filter UserFilter) ([]PublicUser, error) {
	var conditions []string
	var args []interface{}

	// Build WHERE clause
	if filter.Username != "" {
		conditions = append(conditions, "username LIKE ?")
		args = append(args, "%"+filter.Username+"%")
	}

	if filter.MinRep != nil {
		conditions = append(conditions, "reputation >= ?")
		args = append(args, *filter.MinRep)
	}

	if filter.IsActive != nil {
		conditions = append(conditions, "is_active = ?")
		args = append(args, *filter.IsActive)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Set default limits
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT id, username, display_name, bio, avatar_url, skills,
		       reputation, bounties_completed, total_earned, created_at
		FROM users %s
		ORDER BY reputation DESC, username ASC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []PublicUser
	for rows.Next() {
		var user PublicUser
		var skillsJSON string

		err := rows.Scan(
			&user.ID, &user.Username, &user.DisplayName, &user.Bio, &user.AvatarURL,
			&skillsJSON, &user.Reputation, &user.BountiesCompleted, &user.TotalEarned, &user.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		// Unmarshal skills
		user.Skills, err = UnmarshalSkills(skillsJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal skills: %w", err)
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, nil
}

// DeactivateUser marks a user as inactive
func (r *Repository) DeactivateUser(userID string) error {
	query := `UPDATE users SET is_active = false WHERE id = ?`
	result, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check deactivation result: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// scanUser is a helper function to scan a user from a database row
func (r *Repository) scanUser(row *sql.Row) (*User, error) {
	var user User
	var skillsJSON string
	var lastLogin sql.NullTime

	err := row.Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.DisplayName, &user.Bio, &user.AvatarURL, &skillsJSON,
		&user.Reputation, &user.BountiesCompleted, &user.TotalEarned,
		&user.CreatedAt, &user.UpdatedAt, &lastLogin, &user.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	// Handle nullable last login
	if lastLogin.Valid {
		user.LastLogin = &lastLogin.Time
	}

	// Unmarshal skills
	user.Skills, err = UnmarshalSkills(skillsJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal skills: %w", err)
	}

	return &user, nil
}