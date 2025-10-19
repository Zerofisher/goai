package file

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ReadTool implements file reading functionality with security checks.
type ReadTool struct {
	workDir string
	maxSize int64 // max file size in bytes
}

// NewReadTool creates a new file reading tool.
func NewReadTool(workDir string, maxSize int64) *ReadTool {
	if maxSize <= 0 {
		maxSize = 10 * 1024 * 1024 // Default 10MB
	}
	return &ReadTool{
		workDir: workDir,
		maxSize: maxSize,
	}
}

// Name returns the name of the tool.
func (t *ReadTool) Name() string {
	return "read_file"
}

// Description returns the description of the tool.
func (t *ReadTool) Description() string {
	return "Read contents of a file within the work directory"
}

// InputSchema returns the JSON schema for the input.
func (t *ReadTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to read (relative to work directory)",
			},
			"start_line": map[string]interface{}{
				"type":        "integer",
				"description": "Starting line number (1-based, optional, >= 1)",
				"minimum":     1,
			},
			"end_line": map[string]interface{}{
				"type":        "integer",
				"description": "Ending line number (1-based, optional, >= start_line)",
				"minimum":     1,
			},
			"max_bytes": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum bytes to read (optional, default 200KB)",
				"default":     200 * 1024,
			},
		},
		"required": []string{"path"},
	}
}

// Execute reads the file and returns its contents in JSON format.
func (t *ReadTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
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

	// Check file exists
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Error("File not found", fmt.Errorf("file does not exist: %s", path)), nil
		}
		return Error("Failed to access file", err), nil
	}

	// Check if it's a directory
	if fileInfo.IsDir() {
		return Error("Invalid file type", fmt.Errorf("path is a directory, not a file: %s", path)), nil
	}

	// Extract max_bytes parameter (default 200KB)
	maxBytes := int64(200 * 1024)
	if maxBytesRaw, ok := input["max_bytes"]; ok {
		if maxBytesFloat, ok := maxBytesRaw.(float64); ok {
			maxBytes = int64(maxBytesFloat)
		}
	}

	// Check file size against max_bytes
	if fileInfo.Size() > maxBytes {
		return Error("File too large", fmt.Errorf("file size %d bytes exceeds max_bytes %d", fileInfo.Size(), maxBytes)), nil
	}

	// Extract line range if provided (using start_line and end_line)
	startLine := 0
	endLine := 0

	if startRaw, ok := input["start_line"]; ok {
		if startFloat, ok := startRaw.(float64); ok {
			startLine = int(startFloat)
		}
	}

	if endRaw, ok := input["end_line"]; ok {
		if endFloat, ok := endRaw.(float64); ok {
			endLine = int(endFloat)
		}
	}

	// Read the file
	content, readBytes, lineRange, err := t.readFile(absPath, startLine, endLine, maxBytes)
	if err != nil {
		return Error("Failed to read file", err), nil
	}

	// Create response data
	data := ReadFileData{
		Path:    path,
		Content: content,
		Bytes:   readBytes,
	}

	if lineRange != nil {
		data.Range = lineRange
	}

	// Format summary
	summary := fmt.Sprintf("Read %d bytes from %s", readBytes, path)
	if lineRange != nil {
		summary = fmt.Sprintf("Read lines %d-%d from %s (%d bytes)", lineRange.Start, lineRange.End, path, readBytes)
	}

	return Success(summary, data), nil
}

// Validate validates the input parameters.
func (t *ReadTool) Validate(input map[string]interface{}) error {
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

	// Validate line numbers if provided (using start_line and end_line)
	startLine := 0
	endLine := 0

	if startRaw, ok := input["start_line"]; ok {
		if startFloat, ok := startRaw.(float64); ok {
			startLine = int(startFloat)
		} else {
			return fmt.Errorf("start_line must be a number")
		}
	}

	if endRaw, ok := input["end_line"]; ok {
		if endFloat, ok := endRaw.(float64); ok {
			endLine = int(endFloat)
		} else {
			return fmt.Errorf("end_line must be a number")
		}
	}

	if startLine < 0 {
		return fmt.Errorf("start_line must be >= 1")
	}

	if endLine < 0 {
		return fmt.Errorf("end_line must be >= 1")
	}

	if startLine > 0 && endLine > 0 && startLine > endLine {
		return fmt.Errorf("start_line cannot be greater than end_line")
	}

	// Validate max_bytes if provided
	if maxBytesRaw, ok := input["max_bytes"]; ok {
		if maxBytesFloat, ok := maxBytesRaw.(float64); ok {
			if maxBytesFloat <= 0 {
				return fmt.Errorf("max_bytes must be positive")
			}
		} else {
			return fmt.Errorf("max_bytes must be a number")
		}
	}

	return nil
}

// validatePath checks if the path is safe to access.
func (t *ReadTool) validatePath(path string) error {
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
func (t *ReadTool) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(t.workDir, path)
}

// readFile reads the file content, optionally filtering by line range.
// Returns content, byte count, line range info, and error.
func (t *ReadTool) readFile(path string, startLine, endLine int, maxBytes int64) (string, int, *LineRange, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// If no line range specified, read entire file
	if startLine == 0 && endLine == 0 {
		content, err := io.ReadAll(io.LimitReader(file, maxBytes))
		if err != nil {
			return "", 0, nil, fmt.Errorf("failed to read file: %w", err)
		}
		return string(content), len(content), nil, nil
	}

	// Read specific line range
	var result strings.Builder
	scanner := bufio.NewScanner(file)
	lineNum := 0
	actualStart := 0
	actualEnd := 0

	for scanner.Scan() {
		lineNum++

		// Skip lines before start
		if startLine > 0 && lineNum < startLine {
			continue
		}

		// Track actual start line
		if actualStart == 0 {
			actualStart = lineNum
		}

		// Stop after end line
		if endLine > 0 && lineNum > endLine {
			break
		}

		// Add line to result
		if startLine == 0 || (lineNum >= startLine) {
			if result.Len() > 0 {
				result.WriteString("\n")
			}
			result.WriteString(scanner.Text())
			actualEnd = lineNum
		}

		// Check max bytes limit
		if int64(result.Len()) > maxBytes {
			return "", 0, nil, fmt.Errorf("content exceeds max_bytes limit")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", 0, nil, fmt.Errorf("error reading file: %w", err)
	}

	content := result.String()
	lineRange := &LineRange{
		Start: actualStart,
		End:   actualEnd,
	}

	return content, len(content), lineRange, nil
}