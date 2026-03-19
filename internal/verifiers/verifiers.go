package verifiers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// VerificationContext provides context for running verifiers
type VerificationContext struct {
	BountyID       int                    `json:"bounty_id"`
	WorkSubmission string                 `json:"work_submission"`
	BountyTitle    string                 `json:"bounty_title"`
	Requirements   string                 `json:"requirements"`
	ClaimerName    string                 `json:"claimer_name"`
	Config         map[string]interface{} `json:"config"`
}

// VerificationResult represents the result of running a verifier
type VerificationResult struct {
	Status       string                 `json:"status"`        // 'passed', 'failed', 'error'
	Message      string                 `json:"message"`       // human-readable result message
	Details      map[string]interface{} `json:"details"`       // detailed results
	ErrorMessage *string                `json:"error_message"` // error details if status = 'error'
}

// Verifier interface for all verifier types
type Verifier interface {
	GetType() string
	Verify(ctx VerificationContext) VerificationResult
	GetConfigSchema() map[string]interface{}
}

// FileExistsVerifier checks if specified files exist
type FileExistsVerifier struct{}

func (v *FileExistsVerifier) GetType() string {
	return "file_exists"
}

func (v *FileExistsVerifier) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"files":     []string{"Required: List of files/patterns to check"},
		"base_path": "Optional: Base directory to check from (default: current directory)",
	}
}

func (v *FileExistsVerifier) Verify(ctx VerificationContext) VerificationResult {
	files, ok := ctx.Config["files"].([]interface{})
	if !ok {
		return VerificationResult{
			Status:       "error",
			Message:      "Configuration error",
			ErrorMessage: stringPtr("'files' must be an array of strings"),
		}
	}

	basePath, _ := ctx.Config["base_path"].(string)
	if basePath == "" {
		basePath = "."
	}

	var foundFiles []string
	var missingFiles []string

	for _, fileInterface := range files {
		fileName, ok := fileInterface.(string)
		if !ok {
			continue
		}

		fullPath := filepath.Join(basePath, fileName)
		
		// Support glob patterns
		if strings.Contains(fileName, "*") {
			matches, err := filepath.Glob(fullPath)
			if err != nil {
				continue
			}
			if len(matches) > 0 {
				foundFiles = append(foundFiles, fileName)
			} else {
				missingFiles = append(missingFiles, fileName)
			}
		} else {
			if _, err := os.Stat(fullPath); err == nil {
				foundFiles = append(foundFiles, fileName)
			} else {
				missingFiles = append(missingFiles, fileName)
			}
		}
	}

	status := "passed"
	message := fmt.Sprintf("All %d files found", len(foundFiles))
	
	if len(missingFiles) > 0 {
		status = "failed"
		message = fmt.Sprintf("%d files missing: %s", len(missingFiles), strings.Join(missingFiles, ", "))
	}

	return VerificationResult{
		Status:  status,
		Message: message,
		Details: map[string]interface{}{
			"found_files":   foundFiles,
			"missing_files": missingFiles,
			"base_path":     basePath,
		},
	}
}

// TestSuiteVerifier runs a test command and checks the exit code
type TestSuiteVerifier struct{}

func (v *TestSuiteVerifier) GetType() string {
	return "test_suite"
}

func (v *TestSuiteVerifier) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"command":      "Required: Command to run (e.g., 'npm test', 'go test ./...')",
		"working_dir":  "Optional: Directory to run command in",
		"timeout_sec":  "Optional: Timeout in seconds (default: 300)",
		"success_codes": "Optional: Array of exit codes considered success (default: [0])",
	}
}

