package validation

import (
	"strings"
	"testing"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	if v == nil {
		t.Fatal("NewValidator returned nil")
	}
	if v.HasErrors() {
		t.Error("New validator should not have errors")
	}
	if len(v.GetErrors()) != 0 {
		t.Error("New validator should have empty error slice")
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError{
		Field:   "email",
		Message: "Invalid format",
		Code:    "invalid_email",
	}
	
	expected := "validation error on field 'email': Invalid format"
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestValidationErrors(t *testing.T) {
	var errors ValidationErrors
	
	// Test empty errors
	if errors.Error() != "" {
		t.Error("Empty ValidationErrors should return empty string")
	}
	
	// Test single error
	errors = append(errors, ValidationError{Field: "name", Message: "Required", Code: "required"})
	expected := "validation error on field 'name': Required"
	if errors.Error() != expected {
		t.Errorf("Single error: expected %q, got %q", expected, errors.Error())
	}
	
	// Test multiple errors
	errors = append(errors, ValidationError{Field: "email", Message: "Invalid", Code: "invalid"})
	result := errors.Error()
	if !strings.Contains(result, "validation errors:") {
		t.Error("Multiple errors should start with 'validation errors:'")
	}
	if !strings.Contains(result, "name") || !strings.Contains(result, "email") {
		t.Error("Multiple errors should contain all field names")
	}
}

func TestRequired(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		shouldErr bool
	}{
		{"empty string", "", true},
		{"whitespace", "   ", true},
		{"valid value", "test", false},
		{"value with spaces", " test ", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.Required("test_field", tt.value)
			
			hasError := v.HasErrors()
			if hasError != tt.shouldErr {
				t.Errorf("Expected error: %v, got error: %v", tt.shouldErr, hasError)
			}
			
			if tt.shouldErr {
				errors := v.GetErrors()
				if len(errors) != 1 {
					t.Errorf("Expected 1 error, got %d", len(errors))
				}
				if errors[0].Code != "required" {
					t.Errorf("Expected error code 'required', got %q", errors[0].Code)
				}
			}
		})
	}
}

func TestRequiredInt(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		shouldErr bool
	}{
		{"zero", 0, true},
		{"negative", -1, true},
		{"positive", 1, false},
		{"large positive", 999, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.RequiredInt("test_field", tt.value)
			
			hasError := v.HasErrors()
			if hasError != tt.shouldErr {
				t.Errorf("Expected error: %v, got error: %v", tt.shouldErr, hasError)
			}
		})
	}
}

func TestRequiredFloat(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		shouldErr bool
	}{
		{"zero", 0.0, true},
		{"negative", -1.5, true},
		{"positive", 1.5, false},
		{"small positive", 0.01, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.RequiredFloat("test_field", tt.value)
			
			hasError := v.HasErrors()
			if hasError != tt.shouldErr {
				t.Errorf("Expected error: %v, got error: %v", tt.shouldErr, hasError)
			}
		})
	}
}

func TestMaxLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		max       int
		shouldErr bool
	}{
		{"under limit", "test", 10, false},
		{"at limit", "test", 4, false},
		{"over limit", "test", 3, true},
		{"empty string", "", 0, false},
		{"unicode characters", "测试", 2, false},
		{"unicode over limit", "测试测", 2, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.MaxLength("test_field", tt.value, tt.max)
			
			hasError := v.HasErrors()
			if hasError != tt.shouldErr {
				t.Errorf("Expected error: %v, got error: %v", tt.shouldErr, hasError)
			}
		})
	}
}

func TestMinLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		min       int
		shouldErr bool
	}{
		{"over minimum", "testing", 4, false},
		{"at minimum", "test", 4, false},
		{"under minimum", "hi", 4, true},
		{"empty string", "", 1, true},
		{"whitespace only", "   ", 3, true}, // Should trim whitespace
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.MinLength("test_field", tt.value, tt.min)
			
			hasError := v.HasErrors()
			if hasError != tt.shouldErr {
				t.Errorf("Expected error: %v, got error: %v", tt.shouldErr, hasError)
			}
		})
	}
}

func TestRange(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		min, max  int
		shouldErr bool
	}{
		{"in range", 5, 1, 10, false},
		{"at minimum", 1, 1, 10, false},
		{"at maximum", 10, 1, 10, false},
		{"below minimum", 0, 1, 10, true},
		{"above maximum", 11, 1, 10, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.Range("test_field", tt.value, tt.min, tt.max)
			
			hasError := v.HasErrors()
			if hasError != tt.shouldErr {
				t.Errorf("Expected error: %v, got error: %v", tt.shouldErr, hasError)
			}
		})
	}
}

func TestRangeFloat(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		min, max  float64
		shouldErr bool
	}{
		{"in range", 5.5, 1.0, 10.0, false},
		{"at minimum", 1.0, 1.0, 10.0, false},
		{"at maximum", 10.0, 1.0, 10.0, false},
		{"below minimum", 0.9, 1.0, 10.0, true},
		{"above maximum", 10.1, 1.0, 10.0, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.RangeFloat("test_field", tt.value, tt.min, tt.max)
			
			hasError := v.HasErrors()
			if hasError != tt.shouldErr {
				t.Errorf("Expected error: %v, got error: %v", tt.shouldErr, hasError)
			}
		})
	}
}

