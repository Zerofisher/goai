//go:build integration
// +build integration

package openai_test

import (
	"context"
	"os"
	"testing"

	"github.com/Zerofisher/goai/pkg/llm"
	_ "github.com/Zerofisher/goai/pkg/llm/openai" // Register OpenAI factory
	"github.com/Zerofisher/goai/pkg/types"
)

func TestOpenAI_CreateMessage_Integration(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	client, err := llm.CreateClient(llm.ClientConfig{
		Provider: "openai",
		APIKey:   apiKey,
		Model:    "gpt-4o-mini", // Use cheaper model for testing
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	ctx := context.Background()
	resp, err := client.CreateMessage(ctx, llm.MessageRequest{
		Messages: []types.Message{
			types.NewTextMessage("user", "Say hello in one word"),
		},
		MaxTokens:   10,
		Temperature: 0.7,
	})

	if err != nil {
		t.Fatalf("CreateMessage() error = %v", err)
	}

	if resp == nil {
		t.Fatal("CreateMessage() returned nil response")
	}

	if resp.Message.GetText() == "" {
		t.Error("Expected non-empty response text")
	}

	t.Logf("Response: %s", resp.Message.GetText())
	t.Logf("Model: %s", resp.Model)
	t.Logf("Usage: %+v", resp.Usage)
}

func TestOpenAI_StreamMessage_Integration(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	client, err := llm.CreateClient(llm.ClientConfig{
		Provider: "openai",
		APIKey:   apiKey,
		Model:    "gpt-4o-mini",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	ctx := context.Background()
	chunkChan, err := client.StreamMessage(ctx, llm.MessageRequest{
		Messages: []types.Message{
			types.NewTextMessage("user", "Count from 1 to 5"),
		},
		MaxTokens:   50,
		Temperature: 0.7,
	})

	if err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}

	chunkCount := 0
	fullText := ""

	for chunk := range chunkChan {
		if chunk.Error != nil {
			t.Fatalf("Stream error: %v", chunk.Error)
		}

		if chunk.Delta.Type == "text" && chunk.Delta.Text != "" {
			fullText += chunk.Delta.Text
			t.Logf("Chunk %d: %s", chunkCount, chunk.Delta.Text)
		}

		chunkCount++

		if chunk.Done {
			break
		}
	}

	if chunkCount == 0 {
		t.Error("Expected at least one chunk")
	}

	if fullText == "" {
		t.Error("Expected non-empty full text from stream")
	}

	t.Logf("Total chunks: %d", chunkCount)
	t.Logf("Full text: %s", fullText)
}

func TestOpenAI_ToolCalling_Integration(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	client, err := llm.CreateClient(llm.ClientConfig{
		Provider: "openai",
		APIKey:   apiKey,
		Model:    "gpt-4o-mini",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}()

	// Define a simple tool
	tools := []llm.ToolDefinition{
		{
			Name:        "get_weather",
			Description: "Get the current weather for a location",
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
	}

	ctx := context.Background()
	resp, err := client.CreateMessage(ctx, llm.MessageRequest{
		Messages: []types.Message{
			types.NewTextMessage("user", "What's the weather in San Francisco?"),
		},
		Tools:       tools,
		MaxTokens:   100,
		Temperature: 0.7,
	})

	if err != nil {
		t.Fatalf("CreateMessage() error = %v", err)
	}

	if resp == nil {
		t.Fatal("CreateMessage() returned nil response")
	}

	// Check if the model requested to use the tool
	toolUses := resp.Message.GetToolUses()
	if len(toolUses) > 0 {
		t.Logf("Model requested tool call: %s", toolUses[0].Name)
		t.Logf("Tool arguments: %+v", toolUses[0].Input)

		if toolUses[0].Name != "get_weather" {
			t.Errorf("Expected tool name 'get_weather', got '%s'", toolUses[0].Name)
		}
	} else {
		t.Log("Model did not request tool call - this is acceptable behavior")
	}
}
