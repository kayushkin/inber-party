package db

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// StringSlice handles PostgreSQL array serialization for required_skills
type StringSlice []string

func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	
	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into StringSlice", value)
	}
	
	return json.Unmarshal(data, s)
}

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// GetBounties returns all bounties with optional filtering
func (db *DB) GetBounties(status string, limit, offset int) ([]Bounty, error) {
	query := `
		SELECT id, title, description, requirements, payout_amount, status, deadline,
		       creator_id, claimer_id, work_submission, verification_notes,
		       required_skills, tier, claimed_at, submitted_at, completed_at,
		       created_at, updated_at
		FROM bounties
		WHERE ($1 = '' OR status = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := db.Query(query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query bounties: %w", err)
	}
	defer rows.Close()
	
	var bounties []Bounty
	for rows.Next() {
		var b Bounty
		var requiredSkills StringSlice
		
		err := rows.Scan(
			&b.ID, &b.Title, &b.Description, &b.Requirements, &b.PayoutAmount,
			&b.Status, &b.Deadline, &b.CreatorID, &b.ClaimerID,
			&b.WorkSubmission, &b.VerificationNotes, &requiredSkills,
			&b.Tier, &b.ClaimedAt, &b.SubmittedAt, &b.CompletedAt,
			&b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bounty: %w", err)
		}
		
		b.RequiredSkills = []string(requiredSkills)
		bounties = append(bounties, b)
	}
	
	return bounties, nil
}

// GetBountyByID returns a single bounty by ID
func (db *DB) GetBountyByID(id int) (*Bounty, error) {
	query := `
		SELECT id, title, description, requirements, payout_amount, status, deadline,
		       creator_id, claimer_id, work_submission, verification_notes,
		       required_skills, tier, claimed_at, submitted_at, completed_at,
		       created_at, updated_at
		FROM bounties WHERE id = $1
	`
	
	var b Bounty
	var requiredSkills StringSlice
	
	err := db.QueryRow(query, id).Scan(
		&b.ID, &b.Title, &b.Description, &b.Requirements, &b.PayoutAmount,
		&b.Status, &b.Deadline, &b.CreatorID, &b.ClaimerID,
		&b.WorkSubmission, &b.VerificationNotes, &requiredSkills,
		&b.Tier, &b.ClaimedAt, &b.SubmittedAt, &b.CompletedAt,
		&b.CreatedAt, &b.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get bounty: %w", err)
	}
	
	b.RequiredSkills = []string(requiredSkills)
	return &b, nil
}

// CreateBounty creates a new bounty
func (db *DB) CreateBounty(b *Bounty) error {
	// Determine tier based on payout amount
	tier := "bronze"
	if b.PayoutAmount >= 1000 {
		tier = "legendary"
	} else if b.PayoutAmount >= 500 {
		tier = "gold"
	} else if b.PayoutAmount >= 100 {
		tier = "silver"
	}
	
	query := `
		INSERT INTO bounties (title, description, requirements, payout_amount, 
		                     creator_id, required_skills, tier, deadline)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	
	requiredSkills := StringSlice(b.RequiredSkills)
	
	err := db.QueryRow(query, b.Title, b.Description, b.Requirements, 
		b.PayoutAmount, b.CreatorID, requiredSkills, tier, b.Deadline).Scan(
		&b.ID, &b.CreatedAt, &b.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create bounty: %w", err)
	}
	
	b.Tier = tier
	b.Status = "open"
	return nil
}

// ClaimBounty allows an agent to claim a bounty
func (db *DB) ClaimBounty(bountyID, claimerID int) error {
	query := `
		UPDATE bounties 
		SET claimer_id = $1, status = 'claimed', claimed_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND status = 'open'
	`
	
	result, err := db.Exec(query, claimerID, bountyID)
	if err != nil {
		return fmt.Errorf("failed to claim bounty: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("bounty not found or already claimed")
	}
	
	return nil
}

// SubmitWork allows the claimer to submit their completed work
func (db *DB) SubmitWork(bountyID int, workSubmission string) error {
	query := `
		UPDATE bounties 
		SET work_submission = $1, status = 'submitted', submitted_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND status = 'claimed'
	`
	
	result, err := db.Exec(query, workSubmission, bountyID)
	if err != nil {
		return fmt.Errorf("failed to submit work: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("bounty not found or not in claimed status")
	}
	
	return nil
}

// VerifyBounty allows the creator to verify and approve/reject submitted work
func (db *DB) VerifyBounty(bountyID int, approved bool, notes string) error {
	status := "rejected"
	completedAt := (*time.Time)(nil)
	now := time.Now()
	
	if approved {
		status = "completed"
		completedAt = &now
	}
	
	query := `
		UPDATE bounties 
		SET status = $1, verification_notes = $2, completed_at = $3,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $4 AND status = 'submitted'
	`
	
	result, err := db.Exec(query, status, notes, completedAt, bountyID)
	if err != nil {
		return fmt.Errorf("failed to verify bounty: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("bounty not found or not in submitted status")
	}
	
	// If approved, pay out the gold to the claimer
	if approved {
		err = db.payoutBounty(bountyID)
		if err != nil {
			return fmt.Errorf("failed to payout bounty: %w", err)
		}
	}
	
	return nil
}

// payoutBounty transfers the bounty payout to the claimer
func (db *DB) payoutBounty(bountyID int) error {
	query := `
		UPDATE agents 
		SET gold = gold + b.payout_amount,
		    updated_at = CURRENT_TIMESTAMP
		FROM bounties b 
		WHERE agents.id = b.claimer_id AND b.id = $1
	`
	
	_, err := db.Exec(query, bountyID)
	if err != nil {
		return fmt.Errorf("failed to payout gold: %w", err)
	}
	
	// Mark bounty as paid
	_, err = db.Exec(`
		UPDATE bounties 
		SET status = 'paid', updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1
	`, bountyID)
	
	return err
}

// GetBountyDetail returns a bounty with creator and claimer information
func (db *DB) GetBountyDetail(id int) (*BountyDetail, error) {
	bounty, err := db.GetBountyByID(id)
	if err != nil {
		return nil, err
	}
	if bounty == nil {
		return nil, nil
	}
	
	detail := &BountyDetail{Bounty: *bounty}
	
	// Get creator info
	creator, err := db.GetAgentByID(bounty.CreatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get creator: %w", err)
	}
	if creator != nil {
		detail.Creator = *creator
	}
	
	// Get claimer info if exists
	if bounty.ClaimerID != nil {
		claimer, err := db.GetAgentByID(*bounty.ClaimerID)
		if err != nil {
			return nil, fmt.Errorf("failed to get claimer: %w", err)
		}
		detail.Claimer = claimer
	}
	
	return detail, nil
}