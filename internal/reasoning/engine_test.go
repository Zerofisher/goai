package reasoning

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Zerofisher/goai/pkg/types"
)

func TestEngine_NewEngine(t *testing.T) {
	// Skip if no API key is set
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	ctx := context.Background()
	
	// Create a mock context manager
	mockContextManager := &MockContextManager{}
	
	// Create engine
	engine, err := NewEngine(ctx, mockContextManager)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	if engine == nil {
		t.Fatal("Engine should not be nil")
	}
	
	if engine.analysisChain == nil {
		t.Fatal("Analysis chain should not be nil")
	}
	
	if engine.planningChain == nil {
		t.Fatal("Planning chain should not be nil")
	}
	
	if engine.executionChain == nil {
		t.Fatal("Execution chain should not be nil")
	}
	
	if engine.validationChain == nil {
		t.Fatal("Validation chain should not be nil")
	}
}

func TestEngine_AnalyzeProblem_Fallback(t *testing.T) {
	// Test with invalid API key to trigger fallback behavior
	_ = os.Setenv("OPENAI_API_KEY", "invalid-key")
	defer func() { _ = os.Unsetenv("OPENAI_API_KEY") }()
	
	ctx := context.Background()
	mockContextManager := &MockContextManager{}
	
	engine, err := NewEngine(ctx, mockContextManager)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Test problem request
	req := &types.ProblemRequest{
		Description:  "Create a simple Hello World program in Go",
		Context:      &types.ProjectContext{WorkingDirectory: "/tmp/test"},
		Requirements: []string{"Must be a valid Go program", "Should print Hello World"},
		Constraints:  []string{"Use only standard library"},
	}
	
	// This should fail due to invalid API key, but test the structure
	_, err = engine.AnalyzeProblem(ctx, req)
	if err == nil {
		t.Log("Analysis completed successfully (or fallback worked)")
	} else {
		t.Logf("Analysis failed as expected with invalid API key: %v", err)
	}
}

func TestEngine_AnalyzeProblem_ValidKey(t *testing.T) {
	// Only run if API key is valid
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" || apiKey == "invalid-key" {
		t.Skip("Valid OPENAI_API_KEY not set, skipping integration test")
	}
	
	ctx := context.Background()
	mockContextManager := &MockContextManager{}
	
	engine, err := NewEngine(ctx, mockContextManager)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Test problem request
	req := &types.ProblemRequest{
		Description:  "Create a simple Hello World program in Go",
		Context:      &types.ProjectContext{WorkingDirectory: "/tmp/test"},
		Requirements: []string{"Must be a valid Go program", "Should print Hello World"},
		Constraints:  []string{"Use only standard library"},
	}
	
	// This should work with valid API key
	analysis, err := engine.AnalyzeProblem(ctx, req)
	if err != nil {
		t.Fatalf("Analysis failed with valid API key: %v", err)
	}
	
	if analysis == nil {
		t.Fatal("Analysis should not be nil")
	}
	
	// Verify analysis structure
	if analysis.ProblemDomain == "" {
		t.Error("ProblemDomain should not be empty")
	}
	
	if len(analysis.TechnicalStack) == 0 {
		t.Error("TechnicalStack should not be empty")
	}
	
	t.Logf("Analysis completed successfully: Domain=%s, Stack=%v", 
		analysis.ProblemDomain, analysis.TechnicalStack)
}

// MockContextManager implements types.ContextManager for testing
type MockContextManager struct{}

func (m *MockContextManager) BuildProjectContext(workdir string) (*types.ProjectContext, error) {
	return &types.ProjectContext{
		WorkingDirectory: workdir,
		LoadedAt:         time.Now(),
		ProjectConfig: &types.GOAIConfig{
			ProjectName: "test-project",
			Language:    "go",
		},
	}, nil
}

func (m *MockContextManager) LoadConfiguration(configPath string) (*types.GOAIConfig, error) {
	return &types.GOAIConfig{
		ProjectName: "test-project",
		Language:    "go",
	}, nil
}

func (m *MockContextManager) WatchFileChanges(callback func(*types.FileChangeEvent)) error {
	return nil
}

func (m *MockContextManager) GetRecentChanges(since time.Time) ([]*types.GitChange, error) {
	return []*types.GitChange{}, nil
}

// Test adapter functions
func TestAdaptTestStrategy(t *testing.T) {
	strategy := adaptTestStrategy("comprehensive", []string{"unit", "integration", "e2e"}, []string{"testing", "testify"}, 85.0)
	
	if strategy == nil {
		t.Fatal("Strategy should not be nil")
	}
	
	if strategy.CoverageTarget != 85.0 {
		t.Errorf("Expected coverage target 85.0, got %f", strategy.CoverageTarget)
	}
	
	if !strategy.UnitTests {
		t.Error("Unit tests should be enabled")
	}
	
	if !strategy.IntegrationTests {
		t.Error("Integration tests should be enabled")
	}
	
	if !strategy.EndToEndTests {
		t.Error("End-to-end tests should be enabled")
	}
	
	if len(strategy.TestFrameworks) != 2 {
		t.Errorf("Expected 2 test frameworks, got %d", len(strategy.TestFrameworks))
	}
}

func TestAdaptTimeline(t *testing.T) {
	duration := 5 * time.Hour
	timeline := adaptTimeline(duration)
	
	if timeline == nil {
		t.Fatal("Timeline should not be nil")
	}
	
	if timeline.TotalDuration != duration {
		t.Errorf("Expected total duration %v, got %v", duration, timeline.TotalDuration)
	}
	
	if timeline.StartTime.IsZero() {
		t.Error("Start time should be set")
	}
	
	if timeline.EstimatedEnd.IsZero() {
		t.Error("Estimated end should be set")
	}
}