func TestOneOf(t *testing.T) {
	allowed := []string{"red", "green", "blue"}
	
	tests := []struct {
		name      string
		value     string
		shouldErr bool
	}{
		{"valid option", "red", false},
		{"another valid option", "blue", false},
		{"invalid option", "yellow", true},
		{"empty string", "", true},
		{"case sensitive", "RED", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.OneOf("color", tt.value, allowed)
			
			hasError := v.HasErrors()
			if hasError != tt.shouldErr {
				t.Errorf("Expected error: %v, got error: %v", tt.shouldErr, hasError)
			}
		})
	}
}

func TestEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		shouldErr bool
	}{
		{"valid email", "user@example.com", false},
		{"valid email with numbers", "user123@example.com", false},
		{"valid email with dots", "user.name@example.com", false},
		{"empty email", "", false}, // Should use Required() separately
		{"invalid format", "notanemail", true},
		{"missing @", "user.example.com", true},
		{"missing domain", "user@", true},
		{"missing TLD", "user@example", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.Email("email", tt.email)
			
			hasError := v.HasErrors()
			if hasError != tt.shouldErr {
				t.Errorf("Expected error: %v, got error: %v", tt.shouldErr, hasError)
			}
		})
	}
}

func TestURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		shouldErr bool
	}{
		{"valid http", "http://example.com", false},
		{"valid https", "https://example.com", false},
		{"empty URL", "", false}, // Should use Required() separately
		{"invalid format", "notaurl", true},
		{"missing protocol", "example.com", true},
		{"wrong protocol", "ftp://example.com", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.URL("url", tt.url)
			
			hasError := v.HasErrors()
			if hasError != tt.shouldErr {
				t.Errorf("Expected error: %v, got error: %v", tt.shouldErr, hasError)
			}
		})
	}
}

func TestAlphanumeric(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		shouldErr bool
	}{
		{"valid alphanumeric", "test123", false},
		{"with underscore", "test_123", false},
		{"with hyphen", "test-123", false},
		{"empty string", "", false}, // Should use Required() separately
		{"with spaces", "test 123", true},
		{"with special chars", "test@123", true},
		{"only letters", "test", false},
		{"only numbers", "123", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.Alphanumeric("field", tt.value)
			
			hasError := v.HasErrors()
			if hasError != tt.shouldErr {
				t.Errorf("Expected error: %v, got error: %v", tt.shouldErr, hasError)
			}
		})
	}
}

func TestJSON(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		shouldErr bool
	}{
		{"valid object", `{"key": "value"}`, false},
		{"valid array", `[1, 2, 3]`, false},
		{"empty string", "", false}, // Should use Required() separately
		{"invalid json", "not json", true},
		{"incomplete object", `{"key":`, true},
		{"incomplete array", `[1, 2`, true},
		{"with whitespace", `  {"key": "value"}  `, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.JSON("json_field", tt.value)
			
			hasError := v.HasErrors()
			if hasError != tt.shouldErr {
				t.Errorf("Expected error: %v, got error: %v", tt.shouldErr, hasError)
			}
		})
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectedID  int
		shouldError bool
	}{
		{"valid ID", "123", 123, false},
		{"empty string", "", 0, true},
		{"zero", "0", 0, true},
		{"negative", "-1", 0, true},
		{"non-numeric", "abc", 0, true},
		{"float", "1.5", 0, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			id := v.ValidateID("id", tt.value)
			
			hasError := v.HasErrors()
			if hasError != tt.shouldError {
				t.Errorf("Expected error: %v, got error: %v", tt.shouldError, hasError)
			}
			
			if id != tt.expectedID {
				t.Errorf("Expected ID %d, got %d", tt.expectedID, id)
			}
		})
	}
}

func TestValidateOptionalID(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectedID  *int
		shouldError bool
	}{
		{"valid ID", "123", intPtr(123), false},
		{"empty string", "", nil, false},
		{"zero", "0", nil, true},
		{"negative", "-1", nil, true},
		{"non-numeric", "abc", nil, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			id := v.ValidateOptionalID("id", tt.value)
			
			hasError := v.HasErrors()
			if hasError != tt.shouldError {
				t.Errorf("Expected error: %v, got error: %v", tt.shouldError, hasError)
			}
			
			if (id == nil && tt.expectedID != nil) || (id != nil && tt.expectedID == nil) {
				t.Errorf("Expected ID %v, got %v", tt.expectedID, id)
			} else if id != nil && tt.expectedID != nil && *id != *tt.expectedID {
				t.Errorf("Expected ID %d, got %d", *tt.expectedID, *id)
			}
		})
	}
}

