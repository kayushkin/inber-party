package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// This demo shows how orchestrator agents can programmatically create bounties

const baseURL = "http://localhost:8080"

type AutoBountyRequest struct {
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Requirements   string                 `json:"requirements,omitempty"`
	PayoutAmount   float64                `json:"payout_amount"`
	Currency       string                 `json:"currency,omitempty"`
	Deadline       *time.Time             `json:"deadline,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	RequiredSkills []string               `json:"required_skills,omitempty"`
	SourceAgent    string                 `json:"source_agent"`
	Verification   *VerificationConfig    `json:"verification,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type VerificationConfig struct {
	Type      string                 `json:"type"`
	Automated bool                   `json:"automated"`
	Config    map[string]interface{} `json:"config"`
}

func main() {
	fmt.Println("🎯 Auto-Bounty Demo - Inber Party")
	fmt.Println("==================================")

	// Demo 1: Structured auto-bounty creation
	fmt.Println("\n1. Creating structured auto-bounty...")
	structuredBounty := AutoBountyRequest{
		Title:        "Implement user authentication API",
		Description:  "Build a secure JWT-based authentication system with login, logout, and token refresh endpoints",
		Requirements: "Must include unit tests with >90% coverage, API documentation, and rate limiting",
		PayoutAmount: 25.0,
		Currency:     "USD",
		SourceAgent:  "orchestrator-1",
		Tags:         []string{"backend", "security", "api"},
		RequiredSkills: []string{"go", "jwt", "testing"},
		Verification: &VerificationConfig{
			Type:      "test_suite",
			Automated: true,
			Config: map[string]interface{}{
				"test_command": "go test ./auth/...",
				"min_coverage": 90,
			},
		},
		Metadata: map[string]interface{}{
			"priority": "high",
			"repository": "github.com/company/api",
		},
	}

	if bounty, err := createAutoBounty(structuredBounty); err != nil {
		log.Printf("❌ Failed to create structured bounty: %v", err)
	} else {
		fmt.Printf("✅ Created bounty: %s (ID: %s, Tier: %s)\n", bounty.Title, bounty.ID, bounty.Tier)
	}

	// Demo 2: Natural language auto-bounty creation
	fmt.Println("\n2. Creating bounty from natural language...")
	textRequests := []struct {
		agentID string
		text    string
	}{
		{
			agentID: "code-review-bot",
			text:    "Fix the memory leak in the user session handler. This is causing performance issues. Pay $15 for this urgent bug fix.",
		},
		{
			agentID: "feature-planner",
			text:    "Add dark mode support to the frontend React components. Should include theme switching and localStorage persistence. Worth $30.",
		},
		{
			agentID: "docs-bot",
			text:    "Write comprehensive API documentation for the new GraphQL endpoints. Include examples and schema definitions. $20 bounty.",
		},
	}

	for _, req := range textRequests {
		if bounty, err := createAutoBountyFromText(req.agentID, req.text); err != nil {
			log.Printf("❌ Failed to create text bounty: %v", err)
		} else {
			fmt.Printf("✅ Created from text: %s (Payout: $%.2f, Skills: %v)\n", 
				bounty.Title, bounty.PayoutAmount, bounty.RequiredSkills)
		}
	}

	// Demo 3: List auto-generated bounties
	fmt.Println("\n3. Listing auto-generated bounties...")
	if bounties, err := listAutoBounties(); err != nil {
		log.Printf("❌ Failed to list bounties: %v", err)
	} else {
		fmt.Printf("📋 Found %d auto-generated bounties:\n", len(bounties))
		for i, bounty := range bounties {
			fmt.Printf("   %d. %s - $%.2f (%s)\n", i+1, bounty.Title, bounty.PayoutAmount, bounty.Status)
		}
	}

	fmt.Println("\n🎉 Auto-bounty demo completed!")
	fmt.Println("💡 Orchestrator agents can now programmatically create bounties!")
}

func createAutoBounty(req AutoBountyRequest) (*Bounty, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(baseURL+"/api/auto-bounties", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var bounty Bounty
	if err := json.NewDecoder(resp.Body).Decode(&bounty); err != nil {
		return nil, err
	}

	return &bounty, nil
}

func createAutoBountyFromText(agentID, text string) (*Bounty, error) {
	payload := map[string]string{
		"agent_id": agentID,
		"text":     text,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(baseURL+"/api/auto-bounties/from-text", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var bounty Bounty
	if err := json.NewDecoder(resp.Body).Decode(&bounty); err != nil {
		return nil, err
	}

	return &bounty, nil
}

func listAutoBounties() ([]*Bounty, error) {
	resp, err := http.Get(baseURL + "/api/bounties?auto_generated=true")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var bounties []*Bounty
	if err := json.NewDecoder(resp.Body).Decode(&bounties); err != nil {
		return nil, err
	}

	return bounties, nil
}

// Copy of relevant types for the demo
type Bounty struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	PayoutAmount   float64   `json:"payout_amount"`
	Currency       string    `json:"currency"`
	Status         string    `json:"status"`
	Tier           string    `json:"tier"`
	RequiredSkills []string  `json:"required_skills"`
	Tags           []string  `json:"tags"`
	AutoGenerated  bool      `json:"auto_generated"`
	SourceAgent    *string   `json:"source_agent"`
	CreatedAt      time.Time `json:"created_at"`
}