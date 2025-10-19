package tools

import (
	"context"
	"errors"
	"testing"
)

// MockTool is a mock implementation of the Tool interface for testing
type MockTool struct {
	name        string
	description string
	schema      map[string]interface{}
	executeFunc func(ctx context.Context, input map[string]interface{}) (string, error)
	validateFunc func(input map[string]interface{}) error
}

func (m *MockTool) Name() string {
	return m.name
}

func (m *MockTool) Description() string {
	return m.description
}

func (m *MockTool) InputSchema() map[string]interface{} {
	return m.schema
}

func (m *MockTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, input)
	}
	return "mock result", nil
}

func (m *MockTool) Validate(input map[string]interface{}) error {
	if m.validateFunc != nil {
		return m.validateFunc(input)
	}
	return nil
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if registry.tools == nil {
		t.Fatal("Registry tools map not initialized")
	}
	if len(registry.tools) != 0 {
		t.Errorf("Registry should start empty, got %d tools", len(registry.tools))
	}
}

func TestToolRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	tests := []struct {
		name    string
		tool    Tool
		wantErr bool
		errMsg  string
	}{
		{
			name: "register valid tool",
			tool: &MockTool{
				name:        "test_tool",
				description: "A test tool",
			},
			wantErr: false,
		},
		{
			name:    "register nil tool",
			tool:    nil,
			wantErr: true,
			errMsg:  "cannot register nil tool",
		},
		{
			name: "register tool with empty name",
			tool: &MockTool{
				name:        "",
				description: "Tool with no name",
			},
			wantErr: true,
			errMsg:  "tool name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.Register(tt.tool)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Register() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}

	// Test duplicate registration
	tool := &MockTool{name: "duplicate", description: "Test"}
	if err := registry.Register(tool); err != nil {
		t.Errorf("First registration failed: %v", err)
	}
	if err := registry.Register(tool); err == nil {
		t.Error("Duplicate registration should fail")
	}
}

func TestToolRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	tool := &MockTool{
		name:        "test_tool",
		description: "A test tool",
	}
	registry.Register(tool)

	tests := []struct {
		name     string
		toolName string
		wantErr  bool
	}{
		{
			name:     "get existing tool",
			toolName: "test_tool",
			wantErr:  false,
		},
		{
			name:     "get non-existent tool",
			toolName: "non_existent",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := registry.Get(tt.toolName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got.Name() != tt.toolName {
				t.Errorf("Get() returned wrong tool, got %v, want %v", got.Name(), tt.toolName)
			}
		})
	}
}

func TestToolRegistry_List(t *testing.T) {
	registry := NewRegistry()

	// Empty registry
	tools := registry.List()
	if len(tools) != 0 {
		t.Errorf("Empty registry should return empty list, got %d tools", len(tools))
	}

	// Add tools
	tool1 := &MockTool{name: "tool1", description: "First tool"}
	tool2 := &MockTool{name: "tool2", description: "Second tool"}
	registry.Register(tool1)
	registry.Register(tool2)

	tools = registry.List()
	if len(tools) != 2 {
		t.Errorf("Registry should have 2 tools, got %d", len(tools))
	}
}

func TestToolRegistry_Remove(t *testing.T) {
	registry := NewRegistry()
	tool := &MockTool{name: "test_tool", description: "Test"}
	registry.Register(tool)

	// Remove existing tool
	err := registry.Remove("test_tool")
	if err != nil {
		t.Errorf("Remove() failed: %v", err)
	}

	// Verify tool is removed
	if registry.Has("test_tool") {
		t.Error("Tool should be removed")
	}

	// Remove non-existent tool
	err = registry.Remove("non_existent")
	if err == nil {
		t.Error("Removing non-existent tool should fail")
	}
}

func TestToolRegistry_Has(t *testing.T) {
	registry := NewRegistry()
	tool := &MockTool{name: "test_tool", description: "Test"}
	registry.Register(tool)

	if !registry.Has("test_tool") {
		t.Error("Has() should return true for existing tool")
	}

	if registry.Has("non_existent") {
		t.Error("Has() should return false for non-existent tool")
	}
}

func TestToolRegistry_Clear(t *testing.T) {
	registry := NewRegistry()
	tool1 := &MockTool{name: "tool1", description: "First"}
	tool2 := &MockTool{name: "tool2", description: "Second"}
	registry.Register(tool1)
	registry.Register(tool2)

	registry.Clear()

	if len(registry.tools) != 0 {
		t.Errorf("Clear() should remove all tools, got %d", len(registry.tools))
	}

	if registry.Has("tool1") || registry.Has("tool2") {
		t.Error("Clear() should remove all tools")
	}
}

func TestBaseTool(t *testing.T) {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"input": map[string]interface{}{
				"type": "string",
			},
		},
	}

	baseTool := NewBaseTool("test_tool", "A test tool", schema)

	if baseTool.Name() != "test_tool" {
		t.Errorf("Name() = %v, want test_tool", baseTool.Name())
	}

	if baseTool.Description() != "A test tool" {
		t.Errorf("Description() = %v, want 'A test tool'", baseTool.Description())
	}

	if baseTool.InputSchema()["type"] != "object" {
		t.Error("InputSchema() should return the provided schema")
	}
}

