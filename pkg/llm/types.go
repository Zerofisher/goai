package llm

import (
	"time"

	"github.com/Zerofisher/goai/pkg/types"
)

// MessageRequest represents a request to the LLM
type MessageRequest struct {
	Model         string            `json:"model"`
	Messages      []types.Message   `json:"messages"`
	MaxTokens     int               `json:"max_tokens,omitempty"`
	Temperature   float32           `json:"temperature,omitempty"`
	TopP          float32           `json:"top_p,omitempty"`
	Stream        bool              `json:"stream"`
	Tools         []ToolDefinition  `json:"tools,omitempty"`
	ToolChoice    *ToolChoice       `json:"tool_choice,omitempty"`
	SystemPrompt  string            `json:"system,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	ResponseFormat *ResponseFormat  `json:"response_format,omitempty"`
	Seed          *int              `json:"seed,omitempty"`
	StopSequences []string          `json:"stop,omitempty"`
}

// MessageResponse represents a response from the LLM
type MessageResponse struct {
	ID        string        `json:"id"`
	Model     string        `json:"model"`
	Message   types.Message `json:"message"`
	Usage     *TokenUsage   `json:"usage,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
}

// StreamChunk represents a chunk of streaming response
type StreamChunk struct {
	ID    string        `json:"id"`
	Model string        `json:"model"`
	Delta types.Content `json:"delta"`
	Usage *TokenUsage   `json:"usage,omitempty"`
	Error error         `json:"error,omitempty"`
	Done  bool          `json:"done"`
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

// ToolChoice controls how tools are selected
type ToolChoice struct {
	Type     string `json:"type"`           // "auto", "any", "tool", "none"
	ToolName string `json:"name,omitempty"` // For type="tool"
}

// ResponseFormat specifies the output format
type ResponseFormat struct {
	Type       string                 `json:"type"` // "text", "json_object", "json_schema"
	JSONSchema map[string]interface{} `json:"json_schema,omitempty"`
}

// BatchResponse represents a batch processing result
type BatchResponse struct {
	ID             string            `json:"id"`
	Status         string            `json:"status"` // "processing", "completed", "failed"
	TotalRequests  int               `json:"total_requests"`
	CompletedCount int               `json:"completed_count"`
	FailedCount    int               `json:"failed_count"`
	Results        []MessageResponse `json:"results,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	CompletedAt    *time.Time        `json:"completed_at,omitempty"`
}

// ClientConfig represents configuration for an LLM client
type ClientConfig struct {
	Provider    string        `json:"provider"`
	APIKey      string        `json:"api_key"`
	BaseURL     string        `json:"base_url,omitempty"`
	Model       string        `json:"model"`
	MaxTokens   int           `json:"max_tokens"`
	Temperature float32       `json:"temperature"`
	Timeout     time.Duration `json:"timeout"`
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
