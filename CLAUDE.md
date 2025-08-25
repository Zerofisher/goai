# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoAI Coder is a reasoning-based programming assistant CLI tool built in Go. It provides intelligent code generation, analysis, and problem-solving capabilities with a focus on context-aware assistance using the Eino framework for AI reasoning chains.

## Architecture

### Core Components
- **cmd/goai**: Main CLI entry point using Cobra framework with commands: think, plan, analyze, fix
- **internal/reasoning**: Eino-based reasoning engine with four main chains (analysis, planning, execution, validation)
- **pkg/indexing**: Comprehensive codebase indexing and search system with SQLite FTS5
- **pkg/types**: Core data models and interfaces for the entire system
- **pkg/context**: Project context management, Git integration, and file watching
- **pkg/errors**: Error handling and recovery utilities with custom error types

### Key Architectural Patterns
- **Chain-based Reasoning**: Uses Eino framework for sequential AI reasoning operations
- **Interface-driven Design**: Core functionality defined through interfaces (ReasoningEngine, ContextManager, CodeGenerator, Validator)
- **Modular Indexing**: Pluggable indexing system supporting multiple backends (currently FTS5, designed for vector embeddings)
- **Context-Aware Processing**: Deep integration with project context, Git history, and file changes

## Development Commands

**Build:**
```bash
go build ./cmd/goai                    # Build main CLI
go build ./cmd/indexing-example        # Build indexing system demo
```

**Run tests:**
```bash
go test ./...                          # Run all tests (note: some may fail without valid test setup)
go test -v ./pkg/indexing              # Run indexing system tests
go test -v ./internal/reasoning        # Run reasoning engine tests
go test ./pkg/context ./pkg/types ./pkg/errors  # Run core package tests
```

**Test with coverage:**
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

**Run CLI and examples:**
```bash
./goai --help                          # Show help
go run ./cmd/goai think "Create a REST API"  # Problem analysis (placeholder)
go run ./cmd/indexing-example          # Demo indexing system
go run ./examples/full_reasoning_chain.go  # Complete reasoning chain demo
```

**Test reasoning chains (requires OpenAI API key):**
```bash
export OPENAI_API_KEY="your-key"
export OPENAI_MODEL_NAME="gpt-4"       # Optional, defaults to gpt-4
go test ./internal/reasoning/ -v -run TestEngine_AnalyzeProblem_ValidKey
```

**Code Quality Checks:**
```bash
golangci-lint run                      # Run all linters (comprehensive)
golangci-lint run ./pkg/indexing       # Run linters on specific package
golangci-lint run --fix                # Auto-fix issues where possible
go vet ./...                           # Basic static analysis
```

## Core Systems

### Reasoning Engine (internal/reasoning)
Four-chain Eino-based system for AI-powered code assistance:
- **Analysis Chain**: Problem domain analysis, technical stack identification, risk assessment
- **Planning Chain**: Detailed execution plans with dependencies and timelines  
- **Execution Chain**: Working Go code generation with tests and documentation
- **Validation Chain**: Code quality analysis and compliance checking

All chains include fallback mechanisms for when LLM responses can't be parsed as JSON.

### Indexing System (pkg/indexing)  
Complete multi-modal codebase indexing and search system:
- **File Discovery**: Smart filtering with .gitignore support and language detection
- **Document Chunking**: Intelligent segmentation respecting code boundaries (functions, classes)
- **Full-Text Search**: SQLite FTS5 with BM25 ranking (`fts_index.go`)
- **Symbol Indexing**: Go AST parser for functions, types, variables (`symbol_index.go`, `treesitter_parser.go`)
- **Semantic Search**: Vector embeddings with OpenAI integration (`embedding_index.go`, `embedding_provider.go`)
- **Specialized Retrievers**: FTS, semantic, symbol, recent files retrievers (`retrievers.go`)
- **Hybrid Pipeline**: Parallel execution with intelligent reranking (`reranker.go`)
- **Enhanced Manager**: Unified interface for all search capabilities (`enhanced_manager.go`)
- **Index Management**: Concurrent operations, incremental updates, status monitoring

