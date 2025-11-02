package openai

import (
	"context"
	"fmt"

	"github.com/Zerofisher/goai/pkg/llm"
	openaisdk "github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

// Client wraps the official OpenAI SDK client to implement llm.Client interface
type Client struct {
	client *openaisdk.Client
	config llm.ClientConfig
	model  string
}

// NewClient creates a new OpenAI client using the official SDK
func NewClient(config llm.ClientConfig) (llm.Client, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("openai api key is required")
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

	sdkClient := openaisdk.NewClient(opts...)

	model := config.Model
	if model == "" {
		model = "gpt-4.1-mini" // Default to gpt-4.1-mini
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
	model := req.Model
	if model == "" {
		model = c.model
	}

	// Convert request to OpenAI format
	params := convertToOpenAIParams(req, model)

	// Make API call
	completion, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("openai api error: %w", err)
	}

	// Convert response
	return convertFromOpenAIResponse(completion), nil
}

// StreamMessage sends a message to the LLM and streams the response
func (c *Client) StreamMessage(ctx context.Context, req llm.MessageRequest) (<-chan llm.StreamChunk, error) {
	// Use the model from the request if provided, otherwise use the client's default model
	model := req.Model
	if model == "" {
		model = c.model
	}

	// Convert request to OpenAI format
	params := convertToOpenAIParams(req, model)

	// Create stream
	stream := c.client.Chat.Completions.NewStreaming(ctx, params)

	// Create channel
	chunkChan := make(chan llm.StreamChunk)

	// Process stream in goroutine
	go func() {
		defer close(chunkChan)

		acc := openaisdk.ChatCompletionAccumulator{}

		for stream.Next() {
			chunk := stream.Current()
			acc.AddChunk(chunk)

			// Send chunk
			select {
			case chunkChan <- convertStreamChunk(chunk):
			case <-ctx.Done():
				return
			}

			// Handle finished content
			if content, ok := acc.JustFinishedContent(); ok {
				// Content finished - we've already sent the chunks
				_ = content
			}

			// Handle finished tool calls
			if tool, ok := acc.JustFinishedToolCall(); ok {
				// Tool call finished - we've already sent the chunks
				_ = tool
			}
		}

		if err := stream.Err(); err != nil {
			chunkChan <- llm.StreamChunk{
				Error: err,
				Done:  true,
			}
		} else {
			chunkChan <- llm.StreamChunk{Done: true}
		}
	}()

	return chunkChan, nil
}

// GetModel returns the current model being used
func (c *Client) GetModel() string {
	return c.model
}

// SetModel sets the model to use
func (c *Client) SetModel(model string) error {
	if model == "" {
		return fmt.Errorf("model cannot be empty")
	}
	c.model = model
	return nil
}

// IsAvailable checks if the client is available and configured
func (c *Client) IsAvailable() bool {
	return c.config.APIKey != ""
}

// Provider returns the provider name (openai, anthropic, etc.)
func (c *Client) Provider() string {
	return "openai"
}

// Close gracefully closes the client and releases resources
func (c *Client) Close() error {
	// OpenAI SDK doesn't require explicit closing
	return nil
}
