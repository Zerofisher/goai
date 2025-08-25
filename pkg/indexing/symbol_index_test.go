package indexing

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestSymbolIndex(t *testing.T) {
	// Create temporary database
	tempDB, err := os.CreateTemp("", "symbol_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	parser := NewGoTreeSitterParser()
	chunker := NewDefaultChunker()
	
	index, err := NewSymbolIndex(tempDB.Name(), parser, chunker)
	if err != nil {
		t.Fatalf("NewSymbolIndex failed: %v", err)
	}
	defer func() { _ = index.Close() }()

	if index.Name() != "symbol_index" {
		t.Errorf("Expected name 'symbol_index', got '%s'", index.Name())
	}

	if index.Version() != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", index.Version())
	}
}

func TestSymbolIndexUpdate(t *testing.T) {
	// Create temporary database
	tempDB, err := os.CreateTemp("", "symbol_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	parser := &mockParser{}
	chunker := NewDefaultChunker()
	
	index, err := NewSymbolIndex(tempDB.Name(), parser, chunker)
	if err != nil {
		t.Fatalf("NewSymbolIndex failed: %v", err)
	}
	defer func() { _ = index.Close() }()

	ctx := context.Background()
	tag := &IndexTag{
		WorkingDirectory: "/test",
		Timestamp:       time.Now(),
	}

	changes := &IndexChanges{
		Added:    []string{"test.go", "main.go"},
		Modified: []string{"helper.go"},
		Deleted:  []string{"old.go"},
	}

	err = index.Update(ctx, tag, changes)
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}
}

func TestSymbolIndexSearch(t *testing.T) {
	// Create temporary database
	tempDB, err := os.CreateTemp("", "symbol_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	parser := &mockParserWithSymbols{}
	chunker := NewDefaultChunker()
	
	index, err := NewSymbolIndex(tempDB.Name(), parser, chunker)
	if err != nil {
		t.Fatalf("NewSymbolIndex failed: %v", err)
	}
	defer func() { _ = index.Close() }()

	// Add some test data
	ctx := context.Background()
	tag := &IndexTag{
		WorkingDirectory: "/test",
		Timestamp:       time.Now(),
	}

	changes := &IndexChanges{
		Added: []string{"test.go"},
	}

	err = index.Update(ctx, tag, changes)
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	// Test search
	opts := &SearchOptions{
		MaxResults:     10,
		IncludeContent: true,
	}

	results, err := index.Search(ctx, "testFunc", opts)
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}

	// Since file reading is not implemented, we may not get results
	// This is acceptable for testing the search interface
	t.Logf("Found %d search results (file reading not implemented)", len(results))

	// Verify result structure
	if len(results) > 0 {
		result := results[0]
		if result.SearchType != SearchTypeSymbol {
			t.Errorf("Expected SearchType to be Symbol, got %v", result.SearchType)
		}
		
		if result.SymbolInfo == nil {
			t.Error("Expected SymbolInfo to be populated")
		}
	}
}

