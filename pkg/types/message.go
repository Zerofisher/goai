package types

import (
	"encoding/json"
	"fmt"
)

// Message represents a message in the conversation between user, assistant, and system
type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

// Content represents a piece of content within a message
type Content struct {
	Type       string      `json:"type"`
	Text       string      `json:"text,omitempty"`
	ToolUse    *ToolUse    `json:"tool_use,omitempty"`
	ToolResult *ToolResult `json:"tool_result,omitempty"`
}

// NewTextMessage creates a new message with text content
func NewTextMessage(role, text string) Message {
	return Message{
		Role: role,
		Content: []Content{
			{
				Type: "text",
				Text: text,
			},
		},
	}
}

// NewToolUseMessage creates a new message with tool use content
func NewToolUseMessage(toolUse *ToolUse) Message {
	return Message{
		Role: "assistant",
		Content: []Content{
			{
				Type:    "tool_use",
				ToolUse: toolUse,
			},
		},
	}
}

// NewToolResultMessage creates a new message with tool result content
func NewToolResultMessage(toolResult *ToolResult) Message {
	return Message{
		Role: "tool",
		Content: []Content{
			{
				Type:       "tool_result",
				ToolResult: toolResult,
			},
		},
	}
}

// AddContent adds content to the message
func (m *Message) AddContent(content Content) {
	m.Content = append(m.Content, content)
}

// GetText returns all text content concatenated
func (m *Message) GetText() string {
	var text string
	for _, content := range m.Content {
		if content.Type == "text" && content.Text != "" {
			text += content.Text
		}
	}
	return text
}

// HasToolUse checks if the message contains any tool use
func (m *Message) HasToolUse() bool {
	for _, content := range m.Content {
		if content.Type == "tool_use" && content.ToolUse != nil {
			return true
		}
	}
	return false
}

// GetToolUses returns all tool uses in the message
func (m *Message) GetToolUses() []*ToolUse {
	var toolUses []*ToolUse
	for _, content := range m.Content {
		if content.Type == "tool_use" && content.ToolUse != nil {
			toolUses = append(toolUses, content.ToolUse)
		}
	}
	return toolUses
}

// Validate validates the message structure
func (m *Message) Validate() error {
	if m.Role == "" {
		return fmt.Errorf("message role cannot be empty")
	}

	if m.Role != "user" && m.Role != "assistant" && m.Role != "system" && m.Role != "tool" {
		return fmt.Errorf("invalid message role: %s", m.Role)
	}

	if len(m.Content) == 0 {
		return fmt.Errorf("message must have at least one content")
	}

	for i, content := range m.Content {
		if err := content.Validate(); err != nil {
			return fmt.Errorf("content[%d]: %w", i, err)
		}
	}

	return nil
}

// Validate validates the content structure
func (c *Content) Validate() error {
	if c.Type == "" {
		return fmt.Errorf("content type cannot be empty")
	}

	switch c.Type {
	case "text":
		if c.Text == "" {
			return fmt.Errorf("text content cannot be empty")
		}
	case "tool_use":
		if c.ToolUse == nil {
			return fmt.Errorf("tool_use content cannot be nil")
		}
		if err := c.ToolUse.Validate(); err != nil {
			return fmt.Errorf("tool_use validation: %w", err)
		}
	case "tool_result":
		if c.ToolResult == nil {
			return fmt.Errorf("tool_result content cannot be nil")
		}
		if err := c.ToolResult.Validate(); err != nil {
			return fmt.Errorf("tool_result validation: %w", err)
		}
	default:
		return fmt.Errorf("unknown content type: %s", c.Type)
	}

	return nil
}

// MarshalJSON implements json.Marshaler
func (m Message) MarshalJSON() ([]byte, error) {
	type Alias Message
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(m),
	})
}

// UnmarshalJSON implements json.Unmarshaler
func (m *Message) UnmarshalJSON(data []byte) error {
	type Alias Message
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	}
	return json.Unmarshal(data, &aux)
}