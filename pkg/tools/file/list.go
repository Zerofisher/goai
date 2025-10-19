package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ListTool implements directory listing functionality with security checks.
type ListTool struct {
	workDir  string
	maxItems int
}

// NewListTool creates a new directory listing tool.
func NewListTool(workDir string, maxItems int) *ListTool {
	if maxItems <= 0 {
		maxItems = 1000 // Default max items
	}
	return &ListTool{
		workDir:  workDir,
		maxItems: maxItems,
	}
}

// Name returns the name of the tool.
func (t *ListTool) Name() string {
	return "list_files"
}

// Description returns the description of the tool.
func (t *ListTool) Description() string {
	return "List files in a directory with glob pattern support and filtering"
}

// InputSchema returns the JSON schema for the input.
func (t *ListTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"dir": map[string]interface{}{
				"type":        "string",
				"description": "Directory path to list (default: '.' - current directory)",
				"default":     ".",
			},
			"include_globs": map[string]interface{}{
				"type":        "array",
				"description": "Glob patterns to include (e.g., ['*.go', '*.md'])",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"exclude_globs": map[string]interface{}{
				"type":        "array",
				"description": "Glob patterns to exclude (e.g., ['*.test', 'vendor/*'])",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of files to return",
				"default":     1000,
			},
			"recursive": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to list recursively",
				"default":     false,
			},
		},
	}
}

// Execute lists the directory contents and returns JSON formatted result.
func (t *ListTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
	// Extract dir parameter (default to current directory)
	dir := "."
	if dirRaw, ok := input["dir"]; ok {
		if dirStr, ok := dirRaw.(string); ok {
			dir = dirStr
		}
	}

	// Extract include_globs
	var includeGlobs []string
	if includeRaw, ok := input["include_globs"]; ok {
		if includeArray, ok := includeRaw.([]interface{}); ok {
			for _, item := range includeArray {
				if str, ok := item.(string); ok {
					includeGlobs = append(includeGlobs, str)
				}
			}
		}
	}

	// Extract exclude_globs
	var excludeGlobs []string
	if excludeRaw, ok := input["exclude_globs"]; ok {
		if excludeArray, ok := excludeRaw.([]interface{}); ok {
			for _, item := range excludeArray {
				if str, ok := item.(string); ok {
					excludeGlobs = append(excludeGlobs, str)
				}
			}
		}
	}

	// Extract limit (default to maxItems)
	limit := t.maxItems
	if limitRaw, ok := input["limit"]; ok {
		if limitFloat, ok := limitRaw.(float64); ok {
			limit = int(limitFloat)
			if limit > t.maxItems {
				limit = t.maxItems
			}
		}
	}

	// Extract recursive option
	recursive := false
	if recRaw, ok := input["recursive"]; ok {
		if recBool, ok := recRaw.(bool); ok {
			recursive = recBool
		}
	}

	// Validate path security
	if err := t.validatePath(dir); err != nil {
		return Error("Invalid path", err), nil
	}

	// Convert to absolute path
	absPath := t.resolvePath(dir)

	// Check if path exists
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Error("Directory not found", fmt.Errorf("path does not exist: %s", dir)), nil
		}
		return Error("Failed to access directory", err), nil
	}

	// Ensure it's a directory
	if !fileInfo.IsDir() {
		return Error("Invalid directory", fmt.Errorf("path is not a directory: %s", dir)), nil
	}

	// List the directory
	files, err := t.listDirectory(absPath, recursive, includeGlobs, excludeGlobs, limit)
	if err != nil {
		return Error("Failed to list directory", err), nil
	}

	// Create file entries for response
	entries := make([]FileEntry, len(files))
	for i, file := range files {
		entries[i] = FileEntry{
			Path:     file.Path,
			Name:     file.Name,
			Size:     file.Size,
			IsDir:    file.IsDir,
			Modified: file.Modified,
			Mode:     file.Mode,
		}
	}

	// Create response data
	data := ListFilesData{
		Files: entries,
		Count: len(entries),
		Path:  dir,
	}

	summary := fmt.Sprintf("Found %d files in %s", len(entries), dir)
	if recursive {
		summary += " (recursive)"
	}

	return Success(summary, data), nil
}

