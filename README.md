# GoAI Coder

[![Tests](https://github.com/Zerofisher/goai/actions/workflows/test.yml/badge.svg)](https://github.com/Zerofisher/goai/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/Zerofisher/goai/branch/main/graph/badge.svg)](https://codecov.io/gh/Zerofisher/goai)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Zerofisher/goai)](https://github.com/Zerofisher/goai)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

An intelligent programming assistant CLI tool built in Go, following the **"Model as Agent"** philosophy - where the LLM is the intelligent agent and code provides simple, focused tools.

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────┐
│                   User Input                    │
└──────────────────────┬──────────────────────────┘
                       ▼
┌─────────────────────────────────────────────────┐
│              Main Loop (Agent)                  │
│  ┌─────────────────────────────────────────┐    │
│  │      Message History Management         │    │
│  └─────────────────────────────────────────┘    │
│  ┌─────────────────────────────────────────┐    │
│  │      System Prompt & Configuration      │    │
│  └─────────────────────────────────────────┘    │
└──────────────────────┬──────────────────────────┘
                       ▼
┌─────────────────────────────────────────────────┐
│            LLM Client (OpenAI/Claude)           │
└──────────────────────┬──────────────────────────┘
                       ▼
┌─────────────────────────────────────────────────┐
│              Tool Dispatcher                    │
│  ┌──────────┐ ┌──────────┐ ┌──────────────┐     │
│  │   Bash   │ │   File   │ │    Search    │     │
│  │   Tool   │ │   Tools  │ │     Tool     │     │
│  └──────────┘ └──────────┘ └──────────────┘     │
│  ┌──────────┐ ┌──────────┐ ┌──────────────┐     │
│  │   Edit   │ │   Todo   │ │   Security   │     │
│  │   Tool   │ │   Tool   │ │   Validator  │     │
│  └──────────┘ └──────────┘ └──────────────┘     │
└──────────────────────┬──────────────────────────┘
                       ▼
┌─────────────────────────────────────────────────┐
│          Output Formatting & Display            │
└─────────────────────────────────────────────────┘
```

### Core Components

1. **Agent System** (`pkg/agent/`)

   - Main loop orchestration
   - Message history management
   - LLM interaction coordination
   - State management

2. **Tool System** (`pkg/tools/`)

   - **File Operations** (`pkg/tools/file/`): Read, write, list files with security validation
   - **Bash Execution** (`pkg/tools/bash/`): Safe command execution with timeout and filtering
   - **File Editing** (`pkg/tools/edit/`): Text replacement, insertion, deletion with backup
   - **Code Search** (`pkg/tools/search/`): Grep-based code and symbol search with caching
   - **Todo Management** (`pkg/tools/todo/`): Task tracking and progress monitoring
   - **Security**: Path validation, command filtering, permission system

3. **LLM Client** (`pkg/llm/`)

   - **OpenAI Integration** (`pkg/llm/openai/`): Official OpenAI SDK integration
   - **Anthropic Integration** (`pkg/llm/anthropic/`): Official Anthropic SDK integration
   - **Mock Client** (`pkg/llm/mock/`): Testing and development support
   - Factory pattern for extensible provider support
   - Streaming and non-streaming responses
   - Tool calling (function calling) support
   - Error handling and retry logic

4. **Message Management** (`pkg/message/`)

   - Message history with token limits
   - Content formatting (Markdown, code highlighting)
   - Message normalization

5. **Configuration** (`pkg/config/`)

   - YAML-based configuration
   - Environment variable support
   - Validation and defaults

6. **Todo System** (`pkg/todo/`)

   - Task management with status tracking
   - Progress rendering with colors
   - Constraints validation (max 20 items, 1 in-progress)

7. **Reminder System** (`pkg/reminder/`)
   - Periodic reminders for Todo usage
   - Non-intrusive message injection

## Installation

### Prerequisites

- Go 1.24.6 or later
- API key for your chosen LLM provider:
  - OpenAI API key (for GPT models)
  - Anthropic API key (for Claude models)

### Build from Source

```bash
git clone https://github.com/Zerofisher/goai.git
cd goai
go mod tidy
go build ./cmd/goai
```

## Usage

### Quick Start

1. **Set up your API key** (required for LLM features):

For OpenAI:

```bash
export OPENAI_API_KEY="your-api-key-here"
```

For Anthropic:

```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

2. **Build the project**:

```bash
go build ./cmd/goai
```

3. **Run GoAI Coder**:

```bash
./goai
```

### Interactive Mode

When you start GoAI Coder, you'll see an interactive prompt where you can ask the AI to help with programming tasks:

```
============================================================
GoAI Coder 0.1.0
============================================================

Welcome to GoAI Coder - Your intelligent programming assistant

Available commands:
  /help    - Show this help message
  /clear   - Clear the conversation
  /stats   - Show agent statistics
  /reset   - Reset the agent state
  /exit    - Exit the application

Type your query or command and press Enter.
------------------------------------------------------------

>
```

### TUI (Bubble Tea)

GoAI includes a full-screen TUI built with Bubble Tea. It shows assistant replies and real-time tool events side by side.

- Run
  - Default: `./goai` launches the TUI.
  - Fallback legacy prompt: set `GOAI_LEGACY_UI=1` to use the old readline interface.

- Layout
  - Left panel: conversation (assistant/user), with streaming updates.
  - Right panel: tool events and outputs (started/succeeded/failed with truncated output).
  - Bottom: input line; top/right: spinner status (e.g., “Running bash…”, “Thinking…”).

- Keys
  - `Enter`: send the current input
  - `Esc`: cancel the current request
  - `Ctrl+C`: quit

- Tool Events
  - Tools executed by the agent emit events in real time:
    - started: tool name and sanitized arguments
    - succeeded: duration + output (long output truncated)
    - failed: error message
  - Events do not alter conversation history; they are for visibility only.

- Configuration (optional)
  - Configure model/provider via `goai.yaml` or env vars (see below).
  - Tool event output length is limited internally to keep the UI responsive.
  - To force the legacy prompt temporarily: `GOAI_LEGACY_UI=1 ./goai`.

### Usage Examples

#### Example 1: Create a Simple Program

```
> please create a simple Go program that prints "Hello, World!" to the console and save it as hello.go

[AI creates hello.go file]
File created: hello.go

created file: hello.go

package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

#### Example 2: Create and Run a Program

```
> please create a Go program that calculates the Fibonacci sequence up to 10 numbers, save it as fib.go, and run it

[AI creates fib.go and runs it]
File created: fib.go
Running: go run fib.go
Output: [0 1 1 2 3 5 8 13 21 34]
```

#### Example 3: Search and Modify Code

```
> please search for all TODO comments in the codebase

[AI searches using grep-based search tool]
Found 5 TODO comments in the codebase...

> please help me implement the first TODO

[AI implements the TODO and updates the file]
```

### Available Tools

GoAI Coder has access to the following tools to help with your development tasks:

- **bash**: Execute shell commands safely (with timeout and filtering)
- **read_file**: Read file contents
- **write_file**: Create or overwrite files
- **list_files**: List directory contents
- **edit_file**: Make precise edits to existing files
- **search**: Search code and symbols using grep
- **todo**: Manage task lists for complex operations

### Special Commands

Inside the interactive prompt, you can use these commands:

- `/help` or `/h` - Show help message
- `/clear` or `/c` - Clear conversation history
- `/stats` or `/s` - Show agent statistics (messages, tokens, tool calls)
- `/reset` or `/r` - Reset the agent state
- `/exit` or `/quit` - Exit the application

### Configuration

You can customize GoAI Coder's behavior through environment variables:

```bash
# Required
export OPENAI_API_KEY="your-api-key"

# Optional
export OPENAI_MODEL="gpt-4"                          # Model to use (default: gpt-4)
export OPENAI_BASE_URL="https://api.openai.com/v1"  # API endpoint (default: OpenAI)
```

For advanced configuration, create a `goai.yaml` file in your project directory or `~/.config/goai/config.yaml`:

**OpenAI Configuration:**

```yaml
model:
  provider: "openai"
  name: "gpt-4.1-mini" # or "gpt-4", "gpt-3.5-turbo", etc.
  max_tokens: 16000
  timeout: 60

tools:
  enabled:
    - bash
    - file
    - edit
    - search
    - todo

  bash:
    timeout_ms: 30000
    max_output_chars: 100000
    forbidden_commands:
      - "rm -rf /"
      - "mkfs"

output:
  format: "markdown"
  colors: true
  show_spinner: true
```

**Anthropic Configuration:**

```yaml
model:
  provider: "anthropic"
  name: "claude-3-7-sonnet-latest" # or "claude-3-opus-latest", etc.
  max_tokens: 16000
  timeout: 60

tools:
  enabled:
    - bash
    - file
    - edit
    - search
    - todo

output:
  format: "markdown"
  colors: true
  show_spinner: true
```

### Tips for Best Results

1. **Be specific**: Clearly describe what you want the AI to do
2. **Provide context**: Mention file names, paths, and requirements
3. **Use /stats**: Monitor token usage and conversation history
4. **Break down complex tasks**: For large tasks, ask the AI to create a todo list first
5. **Review changes**: Always review generated code before using it in production

## Testing

### Run Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Test specific packages
go test -v ./pkg/tools/search    # Search tool tests
go test -v ./pkg/agent           # Agent tests
go test -v ./pkg/tools/...       # All tool tests

# Run with race detection
go test -race ./pkg/tools

# Run single test
go test -v ./pkg/tools/search -run TestSearchTool
```

### Code Quality

```bash
# Run comprehensive linter
golangci-lint run

# Auto-fix issues
golangci-lint run --fix

# Basic static analysis
go vet ./...
```

## Development

### Project Structure

```
goai/
├── cmd/goai/              # CLI entry point
│   ├── main.go           # Application setup
│   ├── interactive.go    # Interactive loop
│   └── spinner.go        # Loading animations
├── pkg/
│   ├── agent/            # Agent core logic
│   ├── config/           # Configuration system
│   ├── dispatcher/       # Tool dispatcher
│   ├── llm/              # LLM client interface
│   ├── message/          # Message management
│   ├── reminder/         # System reminders
│   ├── todo/             # Todo management
│   ├── tools/            # Tool implementations
│   │   ├── bash/         # Command execution
│   │   ├── edit/         # File editing
│   │   ├── file/         # File operations
│   │   ├── search/       # Code search (grep-based)
│   │   └── todo/         # Todo tool
│   └── types/            # Core data structures
```

## Acknowledgments

This project was inspired by and references the design of [mini_kode](https://github.com/shareAI-lab/mini_kode), an excellent project that demonstrates the "Model as Agent" philosophy in AI-powered development tools.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
