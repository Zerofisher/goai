package anthropic

import (
	"context"
	"fmt"

	"github.com/Zerofisher/goai/pkg/llm"
	anthropicsdk "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Client wraps the official Anthropic SDK client to implement llm.Client interface
type Client struct {
	client *anthropicsdk.Client
	config llm.ClientConfig
	model  anthropicsdk.Model
}

// NewClient creates a new Anthropic client using the official SDK
func NewClient(config llm.ClientConfig) (llm.Client, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("anthropic api key is required")
	}

	opts := []option.RequestOption{
		option.WithAPIKey(config.APIKey),
	}

	if config.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(config.BaseURL))
	}

	if config.Timeout > 0 {
		opts = append(opts, option.WithRequestTimeout(config.Timeout))
	}

	sdkClient := anthropicsdk.NewClient(opts...)

	model := anthropicsdk.Model(config.Model)
	if config.Model == "" {
		model = anthropicsdk.ModelClaude3_7SonnetLatest
	}

	return &Client{
		client: &sdkClient,
		config: config,
		model:  model,
	}, nil
}

// CreateMessage sends a message to the LLM and returns the response
func (c *Client) CreateMessage(ctx context.Context, req llm.MessageRequest) (*llm.MessageResponse, error) {
	// Use the model from the request if provided, otherwise use the client's default model
	model := c.model
	if req.Model != "" {
		model = anthropicsdk.Model(req.Model)
	}

	// Convert request to Anthropic format
	params := convertToAnthropicParams(req, model)

	// Make API call
	message, err := c.client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("anthropic api error: %w", err)
	}

	// Convert response
	return convertFromAnthropicResponse(message), nil
}

// StreamMessage sends a message to the LLM and streams the response
func (c *Client) StreamMessage(ctx context.Context, req llm.MessageRequest) (<-chan llm.StreamChunk, error) {
	// Use the model from the request if provided, otherwise use the client's default model
	model := c.model
	if req.Model != "" {
		model = anthropicsdk.Model(req.Model)
	}

	// Convert request to Anthropic format
	params := convertToAnthropicParams(req, model)

	// Create stream
	stream := c.client.Messages.NewStreaming(ctx, params)

	// Create channel
	chunkChan := make(chan llm.StreamChunk)

	// Process stream in goroutine
	go func() {
		defer close(chunkChan)

		message := anthropicsdk.Message{}

		for stream.Next() {
			event := stream.Current()

			// Accumulate message
			if err := message.Accumulate(event); err != nil {
				chunkChan <- llm.StreamChunk{
					Error: err,
					Done:  true,
				}
				return
			}

			// Process events
			switch eventVariant := event.AsAny().(type) {
			case anthropicsdk.ContentBlockDeltaEvent:
				switch deltaVariant := eventVariant.Delta.AsAny().(type) {
				case anthropicsdk.TextDelta:
					chunkChan <- llm.StreamChunk{
						Delta: convertTextDelta(deltaVariant),
					}
				}
			case anthropicsdk.MessageStopEvent:
				chunkChan <- llm.StreamChunk{Done: true}
			}
		}

		if err := stream.Err(); err != nil {
			chunkChan <- llm.StreamChunk{
				Error: err,
				Done:  true,
			}
		}
	}()

	return chunkChan, nil
}

// GetModel returns the current model being used
func (c *Client) GetModel() string {
	return string(c.model)
}

// SetModel sets the model to use
func (c *Client) SetModel(model string) error {
	if model == "" {
		return fmt.Errorf("model cannot be empty")
	}
	c.model = anthropicsdk.Model(model)
	return nil
}

// IsAvailable checks if the client is available and configured
func (c *Client) IsAvailable() bool {
	return c.config.APIKey != ""
}

// Provider returns the provider name (openai, anthropic, etc.)
func (c *Client) Provider() string {
	return "anthropic"
}

// Close gracefully closes the client and releases resources
func (c *Client) Close() error {
	// Anthropic SDK doesn't require explicit closing
	return nil
}
