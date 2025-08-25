# Examples

This directory contains example programs demonstrating various features of GoAI Coder.

## Full Reasoning Chain Example

**File**: `full_reasoning_chain.go`

A comprehensive integration test that demonstrates the complete four-chain reasoning system:
1. **Analysis Chain**: Problem domain identification and technical challenge analysis
2. **Planning Chain**: Execution plan generation with step-by-step implementation  
3. **Execution Chain**: Code generation with proper structure and testing
4. **Validation Chain**: Quality assurance and compliance checking

**Usage**:
```bash
export OPENAI_API_KEY="your-api-key-here"
go run ./examples/full_reasoning_chain.go
```

**What it does**:
- Creates a mock context manager for testing
- Analyzes the problem: "Create a simple HTTP server in Go that handles GET and POST requests"
- Generates a detailed execution plan
- Executes the plan to generate code
- Validates the generated code for quality and correctness
- Displays a comprehensive summary of the entire reasoning process

This example shows how all components work together and serves as both a test and demonstration of the reasoning capabilities.