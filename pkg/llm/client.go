package llm

import (
	"context"
	"time"

	"github.com/Zerofisher/goai/pkg/types"
)

// MessageRequest represents a request to the LLM
type MessageRequest struct {
	Model       string          `json:"model"`
	Messages    []types.Message `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float32         `json:"temperature,omitempty"`
	Stream      bool            `json:"stream"`
	Tools       []ToolDefinition `json:"tools,omitempty"`
	SystemPrompt string         `json:"system,omitempty"`
}

// MessageResponse represents a response from the LLM
type MessageResponse struct {
	ID        string         `json:"id"`
	Model     string         `json:"model"`
	Message   types.Message  `json:"message"`
	Usage     *TokenUsage    `json:"usage,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// StreamChunk represents a chunk of streaming response
type StreamChunk struct {
	ID      string         `json:"id"`
	Model   string         `json:"model"`
	Delta   types.Content  `json:"delta"`
	Usage   *TokenUsage    `json:"usage,omitempty"`
	Error   error          `json:"error,omitempty"`
	Done    bool           `json:"done"`
}

// TokenUsage represents token usage information
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ToolDefinition represents a tool that can be called by the LLM
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// Client defines the interface for LLM clients
type Client interface {
	// CreateMessage sends a message to the LLM and returns the response
	CreateMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error)

	// StreamMessage sends a message to the LLM and streams the response
	StreamMessage(ctx context.Context, req MessageRequest) (<-chan StreamChunk, error)

	// GetModel returns the current model being used
	GetModel() string

	// SetModel sets the model to use
	SetModel(model string) error

	// IsAvailable checks if the client is available and configured
	IsAvailable() bool
}

// ClientConfig represents configuration for an LLM client
type ClientConfig struct {
	Provider    string  `json:"provider"`
	APIKey      string  `json:"api_key"`
	BaseURL     string  `json:"base_url,omitempty"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float32 `json:"temperature"`
	Timeout     time.Duration `json:"timeout"`
}

// Factory is a function that creates a new LLM client
type Factory func(config ClientConfig) (Client, error)

// Registry for client factories
var clientFactories = make(map[string]Factory)

// RegisterClientFactory registers a factory for a provider
func RegisterClientFactory(provider string, factory Factory) {
	clientFactories[provider] = factory
}

// CreateClient creates a new LLM client based on the provider
func CreateClient(config ClientConfig) (Client, error) {
	factory, exists := clientFactories[config.Provider]
	if !exists {
		return nil, types.NewAgentError(types.ErrCodeLLMConnection, "unknown provider: "+config.Provider)
	}
	return factory(config)
}

// DefaultClientConfig returns a default client configuration
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		Provider:    "openai",
		Model:       "gpt-4",
		MaxTokens:   4096,
		Temperature: 0.7,
		Timeout:     60 * time.Second,
	}
}