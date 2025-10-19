package edit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EditTool implements text file editing functionality with multiple strategies.
type EditTool struct {
	workDir    string
	backup     *BackupManager
	strategies map[string]Strategy
}

// NewEditTool creates a new text editing tool.
func NewEditTool(workDir string) *EditTool {
	tool := &EditTool{
		workDir:    workDir,
		backup:     NewBackupManager(workDir),
		strategies: make(map[string]Strategy),
	}

	// Register available strategies
	tool.strategies["replace"] = NewReplaceStrategy()
	tool.strategies["insert"] = NewInsertStrategy()
	tool.strategies["anchored"] = NewAnchoredStrategy()
	tool.strategies["apply_patch"] = NewApplyPatchStrategy()

	return tool
}

// Name returns the name of the tool.
func (t *EditTool) Name() string {
	return "edit_file"
}

// Description returns the description of the tool.
func (t *EditTool) Description() string {
	return "Edit text files with multiple strategies: replace, insert, anchored, apply_patch"
}

// InputSchema returns the JSON schema for the input.
func (t *EditTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to edit (relative to work directory)",
			},
			"strategy": map[string]interface{}{
				"type":        "string",
				"description": "Edit strategy to use",
				"enum":        []string{"replace", "insert", "anchored", "apply_patch"},
				"default":     "replace",
			},
			// Replace strategy parameters
			"old_text": map[string]interface{}{
				"type":        "string",
				"description": "Text to replace (for replace and anchored strategies)",
			},
			"new_text": map[string]interface{}{
				"type":        "string",
				"description": "Replacement text (for replace and anchored strategies)",
			},
			"line_start": map[string]interface{}{
				"type":        "integer",
				"description": "Start line number for replace (1-based, optional)",
				"minimum":     1,
			},
			"line_end": map[string]interface{}{
				"type":        "integer",
				"description": "End line number for replace (1-based, optional)",
				"minimum":     1,
			},
			"replace_all": map[string]interface{}{
				"type":        "boolean",
				"description": "Replace all occurrences (default: false)",
				"default":     false,
			},
			// Insert strategy parameters
			"text": map[string]interface{}{
				"type":        "string",
				"description": "Text to insert (for insert strategy)",
			},
			"line": map[string]interface{}{
				"type":        "integer",
				"description": "Line number to insert at (0-based, 0=prepend)",
				"minimum":     0,
			},
			"after_anchor": map[string]interface{}{
				"type":        "string",
				"description": "Insert after this text anchor (for insert strategy)",
			},
			// Anchored strategy parameters
			"before_anchor": map[string]interface{}{
				"type":        "string",
				"description": "Text anchor before the edit region (for anchored strategy)",
			},
			// Apply patch strategy parameters
			"patch": map[string]interface{}{
				"type":        "string",
				"description": "Unified diff patch to apply (for apply_patch strategy)",
			},
			// Common options
			"create_backup": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to create a backup before editing (default: true)",
				"default":     true,
			},
			"detect_conflicts": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to detect potential conflicts (default: true)",
				"default":     true,
			},
		},
		"required": []string{"path", "strategy"},
	}
}

// Execute performs the edit operation and returns JSON formatted result.
func (t *EditTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
	// Extract and validate path
	pathRaw, ok := input["path"]
	if !ok {
		return Error("Missing required parameter", fmt.Errorf("path is required")), nil
	}

	path, ok := pathRaw.(string)
	if !ok {
		return Error("Invalid parameter type", fmt.Errorf("path must be a string")), nil
	}

	// Validate path security
	if err := t.validatePath(path); err != nil {
		return Error("Invalid path", err), nil
	}

	// Convert to absolute path
	absPath := t.resolvePath(path)

	// Check if file exists
	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			return Error("File not found", fmt.Errorf("file does not exist: %s", path)), nil
		}
		return Error("Failed to access file", err), nil
	}

	// Extract strategy (default to "replace")
	strategy := "replace"
	if strategyRaw, ok := input["strategy"]; ok {
		if strategyStr, ok := strategyRaw.(string); ok {
			strategy = strategyStr
		}
	}

	// Get the strategy implementation
	strategyImpl, ok := t.strategies[strategy]
	if !ok {
		return Error("Invalid strategy", fmt.Errorf("unknown strategy: %s", strategy)), nil
	}

	// Validate strategy-specific parameters
	if err := strategyImpl.Validate(input); err != nil {
		return Error("Invalid parameters", err), nil
	}

	// Check if backup is requested (default: true)
	createBackup := true
	if backupRaw, ok := input["create_backup"]; ok {
		if backupBool, ok := backupRaw.(bool); ok {
			createBackup = backupBool
		}
	}

	// Create backup if requested
	var backupPath string
	var backupCreated bool
	if createBackup {
		var err error
		backupPath, err = t.backup.CreateBackup(absPath)
		if err != nil {
			return Error("Failed to create backup", err), nil
		}
		backupCreated = true
	}

	// Detect conflicts if requested
	detectConflicts := true
	if detectRaw, ok := input["detect_conflicts"]; ok {
		if detectBool, ok := detectRaw.(bool); ok {
			detectConflicts = detectBool
		}
	}

	var conflicts []string
	if detectConflicts {
		conflicts = t.detectConflicts(absPath, input, strategy)
	}

	// If conflicts detected, warn but don't fail
	if len(conflicts) > 0 {
		// Could optionally fail here if a "fail_on_conflict" flag is set
		// For now, just include conflicts in the response
	}

	// Execute the strategy
	result, err := strategyImpl.Execute(absPath, input)
	if err != nil {
		// Restore from backup if operation failed and backup was created
		if backupPath != "" {
			if restoreErr := t.backup.RestoreBackup(backupPath, absPath); restoreErr != nil {
				return Error("Operation failed and backup restore failed",
					fmt.Errorf("edit error: %v, restore error: %w", err, restoreErr)), nil
			}
			// Clean up failed backup
			_ = os.Remove(backupPath)
		}
		return Error("Edit operation failed", err), nil
	}

	// Update result with backup info
	result.BackupCreated = backupCreated
	result.BackupPath = backupPath
	result.Conflicts = conflicts

	// Generate summary
	summary := fmt.Sprintf("Successfully edited %s using %s strategy (%d lines modified)",
		path, strategy, result.LinesModified)

	if len(conflicts) > 0 {
		summary += fmt.Sprintf(" [%d potential conflicts detected]", len(conflicts))
	}

	return Success(summary, result), nil
}

