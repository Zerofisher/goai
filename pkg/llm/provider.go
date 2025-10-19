package llm

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Zerofisher/goai/pkg/types"
)

// MoonshotClient implements the Client interface for Moonshot API (Kimi).
type MoonshotClient struct {
	*OpenAIClient
}

// NewMoonshotClient creates a new Moonshot client.
func NewMoonshotClient(config ClientConfig) (Client, error) {
	if config.APIKey == "" {
		return nil, types.NewAgentError(types.ErrCodeLLMInvalidKey, "Moonshot API key is required")
	}

	// Moonshot uses OpenAI-compatible API
	config.BaseURL = "https://api.moonshot.cn/v1"

	if config.Model == "" {
		config.Model = "moonshot-v1-8k"
	}

	openAIClient, err := NewOpenAIClient(config)
	if err != nil {
		return nil, err
	}

	return &MoonshotClient{
		OpenAIClient: openAIClient.(*OpenAIClient),
	}, nil
}

// ClaudeClient implements the Client interface for Claude API.
type ClaudeClient struct {
	config     ClientConfig
	httpClient *http.Client
}

// NewClaudeClient creates a new Claude client.
func NewClaudeClient(config ClientConfig) (Client, error) {
	if config.APIKey == "" {
		return nil, types.NewAgentError(types.ErrCodeLLMInvalidKey, "Claude API key is required")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.anthropic.com"
	}

	if config.Model == "" {
		config.Model = "claude-3-opus-20240229"
	}

	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}

	return &ClaudeClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}, nil
}

// CreateMessage implements Client interface for Claude.
func (c *ClaudeClient) CreateMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error) {
	// Claude-specific implementation would go here
	// For now, return a not implemented error
	return nil, types.NewAgentError(types.ErrCodeNotImplemented, "Claude client not fully implemented yet")
}

// StreamMessage implements Client interface for Claude.
func (c *ClaudeClient) StreamMessage(ctx context.Context, req MessageRequest) (<-chan StreamChunk, error) {
	return nil, types.NewAgentError(types.ErrCodeNotImplemented, "Claude streaming not implemented yet")
}

// GetModel returns the current model being used.
func (c *ClaudeClient) GetModel() string {
	return c.config.Model
}

// SetModel sets the model to use.
func (c *ClaudeClient) SetModel(model string) error {
	c.config.Model = model
	return nil
}

// IsAvailable checks if the client is available and configured.
func (c *ClaudeClient) IsAvailable() bool {
	return c.config.APIKey != ""
}

// MockClient implements the Client interface for testing.
type MockClient struct {
	responses []MessageResponse
	index     int
	model     string
}

// NewMockClient creates a new mock client for testing.
func NewMockClient(config ClientConfig) (Client, error) {
	return &MockClient{
		model:     config.Model,
		responses: []MessageResponse{},
	}, nil
}

// CreateMessage returns a mock response.
func (c *MockClient) CreateMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error) {
	if c.index >= len(c.responses) {
		return &MessageResponse{
			ID:    "mock-response-1",
			Model: c.model,
			Message: types.Message{
				Role: "assistant",
				Content: []types.Content{
					{
						Type: "text",
						Text: "This is a mock response",
					},
				},
			},
			CreatedAt: time.Now(),
		}, nil
	}

	resp := c.responses[c.index]
	c.index++
	return &resp, nil
}

// StreamMessage returns a mock stream.
func (c *MockClient) StreamMessage(ctx context.Context, req MessageRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 1)

	go func() {
		defer close(ch)

		// Simulate streaming response
		response := "This is a mock streaming response"
		for i, char := range response {
			select {
			case <-ctx.Done():
				return
			case ch <- StreamChunk{
				ID:    fmt.Sprintf("chunk-%d", i),
				Model: c.model,
				Delta: types.Content{
					Type: "text",
					Text: string(char),
				},
			}:
			}
			time.Sleep(10 * time.Millisecond) // Simulate delay
		}

		ch <- StreamChunk{Done: true}
	}()

	return ch, nil
}

// GetModel returns the current model being used.
func (c *MockClient) GetModel() string {
	return c.model
}

// SetModel sets the model to use.
func (c *MockClient) SetModel(model string) error {
	c.model = model
	return nil
}

// IsAvailable always returns true for mock client.
func (c *MockClient) IsAvailable() bool {
	return true
}

// LocalLLMClient implements the Client interface for local LLMs (like Ollama).
type LocalLLMClient struct {
	config     ClientConfig
	httpClient *http.Client
}

// NewLocalLLMClient creates a new local LLM client.
func NewLocalLLMClient(config ClientConfig) (Client, error) {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:11434" // Default Ollama URL
	}

	if config.Model == "" {
		config.Model = "llama2"
	}

	if config.Timeout == 0 {
		config.Timeout = 300 * time.Second // Longer timeout for local models
	}

	return &LocalLLMClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}, nil
}

// CreateMessage sends a message to the local LLM.
func (c *LocalLLMClient) CreateMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error) {
	// Implementation for local LLM (e.g., Ollama) would go here
	return nil, types.NewAgentError(types.ErrCodeNotImplemented, "Local LLM client not implemented yet")
}

// StreamMessage streams a response from the local LLM.
func (c *LocalLLMClient) StreamMessage(ctx context.Context, req MessageRequest) (<-chan StreamChunk, error) {
	return nil, types.NewAgentError(types.ErrCodeNotImplemented, "Local LLM streaming not implemented yet")
}

// GetModel returns the current model being used.
func (c *LocalLLMClient) GetModel() string {
	return c.config.Model
}

// SetModel sets the model to use.
func (c *LocalLLMClient) SetModel(model string) error {
	c.config.Model = model
	return nil
}

// IsAvailable checks if the local LLM is available.
func (c *LocalLLMClient) IsAvailable() bool {
	// Could ping the local server here
	return true
}

func init() {
	// Register all provider factories
	RegisterClientFactory("moonshot", NewMoonshotClient)
	RegisterClientFactory("claude", NewClaudeClient)
	RegisterClientFactory("mock", NewMockClient)
	RegisterClientFactory("local", NewLocalLLMClient)
	RegisterClientFactory("ollama", NewLocalLLMClient) // Alias for local
}