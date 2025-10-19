package message

import (
	"testing"

	"github.com/Zerofisher/goai/pkg/types"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name      string
		maxTokens int
		expected  int
	}{
		{"default tokens", 0, 16000},
		{"custom tokens", 5000, 5000},
		{"negative tokens", -100, 16000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.maxTokens)
			if m == nil {
				t.Fatal("NewManager returned nil")
			}
			if m.maxTokens != tt.expected {
				t.Errorf("Expected maxTokens %d, got %d", tt.expected, m.maxTokens)
			}
			if len(m.messages) != 0 {
				t.Error("Initial messages should be empty")
			}
		})
	}
}

func TestManager_Add(t *testing.T) {
	m := NewManager(1000)

	// Test adding valid message
	msg := types.NewTextMessage("user", "Hello")
	err := m.Add(msg)
	if err != nil {
		t.Fatalf("Failed to add message: %v", err)
	}

	if len(m.messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(m.messages))
	}

	// Test adding multiple messages
	m.AddAssistantMessage("Hi there!")
	m.AddSystemMessage("You are a helpful assistant")

	if m.Count() != 3 {
		t.Errorf("Expected 3 messages, got %d", m.Count())
	}
}

func TestManager_Truncate(t *testing.T) {
	m := NewManager(100) // Very low token limit

	// Add many messages to exceed token limit
	for i := 0; i < 20; i++ {
		m.AddUserMessage("This is a long message that will consume many tokens")
		m.AddAssistantMessage("This is a long response that will also consume many tokens")
	}

	// Check that truncation happened
	if m.Count() >= 40 {
		t.Error("Messages should have been truncated")
	}

	// Check that at least 2 messages remain
	if m.Count() < 2 {
		t.Error("Should keep at least 2 messages")
	}
}

func TestManager_GetHistory(t *testing.T) {
	m := NewManager(1000)

	m.AddUserMessage("Message 1")
	m.AddAssistantMessage("Response 1")
	m.AddUserMessage("Message 2")

	history := m.GetHistory()
	if len(history) != 3 {
		t.Errorf("Expected 3 messages in history, got %d", len(history))
	}

	// Verify it's a copy
	history[0].Role = "modified"
	if m.messages[0].Role == "modified" {
		t.Error("GetHistory should return a copy, not the original")
	}
}

func TestManager_GetLastMessages(t *testing.T) {
	m := NewManager(1000)

	m.AddSystemMessage("System")
	m.AddUserMessage("User 1")
	m.AddAssistantMessage("Assistant 1")
	m.AddUserMessage("User 2")
	m.AddAssistantMessage("Assistant 2")

	// Test GetLastUserMessage
	lastUser := m.GetLastUserMessage()
	if lastUser == nil {
		t.Fatal("GetLastUserMessage returned nil")
	}
	if lastUser.Content[0].Text != "User 2" {
		t.Errorf("Expected 'User 2', got '%s'", lastUser.Content[0].Text)
	}

	// Test GetLastAssistantMessage
	lastAssistant := m.GetLastAssistantMessage()
	if lastAssistant == nil {
		t.Fatal("GetLastAssistantMessage returned nil")
	}
	if lastAssistant.Content[0].Text != "Assistant 2" {
		t.Errorf("Expected 'Assistant 2', got '%s'", lastAssistant.Content[0].Text)
	}

	// Test GetLastN
	lastThree := m.GetLastN(3)
	if len(lastThree) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(lastThree))
	}
}

func TestManager_Clear(t *testing.T) {
	m := NewManager(1000)

	m.AddUserMessage("Test")
	m.AddAssistantMessage("Response")

	m.Clear()
	if m.Count() != 0 {
		t.Error("Clear should remove all messages")
	}
	if m.GetTokenCount() != 0 {
		t.Error("Clear should reset token count")
	}
}

func TestManager_ClearExceptSystem(t *testing.T) {
	m := NewManager(1000)

	m.AddSystemMessage("System message")
	m.AddUserMessage("User message")
	m.AddAssistantMessage("Assistant message")

	m.ClearExceptSystem()

	if m.Count() != 1 {
		t.Errorf("Expected 1 message after clear, got %d", m.Count())
	}

	history := m.GetHistory()
	if history[0].Role != "system" {
		t.Error("Should keep system message")
	}
}

func TestManager_TokenEstimation(t *testing.T) {
	m := NewManager(1000)

	// Add message with known length
	m.AddUserMessage("1234") // 4 characters ≈ 1 token

	tokens := m.GetTokenCount()
	if tokens < 1 || tokens > 5 {
		t.Errorf("Token estimation seems off: %d tokens for '1234'", tokens)
	}

	// Test with tool use
	msg := types.Message{
		Role: "assistant",
		Content: []types.Content{
			{
				Type: "tool_use",
				ToolUse: &types.ToolUse{
					ID:   "test-id",
					Name: "test-tool",
					Input: map[string]interface{}{
						"param": "value",
					},
				},
			},
		},
	}
	_ = m.Add(msg)

	newTokens := m.GetTokenCount()
	if newTokens <= tokens {
		t.Error("Adding tool use should increase token count")
	}
}

func TestManager_ToolUse(t *testing.T) {
	m := NewManager(1000)

	// Add message without tool use
	m.AddUserMessage("Test")
	if m.HasToolUse() {
		t.Error("Should not have tool use")
	}

	// Add message with tool use
	msg := types.Message{
		Role: "assistant",
		Content: []types.Content{
			{
				Type: "tool_use",
				ToolUse: &types.ToolUse{
					ID:   "tool-1",
					Name: "test_tool",
					Input: map[string]interface{}{
						"param": "value",
					},
				},
			},
		},
	}
	_ = m.Add(msg)

	if !m.HasToolUse() {
		t.Error("Should have tool use")
	}

	toolUses := m.GetToolUses()
	if len(toolUses) != 1 {
		t.Errorf("Expected 1 tool use, got %d", len(toolUses))
	}
	if toolUses[0].Name != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got '%s'", toolUses[0].Name)
	}
}

func TestManager_Summary(t *testing.T) {
	m := NewManager(1000)

	m.AddSystemMessage("System")
	m.AddUserMessage("User 1")
	m.AddAssistantMessage("Assistant 1")
	m.AddUserMessage("User 2")

	// Add tool use
	msg := types.Message{
		Role: "assistant",
		Content: []types.Content{
			{
				Type: "tool_use",
				ToolUse: &types.ToolUse{
					ID:    "tool-1",
					Name:  "test_tool",
					Input: map[string]interface{}{},
				},
			},
		},
	}
	_ = m.Add(msg)

	// Add tool result
	result := types.Message{
		Role: "user",
		Content: []types.Content{
			{
				Type: "tool_result",
				ToolResult: &types.ToolResult{
					ToolUseID: "tool-1",
					Content:   "Tool output",
					IsError:   false,
				},
			},
		},
	}
	_ = m.Add(result)

	summary := m.Summary()

	// Check that summary contains expected counts
	if !contains(summary, "1 system") {
		t.Error("Summary should show system message count")
	}
	if !contains(summary, "3 user") {
		t.Error("Summary should show user message count")
	}
	if !contains(summary, "2 assistant") {
		t.Error("Summary should show assistant message count")
	}
	if !contains(summary, "1 tool uses") {
		t.Error("Summary should show tool use count")
	}
	if !contains(summary, "1 tool results") {
		t.Error("Summary should show tool result count")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}