// Validate validates the input parameters.
func (t *ListTool) Validate(input map[string]interface{}) error {
	// dir is optional, defaults to "."
	if dirRaw, ok := input["dir"]; ok {
		if dir, ok := dirRaw.(string); ok {
			if dir == "" {
				return fmt.Errorf("dir cannot be empty string")
			}
		} else {
			return fmt.Errorf("dir must be a string")
		}
	}

	// Validate include_globs if provided
	if includeRaw, ok := input["include_globs"]; ok {
		if includeArray, ok := includeRaw.([]interface{}); ok {
			for _, item := range includeArray {
				if _, ok := item.(string); !ok {
					return fmt.Errorf("include_globs must be an array of strings")
				}
			}
		} else {
			return fmt.Errorf("include_globs must be an array")
		}
	}

	// Validate exclude_globs if provided
	if excludeRaw, ok := input["exclude_globs"]; ok {
		if excludeArray, ok := excludeRaw.([]interface{}); ok {
			for _, item := range excludeArray {
				if _, ok := item.(string); !ok {
					return fmt.Errorf("exclude_globs must be an array of strings")
				}
			}
		} else {
			return fmt.Errorf("exclude_globs must be an array")
		}
	}

	// Validate limit if provided
	if limitRaw, ok := input["limit"]; ok {
		if limitFloat, ok := limitRaw.(float64); ok {
			if limitFloat <= 0 {
				return fmt.Errorf("limit must be positive")
			}
		} else {
			return fmt.Errorf("limit must be a number")
		}
	}

	// Validate recursive if provided
	if recRaw, ok := input["recursive"]; ok {
		if _, ok := recRaw.(bool); !ok {
			return fmt.Errorf("recursive must be a boolean")
		}
	}

	return nil
}

// validatePath checks if the path is safe to access.
func (t *ListTool) validatePath(path string) error {
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
func (t *ListTool) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(t.workDir, path)
}

// FileInfoInternal represents information about a file or directory (internal use).
type FileInfoInternal struct {
	Name     string
	Path     string
	IsDir    bool
	Size     int64
	Mode     string
	Modified string
}

// listDirectory lists the contents of a directory with filtering.
func (t *ListTool) listDirectory(path string, recursive bool, includeGlobs, excludeGlobs []string, limit int) ([]FileInfoInternal, error) {
	var files []FileInfoInternal
	count := 0

	walkFunc := func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip files we can't access
			return nil
		}

		// Check limit
		if count >= limit {
			if info.IsDir() && filePath != path {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip the root directory itself
		if filePath == path {
			return nil
		}

		// Skip hidden files and directories (starting with .)
		if strings.HasPrefix(filepath.Base(filePath), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Get relative path from work directory
		relPath, err := filepath.Rel(t.workDir, filePath)
		if err != nil {
			return nil
		}

		// Apply exclude patterns
		if t.matchesAnyGlob(relPath, excludeGlobs) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Apply include patterns (if specified)
		if len(includeGlobs) > 0 && !t.matchesAnyGlob(relPath, includeGlobs) {
			return nil
		}

		// Add file to results
		files = append(files, FileInfoInternal{
			Name:     filepath.Base(filePath),
			Path:     relPath,
			IsDir:    info.IsDir(),
			Size:     info.Size(),
			Mode:     info.Mode().String(),
			Modified: info.ModTime().Format(time.RFC3339),
		})
		count++

		// Stop recursion if not recursive mode
		if !recursive && info.IsDir() && filePath != path {
			return filepath.SkipDir
		}

		return nil
	}

	if err := filepath.Walk(path, walkFunc); err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	// Sort files: directories first, then alphabetically
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return files[i].Name < files[j].Name
	})

	return files, nil
}

// matchesAnyGlob checks if the path matches any of the provided glob patterns.
func (t *ListTool) matchesAnyGlob(path string, globs []string) bool {
	for _, glob := range globs {
		// Try matching the full path
		if matched, err := filepath.Match(glob, path); err == nil && matched {
			return true
		}

		// Also try matching just the basename
		if matched, err := filepath.Match(glob, filepath.Base(path)); err == nil && matched {
			return true
		}

		// For patterns like "vendor/*", check if path starts with "vendor/"
		if strings.HasSuffix(glob, "/*") {
			prefix := strings.TrimSuffix(glob, "/*")
			if strings.HasPrefix(path, prefix+string(filepath.Separator)) || path == prefix {
				return true
			}
		}
	}
	return false
}
