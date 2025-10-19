package llm

import (
	"context"
)

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

	// Provider returns the provider name (openai, anthropic, etc.)
	Provider() string

	// Close gracefully closes the client and releases resources
	Close() error
}

// AdvancedClient extends the base Client with advanced features
type AdvancedClient interface {
	Client

	// CountTokens estimates or accurately counts tokens in a message
	CountTokens(ctx context.Context, req MessageRequest) (int, error)

	// CreateBatch creates a batch of messages for processing
	CreateBatch(ctx context.Context, requests []MessageRequest) (*BatchResponse, error)

	// GetBatch retrieves the status and results of a batch
	GetBatch(ctx context.Context, batchID string) (*BatchResponse, error)
}