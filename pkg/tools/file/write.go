package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

// WriteTool implements file writing functionality with security checks.
type WriteTool struct {
	workDir string
	maxSize int64 // max file size in bytes
}

// NewWriteTool creates a new file writing tool.
func NewWriteTool(workDir string, maxSize int64) *WriteTool {
	if maxSize <= 0 {
		maxSize = 10 * 1024 * 1024 // Default 10MB
	}
	return &WriteTool{
		workDir: workDir,
		maxSize: maxSize,
	}
}

// Name returns the name of the tool.
func (t *WriteTool) Name() string {
	return "write_file"
}

// Description returns the description of the tool.
func (t *WriteTool) Description() string {
	return "Write or create a file within the work directory with validation and options"
}

// InputSchema returns the JSON schema for the input.
func (t *WriteTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to write (relative to work directory)",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content to write to the file",
			},
			"overwrite": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to overwrite if file exists (default: true)",
				"default":     true,
			},
			"create_parents": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to create parent directories (default: true)",
				"default":     true,
			},
			"mode_octal": map[string]interface{}{
				"type":        "string",
				"description": "File mode in octal format (e.g., '0644', default: system default)",
			},
			"validate_utf8": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to validate content as UTF-8 (default: true)",
				"default":     true,
			},
		},
		"required": []string{"path", "content"},
	}
}

// Execute writes content to the file and returns JSON formatted result.
func (t *WriteTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
	// Extract and validate path
	pathRaw, ok := input["path"]
	if !ok {
		return Error("Missing required parameter", fmt.Errorf("path is required")), nil
	}

	path, ok := pathRaw.(string)
	if !ok {
		return Error("Invalid parameter type", fmt.Errorf("path must be a string")), nil
	}

	// Extract content
	contentRaw, ok := input["content"]
	if !ok {
		return Error("Missing required parameter", fmt.Errorf("content is required")), nil
	}

	content, ok := contentRaw.(string)
	if !ok {
		return Error("Invalid parameter type", fmt.Errorf("content must be a string")), nil
	}

	// Check content size
	if int64(len(content)) > t.maxSize {
		return Error("Content too large", fmt.Errorf("content size %d bytes exceeds max %d", len(content), t.maxSize)), nil
	}

	// Extract options with defaults
	overwrite := true
	if overwriteRaw, ok := input["overwrite"]; ok {
		if overwriteBool, ok := overwriteRaw.(bool); ok {
			overwrite = overwriteBool
		}
	}

	createParents := true
	if createParentsRaw, ok := input["create_parents"]; ok {
		if createParentsBool, ok := createParentsRaw.(bool); ok {
			createParents = createParentsBool
		}
	}

	validateUTF8 := true
	if validateUTF8Raw, ok := input["validate_utf8"]; ok {
		if validateUTF8Bool, ok := validateUTF8Raw.(bool); ok {
			validateUTF8 = validateUTF8Bool
		}
	}

	// Validate UTF-8 if requested
	if validateUTF8 && !utf8.ValidString(content) {
		return Error("Invalid content encoding", fmt.Errorf("content is not valid UTF-8")), nil
	}

	// Extract file mode (optional)
	var fileMode os.FileMode = 0644 // Default
	if modeOctalRaw, ok := input["mode_octal"]; ok {
		if modeOctalStr, ok := modeOctalRaw.(string); ok {
			mode, err := parseModeOctal(modeOctalStr)
			if err != nil {
				return Error("Invalid file mode", err), nil
			}
			fileMode = mode
		}
	}

	// Validate path security
	if err := t.validatePath(path); err != nil {
		return Error("Invalid path", err), nil
	}

	// Convert to absolute path
	absPath := t.resolvePath(path)

	// Check if file exists
	fileExists := false
	if _, err := os.Stat(absPath); err == nil {
		fileExists = true
	}

	// Check overwrite constraint
	if fileExists && !overwrite {
		return Error("File already exists", fmt.Errorf("file exists and overwrite=false: %s", path)), nil
	}

	// Ensure parent directory exists if requested
	if createParents {
		if err := t.ensureParentDir(absPath); err != nil {
			return Error("Failed to create parent directory", err), nil
		}
	}

	// Write the file
	if err := t.writeFile(absPath, content, fileMode); err != nil {
		return Error("Failed to write file", err), nil
	}

	// Get file info for response
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return Error("Failed to stat written file", err), nil
	}

	// Create response data
	data := WriteFileData{
		Path:    path,
		Bytes:   int(fileInfo.Size()),
		Created: !fileExists,
		Mode:    fileInfo.Mode().String(),
	}

	summary := fmt.Sprintf("Wrote %d bytes to %s", data.Bytes, path)
	if !fileExists {
		summary = fmt.Sprintf("Created %s with %d bytes", path, data.Bytes)
	}

	return Success(summary, data), nil
}

// Validate validates the input parameters.
func (t *WriteTool) Validate(input map[string]interface{}) error {
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

	contentRaw, ok := input["content"]
	if !ok {
		return fmt.Errorf("missing required parameter: content")
	}

	if _, ok := contentRaw.(string); !ok {
		return fmt.Errorf("content must be a string")
	}

	// Validate optional boolean parameters
	if overwriteRaw, ok := input["overwrite"]; ok {
		if _, ok := overwriteRaw.(bool); !ok {
			return fmt.Errorf("overwrite must be a boolean")
		}
	}

	if createParentsRaw, ok := input["create_parents"]; ok {
		if _, ok := createParentsRaw.(bool); !ok {
			return fmt.Errorf("create_parents must be a boolean")
		}
	}

	if validateUTF8Raw, ok := input["validate_utf8"]; ok {
		if _, ok := validateUTF8Raw.(bool); !ok {
			return fmt.Errorf("validate_utf8 must be a boolean")
		}
	}

	// Validate mode_octal if provided
	if modeOctalRaw, ok := input["mode_octal"]; ok {
		modeOctalStr, ok := modeOctalRaw.(string)
		if !ok {
			return fmt.Errorf("mode_octal must be a string")
		}
		if _, err := parseModeOctal(modeOctalStr); err != nil {
			return fmt.Errorf("invalid mode_octal: %w", err)
		}
	}

	return nil
}

// validatePath checks if the path is safe to access.
func (t *WriteTool) validatePath(path string) error {
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

	// Prevent writing to certain sensitive files
	baseName := filepath.Base(absPath)
	forbiddenFiles := []string{".git"}
	for _, forbidden := range forbiddenFiles {
		if baseName == forbidden {
			return fmt.Errorf("cannot write to protected file: %s", baseName)
		}
	}

	return nil
}

// resolvePath converts a relative path to an absolute path within the work directory.
func (t *WriteTool) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(t.workDir, path)
}

// ensureParentDir creates the parent directory if it doesn't exist.
func (t *WriteTool) ensureParentDir(path string) error {
	parentDir := filepath.Dir(path)

	// Check if parent directory exists
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		// Create parent directory with proper permissions
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	return nil
}

// writeFile writes content to the file with the specified mode.
func (t *WriteTool) writeFile(path string, content string, mode os.FileMode) error {
	// Write file with specified mode
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// parseModeOctal parses an octal file mode string (e.g., "0644") into os.FileMode.
func parseModeOctal(modeStr string) (os.FileMode, error) {
	// Remove leading "0" if present
	modeStr = strings.TrimPrefix(modeStr, "0")

	// Parse as octal
	mode, err := strconv.ParseUint(modeStr, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid octal format: %w", err)
	}

	return os.FileMode(mode), nil
}
