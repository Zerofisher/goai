package indexing

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ExampleUsage demonstrates how to use the indexing system
func ExampleUsage() error {
	// Create a temporary directory for this example
	tempDir, err := os.MkdirTemp("", "goai_indexing_example")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	fmt.Printf("Created example directory: %s\n", tempDir)

	// Create some example files to index
	if err := createExampleFiles(tempDir); err != nil {
		return fmt.Errorf("failed to create example files: %w", err)
	}

	// Step 1: Create an index manager
	fmt.Println("\n🔧 Step 1: Creating index manager...")
	manager := NewIndexManager()
	defer func() { _ = manager.Close() }()

	// Step 2: Set up custom file filter if needed
	fmt.Println("\n🔍 Step 2: Configuring file filters...")
	customFilter := NewDefaultFileFilter()
	customFilter.AddIgnorePattern("*.tmp")
	manager.SetFileFilter(customFilter)

	// Step 3: Configure chunker
	fmt.Println("\n📄 Step 3: Configuring document chunker...")
	chunker := NewDefaultChunker()
	chunker.SetMaxChunkSize(800)  // Smaller chunks for this example
	chunker.SetOverlapSize(80)    // Small overlap
	manager.SetChunker(chunker)

	// Step 4: Register an FTS index
	fmt.Println("\n🗃️  Step 4: Setting up FTS index...")
	dbPath := filepath.Join(tempDir, "index.db")
	ftsIndex, err := NewFTSIndex(dbPath, chunker)
	if err != nil {
		return fmt.Errorf("failed to create FTS index: %w", err)
	}
	defer func() { _ = ftsIndex.Close() }()
	
	manager.RegisterIndex(ftsIndex)

	// Step 5: Build the index
	fmt.Println("\n🚀 Step 5: Building index...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	start := time.Now()
	if err := manager.BuildIndex(ctx, tempDir); err != nil {
		return fmt.Errorf("failed to build index: %w", err)
	}
	
	buildTime := time.Since(start)
	fmt.Printf("✅ Index built successfully in %v\n", buildTime)

	// Step 6: Check index status
	fmt.Println("\n📊 Step 6: Checking index status...")
	status, err := manager.GetIndexStatus(tempDir)
	if err != nil {
		return fmt.Errorf("failed to get index status: %w", err)
	}

	fmt.Printf("Index Status:\n")
	fmt.Printf("  - Ready: %v\n", status.IsReady)
	fmt.Printf("  - Total Files: %d\n", status.TotalFiles)
	fmt.Printf("  - Indexed Files: %d\n", status.IndexedFiles)
	fmt.Printf("  - Total Size: %d bytes\n", status.TotalSize)
	fmt.Printf("  - Index Versions: %v\n", status.IndexVersions)

	// Step 7: Perform searches
	fmt.Println("\n🔎 Step 7: Performing searches...")
	
	// Search for Go functions
	searchRequest := &SearchRequest{
		Query:       "func main",
		WorkingDir:  tempDir,
		MaxResults:  10,
		SearchTypes: []SearchType{SearchTypeFullText},
		IncludeContent: true,
	}

	result, err := manager.Search(ctx, searchRequest)
	if err != nil {
		return fmt.Errorf("failed to search: %w", err)
	}

	fmt.Printf("Search results for 'func main':\n")
	fmt.Printf("  - File: %s\n", result.FilePath)
	fmt.Printf("  - Snippet: %s\n", result.Snippet)
	fmt.Printf("  - Score: %.2f\n", result.Score)

	// Step 8: Get context items for reasoning
	fmt.Println("\n🧠 Step 8: Getting context items for reasoning...")
	contextOpts := &ContextOptions{
		MaxItems:           5,
		MaxCharsPerItem:    500,
		ContextTypes:       []ContextType{ContextTypeFunction, ContextTypeSnippet},
		RelevanceThreshold: 0.5,
	}

	contextItems, err := manager.GetContextItems(ctx, "HTTP server setup", contextOpts)
	if err != nil {
		return fmt.Errorf("failed to get context items: %w", err)
	}

	fmt.Printf("Found %d context items:\n", len(contextItems))
	for i, item := range contextItems {
		fmt.Printf("  %d. %s (lines %d-%d, relevance: %.2f)\n", 
			i+1, item.FilePath, item.StartLine, item.EndLine, item.Relevance)
	}

	// Step 9: Demonstrate incremental updates
	fmt.Println("\n🔄 Step 9: Testing incremental updates...")
	
	// Create a new file
	newFile := filepath.Join(tempDir, "new_feature.go")
	newContent := `package main

// NewFeature demonstrates incremental indexing
func NewFeature() {
	fmt.Println("This is a new feature!")
}`

	if err := os.WriteFile(newFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to create new file: %w", err)
	}

	// Refresh index with the new file
	if err := manager.RefreshIndex(ctx, []string{newFile}); err != nil {
		return fmt.Errorf("failed to refresh index: %w", err)
	}

	fmt.Println("✅ Index refreshed with new file")

	fmt.Println("\n✨ Indexing example completed successfully!")
	return nil
}

// createExampleFiles creates sample files for the indexing example
func createExampleFiles(baseDir string) error {
	// Create main.go
	mainGo := `package main

import (
	"fmt"
	"net/http"
	"log"
)

func main() {
	fmt.Println("Starting HTTP server...")
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/api/health", handleHealth)
	
	log.Println("Server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to GoAI HTTP Server!")
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"status\": \"healthy\"}")
}`

	// Create utils/helpers.go
	utilsDir := filepath.Join(baseDir, "utils")
	if err := os.MkdirAll(utilsDir, 0755); err != nil {
		return err
	}
	
	helpersGo := `package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type JSONResponse struct {
	Success bool        json:"success"
	Data    interface{} json:"data,omitempty"
	Error   string      json:"error,omitempty"
}

func WriteJSONResponse(w http.ResponseWriter, response JSONResponse) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(response)
}

func LogError(context string, err error) {
	fmt.Printf("[ERROR] %s: %v\n", context, err)
}`

	// Create README.md
	readmeMd := `# GoAI HTTP Server Example

This is a simple HTTP server built with Go for demonstrating the GoAI indexing system.

## Features

- Basic HTTP routing
- Health check endpoint  
- JSON response utilities
- Error logging

## Usage

Run the server:
go run main.go

## API Endpoints

- GET / - Home page
- GET /api/health - Health check

## Dependencies

This project uses only the Go standard library.`

	// Write files
	files := map[string]string{
		"main.go":           mainGo,
		"utils/helpers.go":  helpersGo,
		"README.md":         readmeMd,
	}

	for filePath, content := range files {
		fullPath := filepath.Join(baseDir, filePath)
		dir := filepath.Dir(fullPath)
		
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", fullPath, err)
		}
	}

	fmt.Printf("Created %d example files\n", len(files))
	return nil
}

// RunIndexingExample runs the indexing system example
func RunIndexingExample() {
	fmt.Println("🚀 GoAI Indexing System Example")
	fmt.Println("================================")
	
	if err := ExampleUsage(); err != nil {
		fmt.Printf("❌ Example failed: %v\n", err)
		os.Exit(1)
	}
}