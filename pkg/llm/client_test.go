package llm

import (
	"context"
	"testing"
	"time"

	"github.com/Zerofisher/goai/pkg/types"
)

func TestNewOpenAIClient(t *testing.T) {
	tests := []struct {
		name    string
		config  ClientConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: ClientConfig{
				Provider: "openai",
				APIKey:   "test-key",
				Model:    "gpt-4",
				Timeout:  60 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing API key",
			config: ClientConfig{
				Provider: "openai",
				Model:    "gpt-4",
			},
			wantErr: true,
		},
		{
			name: "default values",
			config: ClientConfig{
				Provider: "openai",
				APIKey:   "test-key",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewOpenAIClient(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOpenAIClient() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && client == nil {
				t.Error("NewOpenAIClient() returned nil client")
			}
			if !tt.wantErr {
				// Check defaults
				if client.GetModel() == "" {
					t.Error("Model not set")
				}
				if !client.IsAvailable() {
					t.Error("Client should be available")
				}
			}
		})
	}
}

func TestOpenAIClient_buildOpenAIRequest(t *testing.T) {
	config := ClientConfig{
		Provider:    "openai",
		APIKey:      "test-key",
		Model:       "gpt-4",
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	client, err := NewOpenAIClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	openaiClient := client.(*OpenAIClient)

	req := MessageRequest{
		Messages: []types.Message{
			{
				Role: "user",
				Content: []types.Content{
					{
						Type: "text",
						Text: "Hello, world!",
					},
				},
			},
			{
				Role: "assistant",
				Content: []types.Content{
					{
						Type: "text",
						Text: "Hello! How can I help you?",
					},
				},
			},
		},
		Tools: []ToolDefinition{
			{
				Name:        "get_weather",
				Description: "Get weather information",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"location": map[string]interface{}{
							"type":        "string",
							"description": "City name",
						},
					},
				},
			},
		},
	}

	openAIReq := openaiClient.buildOpenAIRequest(req)

	// Check basic fields
	if openAIReq["model"] != "gpt-4" {
		t.Errorf("Expected model gpt-4, got %v", openAIReq["model"])
	}

	// Check messages conversion
	messages, ok := openAIReq["messages"].([]map[string]interface{})
	if !ok {
		t.Fatal("Messages not converted correctly")
	}
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}

	// Check tools conversion
	tools, ok := openAIReq["tools"].([]map[string]interface{})
	if !ok {
		t.Fatal("Tools not converted correctly")
	}
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}
}

func TestMockClient(t *testing.T) {
	config := ClientConfig{
		Provider: "mock",
		Model:    "test-model",
	}

	client, err := NewMockClient(config)
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}

	ctx := context.Background()
	req := MessageRequest{
		Messages: []types.Message{
			{
				Role: "user",
				Content: []types.Content{
					{
						Type: "text",
						Text: "Test message",
					},
				},
			},
		},
	}

	// Test CreateMessage
	resp, err := client.CreateMessage(ctx, req)
	if err != nil {
		t.Fatalf("CreateMessage failed: %v", err)
	}
	if resp == nil {
		t.Fatal("Response is nil")
	}
	if resp.Model != "test-model" {
		t.Errorf("Expected model test-model, got %s", resp.Model)
	}

	// Test StreamMessage
	streamCh, err := client.StreamMessage(ctx, req)
	if err != nil {
		t.Fatalf("StreamMessage failed: %v", err)
	}

	chunks := 0
	for chunk := range streamCh {
		if chunk.Error != nil {
			t.Errorf("Stream error: %v", chunk.Error)
		}
		if chunk.Done {
			break
		}
		chunks++
	}
	if chunks == 0 {
		t.Error("No chunks received from stream")
	}
}

func TestClientRegistry(t *testing.T) {
	// Test OpenAI factory registration
	config := ClientConfig{
		Provider: "openai",
		APIKey:   "test-key",
	}

	client, err := CreateClient(config)
	if err != nil {
		t.Fatalf("Failed to create OpenAI client from registry: %v", err)
	}
	if client == nil {
		t.Fatal("Client is nil")
	}

	// Test unknown provider
	config.Provider = "unknown"
	_, err = CreateClient(config)
	if err == nil {
		t.Error("Expected error for unknown provider")
	}
}

func TestProviders(t *testing.T) {
	tests := []struct {
		provider string
		config   ClientConfig
		wantErr  bool
	}{
		{
			provider: "moonshot",
			config: ClientConfig{
				Provider: "moonshot",
				APIKey:   "test-key",
			},
			wantErr: false,
		},
		{
			provider: "claude",
			config: ClientConfig{
				Provider: "claude",
				APIKey:   "test-key",
			},
			wantErr: false,
		},
		{
			provider: "local",
			config: ClientConfig{
				Provider: "local",
			},
			wantErr: false,
		},
		{
			provider: "ollama",
			config: ClientConfig{
				Provider: "ollama",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			client, err := CreateClient(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateClient(%s) error = %v, wantErr %v", tt.provider, err, tt.wantErr)
			}
			if !tt.wantErr && client == nil {
				t.Errorf("CreateClient(%s) returned nil", tt.provider)
			}
		})
	}
}