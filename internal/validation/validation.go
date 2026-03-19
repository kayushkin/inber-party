package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", ve.Field, ve.Message)
}

// ValidationErrors represents multiple validation failures
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	if len(ve) == 1 {
		return ve[0].Error()
	}
	
	var msgs []string
	for _, err := range ve {
		msgs = append(msgs, err.Error())
	}
	return "validation errors: " + strings.Join(msgs, "; ")
}

// Validator provides validation utilities
type Validator struct {
	errors ValidationErrors
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// HasErrors returns true if validation errors exist
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// GetErrors returns all validation errors
func (v *Validator) GetErrors() ValidationErrors {
	return v.errors
}

// AddError adds a validation error
func (v *Validator) AddError(field, message, code string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	})
}

// Required validates that a field is not empty
func (v *Validator) Required(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, fmt.Sprintf("%s is required", field), "required")
	}
}

// RequiredInt validates that an integer field is greater than 0
func (v *Validator) RequiredInt(field string, value int) {
	if value <= 0 {
		v.AddError(field, fmt.Sprintf("%s is required and must be greater than 0", field), "required")
	}
}

// RequiredFloat validates that a float field is greater than 0
func (v *Validator) RequiredFloat(field string, value float64) {
	if value <= 0 {
		v.AddError(field, fmt.Sprintf("%s is required and must be greater than 0", field), "required")
	}
}

// MaxLength validates string length doesn't exceed maximum
func (v *Validator) MaxLength(field, value string, max int) {
	if utf8.RuneCountInString(value) > max {
		v.AddError(field, fmt.Sprintf("%s cannot exceed %d characters", field, max), "max_length")
	}
}

// MinLength validates string length meets minimum requirement
func (v *Validator) MinLength(field, value string, min int) {
	if utf8.RuneCountInString(strings.TrimSpace(value)) < min {
		v.AddError(field, fmt.Sprintf("%s must be at least %d characters", field, min), "min_length")
	}
}

// Range validates that an integer is within specified range (inclusive)
func (v *Validator) Range(field string, value, min, max int) {
	if value < min || value > max {
		v.AddError(field, fmt.Sprintf("%s must be between %d and %d", field, min, max), "range")
	}
}

// RangeFloat validates that a float is within specified range (inclusive)
func (v *Validator) RangeFloat(field string, value, min, max float64) {
	if value < min || value > max {
		v.AddError(field, fmt.Sprintf("%s must be between %f and %f", field, min, max), "range")
	}
}

// OneOf validates that a value is one of the allowed options
func (v *Validator) OneOf(field, value string, allowed []string) {
	for _, option := range allowed {
		if value == option {
			return
		}
	}
	v.AddError(field, fmt.Sprintf("%s must be one of: %s", field, strings.Join(allowed, ", ")), "invalid_option")
}

// Email validates email format
func (v *Validator) Email(field, email string) {
	if email == "" {
		return // Use Required() separately if email is required
	}
	
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		v.AddError(field, fmt.Sprintf("%s must be a valid email address", field), "invalid_email")
	}
}

// URL validates URL format
func (v *Validator) URL(field, url string) {
	if url == "" {
		return // Use Required() separately if URL is required
	}
	
	// Simple URL validation
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		v.AddError(field, fmt.Sprintf("%s must be a valid URL starting with http:// or https://", field), "invalid_url")
	}
}

// Alphanumeric validates that a string contains only alphanumeric characters and underscores
func (v *Validator) Alphanumeric(field, value string) {
	if value == "" {
		return
	}
	
	alphanumericRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !alphanumericRegex.MatchString(value) {
		v.AddError(field, fmt.Sprintf("%s can only contain letters, numbers, underscores, and hyphens", field), "invalid_format")
	}
}

// JSON validates that a string is valid JSON
func (v *Validator) JSON(field, value string) {
	if value == "" {
		return
	}
	
	// Basic JSON validation - check for proper brackets
	value = strings.TrimSpace(value)
	if !((strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}")) ||
		  (strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]"))) {
		v.AddError(field, fmt.Sprintf("%s must be valid JSON", field), "invalid_json")
	}
}

// ValidateID validates that a string can be converted to a positive integer ID
func (v *Validator) ValidateID(field, value string) int {
	if value == "" {
		v.AddError(field, fmt.Sprintf("%s is required", field), "required")
		return 0
	}
	
	id, err := strconv.Atoi(value)
	if err != nil {
		v.AddError(field, fmt.Sprintf("%s must be a valid integer", field), "invalid_integer")
		return 0
	}
	
	if id <= 0 {
		v.AddError(field, fmt.Sprintf("%s must be a positive integer", field), "invalid_id")
		return 0
	}
	
	return id
}

