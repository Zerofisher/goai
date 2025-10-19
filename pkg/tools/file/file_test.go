package file

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReadTool tests the file reading functionality with JSON output.
func TestReadTool(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "readtool_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test file
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tool := NewReadTool(tempDir, 1024*1024)

	tests := []struct {
		name        string
		input       map[string]interface{}
		wantContent string
		wantOk      bool
	}{
		{
			name: "read_entire_file",
			input: map[string]interface{}{
				"path": "test.txt",
			},
			wantContent: testContent,
			wantOk:      true,
		},
		{
			name: "read_with_line_range",
			input: map[string]interface{}{
				"path":       "test.txt",
				"start_line": float64(2),
				"end_line":   float64(4),
			},
			wantContent: "Line 2\nLine 3\nLine 4",
			wantOk:      true,
		},
		{
			name: "read_single_line",
			input: map[string]interface{}{
				"path":       "test.txt",
				"start_line": float64(3),
				"end_line":   float64(3),
			},
			wantContent: "Line 3",
			wantOk:      true,
		},
		{
			name: "file_not_found",
			input: map[string]interface{}{
				"path": "nonexistent.txt",
			},
			wantContent: "",
			wantOk:      false,
		},
		{
			name: "path_traversal_attempt",
			input: map[string]interface{}{
				"path": "../../../etc/passwd",
			},
			wantContent: "",
			wantOk:      false,
		},
		{
			name: "missing_path_parameter",
			input: map[string]interface{}{
				"start_line": float64(1),
			},
			wantContent: "",
			wantOk:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := tool.Execute(ctx, tt.input)

			// Execute never returns an error now, it returns JSON
			if err != nil {
				t.Errorf("Execute() unexpected error = %v", err)
				return
			}

			// Parse JSON response
			var resp ToolResponse
			if err := json.Unmarshal([]byte(result), &resp); err != nil {
				t.Errorf("Failed to parse JSON response: %v", err)
				return
			}

			// Check success/failure
			if resp.Ok != tt.wantOk {
				t.Errorf("Execute() ok = %v, want %v", resp.Ok, tt.wantOk)
				if !resp.Ok {
					t.Logf("Error message: %s", resp.Error)
				}
				return
			}

			// For successful reads, check content
			if tt.wantOk {
				dataMap, ok := resp.Data.(map[string]interface{})
				if !ok {
					t.Errorf("Data is not a map")
					return
				}

				content, ok := dataMap["content"].(string)
				if !ok {
					t.Errorf("Content is not a string")
					return
				}

				if content != tt.wantContent {
					t.Errorf("Content = %q, want %q", content, tt.wantContent)
				}
			}
		})
	}
}

