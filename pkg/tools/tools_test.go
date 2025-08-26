package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
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