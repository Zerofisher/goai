package search

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Zerofisher/goai/pkg/tools"
)

// TestSearchTool tests the SearchTool implementation
func TestSearchTool(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "search-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir) // Clean up test directory
	}()

	// Create test files
	createTestFiles(t, tempDir)

	// Create search tool
	validator := tools.NewSecurityValidator(tempDir)
	searchTool := NewSearchTool(tempDir, validator)

	// Test basic properties
	t.Run("Properties", func(t *testing.T) {
		if searchTool.Name() != "search" {
			t.Errorf("Expected name 'search', got %s", searchTool.Name())
		}

		if searchTool.Description() == "" {
			t.Error("Expected non-empty description")
		}

		schema := searchTool.InputSchema()
		if schema == nil {
			t.Error("Expected non-nil input schema")
		}
	})

	// Test validation
	t.Run("Validation", func(t *testing.T) {
		tests := []struct {
			name    string
			input   map[string]interface{}
			wantErr bool
		}{
			{
				name: "valid code search",
				input: map[string]interface{}{
					"pattern": "func",
					"type":    "code",
				},
				wantErr: false,
			},
			{
				name: "valid symbol search",
				input: map[string]interface{}{
					"pattern": "TestFunction",
					"type":    "symbol",
				},
				wantErr: false,
			},
			{
				name: "missing pattern",
				input: map[string]interface{}{
					"type": "code",
				},
				wantErr: true,
			},
			{
				name: "missing type",
				input: map[string]interface{}{
					"pattern": "test",
				},
				wantErr: true,
			},
			{
				name: "invalid type",
				input: map[string]interface{}{
					"pattern": "test",
					"type":    "invalid",
				},
				wantErr: true,
			},
			{
				name: "invalid max_results",
				input: map[string]interface{}{
					"pattern":     "test",
					"type":        "code",
					"max_results": 2000.0,
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := searchTool.Validate(tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	// Test code search
	t.Run("CodeSearch", func(t *testing.T) {
		options := SearchOptions{
			CaseSensitive: false,
			MaxResults:    10,
		}

		results, err := searchTool.SearchCode("test", options)
		if err != nil {
			t.Fatalf("SearchCode failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result")
		}

		// Check result structure
		for _, result := range results {
			if result.File == "" {
				t.Error("Result has empty file path")
			}
			if result.Line <= 0 {
				t.Error("Result has invalid line number")
			}
		}
	})

	// Test symbol search
	t.Run("SymbolSearch", func(t *testing.T) {
		locations, err := searchTool.SearchSymbol("TestFunction")
		if err != nil {
			t.Fatalf("SearchSymbol failed: %v", err)
		}

		if len(locations) == 0 {
			t.Error("Expected at least one location")
		}

		// Check location structure
		for _, loc := range locations {
			if loc.File == "" {
				t.Error("Location has empty file path")
			}
			if loc.Line <= 0 {
				t.Error("Location has invalid line number")
			}
			if loc.Type == "" {
				t.Error("Location has empty type")
			}
		}
	})

	// Test Execute method
	t.Run("Execute", func(t *testing.T) {
		ctx := context.Background()

		tests := []struct {
			name    string
			input   map[string]interface{}
			wantErr bool
		}{
			{
				name: "execute code search",
				input: map[string]interface{}{
					"pattern":        "func",
					"type":           "code",
					"case_sensitive": false,
					"max_results":    5.0,
				},
				wantErr: false,
			},
			{
				name: "execute symbol search",
				input: map[string]interface{}{
					"pattern": "TestFunction",
					"type":    "symbol",
				},
				wantErr: false,
			},
			{
				name: "execute with file pattern",
				input: map[string]interface{}{
					"pattern":      "test",
					"type":         "code",
					"file_pattern": "*.go",
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := searchTool.Execute(ctx, tt.input)
				if (err != nil) != tt.wantErr {
					t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				}
				if !tt.wantErr && result == "" {
					t.Error("Expected non-empty result")
				}
			})
		}
	})
}

// TestIndexer tests the Indexer implementation
func TestIndexer(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "indexer-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir) // Clean up test directory
	}()

	// Create test files
	createTestFiles(t, tempDir)

	// Create indexer
	indexer := NewIndexer(tempDir)

	// Test ignore patterns
	t.Run("IgnorePatterns", func(t *testing.T) {
		tests := []struct {
			path       string
			shouldIgnore bool
		}{
			{".git/config", true},
			{"node_modules/package.json", true},
			{"vendor/github.com/test", true},
			{"test.pyc", true},
			{"src/main.go", false},
			{"README.md", false},
		}

		for _, tt := range tests {
			t.Run(tt.path, func(t *testing.T) {
				result := indexer.ShouldIgnore(tt.path)
				if result != tt.shouldIgnore {
					t.Errorf("ShouldIgnore(%s) = %v, want %v", tt.path, result, tt.shouldIgnore)
				}
			})
		}
	})

	// Test refresh index
	t.Run("RefreshIndex", func(t *testing.T) {
		err := indexer.RefreshIndex()
		if err != nil {
			t.Fatalf("RefreshIndex failed: %v", err)
		}

		// Check that files were indexed
		stats := indexer.GetFileStats()
		if stats["total"] == 0 {
			t.Error("Expected at least one indexed file")
		}
	})

	// Test get relevant files
	t.Run("GetRelevantFiles", func(t *testing.T) {
		err := indexer.RefreshIndex()
		if err != nil {
			t.Fatal(err)
		}

		files, err := indexer.GetRelevantFiles("test", "*.go")
		if err != nil {
			t.Fatalf("GetRelevantFiles failed: %v", err)
		}

		if len(files) == 0 {
			t.Error("Expected at least one relevant file")
		}
	})

	// Test file type filter
	t.Run("FileTypeFilter", func(t *testing.T) {
		err := indexer.RefreshIndex()
		if err != nil {
			t.Fatal(err)
		}

		files, err := indexer.FileTypeFilter([]string{"Go"})
		if err != nil {
			t.Fatalf("FileTypeFilter failed: %v", err)
		}

		if len(files) == 0 {
			t.Error("Expected at least one Go file")
		}

		// Check that all returned files are Go files
		for _, file := range files {
			if filepath.Ext(file) != ".go" {
				t.Errorf("Expected .go file, got %s", file)
			}
		}
	})

	// Test recent files
	t.Run("GetRecentFiles", func(t *testing.T) {
		err := indexer.RefreshIndex()
		if err != nil {
			t.Fatal(err)
		}

		// Get files modified in last hour
		files, err := indexer.GetRecentFiles(1 * time.Hour)
		if err != nil {
			t.Fatalf("GetRecentFiles failed: %v", err)
		}

		// Since we just created the files, they should all be recent
		if len(files) == 0 {
			t.Error("Expected recent files")
		}
	})

	// Test quick search
	t.Run("QuickSearch", func(t *testing.T) {
		err := indexer.RefreshIndex()
		if err != nil {
			t.Fatal(err)
		}

		results, err := indexer.QuickSearch("func", "*.go")
		if err != nil {
			t.Fatalf("QuickSearch failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected search results")
		}
	})
}

// TestIndexCache tests the IndexCache implementation
func TestIndexCache(t *testing.T) {
	cache := NewIndexCache(100 * time.Millisecond)

	// Test set and get
	t.Run("SetAndGet", func(t *testing.T) {
		fileInfo := &FileInfo{
			Path:         "/test/file.go",
			Size:         100,
			ModifiedTime: time.Now(),
			Language:     "Go",
		}

		cache.Set("/test/file.go", fileInfo)

		retrieved, exists := cache.Get("/test/file.go")
		if !exists {
			t.Error("Expected file to exist in cache")
		}

		if retrieved.Path != fileInfo.Path {
			t.Errorf("Expected path %s, got %s", fileInfo.Path, retrieved.Path)
		}
	})

	// Test expiration
	t.Run("Expiration", func(t *testing.T) {
		fileInfo := &FileInfo{
			Path: "/test/expired.go",
		}

		cache.Set("/test/expired.go", fileInfo)

		// Wait for cache to expire
		time.Sleep(150 * time.Millisecond)

		if !cache.IsExpired() {
			t.Error("Expected cache to be expired")
		}

		_, exists := cache.Get("/test/expired.go")
		if exists {
			t.Error("Expected expired cache to return no results")
		}
	})

	// Test clear
	t.Run("Clear", func(t *testing.T) {
		fileInfo := &FileInfo{
			Path: "/test/clear.go",
		}

		cache.Set("/test/clear.go", fileInfo)
		cache.Clear()

		_, exists := cache.Get("/test/clear.go")
		if exists {
			t.Error("Expected cleared cache to be empty")
		}
	})
}

// TestLanguageDetection tests the language detection function
func TestLanguageDetection(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"main.go", "Go"},
		{"script.py", "Python"},
		{"app.js", "JavaScript"},
		{"index.ts", "TypeScript"},
		{"Main.java", "Java"},
		{"program.c", "C"},
		{"app.cpp", "C++"},
		{"Program.cs", "C#"},
		{"script.rb", "Ruby"},
		{"index.php", "PHP"},
		{"main.rs", "Rust"},
		{"app.swift", "Swift"},
		{"Main.kt", "Kotlin"},
		{"script.sh", "Shell"},
		{"query.sql", "SQL"},
		{"index.html", "HTML"},
		{"style.css", "CSS"},
		{"data.json", "JSON"},
		{"config.xml", "XML"},
		{"config.yaml", "YAML"},
		{"README.md", "Markdown"},
		{"notes.txt", "Text"},
		{"unknown.xyz", ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := detectLanguage(tt.path)
			if result != tt.expected {
				t.Errorf("detectLanguage(%s) = %s, want %s", tt.path, result, tt.expected)
			}
		})
	}
}

// Helper function to create test files
func createTestFiles(t *testing.T, dir string) {
	// Create test Go file
	goFile := filepath.Join(dir, "test.go")
	goContent := `package main

import "fmt"

func TestFunction() {
	fmt.Println("test")
}

func main() {
	TestFunction()
}
`
	if err := os.WriteFile(goFile, []byte(goContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create test Python file
	pyFile := filepath.Join(dir, "test.py")
	pyContent := `def test_function():
    print("test")

if __name__ == "__main__":
    test_function()
`
	if err := os.WriteFile(pyFile, []byte(pyContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create test JavaScript file
	jsFile := filepath.Join(dir, "test.js")
	jsContent := `function testFunction() {
    console.log("test");
}

testFunction();
`
	if err := os.WriteFile(jsFile, []byte(jsContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create README file
	readmeFile := filepath.Join(dir, "README.md")
	readmeContent := `# Test Project

This is a test project for search functionality.

## Features

- Test search
- Test indexing
`
	if err := os.WriteFile(readmeFile, []byte(readmeContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create subdirectory with file
	subDir := filepath.Join(dir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	subFile := filepath.Join(subDir, "sub.go")
	subContent := `package sub

func SubFunction() string {
	return "test from subdir"
}
`
	if err := os.WriteFile(subFile, []byte(subContent), 0644); err != nil {
		t.Fatal(err)
	}
}