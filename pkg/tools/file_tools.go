package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadFileTool reads the contents of a file
type ReadFileTool struct{}

func NewReadFileTool() *ReadFileTool {
	return &ReadFileTool{}
}

func (t *ReadFileTool) Name() string {
	return "readFile"
}

func (t *ReadFileTool) Description() string {
	return "Read the contents of a file"
}

func (t *ReadFileTool) Parameters() ParameterSchema {
	return ParameterSchema{
		Required: []string{"path"},
		Properties: map[string]ParameterProperty{
			"path": {
				Type:        "string",
				Description: "Path to the file to read",
				Format:      "file-path",
			},
			"encoding": {
				Type:        "string",
				Description: "File encoding (default: utf-8)",
				Default:     "utf-8",
			},
		},
	}
}

func (t *ReadFileTool) Execute(ctx context.Context, params map[string]any) (*ToolResult, error) {
	path, ok := params["path"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "path parameter must be a string",
		}, nil
	}
	
	// Clean and validate the path
	cleanPath := filepath.Clean(path)
	
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read file: %v", err),
		}, nil
	}
	
	return &ToolResult{
		Success: true,
		Data:    string(content),
		Output:  fmt.Sprintf("Successfully read %d bytes from %s", len(content), cleanPath),
		Metadata: map[string]any{
			"file_path": cleanPath,
			"file_size": len(content),
		},
	}, nil
}

func (t *ReadFileTool) RequiresConfirmation() bool {
	return false
}

func (t *ReadFileTool) Category() string {
	return "file"
}

// WriteFileTool writes content to a file
type WriteFileTool struct{}

func NewWriteFileTool() *WriteFileTool {
	return &WriteFileTool{}
}

func (t *WriteFileTool) Name() string {
	return "writeFile"
}

func (t *WriteFileTool) Description() string {
	return "Write content to a file (creates directories if needed)"
}

func (t *WriteFileTool) Parameters() ParameterSchema {
	return ParameterSchema{
		Required: []string{"path", "content"},
		Properties: map[string]ParameterProperty{
			"path": {
				Type:        "string",
				Description: "Path to the file to write",
				Format:      "file-path",
			},
			"content": {
				Type:        "string",
				Description: "Content to write to the file",
			},
			"encoding": {
				Type:        "string",
				Description: "File encoding (default: utf-8)",
				Default:     "utf-8",
			},
			"createDirectories": {
				Type:        "boolean",
				Description: "Create parent directories if they don't exist",
				Default:     true,
			},
		},
	}
}

func (t *WriteFileTool) Execute(ctx context.Context, params map[string]any) (*ToolResult, error) {
	path, ok := params["path"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "path parameter must be a string",
		}, nil
	}
	
	content, ok := params["content"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "content parameter must be a string",
		}, nil
	}
	
	cleanPath := filepath.Clean(path)
	
	// Create directories if needed
	if createDirs, exists := params["createDirectories"]; !exists || createDirs.(bool) {
		if err := os.MkdirAll(filepath.Dir(cleanPath), 0755); err != nil {
			return &ToolResult{
				Success: false,
				Error:   fmt.Sprintf("failed to create directories: %v", err),
			}, nil
		}
	}
	
	err := os.WriteFile(cleanPath, []byte(content), 0644)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to write file: %v", err),
		}, nil
	}
	
	return &ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), cleanPath),
		Metadata: map[string]any{
			"file_path": cleanPath,
			"file_size": len(content),
		},
		ModifiedFiles: []string{cleanPath},
	}, nil
}

func (t *WriteFileTool) RequiresConfirmation() bool {
	return true
}

func (t *WriteFileTool) Category() string {
	return "file"
}

func (t *WriteFileTool) Preview(ctx context.Context, params map[string]any) (*ToolPreview, error) {
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter must be a string")
	}
	
	content, ok := params["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content parameter must be a string")
	}
	
	cleanPath := filepath.Clean(path)
	
	// Check if file exists
	changeType := "create"
	if _, err := os.Stat(cleanPath); err == nil {
		changeType = "modify"
	}
	
	contentPreview := content
	if len(content) > 500 {
		contentPreview = content[:500] + "\n... (truncated)"
	}
	
	return &ToolPreview{
		ToolName:   t.Name(),
		Parameters: params,
		Description: fmt.Sprintf("Write %d bytes to %s", len(content), cleanPath),
		ExpectedChanges: []ExpectedChange{
			{
				Type:        changeType,
				Target:      cleanPath,
				Description: fmt.Sprintf("%s file with %d bytes of content", strings.ToUpper(changeType[:1])+changeType[1:], len(content)),
				Preview:     contentPreview,
			},
		},
		RequiresConfirmation: true,
	}, nil
}

