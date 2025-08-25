package types

import (
	"context"
	"time"
)

// ReasoningEngine is the primary interface for the reasoning system
type ReasoningEngine interface {
	AnalyzeProblem(ctx context.Context, req *ProblemRequest) (*Analysis, error)
	GeneratePlan(ctx context.Context, analysis *Analysis) (*ExecutionPlan, error)
	ExecutePlan(ctx context.Context, plan *ExecutionPlan) (*CodeResult, error)
	ValidateResult(ctx context.Context, result *CodeResult) (*ValidationReport, error)
}

// ContextManager handles project context and configuration
type ContextManager interface {
	BuildProjectContext(workdir string) (*ProjectContext, error)
	LoadConfiguration(configPath string) (*GOAIConfig, error)
	WatchFileChanges(callback func(*FileChangeEvent)) error
	GetRecentChanges(since time.Time) ([]*GitChange, error)
}

// CodeGenerator handles code generation tasks
type CodeGenerator interface {
	GenerateCode(ctx context.Context, spec *CodeSpec) (*GeneratedCode, error)
	GenerateTests(ctx context.Context, code *GeneratedCode) (*TestSuite, error)
	GenerateDocumentation(ctx context.Context, code *GeneratedCode) (*Documentation, error)
}

// Validator performs validation on generated code
type Validator interface {
	StaticAnalysis(code *GeneratedCode) (*StaticReport, error)
	RunTests(testSuite *TestSuite) (*TestResults, error)
	CheckCompliance(code *GeneratedCode, standards *CodingStandards) (*ComplianceReport, error)
}