package message

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Zerofisher/goai/pkg/types"
)

// Normalizer handles content normalization for messages.
type Normalizer struct {
	maxContentLength int
	stripANSI        bool
}

// NewNormalizer creates a new content normalizer.
func NewNormalizer() *Normalizer {
	return &Normalizer{
		maxContentLength: 100000, // Default max content length
		stripANSI:        true,
	}
}

// NormalizeMessage normalizes a single message.
func (n *Normalizer) NormalizeMessage(msg types.Message) types.Message {
	normalized := types.Message{
		Role:    msg.Role,
		Content: make([]types.Content, 0, len(msg.Content)),
	}

	for _, content := range msg.Content {
		normalized.Content = append(normalized.Content, n.normalizeContent(content))
	}

	return normalized
}

// normalizeContent normalizes a single content item.
func (n *Normalizer) normalizeContent(content types.Content) types.Content {
	normalized := types.Content{
		Type: content.Type,
	}

	switch content.Type {
	case "text":
		normalized.Text = n.normalizeText(content.Text)
	case "tool_use":
		if content.ToolUse != nil {
			normalized.ToolUse = &types.ToolUse{
				ID:    content.ToolUse.ID,
				Name:  n.normalizeToolName(content.ToolUse.Name),
				Input: n.normalizeToolInput(content.ToolUse.Input),
			}
		}
	case "tool_result":
		if content.ToolResult != nil {
			normalized.ToolResult = &types.ToolResult{
				ToolUseID: content.ToolResult.ToolUseID,
				Content:   n.normalizeText(content.ToolResult.Content),
				IsError:   content.ToolResult.IsError,
			}
		}
	}

	return normalized
}

// normalizeText normalizes text content.
func (n *Normalizer) normalizeText(text string) string {
	// Remove null bytes
	text = strings.ReplaceAll(text, "\x00", "")

	// Strip ANSI escape codes if needed
	if n.stripANSI {
		text = n.stripANSICodes(text)
	}

	// Ensure valid UTF-8
	if !utf8.ValidString(text) {
		text = n.sanitizeUTF8(text)
	}

	// Truncate if too long
	if n.maxContentLength > 0 && len(text) > n.maxContentLength {
		text = text[:n.maxContentLength] + "\n... [truncated]"
	}

	// Normalize line endings
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	return text
}

// normalizeToolName normalizes a tool name.
func (n *Normalizer) normalizeToolName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace spaces with underscores
	name = strings.ReplaceAll(name, " ", "_")

	// Remove special characters
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// normalizeToolInput normalizes tool input parameters.
func (n *Normalizer) normalizeToolInput(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return make(map[string]interface{})
	}

	normalized := make(map[string]interface{})
	for key, value := range input {
		// Normalize the key
		normalizedKey := n.normalizeKey(key)

		// Normalize the value
		normalizedValue := n.normalizeValue(value)

		normalized[normalizedKey] = normalizedValue
	}

	return normalized
}

// normalizeKey normalizes a parameter key.
func (n *Normalizer) normalizeKey(key string) string {
	// Convert to lowercase
	key = strings.ToLower(key)

	// Replace spaces and hyphens with underscores
	key = strings.ReplaceAll(key, " ", "_")
	key = strings.ReplaceAll(key, "-", "_")

	// Remove special characters
	var result strings.Builder
	for _, r := range key {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// normalizeValue normalizes a parameter value.
func (n *Normalizer) normalizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return n.normalizeText(v)
	case []interface{}:
		normalized := make([]interface{}, len(v))
		for i, item := range v {
			normalized[i] = n.normalizeValue(item)
		}
		return normalized
	case map[string]interface{}:
		return n.normalizeToolInput(v)
	default:
		return value
	}
}

// stripANSICodes removes ANSI escape codes from text.
func (n *Normalizer) stripANSICodes(text string) string {
	var result strings.Builder
	inEscape := false

	for _, r := range text {
		if r == '\033' {
			inEscape = true
		} else if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// sanitizeUTF8 ensures the string is valid UTF-8.
func (n *Normalizer) sanitizeUTF8(text string) string {
	var result strings.Builder
	for _, r := range text {
		if r == utf8.RuneError {
			result.WriteRune('?')
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ConvertFromSDK converts SDK objects to our message format.
func (n *Normalizer) ConvertFromSDK(sdkMessage interface{}) (*types.Message, error) {
	// This would handle conversion from various SDK formats (OpenAI, Claude, etc.)
	// For now, we'll just handle JSON conversion
	data, err := json.Marshal(sdkMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SDK message: %w", err)
	}

	var msg types.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to message: %w", err)
	}

	normalized := n.NormalizeMessage(msg)
	return &normalized, nil
}

// ConvertToSDK converts our message format to SDK objects.
func (n *Normalizer) ConvertToSDK(msg types.Message, sdkType string) (interface{}, error) {
	// Normalize first
	normalized := n.NormalizeMessage(msg)

	// Convert based on SDK type
	switch sdkType {
	case "openai":
		return n.convertToOpenAI(normalized), nil
	case "claude":
		return n.convertToClaude(normalized), nil
	default:
		// Generic JSON conversion
		return normalized, nil
	}
}

// convertToOpenAI converts to OpenAI message format.
func (n *Normalizer) convertToOpenAI(msg types.Message) map[string]interface{} {
	openAIMsg := map[string]interface{}{
		"role": msg.Role,
	}

	// Handle content
	if len(msg.Content) == 1 && msg.Content[0].Type == "text" {
		openAIMsg["content"] = msg.Content[0].Text
	} else {
		var content []map[string]interface{}
		for _, c := range msg.Content {
			switch c.Type {
			case "text":
				content = append(content, map[string]interface{}{
					"type": "text",
					"text": c.Text,
				})
			case "tool_use":
				if c.ToolUse != nil {
					openAIMsg["tool_calls"] = []map[string]interface{}{
						{
							"id":   c.ToolUse.ID,
							"type": "function",
							"function": map[string]interface{}{
								"name":      c.ToolUse.Name,
								"arguments": c.ToolUse.Input,
							},
						},
					}
				}
			case "tool_result":
				if c.ToolResult != nil {
					openAIMsg["role"] = "tool"
					openAIMsg["tool_call_id"] = c.ToolResult.ToolUseID
					openAIMsg["content"] = c.ToolResult.Content
				}
			}
		}
		if len(content) > 0 {
			openAIMsg["content"] = content
		}
	}

	return openAIMsg
}

// convertToClaude converts to Claude message format.
func (n *Normalizer) convertToClaude(msg types.Message) map[string]interface{} {
	// Claude uses a similar format to our internal format
	return map[string]interface{}{
		"role":    msg.Role,
		"content": msg.Content,
	}
}

// SetMaxContentLength sets the maximum content length.
func (n *Normalizer) SetMaxContentLength(length int) {
	if length > 0 {
		n.maxContentLength = length
	}
}

// SetStripANSI sets whether to strip ANSI codes.
func (n *Normalizer) SetStripANSI(strip bool) {
	n.stripANSI = strip
}