// EditFileTool performs targeted edits on existing files
type EditFileTool struct{}

func NewEditFileTool() *EditFileTool {
	return &EditFileTool{}
}

func (t *EditFileTool) Name() string {
	return "editFile"
}

func (t *EditFileTool) Description() string {
	return "Perform targeted edits on a file by replacing specific content"
}

func (t *EditFileTool) Parameters() ParameterSchema {
	return ParameterSchema{
		Required: []string{"path", "oldContent", "newContent"},
		Properties: map[string]ParameterProperty{
			"path": {
				Type:        "string",
				Description: "Path to the file to edit",
				Format:      "file-path",
			},
			"oldContent": {
				Type:        "string",
				Description: "Content to find and replace",
			},
			"newContent": {
				Type:        "string",
				Description: "New content to replace with",
			},
			"all": {
				Type:        "boolean",
				Description: "Replace all occurrences (default: false)",
				Default:     false,
			},
		},
	}
}

func (t *EditFileTool) Execute(ctx context.Context, params map[string]any) (*ToolResult, error) {
	path, ok := params["path"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "path parameter must be a string",
		}, nil
	}
	
	oldContent, ok := params["oldContent"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "oldContent parameter must be a string",
		}, nil
	}
	
	newContent, ok := params["newContent"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "newContent parameter must be a string",
		}, nil
	}
	
	cleanPath := filepath.Clean(path)
	
	// Read the file
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read file: %v", err),
		}, nil
	}
	
	fileContent := string(content)
	
	// Check if old content exists
	if !strings.Contains(fileContent, oldContent) {
		return &ToolResult{
			Success: false,
			Error:   "old content not found in file",
		}, nil
	}
	
	// Replace content
	replaceAll, _ := params["all"].(bool)
	var newFileContent string
	if replaceAll {
		newFileContent = strings.ReplaceAll(fileContent, oldContent, newContent)
	} else {
		newFileContent = strings.Replace(fileContent, oldContent, newContent, 1)
	}
	
	// Write back to file
	err = os.WriteFile(cleanPath, []byte(newFileContent), 0644)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to write file: %v", err),
		}, nil
	}
	
	replacements := strings.Count(fileContent, oldContent)
	if !replaceAll {
		replacements = 1
	}
	
	return &ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Successfully made %d replacement(s) in %s", replacements, cleanPath),
		Metadata: map[string]any{
			"file_path":    cleanPath,
			"replacements": replacements,
			"old_size":     len(content),
			"new_size":     len(newFileContent),
		},
		ModifiedFiles: []string{cleanPath},
	}, nil
}

func (t *EditFileTool) RequiresConfirmation() bool {
	return true
}

func (t *EditFileTool) Category() string {
	return "file"
}

func (t *EditFileTool) Preview(ctx context.Context, params map[string]any) (*ToolPreview, error) {
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter must be a string")
	}
	
	oldContent, ok := params["oldContent"].(string)
	if !ok {
		return nil, fmt.Errorf("oldContent parameter must be a string")
	}
	
	newContent, ok := params["newContent"].(string)
	if !ok {
		return nil, fmt.Errorf("newContent parameter must be a string")
	}
	
	cleanPath := filepath.Clean(path)
	replaceAll, _ := params["all"].(bool)
	
	// Read the file to count potential replacements
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}
	
	fileContent := string(content)
	replacements := strings.Count(fileContent, oldContent)
	if !replaceAll {
		replacements = 1
	}
	
	previewContent := fmt.Sprintf("- %s\n+ %s", oldContent, newContent)
	if len(previewContent) > 200 {
		previewContent = previewContent[:200] + "..."
	}
	
	return &ToolPreview{
		ToolName:   t.Name(),
		Parameters: params,
		Description: fmt.Sprintf("Replace content in %s (%d occurrence(s))", cleanPath, replacements),
		ExpectedChanges: []ExpectedChange{
			{
				Type:        "modify",
				Target:      cleanPath,
				Description: fmt.Sprintf("Replace %d occurrence(s) of content", replacements),
				Preview:     previewContent,
			},
		},
		RequiresConfirmation: true,
	}, nil
}

// MultiEditTool performs multiple edits on a file in a single operation
type MultiEditTool struct{}

