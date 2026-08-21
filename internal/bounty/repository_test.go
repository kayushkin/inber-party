package bounty

import (
	"database/sql"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/kayushkin/inber-party/internal/db"

	_ "github.com/mattn/go-sqlite3"
)

// createTestDB opens a SQLite mirror of the production schema — db.go's own migration
// list, translated, on a connection that enforces foreign keys.
//
// It used to write its own CREATE TABLE bounties. That copy was the fourth hand-written
// mirror in this repository and it had drifted the same way the other three had: three
// columns production has never had (currency, verified_at, auto_generated), three of
// production's missing (work_submission, verification_notes, submitted_at), and no
// REFERENCES clause at all where db.go declares two. So every test in this file could
// name a creator no agents row held, and PostgreSQL refuses that row.
func createTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test_bounties.db")

	sqlDB, err := db.OpenSQLiteMirrorOfProductionSchema(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	cleanup := func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("Failed to close test database: %v", err)
		}
	}

	return sqlDB, cleanup
}

// insertTestAgent creates one agent and returns its id as the string the repository
// stores in creator_id and claimer_id. bounties declares both as REFERENCES agents(id),
// so a test that invents an agent id is writing a row the product cannot hold.
func insertTestAgent(t *testing.T, sqlDB *sql.DB, name string) string {
	t.Helper()

	result, err := sqlDB.Exec(
		`INSERT INTO agents (name, title, class, avatar_emoji) VALUES (?, ?, ?, ?)`,
		name, "Tester", "rogue", "\U0001F9EA")
	if err != nil {
		t.Fatalf("insert test agent %q: %v", name, err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read the id of test agent %q: %v", name, err)
	}
	return strconv.FormatInt(id, 10)
}

func TestNewRepository(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	repo := NewRepository(db)
	if repo == nil {
		t.Fatal("NewRepository returned nil")
	}
	if repo.db != db {
		t.Error("Repository database not set correctly")
	}
}

func TestCreateBounty(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	repo := NewRepository(db)
	creatorID := insertTestAgent(t, db, "create-bounty-creator")

	// Test basic bounty creation
	bounty := &Bounty{
		Title:          "Fix bug in auth module",
		Description:    "There's a memory leak in the auth module",
		Requirements:   "Must include tests and documentation",
		PayoutAmount:   5.50,
		CreatedBy:      creatorID,
		RequiredSkills: []string{"Go", "Testing"},
	}

	err := repo.CreateBounty(bounty)
	if err != nil {
		t.Fatalf("CreateBounty failed: %v", err)
	}

	// Verify ID was set
	if bounty.ID == "" {
		t.Error("Bounty ID not set after creation")
	}

	// Verify defaults were set
	if bounty.Status != StatusOpen {
		t.Errorf("Expected status %s, got %s", StatusOpen, bounty.Status)
	}
	if bounty.Tier != TierGold { // 5.50 should be gold tier (5.0-19.99)
		t.Errorf("Expected tier %s, got %s", TierGold, bounty.Tier)
	}

	// Verify timestamps were set
	if bounty.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}
	if bounty.UpdatedAt.IsZero() {
		t.Error("UpdatedAt not set")
	}
}

func TestGetBounty(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	repo := NewRepository(db)
	creatorID := insertTestAgent(t, db, "get-bounty-creator")

	// Create a test bounty
	original := &Bounty{
		Title:        "Test bounty",
		Description:  "Test description",
		PayoutAmount: 10.00,
		CreatedBy:    creatorID,
	}

	err := repo.CreateBounty(original)
	if err != nil {
		t.Fatalf("Failed to create test bounty: %v", err)
	}

	// Retrieve the bounty
	retrieved, err := repo.GetBounty(original.ID)
	if err != nil {
		t.Fatalf("GetBounty failed: %v", err)
	}

	// Verify fields match
	if retrieved.ID != original.ID {
		t.Errorf("ID mismatch: expected %s, got %s", original.ID, retrieved.ID)
	}
	if retrieved.Title != original.Title {
		t.Errorf("Title mismatch: expected %s, got %s", original.Title, retrieved.Title)
	}
	if retrieved.PayoutAmount != original.PayoutAmount {
		t.Errorf("PayoutAmount mismatch: expected %f, got %f", original.PayoutAmount, retrieved.PayoutAmount)
	}
}

