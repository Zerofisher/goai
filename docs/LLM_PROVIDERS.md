# LLM Provider Integration

This document describes the LLM provider integration in GoAI Coder.

## Supported Providers

GoAI Coder supports multiple LLM providers through a unified interface:

- **OpenAI** (GPT models)
- **Anthropic** (Claude models)

## Architecture

The LLM integration follows a factory pattern with official SDK wrappers:

```
┌─────────────────────────────────────────┐
│          llm.Client Interface           │
│  - CreateMessage()                      │
│  - StreamMessage()                      │
│  - GetModel() / SetModel()              │
│  - IsAvailable()                        │
│  - Provider()                           │
│  - Close()                              │
└─────────────────┬───────────────────────┘
                  │
      ┌───────────┴───────────┐
      ▼                       ▼
┌─────────────┐         ┌─────────────┐
│   OpenAI    │         │  Anthropic  │
│   Client    │         │   Client    │
│             │         │             │
│ (Official   │         │ (Official   │
│  SDK)       │         │  SDK)       │
└─────────────┘         └─────────────┘
```

## OpenAI Provider

### Features

- Official OpenAI Go SDK (github.com/openai/openai-go/v2)
- Support for GPT-4, GPT-3.5, and other models
- Streaming and non-streaming responses
- Tool calling (function calling)
- Full control over parameters (temperature, top_p, etc.)

### Configuration

**Environment Variables:**

```bash
export OPENAI_API_KEY="your-api-key"
export OPENAI_MODEL="gpt-4.1-mini"  # Default model
```

**Configuration File (goai.yaml):**

```yaml
model:
  provider: "openai"
  api_key: "${OPENAI_API_KEY}" # Can use env variable
  name: "gpt-4.1-mini"
  base_url: "https://api.openai.com/v1" # Optional
  max_tokens: 16000
  timeout: 60
```

### Available Models

- `gpt-4.1-mini` (default)
- `gpt-4`
- `gpt-4-turbo`
- `gpt-3.5-turbo`
- Any other OpenAI model

### Usage Example

```go
import (
    "github.com/Zerofisher/goai/pkg/llm"
    _ "github.com/Zerofisher/goai/pkg/llm/openai"
)

config := llm.ClientConfig{
    Provider: "openai",
    APIKey:   os.Getenv("OPENAI_API_KEY"),
    Model:    "gpt-4.1-mini",
    Timeout:  30 * time.Second,
}

client, err := llm.CreateClient(config)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

resp, err := client.CreateMessage(ctx, llm.MessageRequest{
    Messages: []types.Message{
        {
            Role: "user",
            Content: []types.Content{
                {Type: "text", Text: "Hello!"},
            },
        },
    },
    MaxTokens: 100,
})
```

## Anthropic Provider

### Features

- Official Anthropic Go SDK (github.com/anthropics/anthropic-sdk-go)
- Support for Claude 3.7 Sonnet, Claude 3 Opus, and other models
- Streaming and non-streaming responses
- Tool calling (function calling)
- System prompts support

### Configuration

**Environment Variables:**

```bash
export ANTHROPIC_API_KEY="your-api-key"
```

**Configuration File (goai.yaml):**

```yaml
model:
  provider: "anthropic"
  api_key: "${ANTHROPIC_API_KEY}" # Can use env variable
  name: "claude-3-7-sonnet-latest"
  base_url: "https://api.anthropic.com" # Optional
  max_tokens: 16000
  timeout: 60
```

### Available Models

- `claude-3-7-sonnet-latest` (default)
- `claude-3-opus-latest`
- `claude-3-sonnet-20240229`
- `claude-3-haiku-20240307`
- Any other Anthropic model

### Usage Example

```go
import (
    "github.com/Zerofisher/goai/pkg/llm"
    _ "github.com/Zerofisher/goai/pkg/llm/anthropic"
)

config := llm.ClientConfig{
    Provider: "anthropic",
    APIKey:   os.Getenv("ANTHROPIC_API_KEY"),
    Model:    "claude-3-7-sonnet-latest",
    Timeout:  30 * time.Second,
}

client, err := llm.CreateClient(config)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

resp, err := client.CreateMessage(ctx, llm.MessageRequest{
    SystemPrompt: "You are a helpful assistant.",
    Messages: []types.Message{
        {
            Role: "user",
            Content: []types.Content{
                {Type: "text", Text: "Hello!"},
            },
        },
    },
    MaxTokens: 100,
})
```

