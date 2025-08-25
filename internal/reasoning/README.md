# GoAI Reasoning Engine

This package implements the core reasoning engine for GoAI using the Eino framework. It provides structured, AI-powered problem analysis, execution planning, code generation, and validation capabilities.

## Architecture

The reasoning engine consists of four main chains:

1. **Analysis Chain** (`analysis_chain.go`) - Analyzes programming problems and provides structured technical analysis
2. **Planning Chain** (`planning_chain.go`) - Creates detailed execution plans with steps, dependencies, and timelines  
3. **Execution Chain** (`execution_chain.go`) - Generates working Go code based on execution plans
4. **Validation Chain** (`validation_chain.go`) - Validates generated code for quality, correctness, and compliance

## Usage

```go
package main

import (
    "context"
    "github.com/lesion/goai/internal/reasoning"
    "github.com/lesion/goai/pkg/context"
    "github.com/lesion/goai/pkg/types"
)

func main() {
    ctx := context.Background()
    
    // Create context manager
    contextMgr, _ := context.NewContextManager("/path/to/project")
    
    // Create reasoning engine  
    engine, err := reasoning.NewEngine(ctx, contextMgr)
    if err != nil {
        panic(err)
    }
    
    // Analyze problem
    problem := &types.ProblemRequest{
        Description: "Create a REST API server in Go",
        Requirements: []string{
            "Handle GET and POST requests",
            "Include proper error handling", 
            "Use JSON for data exchange",
        },
        Constraints: []string{
            "Use only standard library",
            "Include comprehensive tests",
        },
    }
    
    analysis, err := engine.AnalyzeProblem(ctx, problem)
    if err != nil {
        panic(err)
    }
    
    // Generate execution plan
    plan, err := engine.GeneratePlan(ctx, analysis)
    if err != nil {
        panic(err)
    }
    
    // Execute plan to generate code
    codeResult, err := engine.ExecutePlan(ctx, plan)
    if err != nil {
        panic(err)
    }
    
    // Validate generated code
    validation, err := engine.ValidateResult(ctx, codeResult)
    if err != nil {
        panic(err)
    }
    
    // Process results...
}
```

## Configuration

The reasoning engine requires an OpenAI API key to function:

```bash
export OPENAI_API_KEY="your-api-key-here"
export OPENAI_MODEL_NAME="gpt-4"  # Optional, defaults to gpt-4
export OPENAI_BASE_URL="https://api.openai.com/v1"  # Optional
```

## Features

- **Structured Analysis**: Breaks down problems into domain, technical stack, architecture patterns, risks, and recommendations
- **Detailed Planning**: Creates step-by-step execution plans with dependencies, timelines, and testing strategies
- **Code Generation**: Produces working Go code with tests and documentation
- **Quality Validation**: Performs static analysis, compliance checking, and quality assessment
- **Error Handling**: Comprehensive error handling with fallback mechanisms
- **Extensible**: Built on Eino framework for easy extension and customization

## Testing

Run the test suite:

```bash
go test ./internal/reasoning/ -v
```

Note: Some tests require an OpenAI API key to be set. Tests without API keys will skip integration tests and run unit tests only.

## Integration

The reasoning engine integrates with:

- **Context Management System**: Provides project context and configuration
- **Tool System** (future): Will integrate with file operations and system tools  
- **Indexing System** (future): Will use code search and semantic understanding
- **CLI Interface**: Exposed through `goai think`, `goai plan`, etc. commands

## Error Handling

The engine includes sophisticated error handling:

- **Fallback Parsing**: When LLM responses can't be parsed as JSON, creates reasonable fallback structures
- **Validation**: Input validation at each stage prevents invalid data from propagating
- **Graceful Degradation**: Partial failures are handled gracefully with meaningful error messages
- **Timeout Handling**: Built-in timeouts prevent hanging on API calls

## Performance Considerations

- **Parallel Processing**: Future Graph implementation will enable parallel execution
- **Caching**: LLM responses can be cached to improve performance
- **Incremental Updates**: Context updates are incremental to minimize processing time
- **Resource Management**: Proper context handling and resource cleanup