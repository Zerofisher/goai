package openai

import (
	"testing"

	"github.com/Zerofisher/goai/pkg/llm"
	"github.com/Zerofisher/goai/pkg/types"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  llm.ClientConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: llm.ClientConfig{
				Provider: "openai",
				APIKey:   "test-key",
				Model:    "gpt-4",
			},
			wantErr: false,
		},
		{
			name: "missing api key",
			config: llm.ClientConfig{
				Provider: "openai",
				Model:    "gpt-4",
			},
			wantErr: true,
		},
		{
			name: "empty model defaults to gpt-4.1-mini",
			config: llm.ClientConfig{
				Provider: "openai",
				APIKey:   "test-key",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if client == nil {
					t.Error("NewClient() returned nil client")
					return
				}
				// Check default model
				if tt.config.Model == "" && client.GetModel() != "gpt-4.1-mini" {
					t.Errorf("Expected default model gpt-4.1-mini, got %s", client.GetModel())
				}
			}
		})
	}
}

func TestClient_GetModel(t *testing.T) {
	config := llm.ClientConfig{
		Provider: "openai",
		APIKey:   "test-key",
		Model:    "gpt-4-turbo",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	got := client.GetModel()
	if got != "gpt-4-turbo" {
		t.Errorf("GetModel() = %v, want %v", got, "gpt-4-turbo")
	}
}

func TestClient_SetModel(t *testing.T) {
	config := llm.ClientConfig{
		Provider: "openai",
		APIKey:   "test-key",
		Model:    "gpt-4",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	// Test valid model
	err = client.SetModel("gpt-4-turbo")
	if err != nil {
		t.Errorf("SetModel() error = %v", err)
	}

	if client.GetModel() != "gpt-4-turbo" {
		t.Errorf("GetModel() = %v, want %v", client.GetModel(), "gpt-4-turbo")
	}

	// Test empty model
	err = client.SetModel("")
	if err == nil {
		t.Error("SetModel() with empty string should return error")
	}
}

func TestClient_IsAvailable(t *testing.T) {
	tests := []struct {
		name   string
		config llm.ClientConfig
		want   bool
	}{
		{
			name: "available with api key",
			config: llm.ClientConfig{
				Provider: "openai",
				APIKey:   "test-key",
				Model:    "gpt-4",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			if got := client.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Provider(t *testing.T) {
	config := llm.ClientConfig{
		Provider: "openai",
		APIKey:   "test-key",
		Model:    "gpt-4",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	got := client.Provider()
	want := "openai"
	if got != want {
		t.Errorf("Provider() = %v, want %v", got, want)
	}
}

func TestConvertMessages(t *testing.T) {
	tests := []struct {
		name     string
		messages []types.Message
		wantLen  int
	}{
		{
			name: "user message",
			messages: []types.Message{
				{
					Role: "user",
					Content: []types.Content{
						{Type: "text", Text: "Hello"},
					},
				},
			},
			wantLen: 1,
		},
		{
			name: "system and user messages",
			messages: []types.Message{
				{
					Role: "system",
					Content: []types.Content{
						{Type: "text", Text: "You are a helpful assistant"},
					},
				},
				{
					Role: "user",
					Content: []types.Content{
						{Type: "text", Text: "Hello"},
					},
				},
			},
			wantLen: 2,
		},
		{
			name: "assistant message",
			messages: []types.Message{
				{
					Role: "assistant",
					Content: []types.Content{
						{Type: "text", Text: "Hi there!"},
					},
				},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertMessages(tt.messages)
			if len(got) != tt.wantLen {
				t.Errorf("convertMessages() returned %d messages, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestConvertTools(t *testing.T) {
	tests := []struct {
		name    string
		tools   []llm.ToolDefinition
		wantLen int
	}{
		{
			name:    "empty tools",
			tools:   []llm.ToolDefinition{},
			wantLen: 0,
		},
		{
			name: "single tool",
			tools: []llm.ToolDefinition{
				{
					Name:        "get_weather",
					Description: "Get the weather",
					InputSchema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertTools(tt.tools)
			if len(got) != tt.wantLen {
				t.Errorf("convertTools() returned %d tools, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestConvertToolChoice(t *testing.T) {
	tests := []struct {
		name   string
		choice *llm.ToolChoice
	}{
		{
			name:   "nil choice",
			choice: nil,
		},
		{
			name: "auto choice",
			choice: &llm.ToolChoice{
				Type: "auto",
			},
		},
		{
			name: "none choice",
			choice: &llm.ToolChoice{
				Type: "none",
			},
		},
		{
			name: "required choice",
			choice: &llm.ToolChoice{
				Type: "required",
			},
		},
		{
			name: "specific tool choice",
			choice: &llm.ToolChoice{
				Type:     "tool",
				ToolName: "get_weather",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just ensure it doesn't panic
			_ = convertToolChoice(tt.choice)
		})
	}
}

func TestClient_Close(t *testing.T) {
	config := llm.ClientConfig{
		Provider: "openai",
		APIKey:   "test-key",
		Model:    "gpt-4",
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	err = client.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

// TestCreateMessage_Integration tests the CreateMessage method with a real API call
// This test requires OPENAI_API_KEY environment variable to be set
func TestCreateMessage_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// This is a placeholder - actual integration test should be in integration_test.go
	t.Skip("integration test - requires OPENAI_API_KEY")
}
