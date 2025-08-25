package types

import (
	"testing"
	"time"
)

// TestValidationResult tests ValidationResult functionality
func TestValidationResult(t *testing.T) {
	result := &ValidationResult{Valid: true}
	
	if result.HasErrors() {
		t.Errorf("New validation result should not have errors")
	}
	
	if result.FirstError() != nil {
		t.Errorf("New validation result should not have first error")
	}
	
	// Add an error
	result.AddError("test_field", "test error", "test_value")
	
	if !result.HasErrors() {
		t.Errorf("Validation result should have errors after adding one")
	}
	
	if result.Valid {
		t.Errorf("Validation result should be invalid after adding error")
	}
	
	firstError := result.FirstError()
	if firstError == nil {
		t.Errorf("Should have first error")
	}
	
	if firstError.Field != "test_field" {
		t.Errorf("Expected field 'test_field', got %s", firstError.Field)
	}
}

// TestValidationError tests ValidationError functionality
func TestValidationError(t *testing.T) {
	err := ValidationError{
		Field:   "username",
		Message: "cannot be empty",
		Value:   "",
	}
	
	expected := "validation failed for field 'username': cannot be empty"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

// TestNewDataValidator tests validator creation with defaults
func TestNewDataValidator(t *testing.T) {
	validator := NewDataValidator()
	
	if validator == nil {
		t.Fatalf("Validator should not be nil")
	}
	
	if validator.MaxDescriptionLength != 10000 {
		t.Errorf("Expected MaxDescriptionLength 10000, got %d", validator.MaxDescriptionLength)
	}
	
	if validator.MaxRequirements != 50 {
		t.Errorf("Expected MaxRequirements 50, got %d", validator.MaxRequirements)
	}
}

// TestValidateProblemRequest tests ProblemRequest validation
func TestValidateProblemRequest(t *testing.T) {
	validator := NewDataValidator()
	
	// Test nil request
	result := validator.ValidateProblemRequest(nil)
	if !result.HasErrors() {
		t.Errorf("Should have error for nil request")
	}
	
	// Test valid request
	validRequest := &ProblemRequest{
		Description:  "Create a web API for user management",
		Requirements: []string{"RESTful endpoints", "User authentication"},
		Constraints:  []string{"Use Go", "PostgreSQL database"},
	}
	
	result = validator.ValidateProblemRequest(validRequest)
	if result.HasErrors() {
		t.Errorf("Valid request should not have errors: %v", result.Errors)
	}
	
	// Test empty description
	invalidRequest := &ProblemRequest{
		Description: "",
	}
	
	result = validator.ValidateProblemRequest(invalidRequest)
	if !result.HasErrors() {
		t.Errorf("Should have error for empty description")
	}
	
	// Test description too long
	longDesc := make([]byte, validator.MaxDescriptionLength+1)
	for i := range longDesc {
		longDesc[i] = 'a'
	}
	
	invalidRequest = &ProblemRequest{
		Description: string(longDesc),
	}
	
	result = validator.ValidateProblemRequest(invalidRequest)
	if !result.HasErrors() {
		t.Errorf("Should have error for description too long")
	}
}

// TestValidateAnalysis tests Analysis validation
func TestValidateAnalysis(t *testing.T) {
	validator := NewDataValidator()
	
	// Test nil analysis
	result := validator.ValidateAnalysis(nil)
	if !result.HasErrors() {
		t.Errorf("Should have error for nil analysis")
	}
	
	// Test valid analysis
	validAnalysis := &Analysis{
		ProblemDomain:       "web-development",
		TechnicalStack:      []string{"Go", "PostgreSQL", "Docker"},
		ArchitecturePattern: "REST API",
		Complexity:          ComplexityMedium,
	}
	
	result = validator.ValidateAnalysis(validAnalysis)
	if result.HasErrors() {
		t.Errorf("Valid analysis should not have errors: %v", result.Errors)
	}
	
	// Test empty problem domain
	invalidAnalysis := &Analysis{
		ProblemDomain: "",
		TechnicalStack: []string{"Go"},
		Complexity:    ComplexityLow,
	}
	
	result = validator.ValidateAnalysis(invalidAnalysis)
	if !result.HasErrors() {
		t.Errorf("Should have error for empty problem domain")
	}
	
	// Test invalid complexity
	invalidAnalysis = &Analysis{
		ProblemDomain:  "valid-domain",
		TechnicalStack: []string{"Go"},
		Complexity:     ComplexityLevel("invalid"),
	}
	
	result = validator.ValidateAnalysis(invalidAnalysis)
	if !result.HasErrors() {
		t.Errorf("Should have error for invalid complexity")
	}
}

// TestValidateExecutionPlan tests ExecutionPlan validation
func TestValidateExecutionPlan(t *testing.T) {
	validator := NewDataValidator()
	
	// Test nil plan
	result := validator.ValidateExecutionPlan(nil)
	if !result.HasErrors() {
		t.Errorf("Should have error for nil plan")
	}
	
	// Test valid plan
	validPlan := &ExecutionPlan{
		Steps: []PlanStep{
			{
				ID:            "step-1",
				Name:          "Setup",
				Description:   "Initialize project",
				Dependencies:  []string{},
				EstimatedTime: time.Hour * 2,
				Priority:      1,
			},
		},
	}
	
	result = validator.ValidateExecutionPlan(validPlan)
	if result.HasErrors() {
		t.Errorf("Valid plan should not have errors: %v", result.Errors)
	}
	
	// Test empty steps
	invalidPlan := &ExecutionPlan{
		Steps: []PlanStep{},
	}
	
	result = validator.ValidateExecutionPlan(invalidPlan)
	if !result.HasErrors() {
		t.Errorf("Should have error for empty steps")
	}
	
	// Test duplicate step IDs
	invalidPlan = &ExecutionPlan{
		Steps: []PlanStep{
			{
				ID:            "step-1",
				Name:          "Step 1",
				EstimatedTime: time.Hour,
				Priority:      1,
			},
			{
				ID:            "step-1",
				Name:          "Step 2",
				EstimatedTime: time.Hour,
				Priority:      2,
			},
		},
	}
	
	result = validator.ValidateExecutionPlan(invalidPlan)
	if !result.HasErrors() {
		t.Errorf("Should have error for duplicate step IDs")
	}
}

// TestValidateProjectContext tests ProjectContext validation
func TestValidateProjectContext(t *testing.T) {
	validator := NewDataValidator()
	
	// Test nil context
	result := validator.ValidateProjectContext(nil)
	if !result.HasErrors() {
		t.Errorf("Should have error for nil context")
	}
	
	// Test valid context
	validContext := &ProjectContext{
		WorkingDirectory: "/project/root",
		ProjectConfig: &GOAIConfig{
			ProjectName: "test-project",
			Language:    "go",
		},
	}
	
	result = validator.ValidateProjectContext(validContext)
	if result.HasErrors() {
		t.Errorf("Valid context should not have errors: %v", result.Errors)
	}
	
	// Test empty working directory
	invalidContext := &ProjectContext{
		WorkingDirectory: "",
	}
	
	result = validator.ValidateProjectContext(invalidContext)
	if !result.HasErrors() {
		t.Errorf("Should have error for empty working directory")
	}
}

// TestValidateCodingStyle tests CodingStyle validation
func TestValidateCodingStyle(t *testing.T) {
	validator := NewDataValidator()
	
	// Test valid style
	validStyle := &CodingStyle{
		IndentSize:    4,
		UseSpaces:     true,
		MaxLineLength: 120,
	}
	
	result := validator.ValidateCodingStyle(validStyle)
	if result.HasErrors() {
		t.Errorf("Valid style should not have errors: %v", result.Errors)
	}
	
	// Test invalid indent size
	invalidStyle := &CodingStyle{
		IndentSize:    0,
		MaxLineLength: 120,
	}
	
	result = validator.ValidateCodingStyle(invalidStyle)
	if !result.HasErrors() {
		t.Errorf("Should have error for invalid indent size")
	}
}

// TestValidateString tests string validation utility
func TestValidateString(t *testing.T) {
	// Test valid required string
	err := ValidateString("valid", "test_field", true, 1, 10)
	if err != nil {
		t.Errorf("Valid string should not have error: %v", err)
	}
	
	// Test empty required string
	err = ValidateString("", "test_field", true, 1, 10)
	if err == nil {
		t.Errorf("Should have error for empty required string")
	}
	
	// Test string too short
	err = ValidateString("a", "test_field", false, 5, 10)
	if err == nil {
		t.Errorf("Should have error for string too short")
	}
}

// TestValidateEmail tests email validation utility
func TestValidateEmail(t *testing.T) {
	// Test valid email
	err := ValidateEmail("user@example.com", "email")
	if err != nil {
		t.Errorf("Valid email should not have error: %v", err)
	}
	
	// Test empty email (should be valid)
	err = ValidateEmail("", "email")
	if err != nil {
		t.Errorf("Empty email should be valid: %v", err)
	}
	
	// Test invalid email
	err = ValidateEmail("invalid-email", "email")
	if err == nil {
		t.Errorf("Should have error for invalid email")
	}
}

// TestValidateURL tests URL validation utility
func TestValidateURL(t *testing.T) {
	// Test valid HTTP URL
	err := ValidateURL("http://example.com", "url")
	if err != nil {
		t.Errorf("Valid HTTP URL should not have error: %v", err)
	}
	
	// Test valid HTTPS URL
	err = ValidateURL("https://example.com/path", "url")
	if err != nil {
		t.Errorf("Valid HTTPS URL should not have error: %v", err)
	}
	
	// Test invalid URL
	err = ValidateURL("not-a-url", "url")
	if err == nil {
		t.Errorf("Should have error for invalid URL")
	}
}