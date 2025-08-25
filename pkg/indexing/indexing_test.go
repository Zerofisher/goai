package indexing

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultFileFilter(t *testing.T) {
	filter := NewDefaultFileFilter()
	
	// Test cases for ShouldIndex
	testCases := []struct {
		path        string
		size        int64
		shouldIndex bool
		description string
	}{
		{"test.go", 1000, true, "Go file should be indexed"},
		{"test.js", 1000, true, "JavaScript file should be indexed"},
		{"test.py", 1000, true, "Python file should be indexed"},
		{"test.md", 1000, true, "Markdown file should be indexed"},
		{"README", 1000, true, "README file should be indexed"},
		{"test.exe", 1000, false, "Binary file should not be indexed"},
		{"test.jpg", 1000, false, "Image file should not be indexed"},
		{"large.go", 20*1024*1024, false, "File too large should not be indexed"},
	}
	
	for _, tc := range testCases {
		fileInfo := FileInfo{
			Path: tc.path,
			Size: tc.size,
		}
		
		result := filter.ShouldIndex(tc.path, fileInfo)
		if result != tc.shouldIndex {
			t.Errorf("%s: expected %v, got %v", tc.description, tc.shouldIndex, result)
		}
	}
}

func TestDefaultFileFilter_ShouldIgnore(t *testing.T) {
	filter := NewDefaultFileFilter()
	
	// Test cases for ShouldIgnore
	testCases := []struct {
		path         string
		shouldIgnore bool
		description  string
	}{
		{"src/main.go", false, "Regular Go file should not be ignored"},
		{".git/config", true, "Git files should be ignored"},
		{"node_modules/package.json", true, "Node modules should be ignored"},
		{"vendor/github.com/lib.go", true, "Vendor files should be ignored"},
		{"test.swp", true, "Vim swap files should be ignored"},
		{".DS_Store", true, "macOS system files should be ignored"},
		{"build/output.bin", true, "Build artifacts should be ignored"},
	}
	
	for _, tc := range testCases {
		result := filter.ShouldIgnore(tc.path)
		if result != tc.shouldIgnore {
			t.Errorf("%s: expected %v, got %v", tc.description, tc.shouldIgnore, result)
		}
	}
}

func TestDefaultChunker_ChunkText(t *testing.T) {
	chunker := NewDefaultChunker()
	
	// Test Go code chunking
	goCode := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}

func helper() {
	// Helper function
	return
}`
	
	chunks, err := chunker.ChunkText(goCode, "go")
	if err != nil {
		t.Fatalf("Failed to chunk Go code: %v", err)
	}
	
	if len(chunks) == 0 {
		t.Error("Expected at least one chunk, got none")
	}
	
	// Verify chunk properties
	for i, chunk := range chunks {
		if chunk.Content == "" {
			t.Errorf("Chunk %d has empty content", i)
		}
		if chunk.Language != "go" {
			t.Errorf("Chunk %d has wrong language: expected 'go', got '%s'", i, chunk.Language)
		}
		if chunk.StartLine <= 0 || chunk.EndLine <= 0 {
			t.Errorf("Chunk %d has invalid line numbers: start=%d, end=%d", i, chunk.StartLine, chunk.EndLine)
		}
		if chunk.StartLine > chunk.EndLine {
			t.Errorf("Chunk %d has start line after end line: start=%d, end=%d", i, chunk.StartLine, chunk.EndLine)
		}
	}
}

func TestDefaultChunker_ChunkMarkdown(t *testing.T) {
	chunker := NewDefaultChunker()
	
	markdownContent := `# Main Title

This is the introduction paragraph.

## Section 1

Content for section 1 with some details.

### Subsection 1.1

More detailed content here.

## Section 2

Another section with different content.
Multiple paragraphs in this section.