## Tool Calling Support

Both providers support tool calling (function calling):

```go
req := llm.MessageRequest{
    Messages: messages,
    MaxTokens: 1000,
    Tools: []llm.ToolDefinition{
        {
            Name:        "get_weather",
            Description: "Get the current weather",
            InputSchema: map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "location": map[string]interface{}{
                        "type": "string",
                        "description": "City name",
                    },
                },
                "required": []string{"location"},
            },
        },
    },
}

resp, err := client.CreateMessage(ctx, req)
if err != nil {
    log.Fatal(err)
}

// Check for tool calls
for _, content := range resp.Message.Content {
    if content.Type == "tool_use" {
        // Handle tool call
        toolUse := content.ToolUse
        fmt.Printf("Tool: %s\nInput: %+v\n", toolUse.Name, toolUse.Input)
    }
}
```

## Streaming Support

Both providers support streaming responses:

```go
req := llm.MessageRequest{
    Messages:  messages,
    MaxTokens: 1000,
    Stream:    true,
}

streamChan, err := client.StreamMessage(ctx, req)
if err != nil {
    log.Fatal(err)
}

for chunk := range streamChan {
    if chunk.Error != nil {
        log.Printf("Error: %v\n", chunk.Error)
        break
    }

    if chunk.Delta.Type == "text" {
        fmt.Print(chunk.Delta.Text)
    }

    if chunk.Done {
        fmt.Println()
        break
    }
}
```

## Testing

### Mock Client

For testing, use the mock client:

```go
import (
    "github.com/Zerofisher/goai/pkg/llm/mock"
)

responses := []*llm.MessageResponse{
    {
        ID:    "test-id",
        Model: "mock-model",
        Message: types.Message{
            Role: "assistant",
            Content: []types.Content{
                {Type: "text", Text: "Hello!"},
            },
        },
    },
}

client := mock.NewClient(responses, nil)
```

### Integration Tests

Run integration tests with real API keys:

```bash
# OpenAI integration tests
export OPENAI_API_KEY="your-key"
go test -tags=integration ./pkg/llm/openai/

# Anthropic integration tests
export ANTHROPIC_API_KEY="your-key"
go test -tags=integration ./pkg/llm/anthropic/
```

## Performance

Performance benchmarks for mock client:

```
BenchmarkCreateMessage-12               12256712        92.87 ns/op
BenchmarkStreamMessage-12                 960399      1328 ns/op
BenchmarkToolDefinitionConversion-12     8660696       131.3 ns/op
```

Run benchmarks:

```bash
go test -bench=. -benchmem ./pkg/llm/
```

## Error Handling

Both providers return detailed error information:

```go
resp, err := client.CreateMessage(ctx, req)
if err != nil {
    // Check for specific error types
    if strings.Contains(err.Error(), "rate limit") {
        // Handle rate limiting
    } else if strings.Contains(err.Error(), "timeout") {
        // Handle timeout
    } else {
        // Handle other errors
    }
    return err
}
```

## Adding New Providers

To add a new LLM provider:

1. Create a new package under `pkg/llm/`
2. Implement the `llm.Client` interface
3. Register the factory in an `init()` function:

```go
package myprovider

import "github.com/Zerofisher/goai/pkg/llm"

func init() {
    llm.RegisterClientFactory("myprovider", func(config llm.ClientConfig) (llm.Client, error) {
        return NewClient(config)
    })
}
```

4. Import the package in `cmd/goai/main.go`:

```go
import (
    _ "github.com/Zerofisher/goai/pkg/llm/myprovider"
)
```

## References

- [OpenAI API Documentation](https://platform.openai.com/docs/api-reference)
- [OpenAI Go SDK](https://github.com/openai/openai-go)
- [Anthropic API Documentation](https://docs.anthropic.com/claude/reference/getting-started-with-the-api)
- [Anthropic Go SDK](https://github.com/anthropics/anthropic-sdk-go)