// Validate validates the input parameters.
func (t *EditTool) Validate(input map[string]interface{}) error {
	// Validate path
	pathRaw, ok := input["path"]
	if !ok {
		return fmt.Errorf("missing required parameter: path")
	}

	path, ok := pathRaw.(string)
	if !ok {
		return fmt.Errorf("path must be a string")
	}

	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Validate strategy
	strategyRaw, ok := input["strategy"]
	if !ok {
		return fmt.Errorf("missing required parameter: strategy")
	}

	strategy, ok := strategyRaw.(string)
	if !ok {
		return fmt.Errorf("strategy must be a string")
	}

	// Validate strategy exists
	strategyImpl, ok := t.strategies[strategy]
	if !ok {
		return fmt.Errorf("invalid strategy: %s", strategy)
	}

	// Validate strategy-specific parameters
	return strategyImpl.Validate(input)
}

// validatePath checks if the path is safe to access.
func (t *EditTool) validatePath(path string) error {
	// Prevent empty paths
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Prevent path traversal
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal detected")
	}

	// Ensure path is within work directory
	absPath := t.resolvePath(path)
	if !strings.HasPrefix(absPath, t.workDir) {
		return fmt.Errorf("path escapes work directory")
	}

	return nil
}

// resolvePath converts a relative path to an absolute path within the work directory.
func (t *EditTool) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(t.workDir, path)
}

// detectConflicts checks for potential conflicts before editing.
func (t *EditTool) detectConflicts(path string, input map[string]interface{}, strategy string) []string {
	conflicts := []string{}

	// Read current file content
	content, err := os.ReadFile(path)
	if err != nil {
		return conflicts
	}

	contentStr := string(content)

	switch strategy {
	case "replace":
		oldText, hasOldText := input["old_text"].(string)
		if !hasOldText {
			return conflicts
		}

		// Check for multiple occurrences if not replace_all
		replaceAll := false
		if replaceAllRaw, ok := input["replace_all"]; ok {
			if replaceAllBool, ok := replaceAllRaw.(bool); ok {
				replaceAll = replaceAllBool
			}
		}

		if !replaceAll {
			count := strings.Count(contentStr, oldText)
			if count > 1 {
				conflicts = append(conflicts,
					fmt.Sprintf("Multiple occurrences found (%d), but replace_all=false. Only first will be replaced.", count))
			}
		}

	case "anchored":
		// Check if anchors are unique
		if beforeAnchor, ok := input["before_anchor"].(string); ok {
			count := strings.Count(contentStr, beforeAnchor)
			if count > 1 {
				conflicts = append(conflicts,
					fmt.Sprintf("before_anchor appears %d times (should be unique)", count))
			} else if count == 0 {
				conflicts = append(conflicts, "before_anchor not found in file")
			}
		}

		if afterAnchor, ok := input["after_anchor"].(string); ok {
			count := strings.Count(contentStr, afterAnchor)
			if count > 1 {
				conflicts = append(conflicts,
					fmt.Sprintf("after_anchor appears %d times (should be unique)", count))
			} else if count == 0 {
				conflicts = append(conflicts, "after_anchor not found in file")
			}
		}
	}

	return conflicts
}
