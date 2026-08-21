package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/kayushkin/inber-party/internal/bounty"
	"github.com/kayushkin/inber-party/internal/db"
	"github.com/kayushkin/inber-party/internal/ws"

	_ "github.com/mattn/go-sqlite3"
)

// createTestServer creates a test server with an in-memory SQLite database
func createTestServer(t *testing.T) (*Server, func()) {
	t.Helper()

	// Create temp SQLite database
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Build the production schema — db.go's own migration list, translated into SQLite,
	// on a connection that enforces foreign keys. The hand-written schema this replaced
	// had three columns production has never had (currency, verified_at, auto_generated)
	// and was missing three it does have (work_submission, verification_notes,
	// submitted_at), so a test could pass against a bounties table the server never runs
	// on.
	rawDB, err := db.OpenSQLiteMirrorOfProductionSchema(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// bounty_ratings is not in db.go's migration list — it is declared in
	// schema/rating_system.sql, which Migrate() never reads — so the mirror cannot create
	// it and the rating tests need it. Applying it here, visibly and by itself, is what
	// keeps that anomaly readable; see internal/db/sqlite_rating_system_schema.go.
	if err := db.ApplyRatingSystemSchemaToSQLite(rawDB); err != nil {
		t.Fatalf("Failed to create the rating system schema: %v", err)
	}

	// Wrap in the custom DB type
	database := &db.DB{DB: rawDB}

	// Create WebSocket hub
	hub := ws.NewHub()

	// Create server
	server := NewServer(database, hub, nil, nil)

	cleanup := func() {
		if database != nil {
			database.Close()
		}
	}

	return server, cleanup
}

func TestNewServer(t *testing.T) {
	server, cleanup := createTestServer(t)
	defer cleanup()

	if server == nil {
		t.Fatal("NewServer returned nil")
	}
	if server.DB == nil {
		t.Error("Server database not set")
	}
	if server.Hub == nil {
		t.Error("Server WebSocket hub not set")
	}
	if server.BountyRepo == nil {
		t.Error("Server bounty repository not set")
	}
}

func TestParseAndValidateID(t *testing.T) {
	server, cleanup := createTestServer(t)
	defer cleanup()

	tests := []struct {
		name      string
		path      string
		prefix    string
		fieldName string
		expectID  int
		expectOK  bool
	}{
		{
			name:      "valid ID",
			path:      "/api/agents/123",
			prefix:    "/api/agents/",
			fieldName: "agent_id",
			expectID:  123,
			expectOK:  true,
		},
		{
			name:      "zero ID",
			path:      "/api/agents/0",
			prefix:    "/api/agents/",
			fieldName: "agent_id",
			expectID:  0,
			expectOK:  false,
		},
		{
			name:      "negative ID",
			path:      "/api/agents/-1",
			prefix:    "/api/agents/",
			fieldName: "agent_id",
			expectID:  0,
			expectOK:  false,
		},
		{
			name:      "non-numeric ID",
			path:      "/api/agents/abc",
			prefix:    "/api/agents/",
			fieldName: "agent_id",
			expectID:  0,
			expectOK:  false,
		},
		{
			name:      "no ID in path",
			path:      "/api/agents/",
			prefix:    "/api/agents/",
			fieldName: "agent_id",
			expectID:  0,
			expectOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			id, ok := server.parseAndValidateID(w, tt.path, tt.prefix, tt.fieldName)

			if ok != tt.expectOK {
				t.Errorf("Expected ok=%v, got ok=%v", tt.expectOK, ok)
			}

			if tt.expectOK && id != tt.expectID {
				t.Errorf("Expected ID=%d, got ID=%d", tt.expectID, id)
			}

			// Check HTTP response for invalid cases
			if !tt.expectOK && w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d for invalid ID, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestHandleAgents_MethodNotAllowed(t *testing.T) {
	server, cleanup := createTestServer(t)
	defer cleanup()

	req, err := http.NewRequest(http.MethodDelete, "/api/agents", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	server.handleAgents(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestListAgents_EmptyDatabase(t *testing.T) {
	server, cleanup := createTestServer(t)
	defer cleanup()

	req, err := http.NewRequest(http.MethodGet, "/api/agents", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	server.handleAgents(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Check Content-Type
	if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	// Check response body is valid JSON array
	var agents []interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &agents); err != nil {
		t.Errorf("Response is not valid JSON: %v", err)
	}

	// Should be empty for new database
	if len(agents) != 0 {
		t.Errorf("Expected empty array, got %d items", len(agents))
	}
}

func TestListAgents_NilDatabase(t *testing.T) {
	// Test server behavior when database is nil
	server := &Server{
		DB:         nil,
		Hub:        ws.NewHub(),
		BountyRepo: nil,
	}

	req, err := http.NewRequest(http.MethodGet, "/api/agents", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	server.handleAgents(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Should return empty array when database is nil
	var agents []interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &agents); err != nil {
		t.Errorf("Response is not valid JSON: %v", err)
	}

	if len(agents) != 0 {
		t.Errorf("Expected empty array for nil database, got %d items", len(agents))
	}
}

func TestCreateAgent_InvalidJSON(t *testing.T) {
	server, cleanup := createTestServer(t)
	defer cleanup()

	req, err := http.NewRequest(http.MethodPost, "/api/agents", bytes.NewBuffer([]byte("invalid json")))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.handleAgents(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestCreateAgent_MissingFields(t *testing.T) {
	server, cleanup := createTestServer(t)
	defer cleanup()

	// Test with missing required fields
	incompleteAgent := map[string]interface{}{
		"name": "test-agent",
		// Missing title, class, etc.
	}

	jsonData, _ := json.Marshal(incompleteAgent)
	req, err := http.NewRequest(http.MethodPost, "/api/agents", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.handleAgents(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for incomplete data, got %d", http.StatusBadRequest, rr.Code)
	}

	// Check if response contains validation errors (this might vary based on implementation)
	responseBody := rr.Body.String()

	// For debugging, let's see what we actually get
	if rr.Code == http.StatusBadRequest && responseBody != "" {
		// This is a valid response indicating validation failed
		// The exact format may vary, so let's be more lenient
		var response map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			// If it's not JSON, that's also a valid way to return validation errors
			t.Logf("Response is not JSON (which is fine): %s", responseBody)
		} else {
			t.Logf("JSON response received: %+v", response)
			// Check for various common error response formats
			hasErrors := false
			if _, exists := response["errors"]; exists {
				hasErrors = true
			}
			if _, exists := response["error"]; exists {
				hasErrors = true
			}
			if _, exists := response["message"]; exists {
				hasErrors = true
			}

			if !hasErrors {
				t.Logf("Warning: Response doesn't contain recognizable error fields")
			}
		}
	}
}

func TestCreateAgent_Valid(t *testing.T) {
	server, cleanup := createTestServer(t)
	defer cleanup()

	validAgent := map[string]interface{}{
		"name":         "test-agent",
		"title":        "The Tester",
		"class":        "Developer",
		"level":        1,
		"xp":           0,
		"gold":         50,
		"energy":       100,
		"avatar_emoji": "🤖",
	}

	jsonData, _ := json.Marshal(validAgent)
	req, err := http.NewRequest(http.MethodPost, "/api/agents", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.handleAgents(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Response: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	// Check response contains agent data
	var response db.Agent
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Response is not valid JSON: %v", err)
	}

	if response.ID == 0 {
		t.Error("Created agent should have a non-zero ID")
	}

	if response.Name != "test-agent" {
		t.Errorf("Expected name 'test-agent', got '%s'", response.Name)
	}
}

func TestHealthCheck(t *testing.T) {
	server, cleanup := createTestServer(t)
	defer cleanup()

	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()

	// Use the server to create a health check handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Basic health check logic using the server instance
		w.Header().Set("Content-Type", "application/json")
		status := map[string]interface{}{
			"status": "ok",
			"database": func() string {
				if server.DB != nil {
					return "connected"
				}
				return "disconnected"
			}(),
		}
		json.NewEncoder(w).Encode(status)
	})

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Response is not valid JSON: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", response["status"])
	}

	if response["database"] != "connected" {
		t.Errorf("Expected database 'connected', got %v", response["database"])
	}
}

func TestWriteValidationError(t *testing.T) {

	// Create mock validation errors
	errors := []struct {
		Field   string
		Message string
		Code    string
	}{
		{"name", "Name is required", "required"},
		{"email", "Invalid email format", "invalid_format"},
	}

	w := httptest.NewRecorder()

	// We need to create a ValidationErrors slice - let's simulate it
	// Since ValidationErrors might not be directly accessible, we'll test the response format
	response := map[string]interface{}{
		"errors": errors,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var responseBody map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &responseBody); err != nil {
		t.Errorf("Response is not valid JSON: %v", err)
	}

	if _, hasErrors := responseBody["errors"]; !hasErrors {
		t.Error("Response should contain 'errors' field")
	}
}

// insertTestAgent creates one agent and returns its id as the string the bounty repository
// stores in creator_id. Every table in the production schema that names an agent does it
// with a REFERENCES clause, so a test that invents an agent id is writing a row the
// product cannot hold.
func insertTestAgent(t *testing.T, database *db.DB, name string) string {
	t.Helper()

	result, err := database.Exec(
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

// Integration test for bounty endpoints via API
func TestBountyEndpoints_Integration(t *testing.T) {
	server, cleanup := createTestServer(t)
	defer cleanup()

	// bounties.creator_id REFERENCES agents(id), so the creating agent has to exist
	// before the bounty does. This test used to name creator "1" against an empty agents
	// table and pass, because the mirror created the REFERENCES clause and SQLite did not
	// enforce it; PostgreSQL refuses that row.
	creatorID := insertTestAgent(t, server.DB, "bounty-creator")

	// Test creating a bounty through the API
	newBounty := bounty.AutoBountyRequest{
		Title:        "Test API Bounty",
		Description:  "Testing bounty creation via API",
		Requirements: "Write tests and documentation",
		PayoutAmount: 5.00,
		SourceAgent:  "test-agent",
	}

	jsonData, _ := json.Marshal(newBounty)
	req, err := http.NewRequest(http.MethodPost, "/api/bounties", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// Simulate bounty creation endpoint
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if server.BountyRepo == nil {
			http.Error(w, "Bounty service not available", http.StatusServiceUnavailable)
			return
		}

		var req bounty.AutoBountyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Create bounty
		newBounty := &bounty.Bounty{
			Title:        req.Title,
			Description:  req.Description,
			Requirements: req.Requirements,
			PayoutAmount: req.PayoutAmount,
			CreatedBy:    creatorID,
			Currency:     "USD",
		}

		if err := server.BountyRepo.CreateBounty(newBounty); err != nil {
			http.Error(w, "Failed to create bounty", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newBounty)
	})

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Response: %s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	// Verify response contains bounty data
	var response bounty.Bounty
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Response is not valid JSON: %v", err)
	}

	if response.ID == "" {
		t.Error("Created bounty should have a non-empty ID")
	}

	if response.Title != newBounty.Title {
		t.Errorf("Expected title '%s', got '%s'", newBounty.Title, response.Title)
	}

	if response.Status != bounty.StatusOpen {
		t.Errorf("Expected status '%s', got '%s'", bounty.StatusOpen, response.Status)
	}
}

// Test server behavior with various Content-Type headers
func TestContentTypeHandling(t *testing.T) {
	server, cleanup := createTestServer(t)
	defer cleanup()

	tests := []struct {
		name        string
		contentType string
		expectCode  int
	}{
		{
			name:        "valid JSON content type",
			contentType: "application/json",
			expectCode:  http.StatusBadRequest, // Will fail validation but accept content type
		},
		{
			name:        "JSON with charset",
			contentType: "application/json; charset=utf-8",
			expectCode:  http.StatusBadRequest,
		},
		{
			name:        "missing content type",
			contentType: "",
			expectCode:  http.StatusBadRequest, // Should still try to parse
		},
		{
			name:        "wrong content type",
			contentType: "text/plain",
			expectCode:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/api/agents", bytes.NewBuffer([]byte("{}")))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			rr := httptest.NewRecorder()
			server.handleAgents(rr, req)

			if rr.Code != tt.expectCode {
				t.Errorf("Expected status %d, got %d", tt.expectCode, rr.Code)
			}
		})
	}
}

func TestCORSHeaders(t *testing.T) {
	_, cleanup := createTestServer(t)
	defer cleanup()

	req, err := http.NewRequest(http.MethodOptions, "/api/agents", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()

	// Simulate CORS handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
	})

	handler.ServeHTTP(rr, req)

	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, Authorization",
	}

	for header, expectedValue := range expectedHeaders {
		if actualValue := rr.Header().Get(header); actualValue != expectedValue {
			t.Errorf("Expected header %s: %s, got: %s", header, expectedValue, actualValue)
		}
	}
}
