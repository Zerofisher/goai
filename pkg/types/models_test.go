package types

import (
	"encoding/json"
	"testing"
	"time"
)

// TestProblemRequestSerialization tests JSON serialization/deserialization for ProblemRequest
func TestProblemRequestSerialization(t *testing.T) {
	original := &ProblemRequest{
		Description:  "Test problem description",
		Requirements: []string{"req1", "req2"},
		Constraints:  []string{"constraint1"},
		Context: &ProjectContext{
			WorkingDirectory: "/test/dir",
		},
		PreferredStyle: &CodingStyle{
			IndentSize:    4,
			UseSpaces:     true,
			MaxLineLength: 80,
		},
	}

	// Test marshaling
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal ProblemRequest: %v", err)
	}

	// Test unmarshaling
	var unmarshaled ProblemRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ProblemRequest: %v", err)
	}

	// Verify fields
	if unmarshaled.Description != original.Description {
		t.Errorf("Description mismatch: got %s, want %s", unmarshaled.Description, original.Description)
	}
	if len(unmarshaled.Requirements) != len(original.Requirements) {
		t.Errorf("Requirements length mismatch: got %d, want %d", len(unmarshaled.Requirements), len(original.Requirements))
	}
}

// TestAnalysisSerialization tests JSON serialization/deserialization for Analysis
func TestAnalysisSerialization(t *testing.T) {
	original := &Analysis{
		ProblemDomain:       "web-development",
		TechnicalStack:      []string{"Go", "HTTP", "JSON"},
		ArchitecturePattern: "REST API",
		Complexity:          ComplexityMedium,
		RiskFactors: []RiskFactor{
			{
				Type:        "performance",
				Description: "High load expected",
				Severity:    "medium",
				Mitigation:  "Use caching",
			},
		},
		Recommendations: []Recommendation{
			{
				Category:    "architecture",
				Description: "Use microservices",
				Priority:    "high",
				Impact:      "significant",
			},
		},
	}

	// Test marshaling
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal Analysis: %v", err)
	}

	// Test unmarshaling
	var unmarshaled Analysis
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal Analysis: %v", err)
	}

	// Verify fields
	if unmarshaled.ProblemDomain != original.ProblemDomain {
		t.Errorf("ProblemDomain mismatch: got %s, want %s", unmarshaled.ProblemDomain, original.ProblemDomain)
	}
	if unmarshaled.Complexity != original.Complexity {
		t.Errorf("Complexity mismatch: got %s, want %s", unmarshaled.Complexity, original.Complexity)
	}
	if len(unmarshaled.RiskFactors) != len(original.RiskFactors) {
		t.Errorf("RiskFactors length mismatch: got %d, want %d", len(unmarshaled.RiskFactors), len(original.RiskFactors))
	}
}

// TestExecutionPlanSerialization tests JSON serialization/deserialization for ExecutionPlan
func TestExecutionPlanSerialization(t *testing.T) {
	now := time.Now()
	original := &ExecutionPlan{
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
		Dependencies: []Dependency{
			{
				Name:        "github.com/gin-gonic/gin",
				Version:     "v1.9.1",
				Type:        "library",
				Required:    true,
				Description: "Web framework",
			},
		},
		Timeline: &Timeline{
			StartTime:     now,
			EstimatedEnd:  now.Add(time.Hour * 24),
			TotalDuration: time.Hour * 24,
			Milestones: []Milestone{
				{
					Name:        "MVP",
					Description: "Minimum viable product",
					DueDate:     now.Add(time.Hour * 12),
					Completed:   false,
				},
			},
		},
		TestStrategy: &TestStrategy{
			UnitTests:        true,
			IntegrationTests: true,
			EndToEndTests:    false,
			TestFrameworks:   []string{"testing", "testify"},
			CoverageTarget:   80.0,
		},
	}

	// Test marshaling
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal ExecutionPlan: %v", err)
	}

	// Test unmarshaling
	var unmarshaled ExecutionPlan
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ExecutionPlan: %v", err)
	}

	// Verify fields
	if len(unmarshaled.Steps) != len(original.Steps) {
		t.Errorf("Steps length mismatch: got %d, want %d", len(unmarshaled.Steps), len(original.Steps))
	}
	if unmarshaled.Steps[0].Name != original.Steps[0].Name {
		t.Errorf("Step name mismatch: got %s, want %s", unmarshaled.Steps[0].Name, original.Steps[0].Name)
	}
	if unmarshaled.TestStrategy.CoverageTarget != original.TestStrategy.CoverageTarget {
		t.Errorf("Coverage target mismatch: got %f, want %f", unmarshaled.TestStrategy.CoverageTarget, original.TestStrategy.CoverageTarget)
	}
}

