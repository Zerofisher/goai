package tools

import (
	"fmt"

	"github.com/Zerofisher/goai/pkg/indexing"
)

// ToolFactory creates and configures tool managers with all available tools
type ToolFactory struct{}

// NewToolFactory creates a new tool factory
func NewToolFactory() *ToolFactory {
	return &ToolFactory{}
}

// CreateDefaultToolManager creates a tool manager with all standard tools registered
func (f *ToolFactory) CreateDefaultToolManager(indexManager *indexing.EnhancedIndexManager) (*DefaultToolManager, error) {
	registry := NewToolRegistry()
	confirmationHandler := NewConsoleConfirmationHandler()
	
	manager := NewToolManager(registry, confirmationHandler)
	
	// Register all available tools
	if err := f.registerAllTools(manager, indexManager); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}
	
	return manager, nil
}

// CreateTestToolManager creates a tool manager for testing with mock confirmation handler
func (f *ToolFactory) CreateTestToolManager(indexManager *indexing.EnhancedIndexManager, alwaysConfirm bool) (*DefaultToolManager, error) {
	registry := NewToolRegistry()
	confirmationHandler := NewMockConfirmationHandler(alwaysConfirm)
	
	manager := NewToolManager(registry, confirmationHandler)
	
	// Register all available tools
	if err := f.registerAllTools(manager, indexManager); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}
	
	return manager, nil
}

// registerAllTools registers all available tools with the manager
func (f *ToolFactory) registerAllTools(manager *DefaultToolManager, indexManager *indexing.EnhancedIndexManager) error {
	// File operation tools
	fileTools := []Tool{
		NewReadFileTool(),
		NewWriteFileTool(),
		NewEditFileTool(),
		NewMultiEditTool(),
	}
	
	for _, tool := range fileTools {
		if err := manager.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register file tool %s: %w", tool.Name(), err)
		}
	}
	
	// Search tools
	searchTools := []Tool{
		NewListFilesTool(),
		NewViewDiffTool(),
	}
	
	// Add search code tool if index manager is available
	if indexManager != nil {
		searchTools = append(searchTools, NewSearchCodeTool(indexManager))
	}
	
	for _, tool := range searchTools {
		if err := manager.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register search tool %s: %w", tool.Name(), err)
		}
	}
	
	// System tools
	systemTools := []Tool{
		NewRunCommandTool(),
		NewFetchTool(),
		NewGitTool(),
	}
	
	for _, tool := range systemTools {
		if err := manager.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register system tool %s: %w", tool.Name(), err)
		}
	}
	
	return nil
}

// GetAvailableToolsInfo returns information about all available tools
func (f *ToolFactory) GetAvailableToolsInfo() []ToolInfo {
	return []ToolInfo{
		// File tools
		{
			Name:        "readFile",
			Description: "Read the contents of a file",
			Category:    "file",
			RequiresConfirmation: false,
		},
		{
			Name:        "writeFile",
			Description: "Write content to a file (creates directories if needed)",
			Category:    "file",
			RequiresConfirmation: true,
		},
		{
			Name:        "editFile",
			Description: "Perform targeted edits on a file by replacing specific content",
			Category:    "file",
			RequiresConfirmation: true,
		},
		{
			Name:        "multiEdit",
			Description: "Perform multiple edits on a file in a single operation",
			Category:    "file",
			RequiresConfirmation: true,
		},
		
		// Search tools
		{
			Name:        "searchCode",
			Description: "Search through codebase using full-text, semantic, or symbol-based search",
			Category:    "search",
			RequiresConfirmation: false,
		},
		{
			Name:        "listFiles",
			Description: "List files and directories with optional filtering and sorting",
			Category:    "search",
			RequiresConfirmation: false,
		},
		{
			Name:        "viewDiff",
			Description: "View differences between two files or between a file and its Git version",
			Category:    "search",
			RequiresConfirmation: false,
		},
		
		// System tools
		{
			Name:        "runCommand",
			Description: "Execute system commands and return their output",
			Category:    "system",
			RequiresConfirmation: true,
		},
		{
			Name:        "fetch",
			Description: "Fetch content from HTTP/HTTPS URLs",
			Category:    "system",
			RequiresConfirmation: false,
		},
		{
			Name:        "git",
			Description: "Perform Git operations (status, diff, log, etc.)",
			Category:    "system",
			RequiresConfirmation: false,
		},
	}
}

// ToolInfo provides information about a tool without instantiating it
type ToolInfo struct {
	Name                 string `json:"name"`
	Description          string `json:"description"`
	Category             string `json:"category"`
	RequiresConfirmation bool   `json:"requires_confirmation"`
}