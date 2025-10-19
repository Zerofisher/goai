package message

import (
	"fmt"
	"strings"

	"github.com/Zerofisher/goai/pkg/types"
)

// Manager manages the conversation message history
type Manager struct {
	messages   []types.Message
	maxTokens  int
	tokenCount int
}

// NewManager creates a new message manager
func NewManager(maxTokens int) *Manager {
	if maxTokens <= 0 {
		maxTokens = 16000 // Default to 16k tokens
	}
	return &Manager{
		messages:  []types.Message{},
		maxTokens: maxTokens,
	}
}

// Add adds a message to the history
func (m *Manager) Add(msg types.Message) error {
	if err := msg.Validate(); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}

	m.messages = append(m.messages, msg)
	m.tokenCount += m.estimateTokens(msg)

	// Truncate if necessary
	m.Truncate()

	return nil
}

// AddUserMessage adds a user message
func (m *Manager) AddUserMessage(text string) {
	_ = m.Add(types.NewTextMessage("user", text))
}

// AddAssistantMessage adds an assistant message
func (m *Manager) AddAssistantMessage(text string) {
	_ = m.Add(types.NewTextMessage("assistant", text))
}

// AddSystemMessage adds a system message
func (m *Manager) AddSystemMessage(text string) {
	_ = m.Add(types.NewTextMessage("system", text))
}

// Truncate removes old messages to stay within token limit
func (m *Manager) Truncate() {
	// Keep at least the last 2 messages
	if len(m.messages) <= 2 {
		return
	}

	// Estimate current token count
	totalTokens := 0
	for _, msg := range m.messages {
		totalTokens += m.estimateTokens(msg)
	}

	// Remove messages from the beginning if over limit
	for totalTokens > m.maxTokens && len(m.messages) > 2 {
		// Keep the first message if it's a system message
		startIdx := 0
		if len(m.messages) > 0 && m.messages[0].Role == "system" {
			startIdx = 1
		}

		if startIdx < len(m.messages) {
			removed := m.messages[startIdx]
			totalTokens -= m.estimateTokens(removed)

			// Remove the message
			m.messages = append(m.messages[:startIdx], m.messages[startIdx+1:]...)
		} else {
			break
		}
	}

	m.tokenCount = totalTokens
}

// GetHistory returns the message history
func (m *Manager) GetHistory() []types.Message {
	return append([]types.Message{}, m.messages...)
}

// GetLastN returns the last N messages
func (m *Manager) GetLastN(n int) []types.Message {
	if n <= 0 || n > len(m.messages) {
		return m.GetHistory()
	}
	return append([]types.Message{}, m.messages[len(m.messages)-n:]...)
}

// GetLastUserMessage returns the last user message
func (m *Manager) GetLastUserMessage() *types.Message {
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].Role == "user" {
			msg := m.messages[i]
			return &msg
		}
	}
	return nil
}

// GetLastAssistantMessage returns the last assistant message
func (m *Manager) GetLastAssistantMessage() *types.Message {
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].Role == "assistant" {
			msg := m.messages[i]
			return &msg
		}
	}
	return nil
}

// Clear clears all messages
func (m *Manager) Clear() {
	m.messages = []types.Message{}
	m.tokenCount = 0
}

// ClearExceptSystem clears all messages except system messages
func (m *Manager) ClearExceptSystem() {
	newMessages := []types.Message{}
	for _, msg := range m.messages {
		if msg.Role == "system" {
			newMessages = append(newMessages, msg)
		}
	}
	m.messages = newMessages
	m.recalculateTokens()
}

// Count returns the number of messages
func (m *Manager) Count() int {
	return len(m.messages)
}

// GetTokenCount returns the estimated token count
func (m *Manager) GetTokenCount() int {
	return m.tokenCount
}

// SetMaxTokens sets the maximum token limit
func (m *Manager) SetMaxTokens(maxTokens int) {
	m.maxTokens = maxTokens
	m.Truncate()
}

// estimateTokens estimates the token count for a message
// This is a rough estimation: ~4 characters per token
func (m *Manager) estimateTokens(msg types.Message) int {
	charCount := 0

	// Count role
	charCount += len(msg.Role)

	// Count content
	for _, content := range msg.Content {
		switch content.Type {
		case "text":
			charCount += len(content.Text)
		case "tool_use":
			if content.ToolUse != nil {
				charCount += len(content.ToolUse.Name)
				charCount += len(content.ToolUse.ID)
				// Rough estimate for input map
				charCount += len(fmt.Sprintf("%v", content.ToolUse.Input))
			}
		case "tool_result":
			if content.ToolResult != nil {
				charCount += len(content.ToolResult.ToolUseID)
				charCount += len(content.ToolResult.Content)
			}
		}
	}

	// Estimate tokens (roughly 4 characters per token)
	return (charCount + 3) / 4
}

// recalculateTokens recalculates the total token count
func (m *Manager) recalculateTokens() {
	m.tokenCount = 0
	for _, msg := range m.messages {
		m.tokenCount += m.estimateTokens(msg)
	}
}

// HasToolUse checks if there are any messages with tool uses
func (m *Manager) HasToolUse() bool {
	for _, msg := range m.messages {
		if msg.HasToolUse() {
			return true
		}
	}
	return false
}

// GetToolUses returns all tool uses in the history
func (m *Manager) GetToolUses() []*types.ToolUse {
	var toolUses []*types.ToolUse
	for _, msg := range m.messages {
		toolUses = append(toolUses, msg.GetToolUses()...)
	}
	return toolUses
}

// Summary returns a summary of the message history
func (m *Manager) Summary() string {
	var parts []string

	userCount := 0
	assistantCount := 0
	systemCount := 0
	toolUseCount := 0
	toolResultCount := 0

	for _, msg := range m.messages {
		switch msg.Role {
		case "user":
			userCount++
			// Count tool results in user messages
			for _, content := range msg.Content {
				if content.Type == "tool_result" {
					toolResultCount++
				}
			}
		case "assistant":
			assistantCount++
			// Count tool uses in assistant messages
			for _, content := range msg.Content {
				if content.Type == "tool_use" {
					toolUseCount++
				}
			}
		case "system":
			systemCount++
		}
	}

	if systemCount > 0 {
		parts = append(parts, fmt.Sprintf("%d system", systemCount))
	}
	if userCount > 0 {
		parts = append(parts, fmt.Sprintf("%d user", userCount))
	}
	if assistantCount > 0 {
		parts = append(parts, fmt.Sprintf("%d assistant", assistantCount))
	}
	if toolUseCount > 0 {
		parts = append(parts, fmt.Sprintf("%d tool uses", toolUseCount))
	}
	if toolResultCount > 0 {
		parts = append(parts, fmt.Sprintf("%d tool results", toolResultCount))
	}

	return fmt.Sprintf("Messages: %s (≈%d tokens)", strings.Join(parts, ", "), m.tokenCount)
}