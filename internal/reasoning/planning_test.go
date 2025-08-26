package reasoning

import (
	"testing"
	"time"
)

func TestGetStringFromMap(t *testing.T) {
	testMap := map[string]interface{}{
		"string_key": "test_value",
		"int_key":    42,
		"nil_key":    nil,
	}
	
	// Test existing string value
	result := getStringFromMap(testMap, "string_key", "default")
	if result != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", result)
	}
	
	// Test non-string value
	result = getStringFromMap(testMap, "int_key", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got '%s'", result)
	}
	
	// Test non-existing key
	result = getStringFromMap(testMap, "missing_key", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got '%s'", result)
	}
	
	// Test nil value
	result = getStringFromMap(testMap, "nil_key", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got '%s'", result)
	}
}

func TestGetIntFromMap(t *testing.T) {
	testMap := map[string]interface{}{
		"int_key":    42,
		"float_key":  3.14,
		"string_key": "not_a_number",
		"nil_key":    nil,
	}
	
	// Test existing int value
	result := getIntFromMap(testMap, "int_key", 0)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
	
	// Test float value (should be converted)
	result = getIntFromMap(testMap, "float_key", 0)
	if result != 3 {
		t.Errorf("Expected 3, got %d", result)
	}
	
	// Test non-numeric value
	result = getIntFromMap(testMap, "string_key", 99)
	if result != 99 {
		t.Errorf("Expected 99, got %d", result)
	}
	
	// Test non-existing key
	result = getIntFromMap(testMap, "missing_key", 99)
	if result != 99 {
		t.Errorf("Expected 99, got %d", result)
	}
	
	// Test nil value
	result = getIntFromMap(testMap, "nil_key", 99)
	if result != 99 {
		t.Errorf("Expected 99, got %d", result)
	}
}

func TestCreateFallbackPlan(t *testing.T) {
	plan := createFallbackPlan("some content")
	
	if plan == nil {
		t.Fatal("Plan should not be nil")
	}
	
	if len(plan.Steps) != 5 {
		t.Errorf("Expected 5 steps, got %d", len(plan.Steps))
	}
	
	// Check first step
	firstStep := plan.Steps[0]
	if firstStep.ID != "step_1" {
		t.Errorf("Expected first step ID 'step_1', got '%s'", firstStep.ID)
	}
	
	if firstStep.Name != "Analyze Requirements" {
		t.Errorf("Expected first step name 'Analyze Requirements', got '%s'", firstStep.Name)
	}
	
	if firstStep.EstimatedTime != time.Hour {
		t.Errorf("Expected first step time 1 hour, got %v", firstStep.EstimatedTime)
	}
	
	if firstStep.Priority != 1 {
		t.Errorf("Expected first step priority 1, got %d", firstStep.Priority)
	}
	
	// Check step with dependencies
	thirdStep := plan.Steps[2]
	if len(thirdStep.Dependencies) != 1 {
		t.Errorf("Expected third step to have 1 dependency, got %d", len(thirdStep.Dependencies))
	}
	
	if thirdStep.Dependencies[0] != "step_2" {
		t.Errorf("Expected third step dependency 'step_2', got '%s'", thirdStep.Dependencies[0])
	}
	
	// Check timeline
	if plan.Timeline == nil {
		t.Error("Timeline should not be nil")
	}
	
	if plan.Timeline.TotalDuration != 10*time.Hour {
		t.Errorf("Expected total duration 10 hours, got %v", plan.Timeline.TotalDuration)
	}
	
	// Check test strategy
	if plan.TestStrategy == nil {
		t.Error("Test strategy should not be nil")
	}
	
	if plan.TestStrategy.CoverageTarget != 80.0 {
		t.Errorf("Expected coverage target 80.0, got %f", plan.TestStrategy.CoverageTarget)
	}
	
	if !plan.TestStrategy.UnitTests {
		t.Error("Unit tests should be enabled")
	}
	
	if !plan.TestStrategy.IntegrationTests {
		t.Error("Integration tests should be enabled")
	}
	
	// Check validation rules
	if len(plan.ValidationRules) != 2 {
		t.Errorf("Expected 2 validation rules, got %d", len(plan.ValidationRules))
	}
	
	codeQualityRule := plan.ValidationRules[0]
	if codeQualityRule.Type != "code_quality" {
		t.Errorf("Expected first rule type 'code_quality', got '%s'", codeQualityRule.Type)
	}
}

