// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Zerofisher/goai/internal/reasoning"
	"github.com/Zerofisher/goai/pkg/types"
)

func main() {
	ctx := context.Background()
	
	// Create mock context manager
	mockContextManager := &MockContextManager{}
	
	// Create reasoning engine
	engine, err := reasoning.NewEngine(ctx, mockContextManager)
	if err != nil {
		log.Fatalf("Failed to create engine: %v", err)
	}
	
	// Test problem request
	req := &types.ProblemRequest{
		Description:  "Create a simple HTTP server in Go that handles GET and POST requests",
		Context:      &types.ProjectContext{WorkingDirectory: "/tmp/test"},
		Requirements: []string{
			"Handle GET /health endpoint", 
			"Handle POST /data endpoint",
			"Return JSON responses",
			"Include proper error handling",
		},
		Constraints:  []string{"Use only standard library", "Include tests"},
	}
	
	fmt.Println("🔍 Step 1: Analyzing problem...")
	analysis, err := engine.AnalyzeProblem(ctx, req)
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}
	fmt.Printf("✅ Analysis completed: %s\n", analysis.ProblemDomain)
	
	fmt.Println("\n📋 Step 2: Generating execution plan...")
	plan, err := engine.GeneratePlan(ctx, analysis)
	if err != nil {
		log.Fatalf("Planning failed: %v", err)
	}
	fmt.Printf("✅ Plan generated with %d steps\n", len(plan.Steps))
	
	fmt.Println("\n⚡ Step 3: Executing plan...")
	codeResult, err := engine.ExecutePlan(ctx, plan)
	if err != nil {
		log.Fatalf("Execution failed: %v", err)
	}
	fmt.Printf("✅ Code generated: %d files\n", len(codeResult.GeneratedFiles))
	
	fmt.Println("\n🔍 Step 4: Validating result...")
	validation, err := engine.ValidateResult(ctx, codeResult)
	if err != nil {
		log.Fatalf("Validation failed: %v", err)
	}
	fmt.Printf("✅ Validation completed: %s\n", validation.OverallStatus)
	
	fmt.Println("\n🎉 Full reasoning chain completed successfully!")
	
	// Print summary
	fmt.Println("\n=== SUMMARY ===")
	fmt.Printf("Problem Domain: %s\n", analysis.ProblemDomain)
	fmt.Printf("Architecture: %s\n", analysis.ArchitecturePattern)
	fmt.Printf("Tech Stack: %v\n", analysis.TechnicalStack)
	fmt.Printf("Steps: %d\n", len(plan.Steps))
	fmt.Printf("Generated Files: %d\n", len(codeResult.GeneratedFiles))
	fmt.Printf("Validation Status: %s\n", validation.OverallStatus)
	
	// Show first generated file
	if len(codeResult.GeneratedFiles) > 0 {
		fmt.Printf("\nFirst Generated File (%s):\n", codeResult.GeneratedFiles[0].Path)
		fmt.Printf("```go\n%s\n```\n", codeResult.GeneratedFiles[0].Content[:min(500, len(codeResult.GeneratedFiles[0].Content))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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