// TestProjectContextSerialization tests JSON serialization/deserialization for ProjectContext
func TestProjectContextSerialization(t *testing.T) {
	original := &ProjectContext{
		WorkingDirectory: "/project/root",
		ProjectConfig: &GOAIConfig{
			ProjectName: "test-project",
			Language:    "go",
			Framework:   "gin",
		},
		ProjectStructure: &ProjectStructure{
			RootPath: "/project/root",
			Directories: []Directory{
				{
					Path: "/project/root/cmd",
					Name: "cmd",
					Type: "command",
				},
			},
			Files: []FileInfo{
				{
					Path:         "/project/root/main.go",
					Name:         "main.go",
					Extension:    ".go",
					Size:         1024,
					ModifiedTime: time.Now(),
					IsOpen:       true,
				},
			},
		},
		CodingStandards: &CodingStandards{
			Language:    "go",
			StyleGuide:  "effective-go",
			LintingRules: []string{"gofmt", "golint"},
		},
	}

	// Test marshaling
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal ProjectContext: %v", err)
	}

	// Test unmarshaling
	var unmarshaled ProjectContext
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ProjectContext: %v", err)
	}

	// Verify fields
	if unmarshaled.WorkingDirectory != original.WorkingDirectory {
		t.Errorf("WorkingDirectory mismatch: got %s, want %s", unmarshaled.WorkingDirectory, original.WorkingDirectory)
	}
	if unmarshaled.ProjectConfig.ProjectName != original.ProjectConfig.ProjectName {
		t.Errorf("ProjectName mismatch: got %s, want %s", unmarshaled.ProjectConfig.ProjectName, original.ProjectConfig.ProjectName)
	}
}

// TestComplexityLevel tests ComplexityLevel constants
func TestComplexityLevel(t *testing.T) {
	tests := []struct {
		level    ComplexityLevel
		expected string
	}{
		{ComplexityLow, "low"},
		{ComplexityMedium, "medium"},
		{ComplexityHigh, "high"},
	}

	for _, test := range tests {
		if string(test.level) != test.expected {
			t.Errorf("ComplexityLevel mismatch: got %s, want %s", test.level, test.expected)
		}
	}
}

// TestValidationStatus tests ValidationStatus constants
func TestValidationStatus(t *testing.T) {
	tests := []struct {
		status   ValidationStatus
		expected string
	}{
		{ValidationPassed, "passed"},
		{ValidationFailed, "failed"},
		{ValidationWarning, "warning"},
	}

	for _, test := range tests {
		if string(test.status) != test.expected {
			t.Errorf("ValidationStatus mismatch: got %s, want %s", test.status, test.expected)
		}
	}
}

// TestCodeResultSerialization tests JSON serialization for CodeResult
func TestCodeResultSerialization(t *testing.T) {
	original := &CodeResult{
		GeneratedFiles: []GeneratedFile{
			{
				Path:        "/output/main.go",
				Name:        "main.go",
				Content:     "package main\n\nfunc main() {}",
				Type:        "source",
				Description: "Main application file",
			},
		},
		Tests: &TestSuite{
			Name:        "main-tests",
			Description: "Tests for main package",
			Framework:   "testing",
		},
		Documentation: &Documentation{
			Format: "markdown",
			Files: []DocumentationFile{
				{
					Name:    "README.md",
					Content: "# Test Project",
					Type:    "readme",
				},
			},
		},
	}

	// Test marshaling
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal CodeResult: %v", err)
	}

	// Test unmarshaling
	var unmarshaled CodeResult
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal CodeResult: %v", err)
	}

	// Verify fields
	if len(unmarshaled.GeneratedFiles) != len(original.GeneratedFiles) {
		t.Errorf("GeneratedFiles length mismatch: got %d, want %d", len(unmarshaled.GeneratedFiles), len(original.GeneratedFiles))
	}
	if unmarshaled.Tests.Name != original.Tests.Name {
		t.Errorf("Test suite name mismatch: got %s, want %s", unmarshaled.Tests.Name, original.Tests.Name)
	}
}