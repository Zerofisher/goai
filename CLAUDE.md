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
```

**Test reasoning chains (requires OpenAI API key):**
```bash
export OPENAI_API_KEY="your-key"
go test ./internal/reasoning/ -v -run TestEngine_AnalyzeProblem_ValidKey
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
Production-ready codebase indexing with:
- **File Discovery**: Smart filtering with .gitignore support and language detection
- **Document Chunking**: Intelligent segmentation respecting code boundaries (functions, classes)
- **FTS5 Search**: SQLite-based full-text search with BM25 ranking
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
- Comprehensive indexing system with FTS5 search

**🚧 In Progress/Placeholder:**
- CLI command implementations (skeleton exists)
- Tool system for file operations
- Vector embedding support in indexing
- Parallel processing with Eino Graphs