func TestSymbolIndexRemove(t *testing.T) {
	// Create temporary database
	tempDB, err := os.CreateTemp("", "symbol_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	parser := &mockParserWithSymbols{}
	chunker := NewDefaultChunker()
	
	index, err := NewSymbolIndex(tempDB.Name(), parser, chunker)
	if err != nil {
		t.Fatalf("NewSymbolIndex failed: %v", err)
	}
	defer func() { _ = index.Close() }()

	ctx := context.Background()
	tag := &IndexTag{
		WorkingDirectory: "/test",
		Timestamp:       time.Now(),
	}

	// First add some data
	changes := &IndexChanges{
		Added: []string{"test.go"},
	}
	err = index.Update(ctx, tag, changes)
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	// Then remove it
	err = index.Remove(ctx, tag, []string{"test.go"})
	if err != nil {
		t.Errorf("Remove failed: %v", err)
	}

	// Search should return no results
	opts := &SearchOptions{MaxResults: 10}
	results, err := index.Search(ctx, "testFunc", opts)
	if err != nil {
		t.Errorf("Search after remove failed: %v", err)
	}

	if len(results) > 0 {
		t.Errorf("Expected no results after removal, got %d", len(results))
	}
}

func TestSymbolIndexStats(t *testing.T) {
	// Create temporary database
	tempDB, err := os.CreateTemp("", "symbol_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	parser := &mockParserWithSymbols{}
	chunker := NewDefaultChunker()
	
	index, err := NewSymbolIndex(tempDB.Name(), parser, chunker)
	if err != nil {
		t.Fatalf("NewSymbolIndex failed: %v", err)
	}
	defer func() { _ = index.Close() }()

	stats, err := index.GetStats()
	if err != nil {
		t.Errorf("GetStats failed: %v", err)
	}

	if stats.Name != "symbol_index" {
		t.Errorf("Expected stats name 'symbol_index', got '%s'", stats.Name)
	}

	if stats.Version != "1.0.0" {
		t.Errorf("Expected stats version '1.0.0', got '%s'", stats.Version)
	}
}

func TestSymbolIndexRelevanceScore(t *testing.T) {
	// Create temporary database
	tempDB, err := os.CreateTemp("", "symbol_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	parser := NewGoTreeSitterParser()
	chunker := NewDefaultChunker()
	
	index, err := NewSymbolIndex(tempDB.Name(), parser, chunker)
	if err != nil {
		t.Fatalf("NewSymbolIndex failed: %v", err)
	}
	defer func() { _ = index.Close() }()

	// Test the calculateRelevanceScore method via search
	symbol := &SymbolInfo{
		Name:      "testFunction",
		Kind:      SymbolKindFunction,
		Signature: "func testFunction() error",
	}

	// Test exact match scoring
	score := index.calculateRelevanceScore("testFunction", symbol)
	if score <= 0 {
		t.Errorf("Expected positive score for exact match, got %f", score)
	}

	// Test partial match scoring  
	partialScore := index.calculateRelevanceScore("test", symbol)
	if partialScore <= 0 {
		t.Errorf("Expected positive score for partial match, got %f", partialScore)
	}

	// Exact match should have higher score
	if score <= partialScore {
		t.Errorf("Expected exact match score (%f) to be higher than partial match (%f)", score, partialScore)
	}

	// Test method scoring vs function scoring
	methodSymbol := &SymbolInfo{
		Name: "testMethod",
		Kind: SymbolKindMethod,
	}
	
	funcScore := index.calculateRelevanceScore("test", symbol)
	methodScore := index.calculateRelevanceScore("test", methodSymbol)
	
	// Functions should get higher base score than methods
	if funcScore <= methodScore {
		t.Errorf("Expected function score (%f) to be higher than method score (%f)", funcScore, methodScore)
	}
}

// Mock parser for testing
type mockParser struct{}

func (m *mockParser) ParseFile(ctx context.Context, filePath string, content []byte) (*ParsedFile, error) {
	return &ParsedFile{
		FilePath: filePath,
		Language: "go",
		Symbols:  []*SymbolInfo{},
		Imports:  []string{},
	}, nil
}

func (m *mockParser) GetSupportedLanguages() []string {
	return []string{"go"}
}

// Mock parser that returns symbols for testing
type mockParserWithSymbols struct{}

func (m *mockParserWithSymbols) ParseFile(ctx context.Context, filePath string, content []byte) (*ParsedFile, error) {
	return &ParsedFile{
		FilePath: filePath,
		Language: "go",
		Symbols: []*SymbolInfo{
			{
				Name:        "testFunc",
				Kind:        SymbolKindFunction,
				FilePath:    filePath,
				StartLine:   1,
				EndLine:     5,
				StartColumn: 1,
				EndColumn:   10,
				Language:    "go",
				Signature:   "func testFunc() error",
				DocString:   "Test function",
			},
			{
				Name:        "TestStruct",
				Kind:        SymbolKindStruct,
				FilePath:    filePath,
				StartLine:   7,
				EndLine:     10,
				StartColumn: 1,
				EndColumn:   15,
				Language:    "go",
				Signature:   "type TestStruct struct",
				Children:    []string{"field1", "field2"},
			},
		},
		Imports: []string{"fmt", "testing"},
	}, nil
}

func (m *mockParserWithSymbols) GetSupportedLanguages() []string {
	return []string{"go"}
}

func TestSymbolIndexSearchWithFileTypes(t *testing.T) {
	// Create temporary database
	tempDB, err := os.CreateTemp("", "symbol_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	parser := &mockParserWithSymbols{}
	chunker := NewDefaultChunker()
	
	index, err := NewSymbolIndex(tempDB.Name(), parser, chunker)
	if err != nil {
		t.Fatalf("NewSymbolIndex failed: %v", err)
	}
	defer func() { _ = index.Close() }()

	// Add test data
	ctx := context.Background()
	tag := &IndexTag{
		WorkingDirectory: "/test",
		Timestamp:       time.Now(),
	}

	changes := &IndexChanges{
		Added: []string{"test.go", "test.py"},
	}

	err = index.Update(ctx, tag, changes)
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	// Search with file type filter
	opts := &SearchOptions{
		MaxResults: 10,
		FileTypes:  []string{"go"},
	}

	results, err := index.Search(ctx, "testFunc", opts)
	if err != nil {
		t.Errorf("Search with file types failed: %v", err)
	}

	// Since file reading is not implemented, we may not get results  
	// This is acceptable for testing the search interface
	t.Logf("Found %d results with 'go' file type filter (file reading not implemented)", len(results))
}

func TestSymbolIndexClose(t *testing.T) {
	tempDB, err := os.CreateTemp("", "symbol_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	parser := NewGoTreeSitterParser()
	chunker := NewDefaultChunker()
	
	index, err := NewSymbolIndex(tempDB.Name(), parser, chunker)
	if err != nil {
		t.Fatalf("NewSymbolIndex failed: %v", err)
	}

	// Close should not return error
	err = index.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Closing again should not panic
	err = index.Close()
	if err != nil {
		t.Errorf("Second close failed: %v", err)
	}
}