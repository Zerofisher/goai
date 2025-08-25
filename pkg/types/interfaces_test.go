package types

import (
	"context"
	"testing"
	"time"
)

// TestReasoningEngineInterface tests that ReasoningEngine interface is properly defined
func TestReasoningEngineInterface(t *testing.T) {
	// This test ensures the interface can be implemented
	var _ ReasoningEngine = (*mockReasoningEngine)(nil)
}

// TestContextManagerInterface tests that ContextManager interface is properly defined
func TestContextManagerInterface(t *testing.T) {
	// This test ensures the interface can be implemented
	var _ ContextManager = (*mockContextManager)(nil)
}

// TestCodeGeneratorInterface tests that CodeGenerator interface is properly defined
func TestCodeGeneratorInterface(t *testing.T) {
	// This test ensures the interface can be implemented
	var _ CodeGenerator = (*mockCodeGenerator)(nil)
}

// TestValidatorInterface tests that Validator interface is properly defined
func TestValidatorInterface(t *testing.T) {
	// This test ensures the interface can be implemented
	var _ Validator = (*mockValidator)(nil)
}

// Mock implementations for interface testing

type mockReasoningEngine struct{}

func (m *mockReasoningEngine) AnalyzeProblem(ctx context.Context, req *ProblemRequest) (*Analysis, error) {
	return &Analysis{
		ProblemDomain: "test",
		Complexity:    ComplexityLow,
	}, nil
}

func (m *mockReasoningEngine) GeneratePlan(ctx context.Context, analysis *Analysis) (*ExecutionPlan, error) {
	return &ExecutionPlan{
		Steps: []PlanStep{
			{
				ID:   "test-step",
				Name: "Test Step",
			},
		},
	}, nil
}

func (m *mockReasoningEngine) ExecutePlan(ctx context.Context, plan *ExecutionPlan) (*CodeResult, error) {
	return &CodeResult{
		GeneratedFiles: []GeneratedFile{
			{
				Name:    "test.go",
				Content: "package main",
			},
		},
	}, nil
}

func (m *mockReasoningEngine) ValidateResult(ctx context.Context, result *CodeResult) (*ValidationReport, error) {
	return &ValidationReport{
		OverallStatus: ValidationPassed,
	}, nil
}

type mockContextManager struct{}

func (m *mockContextManager) BuildProjectContext(workdir string) (*ProjectContext, error) {
	return &ProjectContext{
		WorkingDirectory: workdir,
	}, nil
}

func (m *mockContextManager) LoadConfiguration(configPath string) (*GOAIConfig, error) {
	return &GOAIConfig{
		ProjectName: "test",
		Language:    "go",
	}, nil
}

func (m *mockContextManager) WatchFileChanges(callback func(*FileChangeEvent)) error {
	return nil
}

func (m *mockContextManager) GetRecentChanges(since time.Time) ([]*GitChange, error) {
	return []*GitChange{}, nil
}

type mockCodeGenerator struct{}

func (m *mockCodeGenerator) GenerateCode(ctx context.Context, spec *CodeSpec) (*GeneratedCode, error) {
	return &GeneratedCode{
		Files: []GeneratedFile{
			{
				Name:    "generated.go",
				Content: "package main",
			},
		},
	}, nil
}

func (m *mockCodeGenerator) GenerateTests(ctx context.Context, code *GeneratedCode) (*TestSuite, error) {
	return &TestSuite{
		Name: "test-suite",
	}, nil
}

func (m *mockCodeGenerator) GenerateDocumentation(ctx context.Context, code *GeneratedCode) (*Documentation, error) {
	return &Documentation{
		Format: "markdown",
	}, nil
}

type mockValidator struct{}

func (m *mockValidator) StaticAnalysis(code *GeneratedCode) (*StaticReport, error) {
	return &StaticReport{
		OverallScore: 95.0,
	}, nil
}

func (m *mockValidator) RunTests(testSuite *TestSuite) (*TestResults, error) {
	return &TestResults{
		TotalTests:  5,
		PassedTests: 5,
		FailedTests: 0,
	}, nil
}

func (m *mockValidator) CheckCompliance(code *GeneratedCode, standards *CodingStandards) (*ComplianceReport, error) {
	return &ComplianceReport{
		Score: 98.5,
	}, nil
}