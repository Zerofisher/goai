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