// TestWriteTool tests the file writing functionality with JSON output.
func TestWriteTool(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "writetool_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	tool := NewWriteTool(tempDir, 1024*1024)

	tests := []struct {
		name    string
		input   map[string]interface{}
		wantOk  bool
		verify  func(t *testing.T, tempDir string)
	}{
		{
			name: "write_new_file",
			input: map[string]interface{}{
				"path":    "test.txt",
				"content": "Hello, World!",
			},
			wantOk: true,
			verify: func(t *testing.T, tempDir string) {
				content, err := os.ReadFile(filepath.Join(tempDir, "test.txt"))
				if err != nil {
					t.Errorf("Failed to read written file: %v", err)
					return
				}
				if string(content) != "Hello, World!" {
					t.Errorf("File content = %q, want %q", string(content), "Hello, World!")
				}
			},
		},
		{
			name: "overwrite_existing_file",
			input: map[string]interface{}{
				"path":      "existing.txt",
				"content":   "Updated content",
				"overwrite": true,
			},
			wantOk: true,
			verify: func(t *testing.T, tempDir string) {
				// Create existing file first
				existingPath := filepath.Join(tempDir, "existing.txt")
				if err := os.WriteFile(existingPath, []byte("Old content"), 0644); err != nil {
					t.Fatalf("Failed to create existing file: %v", err)
				}

				// Write with overwrite
				ctx := context.Background()
				result, _ := tool.Execute(ctx, map[string]interface{}{
					"path":      "existing.txt",
					"content":   "Updated content",
					"overwrite": true,
				})

				var resp ToolResponse
				if err := json.Unmarshal([]byte(result), &resp); err != nil {
					t.Errorf("Failed to parse JSON: %v", err)
					return
				}

				if !resp.Ok {
					t.Errorf("Overwrite failed: %s", resp.Error)
					return
				}

				content, err := os.ReadFile(existingPath)
				if err != nil {
					t.Errorf("Failed to read file: %v", err)
					return
				}

				if string(content) != "Updated content" {
					t.Errorf("Content = %q, want %q", string(content), "Updated content")
				}
			},
		},
		{
			name: "create_file_in_subdirectory",
			input: map[string]interface{}{
				"path":           "subdir/nested/file.txt",
				"content":        "Nested content",
				"create_parents": true,
			},
			wantOk: true,
			verify: func(t *testing.T, tempDir string) {
				content, err := os.ReadFile(filepath.Join(tempDir, "subdir", "nested", "file.txt"))
				if err != nil {
					t.Errorf("Failed to read nested file: %v", err)
					return
				}
				if string(content) != "Nested content" {
					t.Errorf("File content = %q, want %q", string(content), "Nested content")
				}
			},
		},
		{
			name: "path_traversal_attempt",
			input: map[string]interface{}{
				"path":    "../../../etc/passwd",
				"content": "malicious",
			},
			wantOk: false,
			verify: nil,
		},
		{
			name: "missing_content_parameter",
			input: map[string]interface{}{
				"path": "test.txt",
			},
			wantOk: false,
			verify: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := tool.Execute(ctx, tt.input)

			if err != nil {
				t.Errorf("Execute() unexpected error = %v", err)
				return
			}

			var resp ToolResponse
			if err := json.Unmarshal([]byte(result), &resp); err != nil {
				t.Errorf("Failed to parse JSON response: %v", err)
				return
			}

			if resp.Ok != tt.wantOk {
				t.Errorf("Execute() ok = %v, want %v", resp.Ok, tt.wantOk)
				if !resp.Ok {
					t.Logf("Error message: %s", resp.Error)
				}
				return
			}

			if tt.verify != nil {
				tt.verify(t, tempDir)
			}
		})
	}
}

