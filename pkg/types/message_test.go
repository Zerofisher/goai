package types

import (
	"encoding/json"
	"testing"
)

func TestMessage_NewTextMessage(t *testing.T) {
	msg := NewTextMessage("user", "Hello, world!")

	if msg.Role != "user" {
		t.Errorf("expected role 'user', got '%s'", msg.Role)
	}

	if len(msg.Content) != 1 {
		t.Fatalf("expected 1 content, got %d", len(msg.Content))
	}

	if msg.Content[0].Type != "text" {
		t.Errorf("expected content type 'text', got '%s'", msg.Content[0].Type)
	}

	if msg.Content[0].Text != "Hello, world!" {
		t.Errorf("expected text 'Hello, world!', got '%s'", msg.Content[0].Text)
	}
}

func TestMessage_NewToolUseMessage(t *testing.T) {
	toolUse := &ToolUse{
		ID:   "test_id",
		Name: "test_tool",
		Input: map[string]interface{}{
			"param": "value",
		},
	}

	msg := NewToolUseMessage(toolUse)

	if msg.Role != "assistant" {
		t.Errorf("expected role 'assistant', got '%s'", msg.Role)
	}

	if len(msg.Content) != 1 {
		t.Fatalf("expected 1 content, got %d", len(msg.Content))
	}

	if msg.Content[0].Type != "tool_use" {
		t.Errorf("expected content type 'tool_use', got '%s'", msg.Content[0].Type)
	}

	if msg.Content[0].ToolUse == nil {
		t.Fatal("expected tool use to be non-nil")
	}

	if msg.Content[0].ToolUse.ID != "test_id" {
		t.Errorf("expected tool use ID 'test_id', got '%s'", msg.Content[0].ToolUse.ID)
	}
}

func TestMessage_NewToolResultMessage(t *testing.T) {
	toolResult := &ToolResult{
		ToolUseID: "test_id",
		Content:   "Success",
		IsError:   false,
	}

	msg := NewToolResultMessage(toolResult)

	if msg.Role != "user" {
		t.Errorf("expected role 'user', got '%s'", msg.Role)
	}

	if len(msg.Content) != 1 {
		t.Fatalf("expected 1 content, got %d", len(msg.Content))
	}

	if msg.Content[0].Type != "tool_result" {
		t.Errorf("expected content type 'tool_result', got '%s'", msg.Content[0].Type)
	}

	if msg.Content[0].ToolResult == nil {
		t.Fatal("expected tool result to be non-nil")
	}

	if msg.Content[0].ToolResult.ToolUseID != "test_id" {
		t.Errorf("expected tool result ID 'test_id', got '%s'", msg.Content[0].ToolResult.ToolUseID)
	}
}

func TestMessage_AddContent(t *testing.T) {
	msg := NewTextMessage("user", "First")
	msg.AddContent(Content{
		Type: "text",
		Text: "Second",
	})

	if len(msg.Content) != 2 {
		t.Fatalf("expected 2 contents, got %d", len(msg.Content))
	}

	if msg.Content[1].Text != "Second" {
		t.Errorf("expected second content text 'Second', got '%s'", msg.Content[1].Text)
	}
}

func TestMessage_GetText(t *testing.T) {
	msg := Message{
		Role: "user",
		Content: []Content{
			{Type: "text", Text: "Hello "},
			{Type: "text", Text: "world!"},
			{Type: "tool_use", ToolUse: &ToolUse{ID: "test", Name: "tool"}},
		},
	}

	text := msg.GetText()
	expected := "Hello world!"
	if text != expected {
		t.Errorf("expected text '%s', got '%s'", expected, text)
	}
}

