package types

import (
	"encoding/json"
	"fmt"
)

// ToolUse represents a tool invocation request
type ToolUse struct {
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error,omitempty"`
}

// NewToolUse creates a new tool use with the given parameters
func NewToolUse(id, name string, input map[string]interface{}) *ToolUse {
	return &ToolUse{
		ID:    id,
		Name:  name,
		Input: input,
	}
}

// NewToolResult creates a new tool result
func NewToolResult(toolUseID, content string, isError bool) *ToolResult {
	return &ToolResult{
		ToolUseID: toolUseID,
		Content:   content,
		IsError:   isError,
	}
}

// Validate validates the tool use structure
func (t *ToolUse) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("tool use ID cannot be empty")
	}

	if t.Name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if t.Input == nil {
		return fmt.Errorf("tool input cannot be nil")
	}

	return nil
}

// Validate validates the tool result structure
func (r *ToolResult) Validate() error {
	if r.ToolUseID == "" {
		return fmt.Errorf("tool result must reference a tool use ID")
	}

	if r.Content == "" && !r.IsError {
		return fmt.Errorf("tool result content cannot be empty unless it's an error")
	}

	return nil
}

// GetString gets a string value from the input
func (t *ToolUse) GetString(key string) (string, error) {
	value, exists := t.Input[key]
	if !exists {
		return "", fmt.Errorf("key %s not found in input", key)
	}

	str, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("key %s is not a string", key)
	}

	return str, nil
}

// GetInt gets an integer value from the input
func (t *ToolUse) GetInt(key string) (int, error) {
	value, exists := t.Input[key]
	if !exists {
		return 0, fmt.Errorf("key %s not found in input", key)
	}

	// Handle both int and float64 (JSON numbers are float64)
	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case json.Number:
		i, err := v.Int64()
		return int(i), err
	default:
		return 0, fmt.Errorf("key %s is not a number", key)
	}
}

// GetBool gets a boolean value from the input
func (t *ToolUse) GetBool(key string) (bool, error) {
	value, exists := t.Input[key]
	if !exists {
		return false, fmt.Errorf("key %s not found in input", key)
	}

	b, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("key %s is not a boolean", key)
	}

	return b, nil
}

// GetMap gets a map value from the input
func (t *ToolUse) GetMap(key string) (map[string]interface{}, error) {
	value, exists := t.Input[key]
	if !exists {
		return nil, fmt.Errorf("key %s not found in input", key)
	}

	m, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("key %s is not a map", key)
	}

	return m, nil
}

// GetStringSlice gets a string slice from the input
func (t *ToolUse) GetStringSlice(key string) ([]string, error) {
	value, exists := t.Input[key]
	if !exists {
		return nil, fmt.Errorf("key %s not found in input", key)
	}

	// Handle interface{} slice
	slice, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("key %s is not a slice", key)
	}

	result := make([]string, 0, len(slice))
	for i, item := range slice {
		str, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("item %d in %s is not a string", i, key)
		}
		result = append(result, str)
	}

	return result, nil
}

// Success creates a successful tool result
func (t *ToolUse) Success(content string) *ToolResult {
	return &ToolResult{
		ToolUseID: t.ID,
		Content:   content,
		IsError:   false,
	}
}

// Error creates an error tool result
func (t *ToolUse) Error(err error) *ToolResult {
	return &ToolResult{
		ToolUseID: t.ID,
		Content:   fmt.Sprintf("Error: %v", err),
		IsError:   true,
	}
}

// MarshalJSON implements json.Marshaler
func (t ToolUse) MarshalJSON() ([]byte, error) {
	type Alias ToolUse
	return json.Marshal(&struct {
		Alias
	}{
		Alias: (Alias)(t),
	})
}

// UnmarshalJSON implements json.Unmarshaler
func (t *ToolUse) UnmarshalJSON(data []byte) error {
	type Alias ToolUse
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	return json.Unmarshal(data, &aux)
}