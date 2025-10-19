package tools

import (
	"context"
	"fmt"
)

// Tool represents a tool that can be executed by the agent
type Tool interface {
	// Name returns the unique name of the tool
	Name() string

	// Description returns a human-readable description of what the tool does
	Description() string

	// InputSchema returns the JSON schema for the tool's input parameters
	InputSchema() map[string]interface{}

	// Execute runs the tool with the given input and returns the result
	Execute(ctx context.Context, input map[string]interface{}) (string, error)

	// Validate checks if the input is valid for this tool
	Validate(input map[string]interface{}) error
}

// Registry manages the registration and retrieval of tools
type Registry interface {
	// Register adds a tool to the registry
	Register(tool Tool) error

	// Get retrieves a tool by name
	Get(name string) (Tool, error)

	// List returns all registered tools
	List() []Tool

	// Remove unregisters a tool by name
	Remove(name string) error

	// Has checks if a tool is registered
	Has(name string) bool

	// Clear removes all registered tools
	Clear()
}

// ToolRegistry is the default implementation of Registry
type ToolRegistry struct {
	tools map[string]Tool
}

// NewRegistry creates a new tool registry
func NewRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry
func (r *ToolRegistry) Register(tool Tool) error {
	if tool == nil {
		return fmt.Errorf("cannot register nil tool")
	}

	name := tool.Name()
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}

	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool %s is already registered", name)
	}

	r.tools[name] = tool
	return nil
}

// Get retrieves a tool by name
func (r *ToolRegistry) Get(name string) (Tool, error) {
	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}
	return tool, nil
}

// List returns all registered tools
func (r *ToolRegistry) List() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Remove unregisters a tool by name
func (r *ToolRegistry) Remove(name string) error {
	if _, exists := r.tools[name]; !exists {
		return fmt.Errorf("tool %s not found", name)
	}
	delete(r.tools, name)
	return nil
}

// Has checks if a tool is registered
func (r *ToolRegistry) Has(name string) bool {
	_, exists := r.tools[name]
	return exists
}

// Clear removes all registered tools
func (r *ToolRegistry) Clear() {
	r.tools = make(map[string]Tool)
}

// BaseTool provides common functionality for tools
type BaseTool struct {
	name        string
	description string
	schema      map[string]interface{}
}

// NewBaseTool creates a new base tool
func NewBaseTool(name, description string, schema map[string]interface{}) *BaseTool {
	return &BaseTool{
		name:        name,
		description: description,
		schema:      schema,
	}
}

// Name returns the tool name
func (t *BaseTool) Name() string {
	return t.name
}

// Description returns the tool description
func (t *BaseTool) Description() string {
	return t.description
}

// InputSchema returns the tool's input schema
func (t *BaseTool) InputSchema() map[string]interface{} {
	return t.schema
}

// Helper functions for building schemas

// StringSchema creates a string parameter schema
func StringSchema(description string, required bool) map[string]interface{} {
	schema := map[string]interface{}{
		"type":        "string",
		"description": description,
	}
	if required {
		schema["required"] = true
	}
	return schema
}

// IntegerSchema creates an integer parameter schema
func IntegerSchema(description string, required bool, min, max *int) map[string]interface{} {
	schema := map[string]interface{}{
		"type":        "integer",
		"description": description,
	}
	if required {
		schema["required"] = true
	}
	if min != nil {
		schema["minimum"] = *min
	}
	if max != nil {
		schema["maximum"] = *max
	}
	return schema
}

// BooleanSchema creates a boolean parameter schema
func BooleanSchema(description string, required bool) map[string]interface{} {
	schema := map[string]interface{}{
		"type":        "boolean",
		"description": description,
	}
	if required {
		schema["required"] = true
	}
	return schema
}

// ArraySchema creates an array parameter schema
func ArraySchema(description string, required bool, itemType string) map[string]interface{} {
	schema := map[string]interface{}{
		"type":        "array",
		"description": description,
		"items": map[string]interface{}{
			"type": itemType,
		},
	}
	if required {
		schema["required"] = true
	}
	return schema
}

// ObjectSchema creates an object parameter schema
func ObjectSchema(description string, required bool, properties map[string]interface{}) map[string]interface{} {
	schema := map[string]interface{}{
		"type":        "object",
		"description": description,
		"properties":  properties,
	}
	if required {
		schema["required"] = true
	}
	return schema
}

// BuildToolSchema builds a complete tool schema
func BuildToolSchema(properties map[string]interface{}, required []string) map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}