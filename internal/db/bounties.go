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

// ClaimBounty allows an agent to claim a bounty with reputation-based priority
func (db *DB) ClaimBounty(bountyID, claimerID int) error {
	// First check if bounty exists and is open
	var bounty Bounty
	err := db.QueryRow(`
		SELECT id, title, description, required_skills, tier 
		FROM bounties 
		WHERE id = $1 AND status = 'open'
	`, bountyID).Scan(&bounty.ID, &bounty.Title, &bounty.Description, (*StringSlice)(&bounty.RequiredSkills), &bounty.Tier)
	
	if err != nil {
		return fmt.Errorf("bounty not found or already claimed")
	}
	
	// Check agent's reputation in relevant domain
	domain := InferTaskDomain(bounty.Title, bounty.Description)
	agentReputation, err := db.getAgentReputationInDomain(claimerID, domain)
	if err != nil {
		// If no reputation exists, agent can still claim but gets lower priority
		agentReputation = &Reputation{Score: 100, TaskCount: 0, SuccessRate: 1.0}
	}
	
	// Check if this is a high-tier bounty that requires high reputation
	minReputationRequired := 0
	switch bounty.Tier {
	case "legendary":
		minReputationRequired = 800 // Legendary bounties require very high reputation
	case "gold":
		minReputationRequired = 600 // Gold bounties require high reputation
	case "silver":
		minReputationRequired = 400 // Silver bounties require moderate reputation
	case "bronze":
		minReputationRequired = 100 // Bronze bounties are open to all
	}
	
	// If agent doesn't meet minimum reputation and has completed fewer than 3 tasks in this domain, reject
	if agentReputation.Score < minReputationRequired && agentReputation.TaskCount < 3 {
		return fmt.Errorf("insufficient reputation for this tier of bounty (need %d, have %d)", minReputationRequired, agentReputation.Score)
	}
	
	// Attempt to claim the bounty
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

// getAgentReputationInDomain returns an agent's reputation in a specific domain
func (db *DB) getAgentReputationInDomain(agentID int, domain string) (*Reputation, error) {
	var rep Reputation
	err := db.QueryRow(`
		SELECT id, agent_id, domain, score, task_count, success_rate, last_update
		FROM reputation
		WHERE agent_id = $1 AND domain = $2
	`, agentID, domain).Scan(&rep.ID, &rep.AgentID, &rep.Domain, &rep.Score, &rep.TaskCount, &rep.SuccessRate, &rep.LastUpdate)
	
	if err != nil {
		return nil, err
	}
	
	return &rep, nil
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
	
	// If approved, pay out the gold to the claimer using the new payout tracking system
	if approved {
		// Get bounty details for payout
		bounty, err := db.GetBountyByID(bountyID)
		if err != nil {
			return fmt.Errorf("failed to get bounty for payout: %w", err)
		}
		if bounty != nil && bounty.ClaimerID != nil {
			err = db.RecordBountyPayout(bountyID, *bounty.ClaimerID, bounty.PayoutAmount)
			if err != nil {
				return fmt.Errorf("failed to record bounty payout: %w", err)
			}
			
			// Mark bounty as paid
			_, err = db.Exec(`
				UPDATE bounties 
				SET status = 'paid', updated_at = CURRENT_TIMESTAMP 
				WHERE id = $1
			`, bountyID)
			if err != nil {
				return fmt.Errorf("failed to mark bounty as paid: %w", err)
			}
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

// Verifier management methods

// CreateBountyVerifier adds a verifier to a bounty
func (db *DB) CreateBountyVerifier(verifier *BountyVerifier) error {
	configJSON, err := json.Marshal(verifier.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal verifier config: %w", err)
	}

	err = db.QueryRow(`
		INSERT INTO bounty_verifiers (bounty_id, verifier_type, config, required, weight)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`, verifier.BountyID, verifier.VerifierType, configJSON, verifier.Required, verifier.Weight).Scan(
		&verifier.ID, &verifier.CreatedAt, &verifier.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create bounty verifier: %w", err)
	}

	return nil
}

// GetBountyVerifiers returns all verifiers for a bounty
func (db *DB) GetBountyVerifiers(bountyID int) ([]BountyVerifier, error) {
	query := `
		SELECT id, bounty_id, verifier_type, config, required, weight, created_at, updated_at
		FROM bounty_verifiers
		WHERE bounty_id = $1
		ORDER BY created_at ASC
	`

	rows, err := db.Query(query, bountyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query bounty verifiers: %w", err)
	}
	defer rows.Close()

	var verifiers []BountyVerifier
	for rows.Next() {
		var v BountyVerifier
		var configJSON []byte

		err := rows.Scan(
			&v.ID, &v.BountyID, &v.VerifierType, &configJSON,
			&v.Required, &v.Weight, &v.CreatedAt, &v.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bounty verifier: %w", err)
		}

		if err := json.Unmarshal(configJSON, &v.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal verifier config: %w", err)
		}

		verifiers = append(verifiers, v)
	}

	return verifiers, nil
}

// CreateVerifierResult records the result of running a verifier
func (db *DB) CreateVerifierResult(result *VerifierResult) error {
	resultDataJSON, err := json.Marshal(result.ResultData)
	if err != nil {
		return fmt.Errorf("failed to marshal result data: %w", err)
	}

	err = db.QueryRow(`
		INSERT INTO verifier_results (bounty_id, verifier_id, status, result_data, error_message, checked_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, checked_at
	`, result.BountyID, result.VerifierID, result.Status, resultDataJSON, result.ErrorMessage, result.CheckedBy).Scan(
		&result.ID, &result.CheckedAt)

	if err != nil {
		return fmt.Errorf("failed to create verifier result: %w", err)
	}

	return nil
}

// GetVerifierResults returns all results for a bounty
func (db *DB) GetVerifierResults(bountyID int) ([]VerifierResult, error) {
	query := `
		SELECT id, bounty_id, verifier_id, status, result_data, error_message, checked_at, checked_by
		FROM verifier_results
		WHERE bounty_id = $1
		ORDER BY checked_at DESC
	`

	rows, err := db.Query(query, bountyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query verifier results: %w", err)
	}
	defer rows.Close()

	var results []VerifierResult
	for rows.Next() {
		var r VerifierResult
		var resultDataJSON []byte

		err := rows.Scan(
			&r.ID, &r.BountyID, &r.VerifierID, &r.Status,
			&resultDataJSON, &r.ErrorMessage, &r.CheckedAt, &r.CheckedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan verifier result: %w", err)
		}

		if err := json.Unmarshal(resultDataJSON, &r.ResultData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result data: %w", err)
		}

		results = append(results, r)
	}

	return results, nil
}

// GetBountyWithVerifiers returns a bounty with its verifiers and results
func (db *DB) GetBountyWithVerifiers(id int) (*BountyWithVerifiers, error) {
	detail, err := db.GetBountyDetail(id)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return nil, nil
	}

	verifiers, err := db.GetBountyVerifiers(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get verifiers: %w", err)
	}

	results, err := db.GetVerifierResults(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get verifier results: %w", err)
	}

	return &BountyWithVerifiers{
		BountyDetail: *detail,
		Verifiers:    verifiers,
		Results:      results,
	}, nil
}

// RunBountyVerifiers runs all verifiers for a bounty and records results
func (db *DB) RunBountyVerifiers(bountyID int, registry interface{}) error {
	// This method would coordinate running all verifiers for a bounty
	// Implementation depends on the verifier registry interface
	// For now, it's a placeholder for the verification orchestration
	return fmt.Errorf("RunBountyVerifiers not implemented yet")
}