func NewMultiEditTool() *MultiEditTool {
	return &MultiEditTool{}
}

func (t *MultiEditTool) Name() string {
	return "multiEdit"
}

func (t *MultiEditTool) Description() string {
	return "Perform multiple edits on a file in a single operation"
}

type EditOperation struct {
	OldContent string `json:"oldContent"`
	NewContent string `json:"newContent"`
}

func (t *MultiEditTool) Parameters() ParameterSchema {
	return ParameterSchema{
		Required: []string{"path", "edits"},
		Properties: map[string]ParameterProperty{
			"path": {
				Type:        "string",
				Description: "Path to the file to edit",
				Format:      "file-path",
			},
			"edits": {
				Type:        "array",
				Description: "Array of edit operations with oldContent and newContent",
			},
		},
	}
}

func (t *MultiEditTool) Execute(ctx context.Context, params map[string]any) (*ToolResult, error) {
	path, ok := params["path"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "path parameter must be a string",
		}, nil
	}
	
	editsRaw, ok := params["edits"].([]any)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "edits parameter must be an array",
		}, nil
	}
	
	cleanPath := filepath.Clean(path)
	
	// Read the file
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read file: %v", err),
		}, nil
	}
	
	fileContent := string(content)
	originalContent := fileContent
	totalReplacements := 0
	
	// Apply each edit operation
	for i, editRaw := range editsRaw {
		editMap, ok := editRaw.(map[string]any)
		if !ok {
			return &ToolResult{
				Success: false,
				Error:   fmt.Sprintf("edit %d must be an object", i),
			}, nil
		}
		
		oldContent, ok := editMap["oldContent"].(string)
		if !ok {
			return &ToolResult{
				Success: false,
				Error:   fmt.Sprintf("edit %d: oldContent must be a string", i),
			}, nil
		}
		
		newContent, ok := editMap["newContent"].(string)
		if !ok {
			return &ToolResult{
				Success: false,
				Error:   fmt.Sprintf("edit %d: newContent must be a string", i),
			}, nil
		}
		
		// Check if old content exists
		if !strings.Contains(fileContent, oldContent) {
			return &ToolResult{
				Success: false,
				Error:   fmt.Sprintf("edit %d: old content not found in file", i),
			}, nil
		}
		
		// Apply the replacement (only first occurrence to maintain predictability)
		fileContent = strings.Replace(fileContent, oldContent, newContent, 1)
		totalReplacements++
	}
	
	// Write back to file
	err = os.WriteFile(cleanPath, []byte(fileContent), 0644)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to write file: %v", err),
		}, nil
	}
	
	return &ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Successfully applied %d edit operations to %s", totalReplacements, cleanPath),
		Metadata: map[string]any{
			"file_path":    cleanPath,
			"edits_applied": totalReplacements,
			"old_size":     len(originalContent),
			"new_size":     len(fileContent),
		},
		ModifiedFiles: []string{cleanPath},
	}, nil
}

func (t *MultiEditTool) RequiresConfirmation() bool {
	return true
}

func (t *MultiEditTool) Category() string {
	return "file"
}

func (t *MultiEditTool) Preview(ctx context.Context, params map[string]any) (*ToolPreview, error) {
	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter must be a string")
	}
	
	editsRaw, ok := params["edits"].([]any)
	if !ok {
		return nil, fmt.Errorf("edits parameter must be an array")
	}
	
	cleanPath := filepath.Clean(path)
	
	var expectedChanges []ExpectedChange
	for i, editRaw := range editsRaw {
		editMap, ok := editRaw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("edit %d must be an object", i)
		}
		
		oldContent, _ := editMap["oldContent"].(string)
		newContent, _ := editMap["newContent"].(string)
		
		previewContent := fmt.Sprintf("Edit %d:\n- %s\n+ %s", i+1, oldContent, newContent)
		if len(previewContent) > 150 {
			previewContent = previewContent[:150] + "..."
		}
		
		expectedChanges = append(expectedChanges, ExpectedChange{
			Type:        "modify",
			Target:      cleanPath,
			Description: fmt.Sprintf("Edit operation %d", i+1),
			Preview:     previewContent,
		})
	}
	
	return &ToolPreview{
		ToolName:   t.Name(),
		Parameters: params,
		Description: fmt.Sprintf("Apply %d edit operations to %s", len(editsRaw), cleanPath),
		ExpectedChanges: expectedChanges,
		RequiresConfirmation: true,
	}, nil
}