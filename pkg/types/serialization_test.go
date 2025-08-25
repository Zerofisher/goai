package types

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestValidatedJSON tests the ValidatedJSON functionality
func TestValidatedJSON(t *testing.T) {
	vj := NewValidatedJSON()
	
	if vj == nil {
		t.Fatalf("ValidatedJSON should not be nil")
	}
	
	if vj.validator == nil {
		t.Errorf("Validator should be initialized")
	}
}

// TestMarshalProblemRequest tests marshaling with validation
func TestMarshalProblemRequest(t *testing.T) {
	vj := NewValidatedJSON()
	
	// Test valid request
	validRequest := &ProblemRequest{
		Description:  "Build a REST API",
		Requirements: []string{"Authentication", "Rate limiting"},
		Constraints:  []string{"Use Go", "PostgreSQL"},
	}
	
	data, err := vj.MarshalProblemRequest(validRequest)
	if err != nil {
		t.Errorf("Valid request should marshal without error: %v", err)
	}
	
	if len(data) == 0 {
		t.Errorf("Marshaled data should not be empty")
	}
	
	// Verify it's valid JSON
	var unmarshaled map[string]interface{}
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Errorf("Marshaled data should be valid JSON: %v", err)
	}
	
	// Test invalid request (empty description)
	invalidRequest := &ProblemRequest{
		Description: "",
	}
	
	_, err = vj.MarshalProblemRequest(invalidRequest)
	if err == nil {
		t.Errorf("Invalid request should fail validation")
	}
	
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("Error should mention validation failure: %v", err)
	}
}

// TestUnmarshalProblemRequest tests unmarshaling with validation
func TestUnmarshalProblemRequest(t *testing.T) {
	vj := NewValidatedJSON()
	
	// Test valid JSON
	validJSON := `{
		"description": "Build a REST API",
		"requirements": ["Authentication", "Rate limiting"],
		"constraints": ["Use Go"]
	}`
	
	req, err := vj.UnmarshalProblemRequest([]byte(validJSON))
	if err != nil {
		t.Errorf("Valid JSON should unmarshal without error: %v", err)
	}
	
	if req.Description != "Build a REST API" {
		t.Errorf("Expected description 'Build a REST API', got %s", req.Description)
	}
	
	if len(req.Requirements) != 2 {
		t.Errorf("Expected 2 requirements, got %d", len(req.Requirements))
	}
	
	// Test invalid JSON (malformed)
	invalidJSON := `{"description": "test"`
	
	_, err = vj.UnmarshalProblemRequest([]byte(invalidJSON))
	if err == nil {
		t.Errorf("Malformed JSON should fail")
	}
	
	// Test JSON that fails validation (empty description)
	invalidValidationJSON := `{
		"description": "",
		"requirements": ["test"]
	}`
	
	_, err = vj.UnmarshalProblemRequest([]byte(invalidValidationJSON))
	if err == nil {
		t.Errorf("JSON with validation errors should fail")
	}
	
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("Error should mention validation failure: %v", err)
	}
}

// TestMarshalAnalysis tests Analysis marshaling with validation
func TestMarshalAnalysis(t *testing.T) {
	vj := NewValidatedJSON()
	
	validAnalysis := &Analysis{
		ProblemDomain:       "web-development",
		TechnicalStack:      []string{"Go", "PostgreSQL"},
		ArchitecturePattern: "REST API",
		Complexity:          ComplexityMedium,
	}
	
	data, err := vj.MarshalAnalysis(validAnalysis)
	if err != nil {
		t.Errorf("Valid analysis should marshal without error: %v", err)
	}
	
	if len(data) == 0 {
		t.Errorf("Marshaled data should not be empty")
	}
	
	// Test invalid analysis (empty problem domain)
	invalidAnalysis := &Analysis{
		ProblemDomain:  "",
		TechnicalStack: []string{"Go"},
		Complexity:     ComplexityLow,
	}
	
	_, err = vj.MarshalAnalysis(invalidAnalysis)
	if err == nil {
		t.Errorf("Invalid analysis should fail validation")
	}
}

