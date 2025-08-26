package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestToolRegistry(t *testing.T) {
	registry := NewToolRegistry()
	
	// Test registering a tool
	readTool := NewReadFileTool()
	err := registry.Register(readTool)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}
	
	// Test getting a tool
	tool, exists := registry.Get("readFile")
	if !exists {
		t.Fatal("Tool not found in registry")
	}
	
	if tool.Name() != "readFile" {
		t.Errorf("Expected tool name 'readFile', got '%s'", tool.Name())
	}
	
	// Test listing tools
	tools := registry.List()
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}
	
	// Test duplicate registration
	err = registry.Register(readTool)
	if err == nil {
		t.Error("Expected error when registering duplicate tool")
	}
}

func TestToolManager(t *testing.T) {
	registry := NewToolRegistry()
	confirmationHandler := NewMockConfirmationHandler(true)
	manager := NewToolManager(registry, confirmationHandler)
	
	// Register a tool
	readTool := NewReadFileTool()
	err := manager.RegisterTool(readTool)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}
	
	// Test getting tool
	tool, err := manager.GetTool("readFile")
	if err != nil {
		t.Fatalf("Failed to get tool: %v", err)
	}
	
	if tool.Name() != "readFile" {
		t.Errorf("Expected tool name 'readFile', got '%s'", tool.Name())
	}
	
	// Test listing tools
	tools := manager.ListTools("")
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}
	
	// Test listing by category
	fileTools := manager.ListTools("file")
	if len(fileTools) != 1 {
		t.Errorf("Expected 1 file tool, got %d", len(fileTools))
	}
	
	systemTools := manager.ListTools("system")
	if len(systemTools) != 0 {
		t.Errorf("Expected 0 system tools, got %d", len(systemTools))
	}
}

func TestReadFileTool(t *testing.T) {
	// Create a temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!"
	
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Test reading the file
	tool := NewReadFileTool()
	params := map[string]any{
		"path": testFile,
	}
	
	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Tool execution failed: %v", err)
	}
	
	if !result.Success {
		t.Errorf("Tool execution was not successful: %s", result.Error)
	}
	
	if result.Data != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, result.Data)
	}
}

func TestWriteFileTool(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "subdir", "test.txt")
	testContent := "Hello, World!"
	
	tool := NewWriteFileTool()
	params := map[string]any{
		"path":             testFile,
		"content":          testContent,
		"createDirectories": true,
	}
	
	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Tool execution failed: %v", err)
	}
	
	if !result.Success {
		t.Errorf("Tool execution was not successful: %s", result.Error)
	}
	
	// Verify the file was created
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}
	
	if string(content) != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, string(content))
	}
}

func TestEditFileTool(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	originalContent := "Hello, World!"
	newContent := "Hello, Go!"
	
	// Create original file
	err := os.WriteFile(testFile, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Test editing the file
	tool := NewEditFileTool()
	params := map[string]any{
		"path":       testFile,
		"oldContent": "World",
		"newContent": "Go",
	}
	
	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Tool execution failed: %v", err)
	}
	
	if !result.Success {
		t.Errorf("Tool execution was not successful: %s", result.Error)
	}
	
	// Verify the file was edited
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read edited file: %v", err)
	}
	
	if string(content) != newContent {
		t.Errorf("Expected content '%s', got '%s'", newContent, string(content))
	}
}