func TestValidateLimit(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		defaultLimit int
		maxLimit    int
		expected    int
		shouldError bool
	}{
		{"empty string uses default", "", 10, 100, 10, false},
		{"valid limit", "25", 10, 100, 25, false},
		{"exceeds max gets capped", "150", 10, 100, 100, false},
		{"zero value", "0", 10, 100, 10, true},
		{"negative value", "-5", 10, 100, 10, true},
		{"non-numeric", "abc", 10, 100, 10, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			limit := v.ValidateLimit(tt.value, tt.defaultLimit, tt.maxLimit)
			
			hasError := v.HasErrors()
			if hasError != tt.shouldError {
				t.Errorf("Expected error: %v, got error: %v", tt.shouldError, hasError)
			}
			
			if limit != tt.expected {
				t.Errorf("Expected limit %d, got %d", tt.expected, limit)
			}
		})
	}
}

func TestValidateOffset(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expected    int
		shouldError bool
	}{
		{"empty string", "", 0, false},
		{"valid offset", "25", 25, false},
		{"zero", "0", 0, false},
		{"negative", "-5", 0, true},
		{"non-numeric", "abc", 0, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			offset := v.ValidateOffset(tt.value)
			
			hasError := v.HasErrors()
			if hasError != tt.shouldError {
				t.Errorf("Expected error: %v, got error: %v", tt.shouldError, hasError)
			}
			
			if offset != tt.expected {
				t.Errorf("Expected offset %d, got %d", tt.expected, offset)
			}
		})
	}
}

func TestValidateEnum(t *testing.T) {
	allowed := []string{"open", "closed", "pending"}
	
	tests := []struct {
		name        string
		value       string
		required    bool
		shouldError bool
	}{
		{"valid value", "open", true, false},
		{"empty required", "", true, true},
		{"empty optional", "", false, false},
		{"invalid value", "invalid", true, true},
		{"invalid value optional", "invalid", false, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.ValidateEnum("status", tt.value, allowed, tt.required)
			
			hasError := v.HasErrors()
			if hasError != tt.shouldError {
				t.Errorf("Expected error: %v, got error: %v", tt.shouldError, hasError)
			}
		})
	}
}

func TestValidateBountyData(t *testing.T) {
	tests := []struct {
		name         string
		title        string
		description  string
		requirements string
		payoutAmount float64
		errorCount   int
	}{
		{"valid data", "Test Bounty", "A test description", "Some requirements", 10.0, 0},
		{"empty title", "", "Description", "Requirements", 10.0, 1},
		{"empty description", "Title", "", "Requirements", 10.0, 1},
		{"zero payout", "Title", "Description", "Requirements", 0.0, 1},
		{"negative payout", "Title", "Description", "Requirements", -5.0, 1},
		{"too high payout", "Title", "Description", "Requirements", 15000.0, 1},
		{"long title", strings.Repeat("a", 300), "Description", "Requirements", 10.0, 1},
		{"long description", "Title", strings.Repeat("a", 3000), "Requirements", 10.0, 1},
		{"multiple errors", "", "", "", -1.0, 3}, // title, description, payout
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.ValidateBountyData(tt.title, tt.description, tt.requirements, tt.payoutAmount)
			
			errorCount := len(v.GetErrors())
			if errorCount != tt.errorCount {
				t.Errorf("Expected %d errors, got %d", tt.errorCount, errorCount)
			}
		})
	}
}

func TestValidateAgentData(t *testing.T) {
	tests := []struct {
		name        string
		agentName   string
		title       string
		class       string
		emoji       string
		level       int
		xp          int
		gold        int
		energy      int
		errorCount  int
	}{
		{"valid data", "agent1", "The Brave", "Warrior", "⚔️", 5, 500, 100, 80, 0},
		{"empty name", "", "The Brave", "Warrior", "⚔️", 5, 500, 100, 80, 1},
		{"empty class", "agent1", "The Brave", "", "⚔️", 5, 500, 100, 80, 1},
		{"invalid level", "agent1", "The Brave", "Warrior", "⚔️", 0, 500, 100, 80, 1},
		{"level too high", "agent1", "The Brave", "Warrior", "⚔️", 150, 500, 100, 80, 1},
		{"negative xp", "agent1", "The Brave", "Warrior", "⚔️", 5, -100, 100, 80, 1},
		{"energy out of range", "agent1", "The Brave", "Warrior", "⚔️", 5, 500, 100, 150, 1},
		{"invalid name format", "agent with spaces", "The Brave", "Warrior", "⚔️", 5, 500, 100, 80, 1},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.ValidateAgentData(tt.agentName, tt.title, tt.class, tt.emoji, tt.level, tt.xp, tt.gold, tt.energy)
			
			errorCount := len(v.GetErrors())
			if errorCount != tt.errorCount {
				t.Errorf("Expected %d errors, got %d", tt.errorCount, errorCount)
				for _, err := range v.GetErrors() {
					t.Logf("Error: %s", err.Error())
				}
			}
		})
	}
}

// Helper function for tests
func intPtr(i int) *int {
	return &i
}