// TestMarshalExecutionPlan tests ExecutionPlan marshaling
func TestMarshalExecutionPlan(t *testing.T) {
	vj := NewValidatedJSON()
	
	validPlan := &ExecutionPlan{
		Steps: []PlanStep{
			{
				ID:            "step-1",
				Name:          "Setup",
				Description:   "Initialize project",
				EstimatedTime: time.Hour * 2,
				Priority:      1,
			},
		},
	}
	
	data, err := vj.MarshalExecutionPlan(validPlan)
	if err != nil {
		t.Errorf("Valid plan should marshal without error: %v", err)
	}
	
	if len(data) == 0 {
		t.Errorf("Marshaled data should not be empty")
	}
	
	// Test invalid plan (empty steps)
	invalidPlan := &ExecutionPlan{
		Steps: []PlanStep{},
	}
	
	_, err = vj.MarshalExecutionPlan(invalidPlan)
	if err == nil {
		t.Errorf("Invalid plan should fail validation")
	}
}

// TestConvenienceFunctions tests the convenience functions
func TestConvenienceFunctions(t *testing.T) {
	// Test ProblemRequestToJSON and FromJSON
	request := &ProblemRequest{
		Description:  "Test description",
		Requirements: []string{"req1", "req2"},
	}
	
	jsonStr, err := ProblemRequestToJSON(request)
	if err != nil {
		t.Errorf("ProblemRequestToJSON failed: %v", err)
	}
	
	if jsonStr == "" {
		t.Errorf("JSON string should not be empty")
	}
	
	parsedRequest, err := ProblemRequestFromJSON(jsonStr)
	if err != nil {
		t.Errorf("ProblemRequestFromJSON failed: %v", err)
	}
	
	if parsedRequest.Description != request.Description {
		t.Errorf("Descriptions should match after round-trip")
	}
	
	// Test AnalysisToJSON and FromJSON
	analysis := &Analysis{
		ProblemDomain:  "test-domain",
		TechnicalStack: []string{"Go"},
		Complexity:     ComplexityLow,
	}
	
	jsonStr, err = AnalysisToJSON(analysis)
	if err != nil {
		t.Errorf("AnalysisToJSON failed: %v", err)
	}
	
	parsedAnalysis, err := AnalysisFromJSON(jsonStr)
	if err != nil {
		t.Errorf("AnalysisFromJSON failed: %v", err)
	}
	
	if parsedAnalysis.ProblemDomain != analysis.ProblemDomain {
		t.Errorf("Problem domains should match after round-trip")
	}
}

// TestValidateAndFormat tests the ValidateAndFormat function
func TestValidateAndFormat(t *testing.T) {
	// Test with ProblemRequest pointer
	request := &ProblemRequest{
		Description:  "Test",
		Requirements: []string{"req1"},
	}
	
	result, err := ValidateAndFormat(request)
	if err != nil {
		t.Errorf("ValidateAndFormat failed for ProblemRequest: %v", err)
	}
	
	if result == "" {
		t.Errorf("Result should not be empty")
	}
	
	// Test with ProblemRequest value
	requestValue := ProblemRequest{
		Description:  "Test",
		Requirements: []string{"req1"},
	}
	
	result, err = ValidateAndFormat(requestValue)
	if err != nil {
		t.Errorf("ValidateAndFormat failed for ProblemRequest value: %v", err)
	}
	if result == "" {
		t.Error("Expected non-empty result for ProblemRequest value")
	}
	
	// Test with Analysis
	analysis := &Analysis{
		ProblemDomain:  "test",
		TechnicalStack: []string{"Go"},
		Complexity:     ComplexityLow,
	}
	
	result, err = ValidateAndFormat(analysis)
	if err != nil {
		t.Errorf("ValidateAndFormat failed for Analysis: %v", err)
	}
	if result == "" {
		t.Error("Expected non-empty result for Analysis")
	}
	
	// Test with unsupported type
	unsupported := struct {
		Field string `json:"field"`
	}{Field: "value"}
	
	result, err = ValidateAndFormat(unsupported)
	if err != nil {
		t.Errorf("ValidateAndFormat should handle unsupported types: %v", err)
	}
	
	if !strings.Contains(result, "value") {
		t.Errorf("Result should contain the field value")
	}
}

