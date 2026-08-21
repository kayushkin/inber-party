package db

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// The bounty settlement path — claim, submit, verify, dispute, resolve — and the
// reputation ledger beside it were executed by no test at all. A panic() at the entry
// of any of these functions left `go test ./...` green on main.
//
// ⚠️ These tests deliberately do NOT use db_test.go's migrateSQLite. That helper builds
// a `bounties` table with columns `payout`, `created_by` and `assigned_to`, and a
// `reputation` table keyed on agent_id alone with `total_score`/`average_rating`. The
// code in bounties.go and reputation.go queries `payout_amount`, `creator_id`,
// `claimer_id`, and a reputation row keyed on (agent_id, domain) with
// `score`/`task_count`/`success_rate`. The two schemas are different shapes, so a test
// written against the helper would pin a table this code never reads. The schema below
// is translated from the production migration in db.go — SERIAL to AUTOINCREMENT,
// VARCHAR/JSONB to TEXT, TIMESTAMP to DATETIME — and nothing else.

func setupSettlementDB(t *testing.T) *DB {
	t.Helper()

	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory database: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	database := &DB{sqlDB}

	schema := []string{
		`CREATE TABLE agents (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			title TEXT NOT NULL DEFAULT '',
			class TEXT NOT NULL DEFAULT '',
			level INTEGER DEFAULT 1,
			avatar_emoji TEXT NOT NULL DEFAULT '',
			gold INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE reputation (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			agent_id INTEGER REFERENCES agents(id) ON DELETE CASCADE,
			domain TEXT NOT NULL,
			score INTEGER DEFAULT 100,
			task_count INTEGER DEFAULT 0,
			success_rate REAL DEFAULT 1.0,
			last_update DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(agent_id, domain)
		)`,
		`CREATE TABLE bounties (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			requirements TEXT NOT NULL DEFAULT '',
			payout_amount INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'open',
			deadline DATETIME,
			creator_id INTEGER NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
			claimer_id INTEGER REFERENCES agents(id) ON DELETE SET NULL,
			work_submission TEXT,
			verification_notes TEXT,
			required_skills TEXT DEFAULT '[]',
			tier TEXT NOT NULL DEFAULT 'bronze',
			claimed_at DATETIME,
			submitted_at DATETIME,
			completed_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE payout_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			agent_id INTEGER NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
			amount INTEGER NOT NULL,
			source TEXT NOT NULL,
			source_id INTEGER,
			description TEXT NOT NULL,
			transaction_type TEXT NOT NULL CHECK (transaction_type IN ('credit', 'debit', 'adjustment')),
			balance_before INTEGER NOT NULL DEFAULT 0,
			balance_after INTEGER NOT NULL DEFAULT 0,
			processed_by INTEGER REFERENCES agents(id) ON DELETE SET NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Translated from schema/rating_system.sql, which is where bounty_ratings is
		// declared — it is NOT in db.go's migration list, so Migrate() never creates it.
		// The UNIQUE is on TWO columns, (bounty_id, rater_id): one rating per bounty per
		// rater, whoever is being rated. db_test.go's migrateSQLite builds a differently
		// named `ratings` table with a THREE-column UNIQUE that no production code reads;
		// a test written against that one would pin a constraint this code never meets.
		`CREATE TABLE bounty_ratings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bounty_id INTEGER NOT NULL REFERENCES bounties(id) ON DELETE CASCADE,
			rater_id INTEGER NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
			rated_id INTEGER NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
			rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
			comment TEXT DEFAULT '',
			categories TEXT DEFAULT '{}',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(bounty_id, rater_id)
		)`,
		`CREATE TABLE disputes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			bounty_id INTEGER NOT NULL REFERENCES bounties(id) ON DELETE CASCADE,
			claimer_id INTEGER NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
			creator_id INTEGER NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
			reason TEXT NOT NULL,
			evidence TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'open',
			admin_notes TEXT,
			resolution TEXT,
			resolved_by INTEGER REFERENCES agents(id) ON DELETE SET NULL,
			resolved_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(bounty_id, claimer_id)
		)`,
	}

	for _, stmt := range schema {
		if _, err := sqlDB.Exec(stmt); err != nil {
			t.Fatalf("create schema: %v\n%s", err, stmt)
		}
	}

	return database
}

func makeAgent(t *testing.T, database *DB, name string, gold int) int {
	t.Helper()
	var id int
	err := database.QueryRow(
		`INSERT INTO agents (name, title, class, avatar_emoji, gold) VALUES ($1, '', '', '', $2) RETURNING id`,
		name, gold).Scan(&id)
	if err != nil {
		t.Fatalf("insert agent %s: %v", name, err)
	}
	return id
}

// agentGold reads an agent's balance straight from the table rather than through any
// function under test, so a payout assertion cannot be satisfied by the code that made it.
func agentGold(t *testing.T, database *DB, agentID int) int {
	t.Helper()
	var gold int
	if err := database.QueryRow(`SELECT gold FROM agents WHERE id = $1`, agentID).Scan(&gold); err != nil {
		t.Fatalf("read gold for agent %d: %v", agentID, err)
	}
	return gold
}

func bountyStatus(t *testing.T, database *DB, bountyID int) string {
	t.Helper()
	var status string
	if err := database.QueryRow(`SELECT status FROM bounties WHERE id = $1`, bountyID).Scan(&status); err != nil {
		t.Fatalf("read status for bounty %d: %v", bountyID, err)
	}
	return status
}

func payoutRowCount(t *testing.T, database *DB, agentID int) int {
	t.Helper()
	var n int
	if err := database.QueryRow(`SELECT COUNT(*) FROM payout_entries WHERE agent_id = $1`, agentID).Scan(&n); err != nil {
		t.Fatalf("count payout entries: %v", err)
	}
	return n
}

// newBounty creates a bounty through CreateBounty so the tier rule is exercised on the
// way in, and returns the id.
func newBounty(t *testing.T, database *DB, creatorID, payout int, title, description string) *Bounty {
	t.Helper()
	b := &Bounty{
		Title:          title,
		Description:    description,
		Requirements:   "ship it",
		PayoutAmount:   payout,
		CreatorID:      creatorID,
		RequiredSkills: []string{"go"},
	}
	if err := database.CreateBounty(b); err != nil {
		t.Fatalf("create bounty: %v", err)
	}
	return b
}

// -----------------------------------------------------------------------------
// UpdateReputation
// -----------------------------------------------------------------------------

func TestUpdateReputationSeedsARowOnFirstUse(t *testing.T) {
	database := setupSettlementDB(t)
	agentID := makeAgent(t, database, "brigid", 0)

	if err := database.UpdateReputation(agentID, "coding", true); err != nil {
		t.Fatalf("UpdateReputation: %v", err)
	}

	reps, err := database.GetAgentReputation(agentID)
	if err != nil {
		t.Fatalf("GetAgentReputation: %v", err)
	}
	if len(reps) != 1 {
		t.Fatalf("expected exactly one reputation row, got %d", len(reps))
	}
	if reps[0].Domain != "coding" {
		t.Errorf("domain = %q, want %q", reps[0].Domain, "coding")
	}
	// Seeded at score 100, task_count 0, success_rate 1.0; one success then takes it to
	// 100 + 5 + int(10*1.0) = 115 with one task at a 1.0 rate.
	if reps[0].TaskCount != 1 {
		t.Errorf("task_count = %d, want 1", reps[0].TaskCount)
	}
	if reps[0].Score != 115 {
		t.Errorf("score = %d, want 115", reps[0].Score)
	}
	if reps[0].SuccessRate != 1.0 {
		t.Errorf("success_rate = %v, want 1.0", reps[0].SuccessRate)
	}
}

func TestUpdateReputationKeepsDomainsSeparate(t *testing.T) {
	database := setupSettlementDB(t)
	agentID := makeAgent(t, database, "lugh", 0)

	if err := database.UpdateReputation(agentID, "coding", true); err != nil {
		t.Fatalf("UpdateReputation coding: %v", err)
	}
	if err := database.UpdateReputation(agentID, "testing", false); err != nil {
		t.Fatalf("UpdateReputation testing: %v", err)
	}

	reps, err := database.GetAgentReputation(agentID)
	if err != nil {
		t.Fatalf("GetAgentReputation: %v", err)
	}
	if len(reps) != 2 {
		t.Fatalf("expected two reputation rows, got %d", len(reps))
	}

	byDomain := map[string]Reputation{}
	for _, r := range reps {
		byDomain[r.Domain] = r
	}
	if byDomain["coding"].Score != 115 {
		t.Errorf("coding score = %d, want 115", byDomain["coding"].Score)
	}
	// A single failure: rate becomes 0.0, change is -10 - int(5*1.0) = -15, so 85.
	if byDomain["testing"].Score != 85 {
		t.Errorf("testing score = %d, want 85", byDomain["testing"].Score)
	}
	if byDomain["testing"].SuccessRate != 0.0 {
		t.Errorf("testing success_rate = %v, want 0.0", byDomain["testing"].SuccessRate)
	}
}

// GetAgentReputation orders by score descending. Two domains with different scores
// separate the ordering from insertion order, so removing the ORDER BY is visible.
func TestGetAgentReputationOrdersByScoreDescending(t *testing.T) {
	database := setupSettlementDB(t)
	agentID := makeAgent(t, database, "morrigan", 0)

	// Insert the LOW-scoring domain first, so insertion order is the reverse of the
	// order under test and a dropped ORDER BY cannot pass by luck.
	if err := database.UpdateReputation(agentID, "testing", false); err != nil {
		t.Fatalf("seed testing: %v", err)
	}
	if err := database.UpdateReputation(agentID, "coding", true); err != nil {
		t.Fatalf("seed coding: %v", err)
	}

	reps, err := database.GetAgentReputation(agentID)
	if err != nil {
		t.Fatalf("GetAgentReputation: %v", err)
	}
	if len(reps) != 2 {
		t.Fatalf("expected two rows, got %d", len(reps))
	}
	if reps[0].Domain != "coding" || reps[1].Domain != "testing" {
		t.Errorf("order = [%s %s], want [coding testing] (descending score)",
			reps[0].Domain, reps[1].Domain)
	}
}

func TestUpdateReputationSuccessRateIsARunningAverage(t *testing.T) {
	database := setupSettlementDB(t)
	agentID := makeAgent(t, database, "dagda", 0)

	// Three successes then one failure: 3 of 4 = 0.75.
	for i := 0; i < 3; i++ {
		if err := database.UpdateReputation(agentID, "coding", true); err != nil {
			t.Fatalf("UpdateReputation success %d: %v", i, err)
		}
	}
	if err := database.UpdateReputation(agentID, "coding", false); err != nil {
		t.Fatalf("UpdateReputation failure: %v", err)
	}

	reps, err := database.GetAgentReputation(agentID)
	if err != nil {
		t.Fatalf("GetAgentReputation: %v", err)
	}
	if len(reps) != 1 {
		t.Fatalf("expected one row, got %d", len(reps))
	}
	if reps[0].TaskCount != 4 {
		t.Errorf("task_count = %d, want 4", reps[0].TaskCount)
	}
	if got := reps[0].SuccessRate; got < 0.749 || got > 0.751 {
		t.Errorf("success_rate = %v, want 0.75", got)
	}
}

// The score is clamped to [0, 1000]. The floor is the reachable end: enough failures
// drive the raw arithmetic negative, and the clamp is the only thing stopping it.
func TestUpdateReputationClampsTheScoreAtZero(t *testing.T) {
	database := setupSettlementDB(t)
	agentID := makeAgent(t, database, "balor", 0)

	for i := 0; i < 20; i++ {
		if err := database.UpdateReputation(agentID, "coding", false); err != nil {
			t.Fatalf("UpdateReputation failure %d: %v", i, err)
		}
	}

	reps, err := database.GetAgentReputation(agentID)
	if err != nil {
		t.Fatalf("GetAgentReputation: %v", err)
	}
	if reps[0].Score != 0 {
		t.Errorf("score = %d, want 0 (clamped floor)", reps[0].Score)
	}
}

// The ceiling is the other half of the clamp, and it is reachable — 60 clean successes
// carry a fresh agent from 100 past 1000 at 15 points each.
func TestUpdateReputationClampsTheScoreAtOneThousand(t *testing.T) {
	database := setupSettlementDB(t)
	agentID := makeAgent(t, database, "lugh-the-tireless", 0)

	for i := 0; i < 70; i++ {
		if err := database.UpdateReputation(agentID, "coding", true); err != nil {
			t.Fatalf("UpdateReputation success %d: %v", i, err)
		}
	}

	reps, err := database.GetAgentReputation(agentID)
	if err != nil {
		t.Fatalf("GetAgentReputation: %v", err)
	}
	if reps[0].Score != 1000 {
		t.Errorf("score = %d, want 1000 (clamped ceiling)", reps[0].Score)
	}
}

func TestUpdateReputationRejectsAnUnknownAgent(t *testing.T) {
	database := setupSettlementDB(t)
	if _, err := database.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	err := database.UpdateReputation(4242, "coding", true)
	if err == nil {
		t.Fatal("expected an error for an agent that does not exist, got nil")
	}
}

// TestCreateBountyStoresEachFieldInItsOwnColumn pins where CreateBounty's INSERT puts
// each value. Nothing else in this package reads a written bounty back by value except
// the required_skills round trip, so a value that lands in a neighbouring column is
// invisible to every other test here: the row still inserts, every fixture still gets
// its bounty, and no assertion ever looks.
//
// Every column is read with raw SQL rather than through GetBountyByID. That is the same
// rule agentGold follows — an assertion must not be satisfiable by the code that made
// the value — and it also keeps the check reachable: StringSlice.Scan runs the column
// through json.Unmarshal, so a drift that puts non-JSON in required_skills would fail
// the read and abort before any assertion ran, which is the failure mode this test
// exists to remove.
func TestCreateBountyStoresEachFieldInItsOwnColumn(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)

	// Every value is distinct, so a swap between any two columns changes what is
	// read back. A shared placeholder would make neighbouring columns interchangeable.
	b := &Bounty{
		Title:          "a title and nothing else",
		Description:    "a description and nothing else",
		Requirements:   "acceptance criteria and nothing else",
		PayoutAmount:   250,
		CreatorID:      creatorID,
		RequiredSkills: []string{"go", "sql"},
	}
	if err := database.CreateBounty(b); err != nil {
		t.Fatalf("CreateBounty: %v", err)
	}

	var title, description, requirements, requiredSkills, tier string
	var payoutAmount, storedCreatorID int
	err := database.QueryRow(`
		SELECT title, description, requirements, payout_amount, creator_id, required_skills, tier
		FROM bounties WHERE id = $1`, b.ID).Scan(
		&title, &description, &requirements, &payoutAmount, &storedCreatorID, &requiredSkills, &tier)
	if err != nil {
		t.Fatalf("read the stored row: %v", err)
	}

	if title != "a title and nothing else" {
		t.Errorf("title column = %q, want the title", title)
	}
	if description != "a description and nothing else" {
		t.Errorf("description column = %q, want the description", description)
	}
	if requirements != "acceptance criteria and nothing else" {
		t.Errorf("requirements column = %q, want the requirements", requirements)
	}
	if payoutAmount != 250 {
		t.Errorf("payout_amount column = %d, want 250", payoutAmount)
	}
	if storedCreatorID != creatorID {
		t.Errorf("creator_id column = %d, want %d", storedCreatorID, creatorID)
	}
	if requiredSkills != `["go","sql"]` {
		t.Errorf("required_skills column = %q, want the encoded skills", requiredSkills)
	}
	// 250 gold is gold tier; the value is derived here rather than handed in, and
	// TestCreateBountyDerivesTheTierFromThePayout pins the thresholds themselves.
	if tier != "gold" {
		t.Errorf("tier column = %q, want %q", tier, "gold")
	}
}

// -----------------------------------------------------------------------------
// The claim gate — CreateBounty's tier rule and ClaimBounty's reputation floor
// -----------------------------------------------------------------------------

func TestCreateBountyDerivesTheTierFromThePayout(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)

	cases := []struct {
		payout int
		want   string
	}{
		{0, "bronze"},
		{49, "bronze"},
		{50, "silver"},
		{199, "silver"},
		{200, "gold"},
		{499, "gold"},
		{500, "legendary"},
		{5000, "legendary"},
	}

	for _, c := range cases {
		b := newBounty(t, database, creatorID, c.payout, "task", "a task")
		if b.Tier != c.want {
			t.Errorf("payout %d: tier = %q, want %q", c.payout, b.Tier, c.want)
		}
		// The tier must reach the row, not just the struct the caller handed in.
		var stored string
		if err := database.QueryRow(`SELECT tier FROM bounties WHERE id = $1`, b.ID).Scan(&stored); err != nil {
			t.Fatalf("read stored tier: %v", err)
		}
		if stored != c.want {
			t.Errorf("payout %d: stored tier = %q, want %q", c.payout, stored, c.want)
		}
	}
}

func TestClaimBountyRefusesAnAgentBelowTheTierFloor(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "novice", 0)

	// 500 gold makes this legendary, whose floor is 750. A fresh agent has no
	// reputation row at all, so ClaimBounty falls back to score 100, task_count 0.
	b := newBounty(t, database, creatorID, 500, "rewrite the parser", "implement a new parser")

	err := database.ClaimBounty(b.ID, claimerID)
	if err == nil {
		t.Fatal("expected a legendary bounty to refuse an agent with no reputation")
	}

	// Assert the effect, not the reply: a refused claim must leave the row open and
	// unassigned, or a bounty is locked to an agent that was told it could not have it.
	if got := bountyStatus(t, database, b.ID); got != "open" {
		t.Errorf("status after a refused claim = %q, want %q", got, "open")
	}
	var claimer sql.NullInt64
	if err := database.QueryRow(`SELECT claimer_id FROM bounties WHERE id = $1`, b.ID).Scan(&claimer); err != nil {
		t.Fatalf("read claimer_id: %v", err)
	}
	if claimer.Valid {
		t.Errorf("claimer_id = %d after a refused claim, want NULL", claimer.Int64)
	}
}

func TestClaimBountyAdmitsAnAgentAtTheTierFloor(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "veteran", 0)

	// "implement a new parser" infers the coding domain. Seed a reputation there that
	// clears the silver floor of 250.
	//
	// ⚠️ task_count is deliberately 2, not 10. The gate refuses only when the score is
	// below the floor AND the agent has fewer than 3 tasks in the domain, so an agent
	// with 3 or more tasks is admitted no matter what the floor is — and this test
	// would then pass against any floor value at all, including an absurd one. The
	// score is the only thing admitting this agent.
	if _, err := database.Exec(
		`INSERT INTO reputation (agent_id, domain, score, task_count, success_rate)
		 VALUES ($1, 'coding', 300, 2, 0.9)`, claimerID); err != nil {
		t.Fatalf("seed reputation: %v", err)
	}

	b := newBounty(t, database, creatorID, 100, "implement the parser", "implement a new parser feature")
	if b.Tier != "silver" {
		t.Fatalf("fixture wants a silver bounty, got %q", b.Tier)
	}

	if err := database.ClaimBounty(b.ID, claimerID); err != nil {
		t.Fatalf("ClaimBounty: %v", err)
	}

	if got := bountyStatus(t, database, b.ID); got != "claimed" {
		t.Errorf("status = %q, want %q", got, "claimed")
	}
	stored, err := database.GetBountyByID(b.ID)
	if err != nil {
		t.Fatalf("GetBountyByID: %v", err)
	}
	if stored.ClaimerID == nil || *stored.ClaimerID != claimerID {
		t.Errorf("claimer_id = %v, want %d", stored.ClaimerID, claimerID)
	}
	if stored.ClaimedAt == nil {
		t.Error("claimed_at was not stamped")
	}
}

// The reputation floor has a second door: three completed tasks in the domain admit an
// agent whose score is still under the floor. Both halves of that AND are load-bearing.
func TestClaimBountyAdmitsALowScoreAgentWithEnoughTasks(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "journeyman", 0)

	// Score 10 is far below silver's 250, but task_count 3 opens the second door.
	if _, err := database.Exec(
		`INSERT INTO reputation (agent_id, domain, score, task_count, success_rate)
		 VALUES ($1, 'coding', 10, 3, 0.5)`, claimerID); err != nil {
		t.Fatalf("seed reputation: %v", err)
	}

	b := newBounty(t, database, creatorID, 100, "implement the parser", "implement a new parser feature")

	if err := database.ClaimBounty(b.ID, claimerID); err != nil {
		t.Fatalf("ClaimBounty should admit a low score with 3 tasks, got: %v", err)
	}
	if got := bountyStatus(t, database, b.ID); got != "claimed" {
		t.Errorf("status = %q, want %q", got, "claimed")
	}
}

func TestClaimBountyRefusesAnAlreadyClaimedBounty(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	firstID := makeAgent(t, database, "first", 0)
	secondID := makeAgent(t, database, "second", 0)

	b := newBounty(t, database, creatorID, 10, "small chore", "tidy the readme")
	if err := database.ClaimBounty(b.ID, firstID); err != nil {
		t.Fatalf("first ClaimBounty: %v", err)
	}

	if err := database.ClaimBounty(b.ID, secondID); err == nil {
		t.Fatal("expected the second claim to be refused")
	}

	stored, err := database.GetBountyByID(b.ID)
	if err != nil {
		t.Fatalf("GetBountyByID: %v", err)
	}
	if stored.ClaimerID == nil || *stored.ClaimerID != firstID {
		t.Errorf("claimer_id = %v, want the first claimer %d", stored.ClaimerID, firstID)
	}
}

// -----------------------------------------------------------------------------
// SubmitWork and VerifyBounty
// -----------------------------------------------------------------------------

// claimAndSubmit drives a bounty to the 'submitted' state, which is the only state
// VerifyBounty acts on.
func claimAndSubmit(t *testing.T, database *DB, creatorID, claimerID, payout int) *Bounty {
	t.Helper()
	b := newBounty(t, database, creatorID, payout, "tidy the readme", "documentation chore")
	if err := database.ClaimBounty(b.ID, claimerID); err != nil {
		t.Fatalf("ClaimBounty: %v", err)
	}
	if err := database.SubmitWork(b.ID, "here is the work"); err != nil {
		t.Fatalf("SubmitWork: %v", err)
	}
	return b
}

func TestSubmitWorkOnlyActsOnAClaimedBounty(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)

	// Open, never claimed.
	b := newBounty(t, database, creatorID, 10, "tidy the readme", "documentation chore")

	if err := database.SubmitWork(b.ID, "work from nobody"); err == nil {
		t.Fatal("expected SubmitWork to refuse an unclaimed bounty")
	}
	if got := bountyStatus(t, database, b.ID); got != "open" {
		t.Errorf("status = %q, want %q — a refused submission must not move the bounty", got, "open")
	}
	var submission sql.NullString
	if err := database.QueryRow(`SELECT work_submission FROM bounties WHERE id = $1`, b.ID).Scan(&submission); err != nil {
		t.Fatalf("read work_submission: %v", err)
	}
	if submission.Valid {
		t.Errorf("work_submission = %q, want NULL — refused work must not be stored", submission.String)
	}
}

func TestVerifyBountyApprovedPaysTheClaimerAndMarksItPaid(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "claimer", 40)

	b := claimAndSubmit(t, database, creatorID, claimerID, 25)

	if err := database.VerifyBounty(b.ID, true, "good work"); err != nil {
		t.Fatalf("VerifyBounty: %v", err)
	}

	// Three separate effects, and the reply reports none of them.
	if got := bountyStatus(t, database, b.ID); got != "paid" {
		t.Errorf("status = %q, want %q", got, "paid")
	}
	if got := agentGold(t, database, claimerID); got != 65 {
		t.Errorf("claimer gold = %d, want 65 (40 + 25)", got)
	}
	if got := payoutRowCount(t, database, claimerID); got != 1 {
		t.Errorf("payout entries = %d, want 1", got)
	}

	stored, err := database.GetBountyByID(b.ID)
	if err != nil {
		t.Fatalf("GetBountyByID: %v", err)
	}
	if stored.CompletedAt == nil {
		t.Error("completed_at was not stamped on an approved bounty")
	}
	if stored.VerificationNotes == nil || *stored.VerificationNotes != "good work" {
		t.Errorf("verification_notes = %v, want %q", stored.VerificationNotes, "good work")
	}
}

func TestVerifyBountyRejectedPaysNobody(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "claimer", 40)

	b := claimAndSubmit(t, database, creatorID, claimerID, 25)

	if err := database.VerifyBounty(b.ID, false, "does not meet the requirements"); err != nil {
		t.Fatalf("VerifyBounty: %v", err)
	}

	if got := bountyStatus(t, database, b.ID); got != "rejected" {
		t.Errorf("status = %q, want %q", got, "rejected")
	}
	if got := agentGold(t, database, claimerID); got != 40 {
		t.Errorf("claimer gold = %d, want 40 — a rejection must not pay", got)
	}
	if got := payoutRowCount(t, database, claimerID); got != 0 {
		t.Errorf("payout entries = %d, want 0", got)
	}

	stored, err := database.GetBountyByID(b.ID)
	if err != nil {
		t.Fatalf("GetBountyByID: %v", err)
	}
	if stored.CompletedAt != nil {
		t.Errorf("completed_at = %v on a rejected bounty, want NULL", stored.CompletedAt)
	}
}

func TestVerifyBountyRefusesABountyThatWasNeverSubmitted(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "claimer", 40)

	b := newBounty(t, database, creatorID, 25, "tidy the readme", "documentation chore")
	if err := database.ClaimBounty(b.ID, claimerID); err != nil {
		t.Fatalf("ClaimBounty: %v", err)
	}
	// Claimed but never submitted.

	if err := database.VerifyBounty(b.ID, true, "approved anyway"); err == nil {
		t.Fatal("expected VerifyBounty to refuse a bounty that was never submitted")
	}
	if got := agentGold(t, database, claimerID); got != 40 {
		t.Errorf("claimer gold = %d, want 40 — a refused verification must not pay", got)
	}
	if got := bountyStatus(t, database, b.ID); got != "claimed" {
		t.Errorf("status = %q, want %q", got, "claimed")
	}
}

// Approving twice must not pay twice. The status guard is the only thing standing
// between a double approval and a double payout.
func TestVerifyBountyApprovedTwicePaysOnce(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "claimer", 0)

	b := claimAndSubmit(t, database, creatorID, claimerID, 25)

	if err := database.VerifyBounty(b.ID, true, "good work"); err != nil {
		t.Fatalf("first VerifyBounty: %v", err)
	}
	if err := database.VerifyBounty(b.ID, true, "good work again"); err == nil {
		t.Fatal("expected the second approval to be refused")
	}

	if got := agentGold(t, database, claimerID); got != 25 {
		t.Errorf("claimer gold = %d, want 25 — the payout must not repeat", got)
	}
	if got := payoutRowCount(t, database, claimerID); got != 1 {
		t.Errorf("payout entries = %d, want 1", got)
	}
}

// A bounty in 'submitted' with no claimer cannot happen through the API, but the payout
// branch guards for it explicitly, and what that guard chooses matters: it leaves the
// bounty 'completed' rather than 'paid'.
func TestVerifyBountyLeavesAClaimerlessBountyCompletedNotPaid(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)

	b := newBounty(t, database, creatorID, 25, "tidy the readme", "documentation chore")
	if _, err := database.Exec(
		`UPDATE bounties SET status = 'submitted', work_submission = 'x' WHERE id = $1`, b.ID); err != nil {
		t.Fatalf("force submitted with no claimer: %v", err)
	}

	if err := database.VerifyBounty(b.ID, true, "approved"); err != nil {
		t.Fatalf("VerifyBounty: %v", err)
	}

	if got := bountyStatus(t, database, b.ID); got != "completed" {
		t.Errorf("status = %q, want %q — with no claimer there is nobody to pay", got, "completed")
	}
}

// -----------------------------------------------------------------------------
// CreateDispute
// -----------------------------------------------------------------------------

// rejectedBounty drives a bounty all the way to 'rejected', the only state a dispute
// can be filed against.
func rejectedBounty(t *testing.T, database *DB, creatorID, claimerID, payout int) *Bounty {
	t.Helper()
	b := claimAndSubmit(t, database, creatorID, claimerID, payout)
	if err := database.VerifyBounty(b.ID, false, "rejected"); err != nil {
		t.Fatalf("VerifyBounty reject: %v", err)
	}
	return b
}

func TestCreateDisputeMovesTheBountyToDisputed(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "claimer", 0)

	b := rejectedBounty(t, database, creatorID, claimerID, 25)

	d := &Dispute{BountyID: b.ID, ClaimerID: claimerID, Reason: "the work met the spec", Evidence: "see the diff"}
	if err := database.CreateDispute(d); err != nil {
		t.Fatalf("CreateDispute: %v", err)
	}

	if d.ID == 0 {
		t.Error("CreateDispute did not populate the dispute id")
	}
	// CreateDispute reads the creator off the bounty rather than trusting the caller.
	if d.CreatorID != creatorID {
		t.Errorf("creator_id = %d, want %d (taken from the bounty)", d.CreatorID, creatorID)
	}
	if d.Status != "open" {
		t.Errorf("status = %q, want %q", d.Status, "open")
	}
	if got := bountyStatus(t, database, b.ID); got != "disputed" {
		t.Errorf("bounty status = %q, want %q", got, "disputed")
	}
}

func TestCreateDisputeRefusesABountyThatWasNotRejected(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "claimer", 0)

	// Approved, not rejected.
	b := claimAndSubmit(t, database, creatorID, claimerID, 25)
	if err := database.VerifyBounty(b.ID, true, "good work"); err != nil {
		t.Fatalf("VerifyBounty: %v", err)
	}

	d := &Dispute{BountyID: b.ID, ClaimerID: claimerID, Reason: "greedy", Evidence: ""}
	if err := database.CreateDispute(d); err == nil {
		t.Fatal("expected CreateDispute to refuse a bounty that was not rejected")
	}
	if got := bountyStatus(t, database, b.ID); got != "paid" {
		t.Errorf("bounty status = %q, want %q — a refused dispute must not move it", got, "paid")
	}
}

func TestCreateDisputeRefusesAnAgentWhoIsNotTheClaimer(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "claimer", 0)
	strangerID := makeAgent(t, database, "stranger", 0)

	b := rejectedBounty(t, database, creatorID, claimerID, 25)

	d := &Dispute{BountyID: b.ID, ClaimerID: strangerID, Reason: "not mine but I mind", Evidence: ""}
	err := database.CreateDispute(d)
	if err == nil {
		t.Fatal("expected CreateDispute to refuse an agent who did not claim the bounty")
	}
	// This rejection and the not-rejected rejection above must be distinguishable: they
	// are different faults and the caller can only tell them apart by the message.
	if err.Error() != "only the claimer can dispute this bounty" {
		t.Errorf("error = %q, want the claimer-specific message", err.Error())
	}
	if got := bountyStatus(t, database, b.ID); got != "rejected" {
		t.Errorf("bounty status = %q, want %q", got, "rejected")
	}
}

func TestCreateDisputeRefusesASecondDisputeOnTheSameBounty(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "claimer", 0)

	b := rejectedBounty(t, database, creatorID, claimerID, 25)

	first := &Dispute{BountyID: b.ID, ClaimerID: claimerID, Reason: "first", Evidence: ""}
	if err := database.CreateDispute(first); err != nil {
		t.Fatalf("first CreateDispute: %v", err)
	}

	// The bounty is now 'disputed', so a second attempt is refused by the status check
	// before the duplicate check is ever consulted. That ordering is the behaviour:
	// the caller is told the bounty is not rejected, not that a dispute exists.
	second := &Dispute{BountyID: b.ID, ClaimerID: claimerID, Reason: "second", Evidence: ""}
	err := database.CreateDispute(second)
	if err == nil {
		t.Fatal("expected the second dispute to be refused")
	}

	var n int
	if err := database.QueryRow(`SELECT COUNT(*) FROM disputes WHERE bounty_id = $1`, b.ID).Scan(&n); err != nil {
		t.Fatalf("count disputes: %v", err)
	}
	if n != 1 {
		t.Errorf("disputes on this bounty = %d, want 1", n)
	}
}

// -----------------------------------------------------------------------------
// ResolveDispute
// -----------------------------------------------------------------------------

// disputedBounty drives a bounty to 'disputed' and returns both ids.
func disputedBounty(t *testing.T, database *DB, creatorID, claimerID, payout int) (*Bounty, *Dispute) {
	t.Helper()
	b := rejectedBounty(t, database, creatorID, claimerID, payout)
	d := &Dispute{BountyID: b.ID, ClaimerID: claimerID, Reason: "the work met the spec", Evidence: "see the diff"}
	if err := database.CreateDispute(d); err != nil {
		t.Fatalf("CreateDispute: %v", err)
	}
	return b, d
}

func TestResolveDisputeInFavorOfClaimerPaysOut(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "claimer", 10)
	adminID := makeAgent(t, database, "admin", 0)

	b, d := disputedBounty(t, database, creatorID, claimerID, 25)

	if err := database.ResolveDispute(d.ID, true, "the spec was ambiguous", adminID, "reviewed the diff"); err != nil {
		t.Fatalf("ResolveDispute: %v", err)
	}

	if got := bountyStatus(t, database, b.ID); got != "paid" {
		t.Errorf("bounty status = %q, want %q", got, "paid")
	}
	if got := agentGold(t, database, claimerID); got != 35 {
		t.Errorf("claimer gold = %d, want 35 (10 + 25)", got)
	}
	if got := payoutRowCount(t, database, claimerID); got != 1 {
		t.Errorf("payout entries = %d, want 1", got)
	}

	resolved, err := database.GetDispute(d.ID)
	if err != nil {
		t.Fatalf("GetDispute: %v", err)
	}
	if resolved.Status != "resolved_in_favor" {
		t.Errorf("dispute status = %q, want %q", resolved.Status, "resolved_in_favor")
	}
	if resolved.ResolvedBy == nil || *resolved.ResolvedBy != adminID {
		t.Errorf("resolved_by = %v, want %d", resolved.ResolvedBy, adminID)
	}
	if resolved.ResolvedAt == nil {
		t.Error("resolved_at was not stamped")
	}
	if resolved.AdminNotes == nil || *resolved.AdminNotes != "reviewed the diff" {
		t.Errorf("admin_notes = %v, want %q", resolved.AdminNotes, "reviewed the diff")
	}
}

func TestResolveDisputeAgainstClaimerPaysNobody(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "claimer", 10)
	adminID := makeAgent(t, database, "admin", 0)

	b, d := disputedBounty(t, database, creatorID, claimerID, 25)

	if err := database.ResolveDispute(d.ID, false, "the spec was clear", adminID, "reviewed the diff"); err != nil {
		t.Fatalf("ResolveDispute: %v", err)
	}

	if got := bountyStatus(t, database, b.ID); got != "rejected" {
		t.Errorf("bounty status = %q, want %q", got, "rejected")
	}
	if got := agentGold(t, database, claimerID); got != 10 {
		t.Errorf("claimer gold = %d, want 10 — losing a dispute must not pay", got)
	}
	if got := payoutRowCount(t, database, claimerID); got != 0 {
		t.Errorf("payout entries = %d, want 0", got)
	}

	resolved, err := database.GetDispute(d.ID)
	if err != nil {
		t.Fatalf("GetDispute: %v", err)
	}
	if resolved.Status != "resolved_against" {
		t.Errorf("dispute status = %q, want %q", resolved.Status, "resolved_against")
	}
}

// Both directions append their resolution to the bounty's verification notes, and the
// two banners must differ — they are the record of which way it went.
func TestResolveDisputeAppendsADirectionalBannerToTheNotes(t *testing.T) {
	database := setupSettlementDB(t)
	adminID := makeAgent(t, database, "admin", 0)

	for _, c := range []struct {
		name           string
		inFavor        bool
		wantBanner     string
		unwantedBanner string
	}{
		{"in favor", true, "[DISPUTE RESOLVED IN FAVOR OF CLAIMER]", "[DISPUTE RESOLVED AGAINST CLAIMER]"},
		{"against", false, "[DISPUTE RESOLVED AGAINST CLAIMER]", "[DISPUTE RESOLVED IN FAVOR OF CLAIMER]"},
	} {
		t.Run(c.name, func(t *testing.T) {
			creatorID := makeAgent(t, database, "creator-"+c.name, 0)
			claimerID := makeAgent(t, database, "claimer-"+c.name, 0)
			b, d := disputedBounty(t, database, creatorID, claimerID, 25)

			if err := database.ResolveDispute(d.ID, c.inFavor, "the ruling", adminID, ""); err != nil {
				t.Fatalf("ResolveDispute: %v", err)
			}

			var notes sql.NullString
			if err := database.QueryRow(`SELECT verification_notes FROM bounties WHERE id = $1`, b.ID).Scan(&notes); err != nil {
				t.Fatalf("read verification_notes: %v", err)
			}
			if !notes.Valid {
				t.Fatal("verification_notes is NULL, want the appended banner")
			}
			if !contains(notes.String, c.wantBanner) {
				t.Errorf("notes = %q, want it to contain %q", notes.String, c.wantBanner)
			}
			if contains(notes.String, c.unwantedBanner) {
				t.Errorf("notes = %q, must not contain the opposite banner %q", notes.String, c.unwantedBanner)
			}
			// The append must keep the original rejection note rather than replace it.
			if !contains(notes.String, "rejected") {
				t.Errorf("notes = %q, want the original rejection note preserved", notes.String)
			}
			if !contains(notes.String, "the ruling") {
				t.Errorf("notes = %q, want the resolution text", notes.String)
			}
		})
	}
}

func TestResolveDisputeRefusesADisputeThatIsAlreadyResolved(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "claimer", 0)
	adminID := makeAgent(t, database, "admin", 0)

	_, d := disputedBounty(t, database, creatorID, claimerID, 25)

	if err := database.ResolveDispute(d.ID, true, "first ruling", adminID, ""); err != nil {
		t.Fatalf("first ResolveDispute: %v", err)
	}
	err := database.ResolveDispute(d.ID, true, "second ruling", adminID, "")
	if err == nil {
		t.Fatal("expected a resolved dispute to refuse a second resolution")
	}

	// The guard exists to stop a second payout, so assert the money, not the message.
	if got := agentGold(t, database, claimerID); got != 25 {
		t.Errorf("claimer gold = %d, want 25 — resolving twice must not pay twice", got)
	}
	if got := payoutRowCount(t, database, claimerID); got != 1 {
		t.Errorf("payout entries = %d, want 1", got)
	}
}

func TestResolveDisputeRejectsAnUnknownDispute(t *testing.T) {
	database := setupSettlementDB(t)
	adminID := makeAgent(t, database, "admin", 0)

	err := database.ResolveDispute(9999, true, "ruling", adminID, "")
	if err == nil {
		t.Fatal("expected an error for a dispute that does not exist")
	}
	if err.Error() != "dispute not found" {
		t.Errorf("error = %q, want %q", err.Error(), "dispute not found")
	}
}

// -----------------------------------------------------------------------------
// WithdrawDispute
// -----------------------------------------------------------------------------

func TestWithdrawDisputeRevertsTheBountyToRejected(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "claimer", 0)

	b, d := disputedBounty(t, database, creatorID, claimerID, 25)

	if err := database.WithdrawDispute(d.ID, claimerID); err != nil {
		t.Fatalf("WithdrawDispute: %v", err)
	}

	if got := bountyStatus(t, database, b.ID); got != "rejected" {
		t.Errorf("bounty status = %q, want %q", got, "rejected")
	}
	withdrawn, err := database.GetDispute(d.ID)
	if err != nil {
		t.Fatalf("GetDispute: %v", err)
	}
	if withdrawn.Status != "withdrawn" {
		t.Errorf("dispute status = %q, want %q", withdrawn.Status, "withdrawn")
	}
	if got := agentGold(t, database, claimerID); got != 0 {
		t.Errorf("claimer gold = %d, want 0 — withdrawing must not pay", got)
	}
}

func TestWithdrawDisputeRefusesAnAgentWhoIsNotTheClaimer(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "claimer", 0)
	strangerID := makeAgent(t, database, "stranger", 0)

	b, d := disputedBounty(t, database, creatorID, claimerID, 25)

	if err := database.WithdrawDispute(d.ID, strangerID); err == nil {
		t.Fatal("expected WithdrawDispute to refuse a stranger")
	}

	// The effect that matters: the dispute is still live and the bounty still disputed.
	still, err := database.GetDispute(d.ID)
	if err != nil {
		t.Fatalf("GetDispute: %v", err)
	}
	if still.Status != "open" {
		t.Errorf("dispute status = %q, want %q", still.Status, "open")
	}
	if got := bountyStatus(t, database, b.ID); got != "disputed" {
		t.Errorf("bounty status = %q, want %q", got, "disputed")
	}
}

// The withdrawal is gated on the CLAIMER, not on either party to the bounty. The
// creator is the one plausible wrong answer — they are on the dispute row, so a check
// written against the wrong column still refuses every stranger and admits only them.
func TestWithdrawDisputeRefusesTheBountyCreator(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "claimer", 0)

	b, d := disputedBounty(t, database, creatorID, claimerID, 25)

	if err := database.WithdrawDispute(d.ID, creatorID); err == nil {
		t.Fatal("expected WithdrawDispute to refuse the bounty creator — only the claimer may withdraw")
	}

	still, err := database.GetDispute(d.ID)
	if err != nil {
		t.Fatalf("GetDispute: %v", err)
	}
	if still.Status != "open" {
		t.Errorf("dispute status = %q, want %q", still.Status, "open")
	}
	if got := bountyStatus(t, database, b.ID); got != "disputed" {
		t.Errorf("bounty status = %q, want %q", got, "disputed")
	}
}

func TestWithdrawDisputeRefusesAnAlreadyResolvedDispute(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "claimer", 0)
	adminID := makeAgent(t, database, "admin", 0)

	b, d := disputedBounty(t, database, creatorID, claimerID, 25)
	if err := database.ResolveDispute(d.ID, true, "ruling", adminID, ""); err != nil {
		t.Fatalf("ResolveDispute: %v", err)
	}

	if err := database.WithdrawDispute(d.ID, claimerID); err == nil {
		t.Fatal("expected WithdrawDispute to refuse a resolved dispute")
	}
	// A refused withdrawal must not drag a paid bounty back to rejected.
	if got := bountyStatus(t, database, b.ID); got != "paid" {
		t.Errorf("bounty status = %q, want %q", got, "paid")
	}
}

// ⛔ CHARACTERISATION, not an endorsement. Withdrawing a dispute returns the bounty to
// 'rejected' — precisely the state CreateDispute accepts — and then refuses the claimer
// forever: the withdrawn row still satisfies the duplicate check, and disputes carries
// UNIQUE(bounty_id, claimer_id), so no second dispute on that bounty can ever be
// written. The claimer is told "dispute already exists for this bounty" about a dispute
// they withdrew.
//
// The two halves contradict each other. If withdrawal is meant to be final, reverting
// the bounty to a disputable state is the wrong ending; if it is meant to be a retraction
// the claimer can reconsider, the duplicate check must skip withdrawn rows and the
// UNIQUE constraint must go. Filed for that decision as noteboard card
// `4fa1bc01-a2e3-4b6d-b0b8-a3b8b802f77f`; this test pins today's behaviour so whichever
// repair is chosen has to come here and say so.
func TestWithdrawingADisputeBarsTheClaimerFromEverDisputingAgain(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)
	claimerID := makeAgent(t, database, "claimer", 0)

	b, d := disputedBounty(t, database, creatorID, claimerID, 25)
	if err := database.WithdrawDispute(d.ID, claimerID); err != nil {
		t.Fatalf("WithdrawDispute: %v", err)
	}

	// The bounty is back in the state a dispute is filed against.
	if got := bountyStatus(t, database, b.ID); got != "rejected" {
		t.Fatalf("bounty status after withdrawal = %q, want %q", got, "rejected")
	}

	again := &Dispute{BountyID: b.ID, ClaimerID: claimerID, Reason: "on reflection", Evidence: ""}
	err := database.CreateDispute(again)
	if err == nil {
		t.Fatal("re-filing now succeeds — the withdrawal-is-a-retraction repair has landed; " +
			"delete this characterisation and assert the new behaviour")
	}
	if err.Error() != "dispute already exists for this bounty" {
		t.Fatalf("re-file was refused with %q; the behaviour this pins has changed, so revisit "+
			"card 4fa1bc01-a2e3-4b6d-b0b8-a3b8b802f77f before editing this test", err.Error())
	}
}

// -----------------------------------------------------------------------------
// InferTaskDomain — the router the claim gate depends on
// -----------------------------------------------------------------------------

func TestInferTaskDomainRoutesToTheHighestScoringDomain(t *testing.T) {
	cases := []struct {
		name        string
		title       string
		description string
		want        string
	}{
		{"coding", "implement the api", "fix a bug in the function", "coding"},
		{"testing", "write unit tests", "add coverage and assert the mock", "testing"},
		{"documentation", "update the readme", "write a guide and document it", "documentation"},
		{"devops", "deploy to kubernetes", "docker pipeline for the cloud server", "devops"},
		{"design", "restyle the ui", "css layout and visual theme mockup", "design"},
		{"security", "rotate the auth secret", "encrypt the ssl vulnerability", "security"},
		{"database", "add a migration", "postgres schema index for the table", "database"},
		{"maintenance", "refactor for speed", "optimize memory performance cleanup", "maintenance"},
		{"no keywords falls back", "aaa", "bbb", "general"},
		{"empty falls back", "", "", "general"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := InferTaskDomain(c.title, c.description); got != c.want {
				t.Errorf("InferTaskDomain(%q, %q) = %q, want %q", c.title, c.description, got, c.want)
			}
		})
	}
}

// The router folds case, and the claim gate reads its answer — a title in caps must not
// route an agent's reputation to a different domain than the same title in lower case.
func TestInferTaskDomainIgnoresCase(t *testing.T) {
	if got := InferTaskDomain("IMPLEMENT THE API", "FIX A BUG"); got != "coding" {
		t.Errorf("InferTaskDomain on upper case = %q, want %q", got, "coding")
	}
}

// The title and the description are both read. A keyword in either one must count, or a
// bounty whose domain is only named in its description routes to "general".
func TestInferTaskDomainReadsBothTitleAndDescription(t *testing.T) {
	if got := InferTaskDomain("a chore", "encrypt the auth secret and fix the ssl vulnerability"); got != "security" {
		t.Errorf("description-only keywords = %q, want %q", got, "security")
	}
	if got := InferTaskDomain("encrypt the auth secret, ssl vulnerability", "a chore"); got != "security" {
		t.Errorf("title-only keywords = %q, want %q", got, "security")
	}
}

// -----------------------------------------------------------------------------
// StringSlice — required_skills crosses the driver boundary through this pair
// -----------------------------------------------------------------------------

func TestStringSliceRoundTripsThroughTheDatabase(t *testing.T) {
	database := setupSettlementDB(t)
	creatorID := makeAgent(t, database, "creator", 0)

	b := &Bounty{
		Title:          "task",
		Description:    "a task",
		PayoutAmount:   10,
		CreatorID:      creatorID,
		RequiredSkills: []string{"go", "sql", "a skill with spaces"},
	}
	if err := database.CreateBounty(b); err != nil {
		t.Fatalf("CreateBounty: %v", err)
	}

	stored, err := database.GetBountyByID(b.ID)
	if err != nil {
		t.Fatalf("GetBountyByID: %v", err)
	}
	if len(stored.RequiredSkills) != 3 {
		t.Fatalf("required_skills = %v, want 3 entries", stored.RequiredSkills)
	}
	for i, want := range []string{"go", "sql", "a skill with spaces"} {
		if stored.RequiredSkills[i] != want {
			t.Errorf("required_skills[%d] = %q, want %q", i, stored.RequiredSkills[i], want)
		}
	}
}

func TestStringSliceScanRejectsAValueItCannotRead(t *testing.T) {
	var s StringSlice
	if err := s.Scan(42); err == nil {
		t.Fatal("expected Scan to refuse an int, got nil")
	}
}

func TestStringSliceScanTreatsNilAsNoSkills(t *testing.T) {
	s := StringSlice{"stale"}
	if err := s.Scan(nil); err != nil {
		t.Fatalf("Scan(nil): %v", err)
	}
	if s != nil {
		t.Errorf("Scan(nil) left %v, want nil", s)
	}
}

func TestGetBountyByIDReportsAMissingBountyAsNoBountyNotAnError(t *testing.T) {
	database := setupSettlementDB(t)

	b, err := database.GetBountyByID(9999)
	if err != nil {
		t.Fatalf("GetBountyByID on a missing row returned an error: %v", err)
	}
	if b != nil {
		t.Errorf("GetBountyByID = %+v, want nil", b)
	}
}

func TestGetDisputeReportsAMissingDisputeAsNoDisputeNotAnError(t *testing.T) {
	database := setupSettlementDB(t)

	d, err := database.GetDispute(9999)
	if err != nil {
		t.Fatalf("GetDispute on a missing row returned an error: %v", err)
	}
	if d != nil {
		t.Errorf("GetDispute = %+v, want nil", d)
	}
}

func contains(haystack, needle string) bool {
	return len(needle) == 0 || (len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0)
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