func (v *TestSuiteVerifier) Verify(ctx VerificationContext) VerificationResult {
	command, ok := ctx.Config["command"].(string)
	if !ok || command == "" {
		return VerificationResult{
			Status:       "error",
			Message:      "Configuration error",
			ErrorMessage: stringPtr("'command' is required"),
		}
	}

	workingDir, _ := ctx.Config["working_dir"].(string)
	if workingDir == "" {
		workingDir = "."
	}

	timeoutSec := 300.0 // 5 minutes default
	if t, ok := ctx.Config["timeout_sec"].(float64); ok {
		timeoutSec = t
	}

	var successCodes []int
	if codes, ok := ctx.Config["success_codes"].([]interface{}); ok {
		for _, code := range codes {
			if c, ok := code.(float64); ok {
				successCodes = append(successCodes, int(c))
			}
		}
	}
	if len(successCodes) == 0 {
		successCodes = []int{0}
	}

	// Parse command (basic shell parsing)
	cmdParts := strings.Fields(command)
	if len(cmdParts) == 0 {
		return VerificationResult{
			Status:       "error",
			Message:      "Invalid command",
			ErrorMessage: stringPtr("Command is empty after parsing"),
		}
	}

	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Dir = workingDir

	// Set timeout
	timeout := time.Duration(timeoutSec) * time.Second
	done := make(chan error, 1)
	
	var output []byte
	var err error
	
	go func() {
		output, err = cmd.CombinedOutput()
		done <- err
	}()

	select {
	case err = <-done:
		// Command completed
	case <-time.After(timeout):
		cmd.Process.Kill()
		return VerificationResult{
			Status:       "failed",
			Message:      fmt.Sprintf("Test suite timed out after %v", timeout),
			ErrorMessage: stringPtr("Command execution timed out"),
		}
	}

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return VerificationResult{
				Status:       "error",
				Message:      "Failed to run test command",
				ErrorMessage: stringPtr(err.Error()),
			}
		}
	}

	// Check if exit code is in success list
	success := false
	for _, code := range successCodes {
		if exitCode == code {
			success = true
			break
		}
	}

	status := "failed"
	message := fmt.Sprintf("Tests failed with exit code %d", exitCode)
	if success {
		status = "passed"
		message = "Tests passed successfully"
	}

	return VerificationResult{
		Status:  status,
		Message: message,
		Details: map[string]interface{}{
			"exit_code":     exitCode,
			"command":       command,
			"working_dir":   workingDir,
			"output":        string(output),
			"success_codes": successCodes,
		},
	}
}

// ManualVerifier requires human verification
type ManualVerifier struct{}

func (v *ManualVerifier) GetType() string {
	return "manual"
}

func (v *ManualVerifier) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"instructions": "Required: Instructions for the human reviewer",
		"assignee":     "Optional: Specific person to assign the review to",
	}
}

func (v *ManualVerifier) Verify(ctx VerificationContext) VerificationResult {
	instructions, _ := ctx.Config["instructions"].(string)
	assignee, _ := ctx.Config["assignee"].(string)

	message := "Manual verification required"
	if assignee != "" {
		message = fmt.Sprintf("Manual verification required from %s", assignee)
	}

	return VerificationResult{
		Status:  "pending",
		Message: message,
		Details: map[string]interface{}{
			"instructions":      instructions,
			"assignee":          assignee,
			"requires_human":    true,
			"work_submission":   ctx.WorkSubmission,
		},
	}
}

// BenchmarkVerifier checks if a numeric metric meets a threshold
type BenchmarkVerifier struct{}

func (v *BenchmarkVerifier) GetType() string {
	return "benchmark"
}

func (v *BenchmarkVerifier) GetConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"command":    "Required: Command to run that outputs the metric",
		"metric":     "Required: Name of the metric being checked",
		"threshold":  "Required: Threshold value",
		"comparison": "Required: 'less_than', 'greater_than', 'equals'",
		"extract_regex": "Optional: Regex to extract numeric value from command output",
	}
}

func (v *BenchmarkVerifier) Verify(ctx VerificationContext) VerificationResult {
	// This is a simplified implementation
	// In practice, you'd extract metrics from command output and compare to thresholds
	return VerificationResult{
		Status:  "error",
		Message: "Benchmark verifier not fully implemented",
		ErrorMessage: stringPtr("This verifier type is a placeholder for future development"),
	}
}

// VerifierRegistry manages available verifiers
type VerifierRegistry struct {
	verifiers map[string]Verifier
}

func NewVerifierRegistry() *VerifierRegistry {
	registry := &VerifierRegistry{
		verifiers: make(map[string]Verifier),
	}
	
	// Register built-in verifiers
	registry.Register(&FileExistsVerifier{})
	registry.Register(&TestSuiteVerifier{})
	registry.Register(&ManualVerifier{})
	registry.Register(&BenchmarkVerifier{})
	
	return registry
}

func (r *VerifierRegistry) Register(verifier Verifier) {
	r.verifiers[verifier.GetType()] = verifier
}

func (r *VerifierRegistry) GetVerifier(verifierType string) (Verifier, error) {
	verifier, exists := r.verifiers[verifierType]
	if !exists {
		return nil, fmt.Errorf("verifier type '%s' not found", verifierType)
	}
	return verifier, nil
}

func (r *VerifierRegistry) ListTypes() []string {
	types := make([]string, 0, len(r.verifiers))
	for t := range r.verifiers {
		types = append(types, t)
	}
	return types
}

func (r *VerifierRegistry) GetConfigSchemas() map[string]map[string]interface{} {
	schemas := make(map[string]map[string]interface{})
	for t, v := range r.verifiers {
		schemas[t] = v.GetConfigSchema()
	}
	return schemas
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}