// TestBatchValidate tests batch validation
func TestBatchValidate(t *testing.T) {
	// Test with all valid items
	validItems := map[string]interface{}{
		"request": &ProblemRequest{
			Description:  "Test",
			Requirements: []string{"req1"},
		},
		"analysis": &Analysis{
			ProblemDomain:  "test",
			TechnicalStack: []string{"Go"},
			Complexity:     ComplexityLow,
		},
	}
	
	result := ValidateBatch(validItems)
	if !result.Valid {
		t.Errorf("All valid items should pass batch validation: %v", result.Errors)
	}
	
	if len(result.Errors) != 0 {
		t.Errorf("Should have no errors for valid items")
	}
	
	// Test with mixed valid/invalid items
	mixedItems := map[string]interface{}{
		"valid_request": &ProblemRequest{
			Description:  "Valid",
			Requirements: []string{"req1"},
		},
		"invalid_request": &ProblemRequest{
			Description: "", // Invalid: empty description
		},
		"unsupported": "string", // Will be skipped
	}
	
	result = ValidateBatch(mixedItems)
	if result.Valid {
		t.Errorf("Batch with invalid items should fail validation")
	}
	
	if len(result.Errors) != 1 {
		t.Errorf("Should have exactly 1 error, got %d", len(result.Errors))
	}
	
	if _, exists := result.Errors["invalid_request"]; !exists {
		t.Errorf("Should have error for invalid_request")
	}
}

// TestSafeUnmarshal tests the SafeUnmarshal function
func TestSafeUnmarshal(t *testing.T) {
	// Test ProblemRequest
	requestJSON := `{
		"description": "Test request",
		"requirements": ["req1", "req2"]
	}`
	
	result, err := SafeUnmarshal([]byte(requestJSON), "ProblemRequest")
	if err != nil {
		t.Errorf("SafeUnmarshal failed for ProblemRequest: %v", err)
	}
	
	request, ok := result.(*ProblemRequest)
	if !ok {
		t.Errorf("Result should be *ProblemRequest")
	}
	
	if request.Description != "Test request" {
		t.Errorf("Description should match")
	}
	
	// Test Analysis
	analysisJSON := `{
		"problem_domain": "test-domain",
		"technical_stack": ["Go", "PostgreSQL"],
		"complexity": "low"
	}`
	
	result, err = SafeUnmarshal([]byte(analysisJSON), "Analysis")
	if err != nil {
		t.Errorf("SafeUnmarshal failed for Analysis: %v", err)
	}
	
	analysis, ok := result.(*Analysis)
	if !ok {
		t.Errorf("Result should be *Analysis")
	}
	
	if analysis.ProblemDomain != "test-domain" {
		t.Errorf("Problem domain should match")
	}
	
	// Test unsupported type
	_, err = SafeUnmarshal([]byte("{}"), "UnsupportedType")
	if err == nil {
		t.Errorf("Should fail for unsupported type")
	}
	
	if !strings.Contains(err.Error(), "unsupported target type") {
		t.Errorf("Error should mention unsupported type: %v", err)
	}
	
	// Test invalid JSON
	_, err = SafeUnmarshal([]byte("invalid json"), "ProblemRequest")
	if err == nil {
		t.Errorf("Should fail for invalid JSON")
	}
}

