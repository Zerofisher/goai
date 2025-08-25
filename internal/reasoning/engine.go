package reasoning

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"

	"github.com/Zerofisher/goai/pkg/types"
)

// Engine implements the ReasoningEngine interface using Eino framework
type Engine struct {
	analysisChain   *compose.Chain[*types.ProblemRequest, *types.Analysis]
	planningChain   *compose.Chain[*types.Analysis, *types.ExecutionPlan]
	executionChain  *compose.Chain[*types.ExecutionPlan, *types.CodeResult]
	validationChain *compose.Chain[*types.CodeResult, *types.ValidationReport]
	
	// Dependencies for enhanced reasoning
	contextManager types.ContextManager
	chatModel      *openai.ChatModel
}

// NewEngine creates a new reasoning engine with Eino chains
func NewEngine(ctx context.Context, contextManager types.ContextManager) (*Engine, error) {
	// Create OpenAI chat model
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     getEnvOrDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		APIKey:      os.Getenv("OPENAI_API_KEY"),
		Model:       getEnvOrDefault("OPENAI_MODEL_NAME", "gpt-4"),
		Temperature: &[]float32{0.7}[0],
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}

	engine := &Engine{
		contextManager: contextManager,
		chatModel:      chatModel,
	}

	// Build the reasoning chains
	if err := engine.buildAnalysisChain(ctx); err != nil {
		return nil, fmt.Errorf("failed to build analysis chain: %w", err)
	}
	if err := engine.buildPlanningChain(ctx); err != nil {
		return nil, fmt.Errorf("failed to build planning chain: %w", err)
	}
	if err := engine.buildExecutionChain(ctx); err != nil {
		return nil, fmt.Errorf("failed to build execution chain: %w", err)
	}
	if err := engine.buildValidationChain(ctx); err != nil {
		return nil, fmt.Errorf("failed to build validation chain: %w", err)
	}

	return engine, nil
}

// AnalyzeProblem analyzes a programming problem using structured reasoning
func (e *Engine) AnalyzeProblem(ctx context.Context, req *types.ProblemRequest) (*types.Analysis, error) {
	// Enhance request with relevant code context if available
	if e.contextManager != nil && req.Context != nil {
		projectContext, err := e.contextManager.BuildProjectContext(req.Context.WorkingDirectory)
		if err == nil {
			req.Context = projectContext
		}
	}

	// Execute analysis chain
	runner, err := e.analysisChain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile analysis chain: %w", err)
	}

	analysis, err := runner.Invoke(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute analysis: %w", err)
	}

	return analysis, nil
}

// GeneratePlan creates a detailed execution plan from analysis
func (e *Engine) GeneratePlan(ctx context.Context, analysis *types.Analysis) (*types.ExecutionPlan, error) {
	runner, err := e.planningChain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile planning chain: %w", err)
	}

	plan, err := runner.Invoke(ctx, analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to execute planning: %w", err)
	}

	return plan, nil
}

// ExecutePlan executes the implementation plan
func (e *Engine) ExecutePlan(ctx context.Context, plan *types.ExecutionPlan) (*types.CodeResult, error) {
	runner, err := e.executionChain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile execution chain: %w", err)
	}

	result, err := runner.Invoke(ctx, plan)
	if err != nil {
		return nil, fmt.Errorf("failed to execute plan: %w", err)
	}

	return result, nil
}

// ValidateResult validates the generated code
func (e *Engine) ValidateResult(ctx context.Context, result *types.CodeResult) (*types.ValidationReport, error) {
	runner, err := e.validationChain.Compile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to compile validation chain: %w", err)
	}

	report, err := runner.Invoke(ctx, result)
	if err != nil {
		return nil, fmt.Errorf("failed to execute validation: %w", err)
	}

	return report, nil
}

// Helper function to get environment variable with default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}