func TestListFilesTool(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create some test files
	files := []string{"test1.txt", "test2.go", "README.md"}
	for _, file := range files {
		err := os.WriteFile(filepath.Join(tempDir, file), []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}
	
	// Test listing files
	tool := NewListFilesTool()
	params := map[string]any{
		"path": tempDir,
	}
	
	result, err := tool.Execute(context.Background(), params)
	if err != nil {
		t.Fatalf("Tool execution failed: %v", err)
	}
	
	if !result.Success {
		t.Errorf("Tool execution was not successful: %s", result.Error)
	}
	
	// Verify results
	fileList, ok := result.Data.([]map[string]any)
	if !ok {
		t.Fatal("Expected result data to be a file list")
	}
	
	if len(fileList) != len(files) {
		t.Errorf("Expected %d files, got %d", len(files), len(fileList))
	}
}

func TestToolFactory(t *testing.T) {
	factory := NewToolFactory()
	
	// Create tool manager without index manager
	manager, err := factory.CreateTestToolManager(nil, true)
	if err != nil {
		t.Fatalf("Failed to create tool manager: %v", err)
	}
	
	// Verify tools are registered
	tools := manager.ListTools("")
	if len(tools) == 0 {
		t.Error("Expected some tools to be registered")
	}
	
	// Test getting tool info
	toolsInfo := factory.GetAvailableToolsInfo()
	if len(toolsInfo) == 0 {
		t.Error("Expected some tool info")
	}
}

func TestParameterValidation(t *testing.T) {
	registry := NewToolRegistry()
	manager := NewToolManager(registry, NewMockConfirmationHandler(true))
	
	// Register a tool
	readTool := NewReadFileTool()
	err := manager.RegisterTool(readTool)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}
	
	// Test with missing required parameter
	params := map[string]any{}
	
	err = manager.ValidateParameters("readFile", params)
	if err == nil {
		t.Error("Expected validation error for missing required parameter")
	}
	
	// Test with valid parameters
	params = map[string]any{
		"path": "/some/path",
	}
	
	err = manager.ValidateParameters("readFile", params)
	if err != nil {
		t.Errorf("Unexpected validation error: %v", err)
	}
}

func TestArrayParameterValidation(t *testing.T) {
	// Create a mock tool that expects an array parameter
	mockTool := &MockArrayTool{}
	registry := NewToolRegistry()
	manager := NewToolManager(registry, NewMockConfirmationHandler(true))
	
	err := manager.RegisterTool(mockTool)
	if err != nil {
		t.Fatalf("Failed to register mock tool: %v", err)
	}
	
	// Test various slice types
	testCases := []struct {
		name  string
		value any
		valid bool
	}{
		{"string slice", []string{"a", "b", "c"}, true},
		{"int slice", []int{1, 2, 3}, true},
		{"float64 slice", []float64{1.1, 2.2, 3.3}, true},
		{"any slice", []any{"mixed", 42, true}, true},
		{"empty slice", []string{}, true},
		{"not a slice", "not an array", false},
		{"map", map[string]string{"key": "value"}, false},
		{"nil", nil, false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := map[string]any{
				"items": tc.value,
			}
			
			err := manager.ValidateParameters("mockArrayTool", params)
			
			if tc.valid && err != nil {
				t.Errorf("Expected %s to be valid, but got error: %v", tc.name, err)
			}
			if !tc.valid && err == nil {
				t.Errorf("Expected %s to be invalid, but no error was returned", tc.name)
			}
		})
	}
}

// MockArrayTool for testing array parameter validation
type MockArrayTool struct{}

func (t *MockArrayTool) Name() string {
	return "mockArrayTool"
}

func (t *MockArrayTool) Description() string {
	return "Mock tool for testing array parameters"
}

func (t *MockArrayTool) Parameters() ParameterSchema {
	return ParameterSchema{
		Required: []string{"items"},
		Properties: map[string]ParameterProperty{
			"items": {
				Type:        "array",
				Description: "Array of items",
			},
		},
	}
}

func (t *MockArrayTool) Execute(ctx context.Context, params map[string]any) (*ToolResult, error) {
	return &ToolResult{Success: true}, nil
}

func (t *MockArrayTool) RequiresConfirmation() bool {
	return false
}

func (t *MockArrayTool) Category() string {
	return "test"
}

// Test search tools
func TestSearchCodeTool_WithoutIndexManager(t *testing.T) {
	tool := NewSearchCodeTool(nil)
	
	if tool.Name() != "searchCode" {
		t.Errorf("Expected name 'searchCode', got '%s'", tool.Name())
	}
	
	if tool.Category() != "search" {
		t.Errorf("Expected category 'search', got '%s'", tool.Category())
	}
	
	if tool.RequiresConfirmation() {
		t.Error("SearchCodeTool should not require confirmation")
	}
	
	// Test parameter schema
	params := tool.Parameters()
	if len(params.Required) != 1 || params.Required[0] != "query" {
		t.Errorf("Expected required parameter 'query', got %v", params.Required)
	}
	
	if params.Properties["query"].Type != "string" {
		t.Error("Expected query parameter to be string type")
	}
	
	// Note: Cannot test execution without index manager as it would cause nil pointer dereference
	// This is expected behavior - the tool requires a valid index manager to function
}