func TestParsePlanningResponse_InvalidJSON(t *testing.T) {
	// Test with invalid JSON to trigger fallback
	invalidJSON := "This is not valid JSON"
	
	plan, err := parsePlanningResponse(invalidJSON)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if plan == nil {
		t.Fatal("Plan should not be nil")
	}
	
	// Should return fallback plan
	if len(plan.Steps) != 5 {
		t.Errorf("Expected fallback plan with 5 steps, got %d", len(plan.Steps))
	}
}

func TestParsePlanningResponse_ValidJSON(t *testing.T) {
	validJSON := `{
		"steps": [
			{
				"id": "step_1",
				"name": "Test Step",
				"description": "A test step",
				"dependencies": ["dep1"],
				"estimated_time": "2h",
				"priority": 1
			}
		],
		"dependencies": [
			{
				"name": "testlib",
				"version": "v1.0.0",
				"type": "library",
				"description": "Test library"
			}
		],
		"timeline": {
			"total_estimate": "4h"
		},
		"test_strategy": {
			"approach": "comprehensive",
			"levels": ["unit", "integration"],
			"frameworks": ["testing"],
			"coverage_target": 90
		},
		"validation_rules": [
			{
				"type": "quality",
				"description": "Quality check",
				"criteria": "pass all checks"
			}
		]
	}`
	
	plan, err := parsePlanningResponse(validJSON)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if plan == nil {
		t.Fatal("Plan should not be nil")
	}
	
	if len(plan.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(plan.Steps))
	}
	
	step := plan.Steps[0]
	if step.ID != "step_1" {
		t.Errorf("Expected step ID 'step_1', got '%s'", step.ID)
	}
	
	if step.EstimatedTime != 2*time.Hour {
		t.Errorf("Expected estimated time 2h, got %v", step.EstimatedTime)
	}
	
	if len(step.Dependencies) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(step.Dependencies))
	}
	
	if step.Dependencies[0] != "dep1" {
		t.Errorf("Expected dependency 'dep1', got '%s'", step.Dependencies[0])
	}
	
	// Check test strategy
	if plan.TestStrategy == nil {
		t.Error("Test strategy should not be nil")
	}
	
	if plan.TestStrategy.CoverageTarget != 90.0 {
		t.Errorf("Expected coverage target 90.0, got %f", plan.TestStrategy.CoverageTarget)
	}
	
	if !plan.TestStrategy.UnitTests {
		t.Error("Unit tests should be enabled")
	}
	
	if !plan.TestStrategy.IntegrationTests {
		t.Error("Integration tests should be enabled")
	}
}

func TestParsePlanningResponse_MarkdownWrappedJSON(t *testing.T) {
	markdownJSON := "```json\n" + `{
		"steps": [
			{
				"id": "step_1",
				"name": "Test Step",
				"description": "A test step",
				"estimated_time": "1h",
				"priority": 1
			}
		],
		"dependencies": [],
		"timeline": {"total_estimate": "1h"},
		"test_strategy": {"approach": "basic", "levels": ["unit"], "frameworks": ["testing"], "coverage_target": 75},
		"validation_rules": []
	}` + "\n```"
	
	plan, err := parsePlanningResponse(markdownJSON)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if plan == nil {
		t.Fatal("Plan should not be nil")
	}
	
	if len(plan.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(plan.Steps))
	}
	
	step := plan.Steps[0]
	if step.Name != "Test Step" {
		t.Errorf("Expected step name 'Test Step', got '%s'", step.Name)
	}
}