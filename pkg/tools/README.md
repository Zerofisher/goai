# GoAI Tools System

A Continue-inspired tool system for GoAI that provides a comprehensive set of tools for file operations, code search, and system interactions.

## Overview

The tools system is designed to provide programmatic access to common development operations through a unified interface. It includes:

- **File Operations**: Read, write, edit files with preview and confirmation capabilities
- **Code Search**: Search through codebases using the indexing system
- **System Interactions**: Execute commands, fetch URLs, Git operations
- **Preview & Confirmation**: Safe execution with user confirmation for destructive operations

## Architecture

### Core Components

- **Tool Interface**: Base interface that all tools implement
- **ToolManager**: Manages tool registration, execution, and validation
- **ToolRegistry**: Registry for tool discovery and lookup
- **ConfirmationHandler**: Handles user confirmations for dangerous operations
- **ToolFactory**: Creates pre-configured tool managers

### Tool Categories

1. **File Tools** (`file` category)
   - `readFile`: Read file contents
   - `writeFile`: Write content to files
   - `editFile`: Perform targeted text replacements
   - `multiEdit`: Perform multiple edits in one operation

2. **Search Tools** (`search` category)
   - `searchCode`: Search codebase using various methods
   - `listFiles`: List and filter files in directories
   - `viewDiff`: View differences between files

3. **System Tools** (`system` category)
   - `runCommand`: Execute system commands
   - `fetch`: Fetch content from URLs
   - `git`: Git operations (status, diff, log, etc.)

## Usage Examples

### Basic Setup

```go
package main

import (
    "context"
    "github.com/Zerofisher/goai/pkg/tools"
)

func main() {
    // Create tool manager
    factory := tools.NewToolFactory()
    manager, err := factory.CreateDefaultToolManager(nil)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    
    // Use tools...
}
```

### File Operations

```go
// Write a file
writeParams := map[string]any{
    "path":    "example.txt",
    "content": "Hello, World!",
}
result, err := manager.ExecuteTool(ctx, "writeFile", writeParams)

// Read a file
readParams := map[string]any{
    "path": "example.txt",
}
result, err := manager.ExecuteTool(ctx, "readFile", readParams)

// Edit a file
editParams := map[string]any{
    "path":       "example.txt",
    "oldContent": "Hello",
    "newContent": "Hi",
}
result, err := manager.ExecuteTool(ctx, "editFile", editParams)
```

### Search Operations

```go
// List files
listParams := map[string]any{
    "path":      "/project",
    "pattern":   "*.go",
    "recursive": true,
    "maxResults": 50,
}
result, err := manager.ExecuteTool(ctx, "listFiles", listParams)

// Search code (requires indexing system)
searchParams := map[string]any{
    "query":      "function main",
    "type":       "hybrid",
    "maxResults": 10,
}
result, err := manager.ExecuteTool(ctx, "searchCode", searchParams)
```

### System Operations

```go
// Run command
cmdParams := map[string]any{
    "command": "echo",
    "args":    []string{"Hello, World!"},
    "timeout": 10,
}
result, err := manager.ExecuteTool(ctx, "runCommand", cmdParams)

// Git operations
gitParams := map[string]any{
    "operation": "status",
    "workingDir": "/project",
}
result, err := manager.ExecuteTool(ctx, "git", gitParams)
```

### Preview Mode

Tools that require confirmation support preview mode:

```go
// Get a preview without executing
previewParams := map[string]any{
    "path":    "important.txt",
    "content": "New content",
}
preview, err := manager.ExecuteWithPreview(ctx, "writeFile", previewParams)

fmt.Printf("Operation: %s\n", preview.Description)
fmt.Printf("Requires confirmation: %v\n", preview.RequiresConfirmation)
for _, change := range preview.ExpectedChanges {
    fmt.Printf("Change: [%s] %s -> %s\n", change.Type, change.Description, change.Target)
}
```

## Tool Parameters

Each tool defines its parameter schema using `ParameterSchema`:

```go
type ParameterSchema struct {
    Required   []string                       `json:"required"`
    Properties map[string]ParameterProperty   `json:"properties"`
}

type ParameterProperty struct {
    Type        string      `json:"type"` // "string", "number", "boolean", "array", "object"
    Description string      `json:"description"`
    Default     any         `json:"default,omitempty"`
    Enum        []string    `json:"enum,omitempty"`
    Format      string      `json:"format,omitempty"` // "file-path", "url", etc.
}
```

## Error Handling

All tools return a `ToolResult` structure:

```go
type ToolResult struct {
    Success       bool                   `json:"success"`
    Data          any                   `json:"data,omitempty"`
    Error         string                `json:"error,omitempty"`
    Output        string                `json:"output,omitempty"`
    Metadata      map[string]any `json:"metadata,omitempty"`
    ModifiedFiles []string              `json:"modified_files,omitempty"`
}
```

## Safety Features

- **Parameter Validation**: All parameters are validated against the tool's schema
- **Confirmation System**: Dangerous operations require user confirmation
- **Preview Mode**: See what operations will do before executing them
- **Timeout Support**: Commands can be configured with timeouts
- **Error Recovery**: Comprehensive error handling and reporting

## Extending the System

To create a custom tool:

```go
type MyCustomTool struct{}

func (t *MyCustomTool) Name() string {
    return "myTool"
}

func (t *MyCustomTool) Description() string {
    return "Does something custom"
}

func (t *MyCustomTool) Parameters() ParameterSchema {
    return ParameterSchema{
        Required: []string{"input"},
        Properties: map[string]ParameterProperty{
            "input": {
                Type:        "string",
                Description: "Input parameter",
            },
        },
    }
}

func (t *MyCustomTool) Execute(ctx context.Context, params map[string]any) (*ToolResult, error) {
    input := params["input"].(string)
    // ... implement logic
    
    return &ToolResult{
        Success: true,
        Data:    "result",
        Output:  "Operation completed",
    }, nil
}

func (t *MyCustomTool) RequiresConfirmation() bool {
    return false
}

func (t *MyCustomTool) Category() string {
    return "custom"
}

// Register the tool
manager.RegisterTool(&MyCustomTool{})
```

## Testing

The tools system includes comprehensive tests:

```bash
go test ./pkg/tools -v
```

Run the example program:

```bash
go run ./cmd/tools-example
```

## Integration

The tools system is designed to integrate with:

- GoAI's indexing system for code search capabilities
- GoAI's reasoning chains for intelligent tool selection and parameter generation
- GoAI's context management for project-aware operations

## Task 4.5 Implementation Status ✅

This implements task 4.5 from the project plan:

- [x] Create tool manager interface and registry system
- [x] Implement file operation tools (readFile, writeFile, edit, multiEdit)  
- [x] Build code search tools (searchCode, listFiles, viewDiff)
- [x] Add system interaction tools (runCommand, fetch)
- [x] Implement tool preprocessing and preview mechanism

The system provides a foundation for Continue-inspired tool functionality with safety features, extensibility, and comprehensive testing.