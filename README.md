# GoAI Coder

[![Tests](https://github.com/Zerofisher/goai/actions/workflows/test.yml/badge.svg)](https://github.com/Zerofisher/goai/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/Zerofisher/goai/branch/main/graph/badge.svg)](https://codecov.io/gh/Zerofisher/goai)
[![Go Report Card](https://goreportcard.com/badge/github.com/Zerofisher/goai)](https://goreportcard.com/report/github.com/Zerofisher/goai)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Zerofisher/goai)](https://github.com/Zerofisher/goai)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A reasoning-based programming assistant CLI tool built in Go that provides intelligent code generation, analysis, and problem-solving capabilities with a focus on context-aware assistance.

## Features

- **Reasoning-First Architecture**: Uses structured reasoning chains to understand problems deeply before generating solutions
- **Intelligent Code Analysis**: Analyzes programming problems through multi-step reasoning processes
- **Execution Planning**: Generates detailed implementation plans with dependency management
- **Code Generation**: Creates production-ready code with proper error handling and tests
- **Code Validation**: Performs comprehensive validation including static analysis and testing
- **Project Context**: Git integration, file watching, and project structure analysis
- **Advanced Codebase Indexing**: Multi-modal search system with:
  - **Full-Text Search**: SQLite FTS5 with BM25 ranking for keyword matching
  - **Symbol Search**: Go AST-based symbol indexing for functions, types, and variables
  - **Semantic Search**: Vector embeddings with OpenAI integration for meaning-based retrieval
  - **Hybrid Retrieval**: Parallel multi-retriever system with intelligent reranking
  - **Recent Files**: Git-aware recent file tracking and context prioritization

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

## Code Quality

Ensure code quality with linting and static analysis:

```bash
# Run golangci-lint (recommended)
golangci-lint run

# Run specific linters
golangci-lint run --disable-all --enable=errcheck,staticcheck,gosec

# Run linter on specific package
golangci-lint run ./pkg/indexing

# Auto-fix some issues
golangci-lint run --fix

# Run go vet for basic static analysis
go vet ./...
```

Install golangci-lint if not already available:
```bash
# macOS
brew install golangci-lint

# Linux/Windows
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
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

- **Reasoning Engine**: Four-chain system for problem-solving (Analysis → Planning → Execution → Validation)
- **Context Manager**: Project structure analysis, Git integration, and intelligent file watching
- **Enhanced Indexing System**: Multi-modal search with:
  - File discovery and intelligent chunking
  - SQLite FTS5 full-text search with BM25 ranking
  - Go AST symbol parsing and indexing
  - OpenAI vector embeddings for semantic search
  - Hybrid retrieval with parallel execution and reranking
  - Incremental updates and real-time synchronization
- **CLI Framework**: Cobra-based command structure with comprehensive help system

### Running Examples

Test the indexing system:
```bash
go run ./cmd/indexing-example
```

Test individual indexing components:
```bash
# Run enhanced indexing tests
go test ./pkg/indexing -v -run TestEnhancedIndexManager

# Test embedding functionality
go test ./pkg/indexing -v -run TestEmbeddingProvider

# Test retrieval system
go test ./pkg/indexing -v -run TestRetrievers
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
- [x] **Complete Codebase Indexing and Search System**:
  - [x] Full-text search with SQLite FTS5
  - [x] Symbol indexing with Go AST parser
  - [x] Vector embeddings with OpenAI integration
  - [x] Hybrid retrieval and intelligent reranking
  - [x] Enhanced index manager with all search types
- [ ] Tool system integration
- [ ] Parallel processing with Eino Graphs
- [ ] Web UI interface
- [ ] Plugin system

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built on the [Eino framework](https://github.com/cloudwego/eino) by CloudWeGo
- Inspired by [Continue](https://continue.dev/) for code assistance patterns
- Uses [Cobra](https://github.com/spf13/cobra) for CLI framework