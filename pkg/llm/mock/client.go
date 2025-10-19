// Package mock provides a mock LLM client for testing purposes
package mock

import (
	"context"
	"fmt"
	"time"

	"github.com/Zerofisher/goai/pkg/llm"
	"github.com/Zerofisher/goai/pkg/types"
)

// Client is a mock implementation of llm.Client for testing
type Client struct {
	responses     []*llm.MessageResponse
	streamChunks  []llm.StreamChunk
	responseIndex int
	model         string
	provider      string

	// Optional callbacks for testing
	CreateMessageFunc  func(ctx context.Context, req llm.MessageRequest) (*llm.MessageResponse, error)
	StreamMessageFunc  func(ctx context.Context, req llm.MessageRequest) (<-chan llm.StreamChunk, error)
	CountTokensFunc    func(ctx context.Context, req llm.MessageRequest) (int, error)
	CreateBatchFunc    func(ctx context.Context, requests []llm.MessageRequest) (*llm.BatchResponse, error)
	GetBatchFunc       func(ctx context.Context, batchID string) (*llm.BatchResponse, error)
}

// NewClient creates a new mock client with predefined responses
func NewClient(responses []*llm.MessageResponse, streamChunks []llm.StreamChunk) *Client {
	return &Client{
		responses:    responses,
		streamChunks: streamChunks,
		model:        "mock-model",
		provider:     "mock",
	}
}

// NewSimpleClient creates a mock client with a single text response
func NewSimpleClient(responseText string) *Client {
	return &Client{
		responses: []*llm.MessageResponse{
			{
				ID:      "mock-response-1",
				Model:   "mock-model",
				Message: types.NewTextMessage("assistant", responseText),
				CreatedAt: time.Now(),
			},
		},
		model:    "mock-model",
		provider: "mock",
	}
}

// CreateMessage implements llm.Client
func (c *Client) CreateMessage(ctx context.Context, req llm.MessageRequest) (*llm.MessageResponse, error) {
	if c.CreateMessageFunc != nil {
		return c.CreateMessageFunc(ctx, req)
	}

	if c.responseIndex >= len(c.responses) {
		return nil, fmt.Errorf("no more mock responses (requested index: %d, available: %d)",
			c.responseIndex, len(c.responses))
	}

	resp := c.responses[c.responseIndex]
	c.responseIndex++
	return resp, nil
}

// StreamMessage implements llm.Client
func (c *Client) StreamMessage(ctx context.Context, req llm.MessageRequest) (<-chan llm.StreamChunk, error) {
	if c.StreamMessageFunc != nil {
		return c.StreamMessageFunc(ctx, req)
	}

	ch := make(chan llm.StreamChunk, len(c.streamChunks)+1)

	go func() {
		defer close(ch)

		for _, chunk := range c.streamChunks {
			select {
			case ch <- chunk:
			case <-ctx.Done():
				ch <- llm.StreamChunk{
					Error: ctx.Err(),
					Done:  true,
				}
				return
			}
		}

		// Send final done chunk
		ch <- llm.StreamChunk{Done: true}
	}()

	return ch, nil
}

// GetModel implements llm.Client
func (c *Client) GetModel() string {
	return c.model
}

// SetModel implements llm.Client
func (c *Client) SetModel(model string) error {
	c.model = model
	return nil
}

// IsAvailable implements llm.Client
func (c *Client) IsAvailable() bool {
	return true
}

// Provider implements llm.Client
func (c *Client) Provider() string {
	return c.provider
}

// Close implements llm.Client
func (c *Client) Close() error {
	return nil
}

// CountTokens implements llm.AdvancedClient
func (c *Client) CountTokens(ctx context.Context, req llm.MessageRequest) (int, error) {
	if c.CountTokensFunc != nil {
		return c.CountTokensFunc(ctx, req)
	}

	// Simple mock: count characters / 4 (rough token estimate)
	totalChars := 0
	for _, msg := range req.Messages {
		totalChars += len(msg.GetText())
	}
	return totalChars / 4, nil
}

// CreateBatch implements llm.AdvancedClient
func (c *Client) CreateBatch(ctx context.Context, requests []llm.MessageRequest) (*llm.BatchResponse, error) {
	if c.CreateBatchFunc != nil {
		return c.CreateBatchFunc(ctx, requests)
	}

	// Convert []*MessageResponse to []MessageResponse
	results := make([]llm.MessageResponse, len(c.responses))
	for i, resp := range c.responses {
		if resp != nil {
			results[i] = *resp
		}
	}

	return &llm.BatchResponse{
		ID:             "mock-batch-1",
		Status:         "completed",
		TotalRequests:  len(requests),
		CompletedCount: len(requests),
		FailedCount:    0,
		Results:        results,
		CreatedAt:      time.Now(),
	}, nil
}

// GetBatch implements llm.AdvancedClient
func (c *Client) GetBatch(ctx context.Context, batchID string) (*llm.BatchResponse, error) {
	if c.GetBatchFunc != nil {
		return c.GetBatchFunc(ctx, batchID)
	}

	// Convert []*MessageResponse to []MessageResponse
	results := make([]llm.MessageResponse, len(c.responses))
	for i, resp := range c.responses {
		if resp != nil {
			results[i] = *resp
		}
	}

	now := time.Now()
	return &llm.BatchResponse{
		ID:             batchID,
		Status:         "completed",
		TotalRequests:  1,
		CompletedCount: 1,
		FailedCount:    0,
		Results:        results,
		CreatedAt:      now.Add(-time.Minute),
		CompletedAt:    &now,
	}, nil
}

// AddResponse adds a new response to the mock client
func (c *Client) AddResponse(resp *llm.MessageResponse) {
	c.responses = append(c.responses, resp)
}

// AddStreamChunk adds a new stream chunk to the mock client
func (c *Client) AddStreamChunk(chunk llm.StreamChunk) {
	c.streamChunks = append(c.streamChunks, chunk)
}

// Reset resets the response index
func (c *Client) Reset() {
	c.responseIndex = 0
}

// SetProvider sets the provider name
func (c *Client) SetProvider(provider string) {
	c.provider = provider
}