func TestSchemaHelpers(t *testing.T) {
	t.Run("StringSchema", func(t *testing.T) {
		schema := StringSchema("A string parameter", true)
		if schema["type"] != "string" {
			t.Errorf("Type should be 'string', got %v", schema["type"])
		}
		if schema["description"] != "A string parameter" {
			t.Errorf("Description mismatch")
		}
		if schema["required"] != true {
			t.Error("Required should be true")
		}

		schema = StringSchema("Optional string", false)
		if _, exists := schema["required"]; exists {
			t.Error("Optional parameter should not have required field")
		}
	})

	t.Run("IntegerSchema", func(t *testing.T) {
		min := 0
		max := 100
		schema := IntegerSchema("An integer", true, &min, &max)
		if schema["type"] != "integer" {
			t.Errorf("Type should be 'integer', got %v", schema["type"])
		}
		if schema["minimum"] != 0 {
			t.Errorf("Minimum should be 0, got %v", schema["minimum"])
		}
		if schema["maximum"] != 100 {
			t.Errorf("Maximum should be 100, got %v", schema["maximum"])
		}

		// Test without min/max
		schema = IntegerSchema("An integer", false, nil, nil)
		if _, exists := schema["minimum"]; exists {
			t.Error("Should not have minimum when nil")
		}
		if _, exists := schema["maximum"]; exists {
			t.Error("Should not have maximum when nil")
		}
	})

	t.Run("BooleanSchema", func(t *testing.T) {
		schema := BooleanSchema("A boolean", true)
		if schema["type"] != "boolean" {
			t.Errorf("Type should be 'boolean', got %v", schema["type"])
		}
	})

	t.Run("ArraySchema", func(t *testing.T) {
		schema := ArraySchema("An array of strings", true, "string")
		if schema["type"] != "array" {
			t.Errorf("Type should be 'array', got %v", schema["type"])
		}
		items := schema["items"].(map[string]interface{})
		if items["type"] != "string" {
			t.Errorf("Items type should be 'string', got %v", items["type"])
		}
	})

	t.Run("ObjectSchema", func(t *testing.T) {
		props := map[string]interface{}{
			"field1": StringSchema("Field 1", true),
			"field2": IntegerSchema("Field 2", false, nil, nil),
		}
		schema := ObjectSchema("An object", true, props)
		if schema["type"] != "object" {
			t.Errorf("Type should be 'object', got %v", schema["type"])
		}
		if schema["properties"] == nil {
			t.Error("Properties should not be nil")
		}
	})

	t.Run("BuildToolSchema", func(t *testing.T) {
		props := map[string]interface{}{
			"param1": StringSchema("Parameter 1", true),
			"param2": BooleanSchema("Parameter 2", false),
		}
		required := []string{"param1"}

		schema := BuildToolSchema(props, required)
		if schema["type"] != "object" {
			t.Errorf("Type should be 'object', got %v", schema["type"])
		}
		if schema["properties"] == nil {
			t.Error("Properties should not be nil")
		}
		req := schema["required"].([]string)
		if len(req) != 1 || req[0] != "param1" {
			t.Errorf("Required should be ['param1'], got %v", req)
		}
	})
}

// TestToolIntegration tests the integration of tools with the registry
func TestToolIntegration(t *testing.T) {
	registry := NewRegistry()

	// Create a custom tool
	customTool := &MockTool{
		name:        "calculator",
		description: "Performs calculations",
		schema: BuildToolSchema(
			map[string]interface{}{
				"expression": StringSchema("Math expression to evaluate", true),
			},
			[]string{"expression"},
		),
		executeFunc: func(ctx context.Context, input map[string]interface{}) (string, error) {
			expr, ok := input["expression"].(string)
			if !ok {
				return "", errors.New("expression must be a string")
			}
			// Simplified calculation (in real implementation, use a proper parser)
			if expr == "2+2" {
				return "4", nil
			}
			return "", errors.New("unsupported expression")
		},
		validateFunc: func(input map[string]interface{}) error {
			if _, ok := input["expression"]; !ok {
				return errors.New("expression is required")
			}
			return nil
		},
	}

	// Register the tool
	if err := registry.Register(customTool); err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	// Get and execute the tool
	tool, err := registry.Get("calculator")
	if err != nil {
		t.Fatalf("Failed to get tool: %v", err)
	}

	// Validate input
	input := map[string]interface{}{"expression": "2+2"}
	if err := tool.Validate(input); err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Execute the tool
	ctx := context.Background()
	result, err := tool.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if result != "4" {
		t.Errorf("Expected result '4', got '%s'", result)
	}

	// Test invalid input
	invalidInput := map[string]interface{}{}
	if err := tool.Validate(invalidInput); err == nil {
		t.Error("Validation should fail for invalid input")
	}
}