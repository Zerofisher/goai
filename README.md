# GoAI Coder

A reasoning-based programming assistant CLI tool built in Go that provides intelligent code generation, analysis, and problem-solving capabilities with a focus on context-aware assistance.

## Features

- **Reasoning-First Architecture**: Uses structured reasoning chains to understand problems deeply before generating solutions
- **Intelligent Code Analysis**: Analyzes programming problems through multi-step reasoning processes
- **Execution Planning**: Generates detailed implementation plans with dependency management
- **Code Generation**: Creates production-ready code with proper error handling and tests
- **Code Validation**: Performs comprehensive validation including static analysis and testing
- **Project Context**: Git integration, file watching, and project structure analysis
- **Codebase Indexing**: Full-text search with intelligent chunking and BM25 ranking

## Architecture

GoAI Coder is built on the [Eino framework](https://github.com/cloudwego/eino) and implements a four-chain reasoning system:

1. **Analysis Chain**: Problem domain identification and technical challenge analysis
2. **Planning Chain**: Execution plan generation with step-by-step implementation
3. **Execution Chain**: Code generation with proper structure and testing
4. **Validation Chain**: Quality assurance and compliance checking

## Installation

### Prerequisites

- Go 1.24.6 or later
- OpenAI API key (for LLM reasoning)

### Build from Source

```bash
git clone https://github.com/Zerofisher/goai.git
cd goai
go mod tidy
go build ./cmd/goai
```

## Usage

### Basic Commands

```bash
# Show help
./goai --help

# Analyze programming problems
./goai think "Create a REST API for user management"

# Generate execution plans
./goai plan analysis-result.json

# Analyze project structure
./goai analyze ./my-project

# Debug and fix issues
./goai fix "API returns 500 error on user creation"
```

### Configuration

**Required**: Set your OpenAI API key as an environment variable:

```bash
export OPENAI_API_KEY="your-api-key-here"
```

**Optional**: Configure custom OpenAI settings:

```bash
export OPENAI_BASE_URL="https://api.openai.com/v1"  # Custom API endpoint
export OPENAI_MODEL="gpt-4"                         # Custom model name
```

Create a `GOAI.md` file in your project directory with configuration:

```markdown
# Project Configuration
- **Language**: Go
- **Framework**: Gin/Echo
- **Testing**: Standard library + testify
- **Database**: PostgreSQL
```

### Example Usage

**Note**: Make sure to configure your OpenAI API key before running examples.

1. **Problem Analysis**:
   ```bash
   ./goai think "Build a microservice for processing payments"
   ```

2. **Project Analysis**:
   ```bash
   ./goai analyze ./payment-service
   ```

3. **Bug Fixing**:
   ```bash
   ./goai fix "Database connection pool exhausted"
   ```

## Testing

Run the comprehensive test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Development

### Project Structure

```
goai/
├── cmd/goai/           # CLI entry point
├── internal/reasoning/ # Eino-based reasoning chains
├── pkg/
│   ├── types/         # Core data models and interfaces
│   ├── context/       # Project context management
│   ├── indexing/      # Codebase indexing and search
│   └── errors/        # Error handling utilities
└── .kiro/specs/       # Design documents and tasks
```

### Key Components

- **Reasoning Engine**: Four-chain system for problem-solving
- **Context Manager**: Project structure analysis and Git integration
- **Indexing System**: File discovery, chunking, and full-text search
- **CLI Framework**: Cobra-based command structure

### Running Examples

Test the indexing system:
```bash
go run ./pkg/indexing/example_usage.go
```

Test the full reasoning chain:
```bash
go run ./examples/full_reasoning_chain.go
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Roadmap

- [x] CLI framework with Cobra
- [x] Core data types and interfaces
- [x] Eino-based reasoning chains
- [x] Codebase indexing and search
- [ ] Tool system integration
- [ ] Vector embeddings support
- [ ] Web UI interface
- [ ] Plugin system

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built on the [Eino framework](https://github.com/cloudwego/eino) by CloudWeGo
- Inspired by [Continue](https://continue.dev/) for code assistance patterns
- Uses [Cobra](https://github.com/spf13/cobra) for CLI framework