func TestGetBounty_NotFound(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	repo := NewRepository(db)

	// Try to get non-existent bounty
	_, err := repo.GetBounty("99999")
	if err == nil {
		t.Error("Expected error for non-existent bounty")
	}
	if err != sql.ErrNoRows {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestListBounties_Empty(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	repo := NewRepository(db)

	filter := BountyFilter{}
	bounties, err := repo.ListBounties(filter)
	if err != nil {
		t.Fatalf("ListBounties failed: %v", err)
	}

	if len(bounties) != 0 {
		t.Errorf("Expected 0 bounties, got %d", len(bounties))
	}
}

func TestListBounties_WithData(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	repo := NewRepository(db)

	// Two distinct creators, because the creator filter below has to tell them apart.
	prolificCreatorID := insertTestAgent(t, db, "prolific-creator")
	occasionalCreatorID := insertTestAgent(t, db, "occasional-creator")

	// Create test bounties with different attributes
	bounties := []*Bounty{
		{
			Title:        "High payout bounty",
			Description:  "Expensive task",
			PayoutAmount: 25.00, // Platinum tier
			CreatedBy:    prolificCreatorID,
		},
		{
			Title:        "Low payout bounty",
			Description:  "Cheap task",
			PayoutAmount: 0.50, // Bronze tier
			CreatedBy:    occasionalCreatorID,
		},
		{
			Title:        "Medium bounty",
			Description:  "Medium task",
			PayoutAmount: 3.00, // Silver tier
			CreatedBy:    prolificCreatorID,
		},
	}

	for _, b := range bounties {
		if err := repo.CreateBounty(b); err != nil {
			t.Fatalf("Failed to create test bounty: %v", err)
		}
	}

	// Test listing all
	filter := BountyFilter{}
	result, err := repo.ListBounties(filter)
	if err != nil {
		t.Fatalf("ListBounties failed: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("Expected 3 bounties, got %d", len(result))
	}

	// Test filtering by tier
	filter = BountyFilter{Tier: []BountyTier{TierPlatinum}}
	result, err = repo.ListBounties(filter)
	if err != nil {
		t.Fatalf("ListBounties with tier filter failed: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 platinum bounty, got %d", len(result))
	}
	if result[0].Tier != TierPlatinum {
		t.Errorf("Expected platinum tier, got %s", result[0].Tier)
	}

	// Test filtering by creator
	creator := prolificCreatorID
	filter = BountyFilter{CreatedBy: &creator}
	result, err = repo.ListBounties(filter)
	if err != nil {
		t.Fatalf("ListBounties with creator filter failed: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 bounties from the prolific creator, got %d", len(result))
	}

	// Test payout range filtering
	minPayout := 1.0
	maxPayout := 10.0
	filter = BountyFilter{MinPayout: &minPayout, MaxPayout: &maxPayout}
	result, err = repo.ListBounties(filter)
	if err != nil {
		t.Fatalf("ListBounties with payout filter failed: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 bounty in payout range, got %d", len(result))
	}
	if result[0].PayoutAmount != 3.00 {
		t.Errorf("Expected payout 3.00, got %f", result[0].PayoutAmount)
	}

	// Test limit
	filter = BountyFilter{Limit: 2}
	result, err = repo.ListBounties(filter)
	if err != nil {
		t.Fatalf("ListBounties with limit failed: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 bounties with limit, got %d", len(result))
	}
}

func TestListBounties_StatusFilter(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	repo := NewRepository(db)
	creatorID := insertTestAgent(t, db, "status-filter-creator")
	claimerID := insertTestAgent(t, db, "status-filter-claimer")

	// Create bounties and claim one
	bounty1 := &Bounty{
		Title:        "Open bounty",
		Description:  "Still available",
		PayoutAmount: 1.00,
		CreatedBy:    creatorID,
	}
	bounty2 := &Bounty{
		Title:        "Will be claimed",
		Description:  "Will be claimed soon",
		PayoutAmount: 2.00,
		CreatedBy:    creatorID,
	}

	err := repo.CreateBounty(bounty1)
	if err != nil {
		t.Fatalf("Failed to create bounty1: %v", err)
	}
	err = repo.CreateBounty(bounty2)
	if err != nil {
		t.Fatalf("Failed to create bounty2: %v", err)
	}

	// Claim bounty2
	err = repo.ClaimBounty(bounty2.ID, claimerID)
	if err != nil {
		t.Fatalf("Failed to claim bounty: %v", err)
	}

	// Test filtering by status
	filter := BountyFilter{Status: []BountyStatus{StatusOpen}}
	result, err := repo.ListBounties(filter)
	if err != nil {
		t.Fatalf("ListBounties with status filter failed: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 open bounty, got %d", len(result))
	}

	filter = BountyFilter{Status: []BountyStatus{StatusClaimed}}
	result, err = repo.ListBounties(filter)
	if err != nil {
		t.Fatalf("ListBounties with claimed status filter failed: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 claimed bounty, got %d", len(result))
	}
}

func TestClaimBounty(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	repo := NewRepository(db)
	creatorID := insertTestAgent(t, db, "claim-creator")

	// Create a test bounty
	bounty := &Bounty{
		Title:        "Test claim",
		Description:  "Test claiming",
		PayoutAmount: 1.00,
		CreatedBy:    creatorID,
	}

	err := repo.CreateBounty(bounty)
	if err != nil {
		t.Fatalf("Failed to create bounty: %v", err)
	}

	// Claim the bounty
	claimerID := insertTestAgent(t, db, "first-claimer")
	err = repo.ClaimBounty(bounty.ID, claimerID)
	if err != nil {
		t.Fatalf("ClaimBounty failed: %v", err)
	}

	// Verify bounty was claimed
	claimed, err := repo.GetBounty(bounty.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve claimed bounty: %v", err)
	}

	if claimed.Status != StatusClaimed {
		t.Errorf("Expected status %s, got %s", StatusClaimed, claimed.Status)
	}
	if claimed.ClaimedBy == nil || *claimed.ClaimedBy != claimerID {
		t.Errorf("Expected claimer %s, got %v", claimerID, claimed.ClaimedBy)
	}
	if claimed.ClaimedAt == nil || claimed.ClaimedAt.IsZero() {
		t.Error("ClaimedAt should be set")
	}

	// Try to claim already claimed bounty. The second claimer exists, so the only reason
	// this can fail is the one the test is about.
	secondClaimerID := insertTestAgent(t, db, "second-claimer")
	err = repo.ClaimBounty(bounty.ID, secondClaimerID)
	if err == nil {
		t.Error("Expected error when claiming already claimed bounty")
	}
}

func TestClaimBounty_NotFound(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	repo := NewRepository(db)

	// Try to claim non-existent bounty. The claimer is real, so a missing bounty is the
	// only thing left that can produce the error.
	claimerID := insertTestAgent(t, db, "claim-not-found-claimer")
	err := repo.ClaimBounty("99999", claimerID)
	if err == nil {
		t.Error("Expected error for non-existent bounty")
	}
}

func TestUpdateBountyStatus(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	repo := NewRepository(db)
	creatorID := insertTestAgent(t, db, "status-update-creator")

	// Create a test bounty
	bounty := &Bounty{
		Title:        "Test status update",
		Description:  "Test updating status",
		PayoutAmount: 1.00,
		CreatedBy:    creatorID,
	}

	err := repo.CreateBounty(bounty)
	if err != nil {
		t.Fatalf("Failed to create bounty: %v", err)
	}

	// Update status to completed
	err = repo.UpdateBountyStatus(bounty.ID, StatusCompleted)
	if err != nil {
		t.Fatalf("UpdateBountyStatus failed: %v", err)
	}

	// Verify status was updated
	updated, err := repo.GetBounty(bounty.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated bounty: %v", err)
	}

	if updated.Status != StatusCompleted {
		t.Errorf("Expected status %s, got %s", StatusCompleted, updated.Status)
	}

	// UpdatedAt should be more recent
	if !updated.UpdatedAt.After(updated.CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt")
	}
}

func TestUpdateBountyStatus_NotFound(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	repo := NewRepository(db)

	// Try to update non-existent bounty
	err := repo.UpdateBountyStatus("99999", StatusCompleted)
	if err == nil {
		t.Error("Expected error for non-existent bounty")
	}
}

// Test edge cases and error handling
func TestBountyRepository_EdgeCases(t *testing.T) {
	db, cleanup := createTestDB(t)
	defer cleanup()

	repo := NewRepository(db)
	creatorID := insertTestAgent(t, db, "edge-case-creator")

	t.Run("CreateBounty with nil deadline", func(t *testing.T) {
		bounty := &Bounty{
			Title:        "No deadline",
			Description:  "Test nil deadline",
			PayoutAmount: 1.00,
			CreatedBy:    creatorID,
			Deadline:     nil,
		}

		err := repo.CreateBounty(bounty)
		if err != nil {
			t.Fatalf("CreateBounty with nil deadline failed: %v", err)
		}

		retrieved, err := repo.GetBounty(bounty.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve bounty: %v", err)
		}
		if retrieved.Deadline != nil {
			t.Error("Expected nil deadline to remain nil")
		}
	})

	t.Run("CreateBounty with future deadline", func(t *testing.T) {
		future := time.Now().Add(24 * time.Hour)
		bounty := &Bounty{
			Title:        "Future deadline",
			Description:  "Test future deadline",
			PayoutAmount: 1.00,
			CreatedBy:    creatorID,
			Deadline:     &future,
		}

		err := repo.CreateBounty(bounty)
		if err != nil {
			t.Fatalf("CreateBounty with future deadline failed: %v", err)
		}

		retrieved, err := repo.GetBounty(bounty.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve bounty: %v", err)
		}
		if retrieved.Deadline == nil || !retrieved.Deadline.Equal(future) {
			t.Error("Deadline not preserved correctly")
		}
	})

	t.Run("ListBounties with multiple filters", func(t *testing.T) {
		// A creator of its own, so the creator filter selects this bounty alone out of the
		// ones the earlier subtests left behind.
		filteredCreatorID := insertTestAgent(t, db, "multi-filter-creator")

		// Create bounties with specific characteristics
		bounty := &Bounty{
			Title:        "Filtered bounty",
			Description:  "Should match multiple filters",
			PayoutAmount: 15.00, // Gold tier (5.0 <= amount < 20.0)
			CreatedBy:    filteredCreatorID,
		}

		err := repo.CreateBounty(bounty)
		if err != nil {
			t.Fatalf("Failed to create test bounty: %v", err)
		}

		creator := filteredCreatorID
		minPayout := 10.0
		filter := BountyFilter{
			Status:    []BountyStatus{StatusOpen},
			Tier:      []BountyTier{TierGold},
			CreatedBy: &creator,
			MinPayout: &minPayout,
		}

		result, err := repo.ListBounties(filter)
		if err != nil {
			t.Fatalf("ListBounties with multiple filters failed: %v", err)
		}

		if len(result) != 1 {
			t.Errorf("Expected 1 bounty matching all filters, got %d", len(result))
		}
	})
}

func TestGetTierFromPayout(t *testing.T) {
	tests := []struct {
		payout float64
		tier   BountyTier
	}{
		{0.50, TierBronze},
		{0.99, TierBronze},
		{1.00, TierSilver},
		{4.99, TierSilver},
		{5.00, TierGold},
		{19.99, TierGold},
		{20.00, TierPlatinum},
		{99.99, TierPlatinum},
		{100.00, TierLegendary},
		{1000.00, TierLegendary},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			tier := GetTierFromPayout(tt.payout)
			if tier != tt.tier {
				t.Errorf("GetTierFromPayout(%.2f) = %s, want %s", tt.payout, tier, tt.tier)
			}
		})
	}
}