### Context Management (pkg/context)
Project-aware context system:
- **Git Integration**: Branch tracking, commit history, change detection
- **File Watching**: Real-time file system monitoring
- **Project Structure Analysis**: Dependency mapping and architecture evaluation
- **Configuration Loading**: GOAI.md file parsing for project-specific settings

## Environment Requirements

**Required for reasoning chains:**
```bash
export OPENAI_API_KEY="your-api-key"
export OPENAI_MODEL_NAME="gpt-4"  # Optional, defaults to gpt-4
```

**Dependencies:**
- Go 1.24.6+
- SQLite (for indexing)
- OpenAI API access (for reasoning)

## Current Implementation Status

**✅ Complete:**
- Project foundation and core interfaces
- Data models and validation
- Context management system  
- Eino-based reasoning chains with OpenAI integration
- **Complete multi-modal indexing system**:
  - Full-text search with SQLite FTS5
  - Symbol indexing with Go AST parser  
  - Vector embeddings with OpenAI integration
  - Specialized retrievers (FTS, semantic, symbol, recent files)
  - Hybrid retrieval pipeline with intelligent reranking
  - Enhanced index manager with unified API

**🚧 In Progress/Placeholder:**
- CLI command implementations (skeleton exists) 
- Tool system for file operations
- Vector embedding support in indexing
- Parallel processing with Eino Graphs

## Development Workflow

**Full Quality Check Pipeline:**
```bash
# 1. Run tests with coverage
go test -coverprofile=coverage.out ./...

# 2. Check code quality
golangci-lint run

# 3. Verify build
go build ./cmd/goai

# 4. Generate coverage report (optional)
go tool cover -html=coverage.out -o coverage.html
```

**Pre-commit Quality Gate:**
```bash
# Quick quality check before committing
go test ./... && golangci-lint run && go build ./cmd/goai
```

## Development Tips

**Key Interface Locations:**
- Core interfaces defined in `pkg/types/interfaces.go`
- Reasoning engine interface: `ReasoningEngine` with 4-chain methods
- Context management: `ContextManager` for project context and Git integration
- Indexing: `CodebaseIndexer` for file discovery and search

**Important Implementation Details:**
- Context manager creates `.goai/` directory with multiple index databases:
  - `fts_index.db`: SQLite FTS5 for full-text search
  - `symbol_index.db`: SQLite for symbol indexing  
  - `embedding_index.db`: SQLite for vector embeddings
- Enhanced index manager (`EnhancedIndexManager`) provides unified access to all search types
- Reasoning chains include JSON parsing fallbacks for robustness
- File filtering respects `.gitignore` patterns automatically

**Code Quality Standards:**
- All code must pass `golangci-lint run` with 0 issues
- Maintain test coverage above 60% for core packages (`pkg/indexing`, `pkg/context`, `pkg/types`)
- Use proper error handling with `defer func() { _ = resource.Close() }()` pattern for non-critical errors
- Follow Go best practices: interfaces in consumer packages, concrete types in provider packages
- All indexes support concurrent operations with proper mutex locking
- Vector embeddings use Eino's OpenAI integration with fallback to mock provider
- Hybrid retrieval combines multiple retrievers with intelligent reranking

**Running Individual Components:**
```bash
# Test enhanced indexing system
go run ./cmd/indexing-example

# Test all indexing functionality
go test ./pkg/indexing -v

# Test specific indexing components
go test ./pkg/indexing -v -run TestEnhancedIndexManager
go test ./pkg/indexing -v -run TestEmbeddingProvider
go test ./pkg/indexing -v -run TestRetrievers

# Run single test file
go test ./pkg/context -v -run TestContextManager

# Test with race detection
go test -race ./pkg/indexing
```