// TestApplyJSONPatch tests the JSON patch functionality
func TestApplyJSONPatch(t *testing.T) {
	original := struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
		Flag  bool   `json:"flag"`
	}{
		Name:  "original",
		Value: 10,
		Flag:  true,
	}
	
	patches := []JSONPatch{
		{Op: "replace", Path: "/name", Value: "updated"},
		{Op: "replace", Path: "/value", Value: 20},
		{Op: "add", Path: "/new_field", Value: "new_value"},
		{Op: "remove", Path: "/flag"},
	}
	
	result, err := ApplyJSONPatch(original, patches)
	if err != nil {
		t.Errorf("ApplyJSONPatch failed: %v", err)
	}
	
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Errorf("Result should be map[string]interface{}")
	}
	
	// Check replace operations
	if resultMap["name"] != "updated" {
		t.Errorf("Expected name 'updated', got %v", resultMap["name"])
	}
	
	if resultMap["value"] != 20 {
		t.Errorf("Expected value 20, got %v", resultMap["value"])
	}
	
	// Check add operation
	if resultMap["new_field"] != "new_value" {
		t.Errorf("Expected new_field 'new_value', got %v", resultMap["new_field"])
	}
	
	// Check remove operation
	if _, exists := resultMap["flag"]; exists {
		t.Errorf("Flag field should be removed")
	}
	
	// Test unsupported operation
	invalidPatches := []JSONPatch{
		{Op: "move", Path: "/name", Value: "test"},
	}
	
	_, err = ApplyJSONPatch(original, invalidPatches)
	if err == nil {
		t.Errorf("Should fail for unsupported patch operation")
	}
	
	if !strings.Contains(err.Error(), "unsupported patch operation") {
		t.Errorf("Error should mention unsupported operation: %v", err)
	}
}

// TestRoundTripSerialization tests complete serialization round trips
func TestRoundTripSerialization(t *testing.T) {
	// Test ExecutionPlan round trip
	originalPlan := &ExecutionPlan{
		Steps: []PlanStep{
			{
				ID:            "step-1",
				Name:          "Setup",
				Description:   "Initialize",
				EstimatedTime: time.Hour * 2,
				Priority:      1,
			},
		},
		Dependencies: []Dependency{
			{
				Name:     "gin",
				Version:  "v1.9.1",
				Type:     "library",
				Required: true,
			},
		},
	}
	
	jsonStr, err := ExecutionPlanToJSON(originalPlan)
	if err != nil {
		t.Errorf("ExecutionPlanToJSON failed: %v", err)
	}
	
	parsedPlan, err := ExecutionPlanFromJSON(jsonStr)
	if err != nil {
		t.Errorf("ExecutionPlanFromJSON failed: %v", err)
	}
	
	// Compare key fields
	if len(parsedPlan.Steps) != len(originalPlan.Steps) {
		t.Errorf("Steps count should match")
	}
	
	if parsedPlan.Steps[0].ID != originalPlan.Steps[0].ID {
		t.Errorf("Step ID should match")
	}
	
	if len(parsedPlan.Dependencies) != len(originalPlan.Dependencies) {
		t.Errorf("Dependencies count should match")
	}
	
	// Test ProjectContext round trip
	originalContext := &ProjectContext{
		WorkingDirectory: "/project/root",
		ProjectConfig: &GOAIConfig{
			ProjectName: "test-project",
			Language:    "go",
		},
	}
	
	jsonStr, err = ProjectContextToJSON(originalContext)
	if err != nil {
		t.Errorf("ProjectContextToJSON failed: %v", err)
	}
	
	parsedContext, err := ProjectContextFromJSON(jsonStr)
	if err != nil {
		t.Errorf("ProjectContextFromJSON failed: %v", err)
	}
	
	if parsedContext.WorkingDirectory != originalContext.WorkingDirectory {
		t.Errorf("Working directory should match")
	}
	
	if parsedContext.ProjectConfig.ProjectName != originalContext.ProjectConfig.ProjectName {
		t.Errorf("Project name should match")
	}
}