func TestMessage_HasToolUse(t *testing.T) {
	tests := []struct {
		name     string
		message  Message
		expected bool
	}{
		{
			name:     "message with tool use",
			message:  NewToolUseMessage(&ToolUse{ID: "test", Name: "tool"}),
			expected: true,
		},
		{
			name:     "message without tool use",
			message:  NewTextMessage("user", "text"),
			expected: false,
		},
		{
			name: "mixed message with tool use",
			message: Message{
				Role: "assistant",
				Content: []Content{
					{Type: "text", Text: "Using tool"},
					{Type: "tool_use", ToolUse: &ToolUse{ID: "test", Name: "tool"}},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.message.HasToolUse(); got != tt.expected {
				t.Errorf("HasToolUse() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMessage_GetToolUses(t *testing.T) {
	toolUse1 := &ToolUse{ID: "tool1", Name: "first"}
	toolUse2 := &ToolUse{ID: "tool2", Name: "second"}

	msg := Message{
		Role: "assistant",
		Content: []Content{
			{Type: "text", Text: "Using tools"},
			{Type: "tool_use", ToolUse: toolUse1},
			{Type: "tool_use", ToolUse: toolUse2},
		},
	}

	toolUses := msg.GetToolUses()
	if len(toolUses) != 2 {
		t.Fatalf("expected 2 tool uses, got %d", len(toolUses))
	}

	if toolUses[0].ID != "tool1" {
		t.Errorf("expected first tool ID 'tool1', got '%s'", toolUses[0].ID)
	}

	if toolUses[1].ID != "tool2" {
		t.Errorf("expected second tool ID 'tool2', got '%s'", toolUses[1].ID)
	}
}

func TestMessage_Validate(t *testing.T) {
	tests := []struct {
		name    string
		message Message
		wantErr bool
	}{
		{
			name:    "valid text message",
			message: NewTextMessage("user", "Hello"),
			wantErr: false,
		},
		{
			name: "empty role",
			message: Message{
				Role:    "",
				Content: []Content{{Type: "text", Text: "Hello"}},
			},
			wantErr: true,
		},
		{
			name: "invalid role",
			message: Message{
				Role:    "invalid",
				Content: []Content{{Type: "text", Text: "Hello"}},
			},
			wantErr: true,
		},
		{
			name: "no content",
			message: Message{
				Role:    "user",
				Content: []Content{},
			},
			wantErr: true,
		},
		{
			name: "empty text content",
			message: Message{
				Role:    "user",
				Content: []Content{{Type: "text", Text: ""}},
			},
			wantErr: true,
		},
		{
			name: "nil tool use",
			message: Message{
				Role:    "assistant",
				Content: []Content{{Type: "tool_use", ToolUse: nil}},
			},
			wantErr: true,
		},
		{
			name: "nil tool result",
			message: Message{
				Role:    "user",
				Content: []Content{{Type: "tool_result", ToolResult: nil}},
			},
			wantErr: true,
		},
		{
			name: "unknown content type",
			message: Message{
				Role:    "user",
				Content: []Content{{Type: "unknown"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.message.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		content Content
		wantErr bool
	}{
		{
			name:    "valid text content",
			content: Content{Type: "text", Text: "Hello"},
			wantErr: false,
		},
		{
			name:    "empty type",
			content: Content{Type: "", Text: "Hello"},
			wantErr: true,
		},
		{
			name:    "empty text",
			content: Content{Type: "text", Text: ""},
			wantErr: true,
		},
		{
			name: "valid tool use",
			content: Content{
				Type:    "tool_use",
				ToolUse: &ToolUse{ID: "id", Name: "name", Input: map[string]interface{}{}},
			},
			wantErr: false,
		},
		{
			name: "invalid tool use",
			content: Content{
				Type:    "tool_use",
				ToolUse: &ToolUse{ID: "", Name: "name", Input: map[string]interface{}{}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.content.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMessage_JSONSerialization(t *testing.T) {
	original := Message{
		Role: "user",
		Content: []Content{
			{Type: "text", Text: "Hello"},
			{
				Type: "tool_use",
				ToolUse: &ToolUse{
					ID:   "test",
					Name: "tool",
					Input: map[string]interface{}{
						"key": "value",
					},
				},
			},
		},
	}

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal message: %v", err)
	}

	// Unmarshal
	var decoded Message
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("failed to unmarshal message: %v", err)
	}

	// Compare
	if decoded.Role != original.Role {
		t.Errorf("role mismatch: got %s, want %s", decoded.Role, original.Role)
	}

	if len(decoded.Content) != len(original.Content) {
		t.Fatalf("content length mismatch: got %d, want %d", len(decoded.Content), len(original.Content))
	}

	if decoded.Content[0].Text != original.Content[0].Text {
		t.Errorf("text mismatch: got %s, want %s", decoded.Content[0].Text, original.Content[0].Text)
	}

	if decoded.Content[1].ToolUse.ID != original.Content[1].ToolUse.ID {
		t.Errorf("tool use ID mismatch: got %s, want %s", decoded.Content[1].ToolUse.ID, original.Content[1].ToolUse.ID)
	}
}