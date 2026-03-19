package bounty

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateBounty creates a new bounty in the database
func (r *Repository) CreateBounty(bounty *Bounty) error {
	// Set defaults
	if bounty.Status == "" {
		bounty.Status = StatusOpen
	}
	bounty.Tier = GetTierFromPayout(bounty.PayoutAmount)
	bounty.CreatedAt = time.Now()
	bounty.UpdatedAt = time.Now()

	// Serialize JSON fields
	skillsJSON, _ := json.Marshal(bounty.RequiredSkills)

	query := `
		INSERT INTO bounties (
			title, description, requirements, payout_amount, deadline,
			status, creator_id, required_skills, tier
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	// Convert payout to integer cents for database
	payoutCents := int(bounty.PayoutAmount * 100)

	row := r.db.QueryRow(query,
		bounty.Title, bounty.Description, bounty.Requirements,
		payoutCents, bounty.Deadline,
		bounty.Status, bounty.CreatedBy, string(skillsJSON),
		bounty.Tier)

	var id int
	if err := row.Scan(&id, &bounty.CreatedAt, &bounty.UpdatedAt); err != nil {
		return err
	}
	
	// Convert ID to string for consistency with API
	bounty.ID = fmt.Sprintf("%d", id)
	
	return nil
}

// GetBounty retrieves a bounty by ID
func (r *Repository) GetBounty(id string) (*Bounty, error) {
	query := `
		SELECT id, title, description, requirements, payout_amount,
			   deadline, status, created_at, updated_at, creator_id, claimer_id,
			   claimed_at, completed_at, required_skills, tier
		FROM bounties WHERE id = $1
	`

	row := r.db.QueryRow(query, id)
	return r.scanBounty(row)
}

// ListBounties retrieves bounties with filtering and pagination
func (r *Repository) ListBounties(filter BountyFilter) ([]*Bounty, error) {
	query := "SELECT id, title, description, requirements, payout_amount, deadline, status, created_at, updated_at, creator_id, claimer_id, claimed_at, completed_at, required_skills, tier FROM bounties"
	
	var conditions []string
	var args []interface{}

	// Build WHERE conditions
	paramCounter := 1
	
	if len(filter.Status) > 0 {
		placeholders := make([]string, len(filter.Status))
		for i, status := range filter.Status {
			placeholders[i] = fmt.Sprintf("$%d", paramCounter)
			args = append(args, status)
			paramCounter++
		}
		conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.Tier) > 0 {
		placeholders := make([]string, len(filter.Tier))
		for i, tier := range filter.Tier {
			placeholders[i] = fmt.Sprintf("$%d", paramCounter)
			args = append(args, tier)
			paramCounter++
		}
		conditions = append(conditions, fmt.Sprintf("tier IN (%s)", strings.Join(placeholders, ",")))
	}

	if filter.MinPayout != nil {
		conditions = append(conditions, fmt.Sprintf("payout_amount >= $%d", paramCounter))
		args = append(args, int(*filter.MinPayout * 100)) // Convert to cents
		paramCounter++
	}

	if filter.MaxPayout != nil {
		conditions = append(conditions, fmt.Sprintf("payout_amount <= $%d", paramCounter))
		args = append(args, int(*filter.MaxPayout * 100)) // Convert to cents
		paramCounter++
	}

	if filter.CreatedBy != nil {
		conditions = append(conditions, fmt.Sprintf("creator_id = $%d", paramCounter))
		args = append(args, *filter.CreatedBy)
		paramCounter++
	}

	// Add WHERE clause if we have conditions
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Add ORDER BY and LIMIT
	query += " ORDER BY created_at DESC"
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", paramCounter)
		args = append(args, filter.Limit)
		paramCounter++
		if filter.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", paramCounter)
			args = append(args, filter.Offset)
			paramCounter++
		}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bounties []*Bounty
	for rows.Next() {
		bounty, err := r.scanBounty(rows)
		if err != nil {
			return nil, err
		}
		bounties = append(bounties, bounty)
	}

	return bounties, rows.Err()
}

// ClaimBounty claims a bounty for a user
func (r *Repository) ClaimBounty(id string, claimedBy string) error {
	now := time.Now()
	query := `
		UPDATE bounties 
		SET status = $1, claimer_id = $2, claimed_at = $3, updated_at = $4
		WHERE id = $5 AND status = 'open'
	`

	result, err := r.db.Exec(query, StatusClaimed, claimedBy, now, now, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("bounty not found or not available for claiming")
	}

	return nil
}

// UpdateBountyStatus updates the status of a bounty
func (r *Repository) UpdateBountyStatus(id string, status BountyStatus) error {
	now := time.Now()
	query := `UPDATE bounties SET status = $1, updated_at = $2 WHERE id = $3`

	result, err := r.db.Exec(query, status, now, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("bounty not found")
	}

	return nil
}

// scanBounty is a helper to scan a row into a Bounty struct
func (r *Repository) scanBounty(scanner interface{}) (*Bounty, error) {
	var bounty Bounty
	var skillsJSON sql.NullString
	var id int
	var payoutCents int
	var creatorID, claimerID sql.NullInt64

	var row interface{ Scan(...interface{}) error }
	if r, ok := scanner.(*sql.Row); ok {
		row = r
	} else if rows, ok := scanner.(*sql.Rows); ok {
		row = rows
	} else {
		return nil, fmt.Errorf("invalid scanner type")
	}

	err := row.Scan(
		&id, &bounty.Title, &bounty.Description, &bounty.Requirements,
		&payoutCents, &bounty.Deadline, &bounty.Status,
		&bounty.CreatedAt, &bounty.UpdatedAt, &creatorID, &claimerID,
		&bounty.ClaimedAt, &bounty.CompletedAt, &skillsJSON, &bounty.Tier,
	)

	if err != nil {
		return nil, err
	}

	// Convert database values to API format
	bounty.ID = fmt.Sprintf("%d", id)
	bounty.PayoutAmount = float64(payoutCents) / 100.0 // Convert cents to dollars
	bounty.Currency = "USD" // Default currency

	if creatorID.Valid {
		creatorIDStr := fmt.Sprintf("%d", creatorID.Int64)
		bounty.CreatedBy = creatorIDStr
	}
	
	if claimerID.Valid {
		claimerIDStr := fmt.Sprintf("%d", claimerID.Int64)
		bounty.ClaimedBy = &claimerIDStr
	}

	// Set default auto-generated flag based on convention
	bounty.AutoGenerated = strings.Contains(strings.ToLower(bounty.Description), "auto-generated") || 
						  strings.Contains(strings.ToLower(bounty.Title), "auto:")

	// Parse JSON fields
	if skillsJSON.Valid {
		json.Unmarshal([]byte(skillsJSON.String), &bounty.RequiredSkills)
	}

	return &bounty, nil
}