// TestListTool tests the directory listing functionality with JSON output.
func TestListTool(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "listtool_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test file structure
	files := map[string]string{
		"file1.txt":              "content1",
		"file2.go":               "package main",
		"subdir/file3.txt":       "content3",
		"subdir/file4.go":        "package sub",
		"subdir/nested/file5.md": "# Nested",
		".hidden":                "hidden content",
	}

	for path, content := range files {
		fullPath := filepath.Join(tempDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", path, err)
		}
	}

	tool := NewListTool(tempDir, 1000)

	tests := []struct {
		name          string
		input         map[string]interface{}
		wantOk        bool
		wantFileCount int // Minimum expected file count
		mustContain   []string
		mustNotContain []string
	}{
		{
			name: "list_root_directory",
			input: map[string]interface{}{
				"dir": ".",
			},
			wantOk:        true,
			wantFileCount: 2, // file1.txt, file2.go, subdir (at least 2 non-hidden files)
			mustContain:   []string{"file1.txt", "file2.go"},
			mustNotContain: []string{".hidden"},
		},
		{
			name: "list_recursively",
			input: map[string]interface{}{
				"dir":       ".",
				"recursive": true,
			},
			wantOk:        true,
			wantFileCount: 5, // All non-hidden files
			mustContain:   []string{"file1.txt", "file2.go", "file3.txt", "file4.go", "file5.md"},
			mustNotContain: []string{".hidden"},
		},
		{
			name: "list_with_include_glob",
			input: map[string]interface{}{
				"dir":           ".",
				"recursive":     true,
				"include_globs": []interface{}{"*.go"},
			},
			wantOk:        true,
			wantFileCount: 2, // file2.go and file4.go
			mustContain:   []string{"file2.go", "file4.go"},
			mustNotContain: []string{"file1.txt", "file3.txt"},
		},
		{
			name: "list_with_exclude_glob",
			input: map[string]interface{}{
				"dir":           ".",
				"recursive":     true,
				"exclude_globs": []interface{}{"*.txt"},
			},
			wantOk:        true,
			wantFileCount: 2, // file2.go, file4.go, file5.md (at least 2)
			mustContain:   []string{"file2.go", "file4.go"},
			mustNotContain: []string{"file1.txt", "file3.txt"},
		},
		{
			name: "list_subdirectory",
			input: map[string]interface{}{
				"dir": "subdir",
			},
			wantOk:        true,
			wantFileCount: 2, // file3.txt, file4.go (non-recursive)
			mustContain:   []string{"file3.txt", "file4.go"},
			mustNotContain: []string{"file1.txt", "file5.md"}, // file5.md is in nested
		},
		{
			name: "list_non_existent_directory",
			input: map[string]interface{}{
				"dir": "nonexistent",
			},
			wantOk: false,
		},
		{
			name: "path_traversal_attempt",
			input: map[string]interface{}{
				"dir": "../../../etc",
			},
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := tool.Execute(ctx, tt.input)

			if err != nil {
				t.Errorf("Execute() unexpected error = %v", err)
				return
			}

			var resp ToolResponse
			if err := json.Unmarshal([]byte(result), &resp); err != nil {
				t.Errorf("Failed to parse JSON response: %v", err)
				return
			}

			if resp.Ok != tt.wantOk {
				t.Errorf("Execute() ok = %v, want %v", resp.Ok, tt.wantOk)
				if !resp.Ok {
					t.Logf("Error message: %s", resp.Error)
				}
				return
			}

			if tt.wantOk {
				dataMap, ok := resp.Data.(map[string]interface{})
				if !ok {
					t.Errorf("Data is not a map")
					return
				}

				filesArray, ok := dataMap["files"].([]interface{})
				if !ok {
					t.Errorf("Files is not an array")
					return
				}

				// Check file count
				if len(filesArray) < tt.wantFileCount {
					t.Errorf("File count = %d, want at least %d", len(filesArray), tt.wantFileCount)
				}

				// Extract file names
				var fileNames []string
				for _, fileItem := range filesArray {
					fileMap, ok := fileItem.(map[string]interface{})
					if !ok {
						continue
					}
					if name, ok := fileMap["name"].(string); ok {
						fileNames = append(fileNames, name)
					}
				}

				// Check must contain
				for _, mustHave := range tt.mustContain {
					found := false
					for _, name := range fileNames {
						if strings.Contains(name, mustHave) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Output does not contain expected file %q. Files: %v", mustHave, fileNames)
					}
				}

				// Check must not contain
				for _, mustNotHave := range tt.mustNotContain {
					for _, name := range fileNames {
						if strings.Contains(name, mustNotHave) {
							t.Errorf("Output contains unexpected file %q", mustNotHave)
						}
					}
				}
			}
		})
	}
}

// TestPathSecurity tests path security validation across all tools.
func TestPathSecurity(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "security_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	readTool := NewReadTool(tempDir, 1024*1024)
	writeTool := NewWriteTool(tempDir, 1024*1024)
	listTool := NewListTool(tempDir, 1000)

	dangerousPaths := []string{
		"../../../etc/passwd",
		"../../.ssh/id_rsa",
		"/etc/passwd",
		"./../../secret.txt",
	}

	for _, path := range dangerousPaths {
		t.Run("read_"+path, func(t *testing.T) {
			ctx := context.Background()
			result, _ := readTool.Execute(ctx, map[string]interface{}{"path": path})

			var resp ToolResponse
			if err := json.Unmarshal([]byte(result), &resp); err != nil {
				t.Errorf("Failed to parse JSON: %v", err)
				return
			}

			if resp.Ok {
				t.Errorf("Expected security error for path %s", path)
			}
		})

		t.Run("write_"+path, func(t *testing.T) {
			ctx := context.Background()
			result, _ := writeTool.Execute(ctx, map[string]interface{}{
				"path":    path,
				"content": "test",
			})

			var resp ToolResponse
			if err := json.Unmarshal([]byte(result), &resp); err != nil {
				t.Errorf("Failed to parse JSON: %v", err)
				return
			}

			if resp.Ok {
				t.Errorf("Expected security error for path %s", path)
			}
		})

		t.Run("list_"+path, func(t *testing.T) {
			ctx := context.Background()
			result, _ := listTool.Execute(ctx, map[string]interface{}{"dir": path})

			var resp ToolResponse
			if err := json.Unmarshal([]byte(result), &resp); err != nil {
				t.Errorf("Failed to parse JSON: %v", err)
				return
			}

			if resp.Ok {
				t.Errorf("Expected security error for path %s", path)
			}
		})
	}
}