func TestSearchCodeTool_InvalidParameters(t *testing.T) {
	tool := NewSearchCodeTool(nil)
	
	// Test with missing query parameter
	result, err := tool.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Expected execution to fail with missing query")
	}
	if !strings.Contains(result.Error, "query parameter must be a string") {
		t.Errorf("Expected query error, got: %s", result.Error)
	}
	
	// Test with invalid search type
	result, err = tool.Execute(context.Background(), map[string]any{
		"query": "test",
		"type":  "invalid",
	})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Expected execution to fail with invalid search type")
	}
	if !strings.Contains(result.Error, "unsupported search type") {
		t.Errorf("Expected unsupported search type error, got: %s", result.Error)
	}
}

func TestListFilesTool_BasicFunctionality(t *testing.T) {
	tool := NewListFilesTool()
	
	if tool.Name() != "listFiles" {
		t.Errorf("Expected name 'listFiles', got '%s'", tool.Name())
	}
	
	if tool.Category() != "search" {
		t.Errorf("Expected category 'search', got '%s'", tool.Category())
	}
	
	if tool.RequiresConfirmation() {
		t.Error("ListFilesTool should not require confirmation")
	}
	
	// Test parameter schema
	params := tool.Parameters()
	if len(params.Required) != 0 {
		t.Errorf("Expected no required parameters, got %v", params.Required)
	}
	
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	// Create test files
	testFiles := []string{"test1.txt", "test2.go", "hidden.txt", ".hidden"}
	for _, file := range testFiles {
		content := "test content"
		if strings.HasPrefix(file, ".") {
			content = "hidden content"
		}
		err := os.WriteFile(filepath.Join(tempDir, file), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}
	
	// Test basic listing
	result, err := tool.Execute(context.Background(), map[string]any{
		"path": tempDir,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}
	
	fileList, ok := result.Data.([]map[string]any)
	if !ok {
		t.Fatal("Expected result data to be a file list")
	}
	
	// Should find 3 files (excluding hidden by default)
	expectedCount := 3
	if len(fileList) != expectedCount {
		t.Errorf("Expected %d files, got %d", expectedCount, len(fileList))
	}
}

func TestListFilesTool_WithOptions(t *testing.T) {
	tool := NewListFilesTool()
	tempDir := t.TempDir()
	
	// Create test files with different extensions and a subdirectory
	files := []string{"test1.txt", "test2.go", ".hidden", "subdir/nested.txt"}
	for _, file := range files {
		filePath := filepath.Join(tempDir, file)
		// Create directory if needed
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		err := os.WriteFile(filePath, []byte("content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}
	
	// Test with includeHidden
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":          tempDir,
		"includeHidden": true,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}
	
	fileList, ok := result.Data.([]map[string]any)
	if !ok {
		t.Fatal("Expected result data to be a file list")
	}
	
	// Should find 4 items (3 files + 1 directory, including hidden)
	expectedCount := 4
	if len(fileList) != expectedCount {
		t.Errorf("Expected %d items with hidden files, got %d", expectedCount, len(fileList))
	}
	
	// Test with pattern filter
	result, err = tool.Execute(context.Background(), map[string]any{
		"path":    tempDir,
		"pattern": "*.go",
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}
	
	fileList, ok = result.Data.([]map[string]any)
	if !ok {
		t.Fatal("Expected result data to be a file list")
	}
	
	// Should find 1 .go file
	if len(fileList) != 1 {
		t.Errorf("Expected 1 .go file, got %d", len(fileList))
	}
	
	// Test recursive option
	result, err = tool.Execute(context.Background(), map[string]any{
		"path":      tempDir,
		"recursive": true,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}
	
	fileList, ok = result.Data.([]map[string]any)
	if !ok {
		t.Fatal("Expected result data to be a file list")
	}
	
	// Should find more files with recursive search (including nested.txt)
	if len(fileList) < 4 {
		t.Errorf("Expected at least 4 files with recursive search, got %d", len(fileList))
	}
}

func TestListFilesTool_InvalidPath(t *testing.T) {
	tool := NewListFilesTool()
	
	// Test with non-existent path
	result, err := tool.Execute(context.Background(), map[string]any{
		"path": "/non/existent/path",
	})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Expected execution to fail with invalid path")
	}
	if !strings.Contains(result.Error, "path does not exist") {
		t.Errorf("Expected path error, got: %s", result.Error)
	}
}

func TestViewDiffTool_BasicFunctionality(t *testing.T) {
	tool := NewViewDiffTool()
	
	if tool.Name() != "viewDiff" {
		t.Errorf("Expected name 'viewDiff', got '%s'", tool.Name())
	}
	
	if tool.Category() != "search" {
		t.Errorf("Expected category 'search', got '%s'", tool.Category())
	}
	
	if tool.RequiresConfirmation() {
		t.Error("ViewDiffTool should not require confirmation")
	}
	
	// Test parameter schema
	params := tool.Parameters()
	if len(params.Required) != 1 || params.Required[0] != "path" {
		t.Errorf("Expected required parameter 'path', got %v", params.Required)
	}
}

func TestViewDiffTool_InvalidParameters(t *testing.T) {
	tool := NewViewDiffTool()
	
	// Test with missing path parameter
	result, err := tool.Execute(context.Background(), map[string]any{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Expected execution to fail with missing path")
	}
	if !strings.Contains(result.Error, "path parameter must be a string") {
		t.Errorf("Expected path error, got: %s", result.Error)
	}
	
	// Test with non-existent file
	result, err = tool.Execute(context.Background(), map[string]any{
		"path": "/non/existent/file.txt",
	})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result.Success {
		t.Error("Expected execution to fail with non-existent file")
	}
	if !strings.Contains(result.Error, "file does not exist") {
		t.Errorf("Expected file not found error, got: %s", result.Error)
	}
}

func TestViewDiffTool_TwoFiles(t *testing.T) {
	tool := NewViewDiffTool()
	tempDir := t.TempDir()
	
	// Create two test files with different content
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")
	
	content1 := "line 1\nline 2\nline 3\n"
	content2 := "line 1\nmodified line 2\nline 3\nline 4\n"
	
	err := os.WriteFile(file1, []byte(content1), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	err = os.WriteFile(file2, []byte(content2), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Test diff between two files
	result, err := tool.Execute(context.Background(), map[string]any{
		"path":      file1,
		"compareTo": file2,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got error: %s", result.Error)
	}
	
	diffOutput, ok := result.Data.(string)
	if !ok {
		t.Fatal("Expected diff output to be a string")
	}
	
	// Check that diff contains expected markers
	if !strings.Contains(diffOutput, "---") || !strings.Contains(diffOutput, "+++") {
		t.Error("Expected diff to contain unified diff headers")
	}
	
	if !strings.Contains(diffOutput, "-line 2") || !strings.Contains(diffOutput, "+modified line 2") {
		t.Error("Expected diff to show the modified line")
	}
}

func TestViewDiffTool_GenerateUnifiedDiff(t *testing.T) {
	tool := NewViewDiffTool()
	
	lines1 := []string{"line 1", "line 2", "line 3"}
	lines2 := []string{"line 1", "modified line 2", "line 3", "line 4"}
	
	diff := tool.generateUnifiedDiff("file1.txt", "file2.txt", lines1, lines2)
	
	if !strings.Contains(diff, "--- file1.txt") {
		t.Error("Expected diff to contain source file header")
	}
	
	if !strings.Contains(diff, "+++ file2.txt") {
		t.Error("Expected diff to contain target file header")
	}
	
	if !strings.Contains(diff, "-line 2") {
		t.Error("Expected diff to show removed line")
	}
	
	if !strings.Contains(diff, "+modified line 2") {
		t.Error("Expected diff to show added line")
	}
	
	if !strings.Contains(diff, "+line 4") {
		t.Error("Expected diff to show new line")
	}
}

func TestFileInfo_Structure(t *testing.T) {
	now := time.Now()
	fileInfo := FileInfo{
		Path:    "/test/path/file.txt",
		Name:    "file.txt",
		Size:    1024,
		IsDir:   false,
		ModTime: now,
	}
	
	if fileInfo.Path != "/test/path/file.txt" {
		t.Errorf("Expected path '/test/path/file.txt', got '%s'", fileInfo.Path)
	}
	
	if fileInfo.Name != "file.txt" {
		t.Errorf("Expected name 'file.txt', got '%s'", fileInfo.Name)
	}
	
	if fileInfo.Size != 1024 {
		t.Errorf("Expected size 1024, got %d", fileInfo.Size)
	}
	
	if fileInfo.IsDir {
		t.Error("Expected IsDir to be false")
	}
	
	if !fileInfo.ModTime.Equal(now) {
		t.Errorf("Expected ModTime %v, got %v", now, fileInfo.ModTime)
	}
}