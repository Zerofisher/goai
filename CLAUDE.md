# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

GoAI Coder is an intelligent programming assistant CLI tool built in Go, following the **"Model as Agent"** philosophy - where the LLM is the intelligent agent and code provides simple, focused tools.

**Current Status**: Production-ready with 55.4% test coverage. Successfully refactored to mini-kode architecture.

## Project Information

- **Go Version**: 1.24.6 (specified in go.mod)
- **Main Branch**: `refactor/mini-kode-architecture`
- **Test Coverage**: 55.4% overall
- **Dependencies**: Minimal - `gopkg.in/yaml.v3` for configuration, `github.com/chzyer/readline` for interactive input

## Development Commands

### Build and Run

```bash
go build ./cmd/goai                    # Build main CLI
./goai --help                          # Show help
./goai                                 # Interactive mode (requires OPENAI_API_KEY)
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage (current: 55.4%)
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Test specific packages
go test -v ./pkg/tools/search          # Search tool (80.9% coverage)
go test -v ./pkg/agent                 # Agent (63.3% coverage)
go test -v ./pkg/todo                  # Todo system (89.4% coverage)

# Run single test
go test -v ./pkg/tools/search -run TestSearchTool

# Test with race detection
go test -race ./pkg/tools
```

### Code Quality

```bash
# Run comprehensive linter
golangci-lint run

# Auto-fix issues
golangci-lint run --fix

# Full quality check (pre-commit)
go test ./... && golangci-lint run && go build ./cmd/goai
```

## Architecture

### High-Level Structure

```
User Input → Agent (Main Loop) → LLM Client → Tool Dispatcher → Output
                                      ↓
                              Message Management
```

### Core Components

1. **Agent System** (`pkg/agent/`)
   - Main loop orchestration with **multi-round tool call support** (up to 10 rounds)
   - Message history management with token limits
   - LLM interaction coordination
   - State management

2. **Tool System** (`pkg/tools/`)
   - **Interface**: All tools implement 5 methods: `Name()`, `Description()`, `InputSchema()`, `Execute()`, `Validate()`
   - **File Operations** (`pkg/tools/file/`): Read, write, list with security validation
   - **Bash Execution** (`pkg/tools/bash/`): Safe command execution with timeout and filtering
   - **File Editing** (`pkg/tools/edit/`): Text replacement, insertion, deletion with backups
   - **Code Search** (`pkg/tools/search/`): Grep-based search with caching (80.9% coverage)
   - **Todo Management** (`pkg/tools/todo/`): Task tracking and progress monitoring
   - **Security**: Path validation, command filtering via `SecurityValidator`

3. **LLM Client** (`pkg/llm/`)
   - OpenAI integration with factory pattern
   - Support for streaming and non-streaming responses
   - Error handling and retry logic

4. **Message Management** (`pkg/message/`)
   - Message history with token limits
   - Content formatting and normalization

5. **Configuration** (`pkg/config/`)
   - YAML/JSON-based configuration
   - Environment variable support with expansion
   - Validation and defaults

6. **Todo System** (`pkg/todo/`)
   - Task management with status tracking (pending, in_progress, completed)
   - Progress rendering with colors
   - Constraints: max 20 items, 1 in-progress at a time

7. **Reminder System** (`pkg/reminder/`)
   - Periodic reminders for Todo usage
   - Non-intrusive message injection

## Key Implementation Details

### Multi-Round Tool Calls

The agent supports **multiple rounds** of tool calls per query (up to 10 rounds):
- LLM can chain tool calls: `list_files` → `read_file` → return analysis
- Each round executes tools, adds results to message history, and requests next LLM response
- Loop continues until LLM returns text response or max rounds reached
- Location: `pkg/agent/agent.go:117-143`

### Tool Registration

Tools are registered in `cmd/goai/main.go:registerTools()`:
- Each tool is created with appropriate configuration
- Registered with the dispatcher's registry
- Tool definitions automatically provided to LLM via `agent.getToolDefinitions()`

### Interactive Input

