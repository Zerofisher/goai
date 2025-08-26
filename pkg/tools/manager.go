package tools

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// DefaultToolManager is the default implementation of ToolManager
type DefaultToolManager struct {
	registry           ToolRegistry
	confirmationHandler ConfirmationHandler
	mu                 sync.RWMutex
}

// NewToolManager creates a new tool manager instance
func NewToolManager(registry ToolRegistry, confirmationHandler ConfirmationHandler) *DefaultToolManager {
	return &DefaultToolManager{
		registry:           registry,
		confirmationHandler: confirmationHandler,
	}
}

// RegisterTool adds a new tool to the registry
func (tm *DefaultToolManager) RegisterTool(tool Tool) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	return tm.registry.Register(tool)
}

// GetTool retrieves a tool by name
func (tm *DefaultToolManager) GetTool(name string) (Tool, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	tool, exists := tm.registry.Get(name)
	if !exists {
		return nil, fmt.Errorf("tool '%s' not found", name)
	}
	
	return tool, nil
}

// ListTools returns all registered tools, optionally filtered by category
func (tm *DefaultToolManager) ListTools(category string) []Tool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	if category == "" {
		return tm.registry.List()
	}
	
	return tm.registry.ListByCategory(category)
}

// ExecuteTool executes a tool with the given parameters
func (tm *DefaultToolManager) ExecuteTool(ctx context.Context, name string, params map[string]any) (*ToolResult, error) {
	tool, err := tm.GetTool(name)
	if err != nil {
		return nil, err
	}
	
	// Validate parameters
	if err := tm.ValidateParameters(name, params); err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("parameter validation failed: %v", err),
		}, nil
	}
	
	// Check if confirmation is required
	if tool.RequiresConfirmation() && tm.confirmationHandler != nil {
		preview, err := tm.ExecuteWithPreview(ctx, name, params)
		if err != nil {
			return &ToolResult{
				Success: false,
				Error:   fmt.Sprintf("failed to generate preview: %v", err),
			}, nil
		}
		
		confirmed, err := tm.confirmationHandler.RequestConfirmation(ctx, preview)
		if err != nil {
			return &ToolResult{
				Success: false,
				Error:   fmt.Sprintf("confirmation failed: %v", err),
			}, nil
		}
		
		if !confirmed {
			return &ToolResult{
				Success: false,
				Error:   "operation cancelled by user",
			}, nil
		}
	}
	
	// Execute the tool
	return tool.Execute(ctx, params)
}

// ExecuteWithPreview executes a tool in preview mode (dry-run)
func (tm *DefaultToolManager) ExecuteWithPreview(ctx context.Context, name string, params map[string]any) (*ToolPreview, error) {
	tool, err := tm.GetTool(name)
	if err != nil {
		return nil, err
	}
	
	// Validate parameters
	if err := tm.ValidateParameters(name, params); err != nil {
		return nil, fmt.Errorf("parameter validation failed: %v", err)
	}
	
	// Check if the tool supports preview (implements PreviewableTool interface)
	if previewable, ok := tool.(PreviewableTool); ok {
		return previewable.Preview(ctx, params)
	}
	
	// Generate a basic preview for tools that don't implement PreviewableTool
	return &ToolPreview{
		ToolName:        name,
		Parameters:      params,
		Description:     tool.Description(),
		ExpectedChanges: []ExpectedChange{},
		RequiresConfirmation: tool.RequiresConfirmation(),
	}, nil
}

// ValidateParameters validates parameters against a tool's schema
func (tm *DefaultToolManager) ValidateParameters(toolName string, params map[string]any) error {
	tool, err := tm.GetTool(toolName)
	if err != nil {
		return err
	}
	
	schema := tool.Parameters()
	
	// Check required parameters
	for _, required := range schema.Required {
		if _, exists := params[required]; !exists {
			return fmt.Errorf("required parameter '%s' is missing", required)
		}
	}
	
	// Validate parameter types and constraints
	for paramName, paramValue := range params {
		if property, exists := schema.Properties[paramName]; exists {
			if err := validateParameterValue(paramName, paramValue, property); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// validateParameterValue validates a single parameter value against its schema
func validateParameterValue(name string, value any, property ParameterProperty) error {
	switch property.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("parameter '%s' must be a string", name)
		}
		
		// Check enum constraints
		if len(property.Enum) > 0 {
			strValue := value.(string)
			for _, enumValue := range property.Enum {
				if strValue == enumValue {
					return nil
				}
			}
			return fmt.Errorf("parameter '%s' must be one of: %v", name, property.Enum)
		}
		
	case "number":
		switch value.(type) {
		case int, int64, float64:
			// Valid number types
		default:
			return fmt.Errorf("parameter '%s' must be a number", name)
		}
		
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("parameter '%s' must be a boolean", name)
		}
		
	case "array":
		// Use reflection to check if value is any slice type
		// This is more robust than type switching on specific slice types like []string, []int
		// because it accepts any slice type ([]T where T is any type)
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Slice {
			return fmt.Errorf("parameter '%s' must be an array (slice)", name)
		}
		
	case "object":
		// Basic object validation - could be expanded
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("parameter '%s' must be an object", name)
		}
	}
	
	return nil
}

// PreviewableTool is an optional interface that tools can implement to provide custom previews
type PreviewableTool interface {
	Tool
	Preview(ctx context.Context, params map[string]any) (*ToolPreview, error)
}