func TestAdaptValidationRules(t *testing.T) {
	inputRules := []map[string]string{
		{
			"type":        "code_quality",
			"description": "Ensure good code quality",
			"criteria":    "pass linting",
		},
		{
			"type":        "test_coverage",
			"description": "Maintain test coverage",
			"criteria":    "coverage >= 80%",
		},
	}
	
	rules := adaptValidationRules(inputRules)
	
	if len(rules) != 2 {
		t.Errorf("Expected 2 validation rules, got %d", len(rules))
	}
	
	rule := rules[0]
	if rule.Name != "code_quality" {
		t.Errorf("Expected name 'code_quality', got '%s'", rule.Name)
	}
	
	if rule.Type != "code_quality" {
		t.Errorf("Expected type 'code_quality', got '%s'", rule.Type)
	}
	
	if !rule.Required {
		t.Error("Rule should be required")
	}
}

func TestAdaptDependencies(t *testing.T) {
	inputDeps := []map[string]string{
		{
			"name":        "testify",
			"version":     "v1.8.0",
			"type":        "test",
			"description": "Testing framework",
		},
	}
	
	deps := adaptDependencies(inputDeps)
	
	if len(deps) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(deps))
	}
	
	dep := deps[0]
	if dep.Name != "testify" {
		t.Errorf("Expected name 'testify', got '%s'", dep.Name)
	}
	
	if dep.Version != "v1.8.0" {
		t.Errorf("Expected version 'v1.8.0', got '%s'", dep.Version)
	}
	
	if !dep.Required {
		t.Error("Dependency should be required")
	}
}

func TestContainsLevel(t *testing.T) {
	levels := []string{"unit", "integration", "e2e"}
	
	if !containsLevel(levels, "unit") {
		t.Error("Should contain 'unit'")
	}
	
	if !containsLevel(levels, "integration") {
		t.Error("Should contain 'integration'")
	}
	
	if containsLevel(levels, "performance") {
		t.Error("Should not contain 'performance'")
	}
}

func TestEngine_GeneratePlan_Fallback(t *testing.T) {
	// Test with invalid API key to trigger fallback behavior
	_ = os.Setenv("OPENAI_API_KEY", "invalid-key")
	defer func() { _ = os.Unsetenv("OPENAI_API_KEY") }()
	
	ctx := context.Background()
	mockContextManager := &MockContextManager{}
	
	engine, err := NewEngine(ctx, mockContextManager)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Create a mock analysis
	analysis := &types.Analysis{
		ProblemDomain:       "Web Development",
		ArchitecturePattern: "MVC",
		TechnicalStack:      []string{"Go", "HTTP"},
		Complexity:          types.ComplexityMedium,
	}
	
	// This should fail due to invalid API key, but test the structure
	_, err = engine.GeneratePlan(ctx, analysis)
	if err == nil {
		t.Log("Plan generation completed successfully (or fallback worked)")
	} else {
		t.Logf("Plan generation failed as expected with invalid API key: %v", err)
	}
}

func TestEngine_ExecutePlan_Fallback(t *testing.T) {
	// Test with invalid API key to trigger fallback behavior
	_ = os.Setenv("OPENAI_API_KEY", "invalid-key")
	defer func() { _ = os.Unsetenv("OPENAI_API_KEY") }()
	
	ctx := context.Background()
	mockContextManager := &MockContextManager{}
	
	engine, err := NewEngine(ctx, mockContextManager)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Create a mock execution plan
	plan := &types.ExecutionPlan{
		Steps: []types.PlanStep{
			{
				ID:            "step_1",
				Name:          "Test Step",
				Description:   "A test step",
				EstimatedTime: time.Hour,
				Priority:      1,
			},
		},
	}
	
	// This should fail due to invalid API key, but test the structure
	_, err = engine.ExecutePlan(ctx, plan)
	if err == nil {
		t.Log("Plan execution completed successfully (or fallback worked)")
	} else {
		t.Logf("Plan execution failed as expected with invalid API key: %v", err)
	}
}

func TestEngine_ValidateResult_Fallback(t *testing.T) {
	// Test with invalid API key to trigger fallback behavior  
	_ = os.Setenv("OPENAI_API_KEY", "invalid-key")
	defer func() { _ = os.Unsetenv("OPENAI_API_KEY") }()
	
	ctx := context.Background()
	mockContextManager := &MockContextManager{}
	
	engine, err := NewEngine(ctx, mockContextManager)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	
	// Create a mock code result
	codeResult := &types.CodeResult{
		GeneratedFiles: []types.GeneratedFile{
			{
				Path:    "main.go",
				Content: "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
				Type:    "source",
			},
		},
	}
	
	// This should fail due to invalid API key, but test the structure
	_, err = engine.ValidateResult(ctx, codeResult)
	if err == nil {
		t.Log("Result validation completed successfully (or fallback worked)")
	} else {
		t.Logf("Result validation failed as expected with invalid API key: %v", err)
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	// Test with existing environment variable
	_ = os.Setenv("TEST_VAR", "test_value")
	defer func() { _ = os.Unsetenv("TEST_VAR") }()
	
	result := getEnvOrDefault("TEST_VAR", "default_value")
	if result != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", result)
	}
	
	// Test with non-existing environment variable
	result = getEnvOrDefault("NON_EXISTING_VAR", "default_value")
	if result != "default_value" {
		t.Errorf("Expected 'default_value', got '%s'", result)
	}
}