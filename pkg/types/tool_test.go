package types

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestToolUse_NewToolUse(t *testing.T) {
	input := map[string]interface{}{
		"path": "/test/file.txt",
		"line": 10,
	}

	toolUse := NewToolUse("test_id", "read_file", input)

	if toolUse.ID != "test_id" {
		t.Errorf("expected ID 'test_id', got '%s'", toolUse.ID)
	}

	if toolUse.Name != "read_file" {
		t.Errorf("expected name 'read_file', got '%s'", toolUse.Name)
	}

	if toolUse.Input["path"] != "/test/file.txt" {
		t.Errorf("expected input path '/test/file.txt', got '%v'", toolUse.Input["path"])
	}
}

func TestToolResult_NewToolResult(t *testing.T) {
	result := NewToolResult("test_id", "Success", false)

	if result.ToolUseID != "test_id" {
		t.Errorf("expected ToolUseID 'test_id', got '%s'", result.ToolUseID)
	}

	if result.Content != "Success" {
		t.Errorf("expected content 'Success', got '%s'", result.Content)
	}

	if result.IsError {
		t.Error("expected IsError to be false")
	}
}

func TestToolUse_Validate(t *testing.T) {
	tests := []struct {
		name    string
		toolUse ToolUse
		wantErr bool
	}{
		{
			name: "valid tool use",
			toolUse: ToolUse{
				ID:    "test",
				Name:  "tool",
				Input: map[string]interface{}{},
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			toolUse: ToolUse{
				ID:    "",
				Name:  "tool",
				Input: map[string]interface{}{},
			},
			wantErr: true,
		},
		{
			name: "empty name",
			toolUse: ToolUse{
				ID:    "test",
				Name:  "",
				Input: map[string]interface{}{},
			},
			wantErr: true,
		},
		{
			name: "nil input",
			toolUse: ToolUse{
				ID:    "test",
				Name:  "tool",
				Input: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.toolUse.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestToolResult_Validate(t *testing.T) {
	tests := []struct {
		name       string
		toolResult ToolResult
		wantErr    bool
	}{
		{
			name: "valid result",
			toolResult: ToolResult{
				ToolUseID: "test",
				Content:   "Success",
				IsError:   false,
			},
			wantErr: false,
		},
		{
			name: "empty tool use ID",
			toolResult: ToolResult{
				ToolUseID: "",
				Content:   "Success",
				IsError:   false,
			},
			wantErr: true,
		},
		{
			name: "empty content non-error",
			toolResult: ToolResult{
				ToolUseID: "test",
				Content:   "",
				IsError:   false,
			},
			wantErr: true,
		},
		{
			name: "empty content with error",
			toolResult: ToolResult{
				ToolUseID: "test",
				Content:   "",
				IsError:   true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.toolResult.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestToolUse_GetString(t *testing.T) {
	toolUse := &ToolUse{
		ID:   "test",
		Name: "tool",
		Input: map[string]interface{}{
			"string_key": "value",
			"number_key": 123,
		},
	}

	// Test existing string key
	val, err := toolUse.GetString("string_key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val != "value" {
		t.Errorf("expected 'value', got '%s'", val)
	}

	// Test non-existent key
	_, err = toolUse.GetString("missing")
	if err == nil {
		t.Error("expected error for missing key")
	}

	// Test wrong type
	_, err = toolUse.GetString("number_key")
	if err == nil {
		t.Error("expected error for wrong type")
	}
}

func TestToolUse_GetInt(t *testing.T) {
	toolUse := &ToolUse{
		ID:   "test",
		Name: "tool",
		Input: map[string]interface{}{
			"int_key":     42,
			"float_key":   42.5,
			"string_key":  "not_a_number",
			"json_number": json.Number("100"),
		},
	}

	// Test int value
	val, err := toolUse.GetInt("int_key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val != 42 {
		t.Errorf("expected 42, got %d", val)
	}

	// Test float64 value
	val, err = toolUse.GetInt("float_key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val != 42 {
		t.Errorf("expected 42, got %d", val)
	}

	// Test json.Number
	val, err = toolUse.GetInt("json_number")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val != 100 {
		t.Errorf("expected 100, got %d", val)
	}

	// Test non-existent key
	_, err = toolUse.GetInt("missing")
	if err == nil {
		t.Error("expected error for missing key")
	}

	// Test wrong type
	_, err = toolUse.GetInt("string_key")
	if err == nil {
		t.Error("expected error for wrong type")
	}
}

func TestToolUse_GetBool(t *testing.T) {
	toolUse := &ToolUse{
		ID:   "test",
		Name: "tool",
		Input: map[string]interface{}{
			"bool_true":  true,
			"bool_false": false,
			"string_key": "not_a_bool",
		},
	}

	// Test true value
	val, err := toolUse.GetBool("bool_true")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !val {
		t.Error("expected true")
	}

	// Test false value
	val, err = toolUse.GetBool("bool_false")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val {
		t.Error("expected false")
	}

	// Test non-existent key
	_, err = toolUse.GetBool("missing")
	if err == nil {
		t.Error("expected error for missing key")
	}

	// Test wrong type
	_, err = toolUse.GetBool("string_key")
	if err == nil {
		t.Error("expected error for wrong type")
	}
}

func TestToolUse_GetMap(t *testing.T) {
	innerMap := map[string]interface{}{
		"nested": "value",
	}

	toolUse := &ToolUse{
		ID:   "test",
		Name: "tool",
		Input: map[string]interface{}{
			"map_key":    innerMap,
			"string_key": "not_a_map",
		},
	}

	// Test valid map
	val, err := toolUse.GetMap("map_key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if val["nested"] != "value" {
		t.Errorf("expected nested value 'value', got '%v'", val["nested"])
	}

	// Test non-existent key
	_, err = toolUse.GetMap("missing")
	if err == nil {
		t.Error("expected error for missing key")
	}

	// Test wrong type
	_, err = toolUse.GetMap("string_key")
	if err == nil {
		t.Error("expected error for wrong type")
	}
}

func TestToolUse_GetStringSlice(t *testing.T) {
	toolUse := &ToolUse{
		ID:   "test",
		Name: "tool",
		Input: map[string]interface{}{
			"string_slice": []interface{}{"a", "b", "c"},
			"mixed_slice":  []interface{}{"a", 1, "c"},
			"string_key":   "not_a_slice",
		},
	}

	// Test valid string slice
	val, err := toolUse.GetStringSlice("string_slice")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(val) != 3 {
		t.Fatalf("expected 3 items, got %d", len(val))
	}
	if val[0] != "a" || val[1] != "b" || val[2] != "c" {
		t.Errorf("unexpected values: %v", val)
	}

	// Test non-existent key
	_, err = toolUse.GetStringSlice("missing")
	if err == nil {
		t.Error("expected error for missing key")
	}

	// Test wrong type
	_, err = toolUse.GetStringSlice("string_key")
	if err == nil {
		t.Error("expected error for wrong type")
	}

	// Test mixed slice
	_, err = toolUse.GetStringSlice("mixed_slice")
	if err == nil {
		t.Error("expected error for mixed slice")
	}
}

func TestToolUse_Success(t *testing.T) {
	toolUse := &ToolUse{
		ID:   "test_id",
		Name: "tool",
		Input: map[string]interface{}{},
	}

	result := toolUse.Success("Operation completed")

	if result.ToolUseID != "test_id" {
		t.Errorf("expected ToolUseID 'test_id', got '%s'", result.ToolUseID)
	}

	if result.Content != "Operation completed" {
		t.Errorf("expected content 'Operation completed', got '%s'", result.Content)
	}

	if result.IsError {
		t.Error("expected IsError to be false")
	}
}

func TestToolUse_Error(t *testing.T) {
	toolUse := &ToolUse{
		ID:   "test_id",
		Name: "tool",
		Input: map[string]interface{}{},
	}

	testErr := errors.New("something went wrong")
	result := toolUse.Error(testErr)

	if result.ToolUseID != "test_id" {
		t.Errorf("expected ToolUseID 'test_id', got '%s'", result.ToolUseID)
	}

	if result.Content != "Error: something went wrong" {
		t.Errorf("expected content 'Error: something went wrong', got '%s'", result.Content)
	}

	if !result.IsError {
		t.Error("expected IsError to be true")
	}
}

func TestToolUse_JSONSerialization(t *testing.T) {
	original := ToolUse{
		ID:   "test_id",
		Name: "test_tool",
		Input: map[string]interface{}{
			"string": "value",
			"number": float64(42),
			"bool":   true,
			"map": map[string]interface{}{
				"nested": "data",
			},
		},
	}

	// Marshal
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal tool use: %v", err)
	}

	// Unmarshal
	var decoded ToolUse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("failed to unmarshal tool use: %v", err)
	}

	// Compare
	if decoded.ID != original.ID {
		t.Errorf("ID mismatch: got %s, want %s", decoded.ID, original.ID)
	}

	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %s, want %s", decoded.Name, original.Name)
	}

	if decoded.Input["string"] != original.Input["string"] {
		t.Errorf("Input string mismatch: got %v, want %v", decoded.Input["string"], original.Input["string"])
	}

	if decoded.Input["number"] != original.Input["number"] {
		t.Errorf("Input number mismatch: got %v, want %v", decoded.Input["number"], original.Input["number"])
	}
}