Uses `readline` library for enhanced terminal input:
- UTF-8 support
- Arrow key navigation and history
- Command history saved to `/tmp/.goai_history`
- Special commands: `/help`, `/stats`, `/clear`, `/reset`, `/exit`

### Security Model

- **Path Validation**: All file paths validated against work directory, no `..` traversal
- **Command Filtering**: Bash tool blocks dangerous commands (`rm -rf /`, `mkfs`, etc.)
- **Resource Limits**: File size (10MB), output (100K chars), timeout (30s), parallel tools (5)

## Module-Specific Notes

### Areas Needing Test Coverage Improvement

**High Priority** (below target coverage):
- `pkg/dispatcher/`: 0.0% - Critical infrastructure, needs comprehensive tests
- `pkg/tools/todo/`: 31.4% - Core functionality, target >70%
- `pkg/message/`: 30.4% - Message management needs better coverage
- `pkg/llm/`: 38.4% - LLM integration should be well-tested

**Well-Tested** (meeting/exceeding targets):
- `pkg/types/`: 97.0% ✅
- `pkg/reminder/`: 91.1% ✅
- `pkg/todo/`: 89.4% ✅
- `pkg/tools/search/`: 80.9% ✅
- `pkg/config/`: 78.9% ✅

## Common Patterns

### Adding a New Tool

1. Create file in `pkg/tools/<toolname>/`
2. Implement Tool interface (5 methods)
3. Add comprehensive tests (target >70% coverage)
4. Register in `cmd/goai/main.go:registerTools()`
5. Run quality checks: `golangci-lint run ./pkg/tools/<toolname>/`

Example tool structure:
```go
type MyTool struct {
    workDir string
}

func NewMyTool(workDir string) *MyTool {
    return &MyTool{workDir: workDir}
}

func (t *MyTool) Name() string { return "my_tool" }
func (t *MyTool) Description() string { return "Does something useful" }
func (t *MyTool) InputSchema() map[string]interface{} { /* ... */ }
func (t *MyTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) { /* ... */ }
func (t *MyTool) Validate(input map[string]interface{}) error { /* ... */ }
```

## Environment Setup

**Required for LLM features:**
```bash
export OPENAI_API_KEY="your-api-key"
export OPENAI_MODEL="gpt-4"  # Optional, defaults to gpt-4
```

**Configuration files** (optional, checked in order):
- `goai.yaml` or `goai.yml` (current directory)
- `.goai.yaml` or `.goai.yml` (current directory)
- `~/.config/goai/config.yaml`

## Important Constraints

1. **No Over-Engineering**: Keep implementations simple. The model is smart; code should just provide tools.
2. **Test First**: Use TDD approach where appropriate. Tests drive design.
3. **Security First**: All file operations must validate paths, all commands must be checked.
4. **Interface Driven**: Define interfaces in consumer packages, implementations in provider packages.
5. **Concurrent Safe**: Use mutexes for shared state, design for concurrent access.

## Code Quality Standards

**Mandatory Requirements:**
- All code must pass `golangci-lint run` with 0 issues
- Test coverage: Core packages >80%, Tool packages >70%
- All exported types, functions, methods must have documentation comments
- Error handling: Never ignore errors, use `fmt.Errorf` with `%w` for wrapping

**Error Handling Pattern:**
```go
// ✅ Good
func ReadFile(path string) ([]byte, error) {
    if err := ValidatePath(path); err != nil {
        return nil, fmt.Errorf("invalid path: %w", err)
    }
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read file %s: %w", path, err)
    }
    return data, nil
}
```

**File Close Pattern:**
```go
file, err := os.Open(path)
if err != nil {
    return err
}
defer func() {
    _ = file.Close() // Non-critical error, can be ignored with comment
}()
```

## GitHub Actions

Release workflow (`.github/workflows/release.yml`) builds for:
- Linux amd64
- Linux arm64
- macOS amd64 (Intel)
- macOS arm64 (Apple Silicon)

Triggered by pushing version tags: `git tag v0.1.0 && git push origin v0.1.0`
