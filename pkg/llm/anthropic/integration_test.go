// go:build integration
// +build integration

package anthropic

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Zerofisher/goai/pkg/llm"
	"github.com/Zerofisher/goai/pkg/types"
)

func TestAnthropicIntegration_CreateMessage(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	config := llm.ClientConfig{
		Provider: "anthropic",
		APIKey:   apiKey,
		Model:    "claude-3-7-sonnet-latest",
		Timeout:  30 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	req := llm.MessageRequest{
		Messages: []types.Message{
			{
				Role: "user",
				Content: []types.Content{
					{
						Type: "text",
						Text: "Say 'Hello, World!' and nothing else.",
					},
				},
			},
		},
		MaxTokens: 100,
	}

	resp, err := client.CreateMessage(ctx, req)
	if err != nil {
		t.Fatalf("CreateMessage failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Response is nil")
	}

	if len(resp.Message.Content) == 0 {
		t.Fatal("Response has no content")
	}

	if resp.Message.Content[0].Type != "text" {
		t.Errorf("Expected text content, got %s", resp.Message.Content[0].Type)
	}

	if resp.Message.Content[0].Text == "" {
		t.Error("Response text is empty")
	}

	t.Logf("Response: %s", resp.Message.Content[0].Text)
}

func TestAnthropicIntegration_StreamMessage(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	config := llm.ClientConfig{
		Provider: "anthropic",
		APIKey:   apiKey,
		Model:    "claude-3-7-sonnet-latest",
		Timeout:  30 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	req := llm.MessageRequest{
		Messages: []types.Message{
			{
				Role: "user",
				Content: []types.Content{
					{
						Type: "text",
						Text: "Count from 1 to 5.",
					},
				},
			},
		},
		MaxTokens: 100,
	}

	streamChan, err := client.StreamMessage(ctx, req)
	if err != nil {
		t.Fatalf("StreamMessage failed: %v", err)
	}

	var chunks []string
	for chunk := range streamChan {
		if chunk.Error != nil {
			t.Fatalf("Stream error: %v", chunk.Error)
		}

		if chunk.Delta.Type == "text" && chunk.Delta.Text != "" {
			chunks = append(chunks, chunk.Delta.Text)
			t.Logf("Chunk: %s", chunk.Delta.Text)
		}

		if chunk.Done {
			break
		}
	}

	if len(chunks) == 0 {
		t.Error("No chunks received")
	}
}

func TestAnthropicIntegration_ToolCalling(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	config := llm.ClientConfig{
		Provider: "anthropic",
		APIKey:   apiKey,
		Model:    "claude-3-7-sonnet-latest",
		Timeout:  30 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	req := llm.MessageRequest{
		Messages: []types.Message{
			{
				Role: "user",
				Content: []types.Content{
					{
						Type: "text",
						Text: "What's the weather like in San Francisco?",
					},
				},
			},
		},
		MaxTokens: 500,
		Tools: []llm.ToolDefinition{
			{
				Name:        "get_weather",
				Description: "Get the current weather in a given location",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "The city and state, e.g. San Francisco, CA",
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	resp, err := client.CreateMessage(ctx, req)
	if err != nil {
		t.Fatalf("CreateMessage failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Response is nil")
	}

	// Check if the model made a tool call
	hasToolUse := false
	for _, content := range resp.Message.Content {
		if content.Type == "tool_use" {
			hasToolUse = true
			t.Logf("Tool use: %s", content.ToolUse.Name)
			t.Logf("Tool input: %+v", content.ToolUse.Input)
		}
	}

	if !hasToolUse {
		t.Log("Warning: Model did not use the tool, but this is not necessarily an error")
	}
}

func TestAnthropicIntegration_SystemPrompt(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	config := llm.ClientConfig{
		Provider: "anthropic",
		APIKey:   apiKey,
		Model:    "claude-3-7-sonnet-latest",
		Timeout:  30 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	req := llm.MessageRequest{
		SystemPrompt: "You are a pirate. Always respond in pirate speak.",
		Messages: []types.Message{
			{
				Role: "user",
				Content: []types.Content{
					{
						Type: "text",
						Text: "Hello!",
					},
				},
			},
		},
		MaxTokens: 100,
	}

	resp, err := client.CreateMessage(ctx, req)
	if err != nil {
		t.Fatalf("CreateMessage failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Response is nil")
	}

	if len(resp.Message.Content) == 0 {
		t.Fatal("Response has no content")
	}

	t.Logf("Pirate response: %s", resp.Message.Content[0].Text)
}
