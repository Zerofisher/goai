package llm_test

import (
	"context"
	"testing"
	"time"

	"github.com/Zerofisher/goai/pkg/llm"
	"github.com/Zerofisher/goai/pkg/llm/mock"
	"github.com/Zerofisher/goai/pkg/types"
)

// BenchmarkCreateMessage benchmarks the CreateMessage method
func BenchmarkCreateMessage(b *testing.B) {
	req := llm.MessageRequest{
		Messages: []types.Message{
			{
				Role: "user",
				Content: []types.Content{
					{
						Type: "text",
						Text: "Hello",
					},
				},
			},
		},
		MaxTokens: 100,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create new mock client for each iteration
		responses := []*llm.MessageResponse{
			{
				ID:    "test-id",
				Model: "mock-model",
				Message: types.Message{
					Role: "assistant",
					Content: []types.Content{
						{
							Type: "text",
							Text: "Hello, this is a test response!",
						},
					},
				},
				Usage: &llm.TokenUsage{
					PromptTokens:     10,
					CompletionTokens: 20,
					TotalTokens:      30,
				},
				CreatedAt: time.Now(),
			},
		}
		client := mock.NewClient(responses, nil)

		_, err := client.CreateMessage(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStreamMessage benchmarks the StreamMessage method
func BenchmarkStreamMessage(b *testing.B) {
	req := llm.MessageRequest{
		Messages: []types.Message{
			{
				Role: "user",
				Content: []types.Content{
					{
						Type: "text",
						Text: "Hello",
					},
				},
			},
		},
		MaxTokens: 100,
		Stream:    true,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create mock client with stream chunks
		chunks := []llm.StreamChunk{
			{
				Delta: types.Content{
					Type: "text",
					Text: "Hello",
				},
			},
			{
				Delta: types.Content{
					Type: "text",
					Text: " world",
				},
			},
			{
				Done: true,
			},
		}
		client := mock.NewClient(nil, chunks)

		streamChan, err := client.StreamMessage(ctx, req)
		if err != nil {
			b.Fatal(err)
		}

		// Consume all chunks
		for range streamChan {
			// Just consume
		}
	}
}

// BenchmarkToolDefinitionConversion benchmarks tool definition handling
func BenchmarkToolDefinitionConversion(b *testing.B) {
	tools := []llm.ToolDefinition{
		{
			Name:        "test_tool",
			Description: "A test tool",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"param1": map[string]interface{}{
						"type":        "string",
						"description": "First parameter",
					},
					"param2": map[string]interface{}{
						"type":        "number",
						"description": "Second parameter",
					},
				},
				"required": []string{"param1"},
			},
		},
	}

	req := llm.MessageRequest{
		Messages: []types.Message{
			{
				Role: "user",
				Content: []types.Content{
					{
						Type: "text",
						Text: "Use the test tool",
					},
				},
			},
		},
		MaxTokens: 100,
		Tools:     tools,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		responses := []*llm.MessageResponse{
			{
				ID:    "test-id",
				Model: "mock-model",
				Message: types.Message{
					Role: "assistant",
					Content: []types.Content{
						{
							Type: "tool_use",
							ToolUse: &types.ToolUse{
								ID:   "tool-1",
								Name: "test_tool",
								Input: map[string]interface{}{
									"param1": "value1",
									"param2": 42,
								},
							},
						},
					},
				},
				CreatedAt: time.Now(),
			},
		}
		client := mock.NewClient(responses, nil)

		_, err := client.CreateMessage(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}