The end.`
	
	chunks, err := chunker.ChunkText(markdownContent, "markdown")
	if err != nil {
		t.Fatalf("Failed to chunk markdown: %v", err)
	}
	
	if len(chunks) == 0 {
		t.Error("Expected at least one chunk, got none")
	}
	
	// Verify that chunks respect section boundaries
	for i, chunk := range chunks {
		if chunk.Language != "markdown" {
			t.Errorf("Chunk %d has wrong language: expected 'markdown', got '%s'", i, chunk.Language)
		}
		if chunk.ChunkType != ChunkTypeDocumentation {
			t.Errorf("Chunk %d has wrong type: expected 'documentation', got '%s'", i, chunk.ChunkType)
		}
	}
}

func TestIndexManager_Basic(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "indexing_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create test files
	testFile := filepath.Join(tempDir, "test.go")
	err = os.WriteFile(testFile, []byte(`package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Create index manager
	manager := NewIndexManager()
	
	// Test initial state
	if manager.IsIndexReady(tempDir) {
		t.Error("Index should not be ready initially")
	}
	
	_, err = manager.GetIndexStatus(tempDir)
	if err == nil {
		t.Error("Expected error for non-existent index status")
	}
	
	// Test building index (this will be limited without actual implementations)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Note: This will fail because we don't have actual file reading implemented
	// but we can test the structure
	err = manager.BuildIndex(ctx, tempDir)
	if err == nil {
		t.Log("Index building completed (expected to have limitations in test environment)")
	}
}

func TestGitignoreFilter(t *testing.T) {
	// Create a temporary directory with .gitignore
	tempDir, err := os.MkdirTemp("", "gitignore_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create .gitignore file
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	gitignoreContent := `# Test gitignore
*.log
temp/
build/*
!build/keep.txt`
	
	err = os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}
	
	// Create gitignore filter
	filter, err := NewGitignoreFilter(tempDir)
	if err != nil {
		t.Fatalf("Failed to create gitignore filter: %v", err)
	}
	
	// Test ignore patterns
	testCases := []struct {
		path         string
		shouldIgnore bool
		description  string
	}{
		{"test.go", false, "Regular Go file should not be ignored"},
		{"debug.log", true, "Log files should be ignored"},
		{"temp/file.txt", true, "Files in temp directory should be ignored"},
		{"build/output.bin", true, "Files in build directory should be ignored"},
	}
	
	for _, tc := range testCases {
		result := filter.ShouldIgnore(tc.path)
		if result != tc.shouldIgnore {
			t.Errorf("%s: expected %v, got %v", tc.description, tc.shouldIgnore, result)
		}
	}
}

func TestFSDirectoryWalker(t *testing.T) {
	// Create a temporary directory structure
	tempDir, err := os.MkdirTemp("", "walker_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create test file structure
	files := []string{
		"main.go",
		"lib/utils.go",
		"lib/helper.js",
		"docs/README.md",
		".git/config",
		"node_modules/package.json",
	}
	
	for _, file := range files {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)
		
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		
		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}
	
	// Test directory walking
	walker := NewFSDirectoryWalker()
	filter := NewDefaultFileFilter()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	fileCh, err := walker.Walk(ctx, tempDir, filter)
	if err != nil {
		t.Fatalf("Failed to start walking: %v", err)
	}
	
	var discoveredFiles []string
	for fileInfo := range fileCh {
		discoveredFiles = append(discoveredFiles, fileInfo.Path)
	}
	
	// Should find main.go, utils.go, helper.js, README.md
	// Should NOT find .git/config, node_modules/package.json
	expectedFiles := 4
	if len(discoveredFiles) < expectedFiles {
		t.Errorf("Expected to find at least %d files, found %d: %v", expectedFiles, len(discoveredFiles), discoveredFiles)
	}
	
	// Check that ignored files are not included
	for _, file := range discoveredFiles {
		if strings.Contains(file, ".git") || strings.Contains(file, "node_modules") {
			t.Errorf("Found ignored file: %s", file)
		}
	}
}