// ValidateOptionalID validates an optional ID parameter
func (v *Validator) ValidateOptionalID(field, value string) *int {
	if value == "" {
		return nil
	}
	
	id, err := strconv.Atoi(value)
	if err != nil {
		v.AddError(field, fmt.Sprintf("%s must be a valid integer", field), "invalid_integer")
		return nil
	}
	
	if id <= 0 {
		v.AddError(field, fmt.Sprintf("%s must be a positive integer", field), "invalid_id")
		return nil
	}
	
	return &id
}

// ValidateLimit validates pagination limit parameter
func (v *Validator) ValidateLimit(value string, defaultLimit, maxLimit int) int {
	if value == "" {
		return defaultLimit
	}
	
	limit, err := strconv.Atoi(value)
	if err != nil {
		v.AddError("limit", "limit must be a valid integer", "invalid_integer")
		return defaultLimit
	}
	
	if limit <= 0 {
		v.AddError("limit", "limit must be greater than 0", "invalid_range")
		return defaultLimit
	}
	
	if limit > maxLimit {
		return maxLimit // Silently cap at max
	}
	
	return limit
}

// ValidateOffset validates pagination offset parameter
func (v *Validator) ValidateOffset(value string) int {
	if value == "" {
		return 0
	}
	
	offset, err := strconv.Atoi(value)
	if err != nil {
		v.AddError("offset", "offset must be a valid integer", "invalid_integer")
		return 0
	}
	
	if offset < 0 {
		v.AddError("offset", "offset cannot be negative", "invalid_range")
		return 0
	}
	
	return offset
}

// ValidateEnum validates that a value is one of the allowed enum values
func (v *Validator) ValidateEnum(field, value string, allowedValues []string, required bool) {
	if !required && value == "" {
		return
	}
	
	if required && value == "" {
		v.AddError(field, fmt.Sprintf("%s is required", field), "required")
		return
	}
	
	v.OneOf(field, value, allowedValues)
}

// Bounty-specific validation constants and functions
var (
	ValidBountyStatuses = []string{"open", "claimed", "submitted", "completed", "rejected", "disputed"}
	ValidBountyTiers    = []string{"bronze", "silver", "gold", "platinum", "legendary"}
	ValidTaskStatuses   = []string{"available", "assigned", "in_progress", "completed", "failed", "cancelled"}
	ValidAgentStatuses  = []string{"idle", "working", "sleeping", "offline", "error"}
	ValidPartyStatuses  = []string{"active", "on_quest", "disbanded", "resting"}
	ValidDifficulties   = []string{"easy", "medium", "hard", "expert", "legendary", "1", "2", "3", "4", "5"}
)

// ValidateBountyData validates bounty creation/update data
func (v *Validator) ValidateBountyData(title, description, requirements string, payoutAmount float64) {
	v.Required("title", title)
	v.MaxLength("title", title, 200)
	
	v.Required("description", description)
	v.MaxLength("description", description, 2000)
	
	v.MaxLength("requirements", requirements, 1000)
	
	if payoutAmount <= 0 {
		v.AddError("payout_amount", "payout amount must be greater than 0", "invalid_amount")
	}
	if payoutAmount > 10000 {
		v.AddError("payout_amount", "payout amount cannot exceed $10,000", "amount_too_high")
	}
}

// ValidateTaskData validates task creation/update data
func (v *Validator) ValidateTaskData(name, description, difficulty string, xpReward, goldReward int) {
	v.Required("name", name)
	v.MaxLength("name", name, 200)
	
	v.Required("description", description)
	v.MaxLength("description", description, 2000)
	
	v.ValidateEnum("difficulty", difficulty, ValidDifficulties, true)
	
	v.Range("xp_reward", xpReward, 1, 10000)
	v.Range("gold_reward", goldReward, 1, 50000)
}

// ValidateAgentData validates agent creation/update data
func (v *Validator) ValidateAgentData(name, title, class, avatarEmoji string, level, xp, gold, energy int) {
	v.Required("name", name)
	v.MaxLength("name", name, 100)
	v.Alphanumeric("name", name)
	
	v.MaxLength("title", title, 200)
	
	v.Required("class", class)
	v.MaxLength("class", class, 50)
	
	v.MaxLength("avatar_emoji", avatarEmoji, 10)
	
	v.Range("level", level, 1, 100)
	v.Range("xp", xp, 0, 1000000)
	v.Range("gold", gold, 0, 1000000)
	v.Range("energy", energy, 0, 100)
}

// ValidatePartyData validates party creation/update data
func (v *Validator) ValidatePartyData(name, description string, leaderID int) {
	v.Required("name", name)
	v.MaxLength("name", name, 100)
	
	v.MaxLength("description", description, 500)
	
	v.RequiredInt("leader_id", leaderID)
}

// ValidateRatingData validates rating submission data
func (v *Validator) ValidateRatingData(rating int, comment string, bountyID, raterID, ratedID int) {
	v.Range("rating", rating, 1, 5)
	v.MaxLength("comment", comment, 1000)
	v.RequiredInt("bounty_id", bountyID)
	v.RequiredInt("rater_id", raterID)
	v.RequiredInt("rated_id", ratedID)
}