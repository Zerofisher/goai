package tools

import (
	"context"
)

// Tool represents a single executable tool with input/output capabilities
type Tool interface {
	// Name returns the unique identifier for this tool
	Name() string
	
	// Description returns a human-readable description of what this tool does
	Description() string
	
	// Parameters returns the schema for input parameters this tool expects
	Parameters() ParameterSchema
	
	// Execute runs the tool with the provided parameters and returns the result
	Execute(ctx context.Context, params map[string]any) (*ToolResult, error)
	
	// RequiresConfirmation indicates if this tool needs user confirmation before execution
	RequiresConfirmation() bool
	
	// Category returns the category this tool belongs to (e.g., "file", "search", "system")
	Category() string
}

// ToolManager manages the registry and execution of tools
type ToolManager interface {
	// RegisterTool adds a new tool to the registry
	RegisterTool(tool Tool) error
	
	// GetTool retrieves a tool by name
	GetTool(name string) (Tool, error)
	
	// ListTools returns all registered tools, optionally filtered by category
	ListTools(category string) []Tool
	
	// ExecuteTool executes a tool with the given parameters
	ExecuteTool(ctx context.Context, name string, params map[string]any) (*ToolResult, error)
	
	// ExecuteWithPreview executes a tool in preview mode (dry-run)
	ExecuteWithPreview(ctx context.Context, name string, params map[string]any) (*ToolPreview, error)
	
	// ValidateParameters validates parameters against a tool's schema
	ValidateParameters(toolName string, params map[string]any) error
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Success     bool                   `json:"success"`
	Data        any                   `json:"data,omitempty"`
	Error       string                `json:"error,omitempty"`
	Output      string                `json:"output,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	ModifiedFiles []string            `json:"modified_files,omitempty"`
}

// ToolPreview represents a preview of what a tool would do without executing
type ToolPreview struct {
	ToolName        string                 `json:"tool_name"`
	Parameters      map[string]any `json:"parameters"`
	Description     string                 `json:"description"`
	ExpectedChanges []ExpectedChange       `json:"expected_changes"`
	RequiresConfirmation bool              `json:"requires_confirmation"`
	EstimatedTime   string                 `json:"estimated_time,omitempty"`
}

// ExpectedChange describes what a tool execution would change
type ExpectedChange struct {
	Type        string `json:"type"` // "create", "modify", "delete"
	Target      string `json:"target"` // file path, command, etc.
	Description string `json:"description"`
	Preview     string `json:"preview,omitempty"` // content preview for file operations
}

// ParameterSchema defines the input schema for a tool
type ParameterSchema struct {
	Required   []string                       `json:"required"`
	Properties map[string]ParameterProperty   `json:"properties"`
}

// ParameterProperty defines a single parameter
type ParameterProperty struct {
	Type        string      `json:"type"` // "string", "number", "boolean", "array", "object"
	Description string      `json:"description"`
	Default     any `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Format      string      `json:"format,omitempty"` // "file-path", "url", etc.
}

// ToolRegistry provides registration and lookup for tools
type ToolRegistry interface {
	Register(tool Tool) error
	Get(name string) (Tool, bool)
	List() []Tool
	ListByCategory(category string) []Tool
}

// ConfirmationHandler handles user confirmations for dangerous operations
type ConfirmationHandler interface {
	RequestConfirmation(ctx context.Context, preview *ToolPreview) (bool, error)
}