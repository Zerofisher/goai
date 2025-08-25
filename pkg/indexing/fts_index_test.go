package indexing

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestNewFTSIndex(t *testing.T) {
	tempDB, err := os.CreateTemp("", "fts_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	chunker := NewDefaultChunker()
	index, err := NewFTSIndex(tempDB.Name(), chunker)
	if err != nil {
		t.Fatalf("NewFTSIndex failed: %v", err)
	}
	defer func() { _ = index.Close() }()

	if index.Name() != "fts_index" {
		t.Errorf("Expected name 'fts_index', got '%s'", index.Name())
	}

	if index.Version() != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", index.Version())
	}
}

func TestFTSIndex_Update(t *testing.T) {
	tempDB, err := os.CreateTemp("", "fts_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	chunker := &mockChunker{}
	index, err := NewFTSIndex(tempDB.Name(), chunker)
	if err != nil {
		t.Fatalf("NewFTSIndex failed: %v", err)
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

func TestFTSIndex_Remove(t *testing.T) {
	tempDB, err := os.CreateTemp("", "fts_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	chunker := &mockChunker{}
	index, err := NewFTSIndex(tempDB.Name(), chunker)
	if err != nil {
		t.Fatalf("NewFTSIndex failed: %v", err)
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
}

func TestFTSIndex_RemoveAll(t *testing.T) {
	tempDB, err := os.CreateTemp("", "fts_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	chunker := &mockChunker{}
	index, err := NewFTSIndex(tempDB.Name(), chunker)
	if err != nil {
		t.Fatalf("NewFTSIndex failed: %v", err)
	}
	defer func() { _ = index.Close() }()

	ctx := context.Background()
	tag := &IndexTag{
		WorkingDirectory: "/test",
		Timestamp:       time.Now(),
	}

	// Add some test data
	changes := &IndexChanges{
		Added: []string{"test.go", "main.go"},
	}
	err = index.Update(ctx, tag, changes)
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	// Remove all using wildcard
	err = index.Remove(ctx, tag, []string{"*"})
	if err != nil {
		t.Errorf("Remove all failed: %v", err)
	}
}

func TestFTSIndex_Search(t *testing.T) {
	tempDB, err := os.CreateTemp("", "fts_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	chunker := &mockSearchChunker{}
	index, err := NewFTSIndex(tempDB.Name(), chunker)
	if err != nil {
		t.Fatalf("NewFTSIndex failed: %v", err)
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

	results, err := index.Search(ctx, "function", opts)
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}

	// Results depend on our mock chunker implementation
	t.Logf("Found %d results for 'function'", len(results))

	// Verify result structure if we have results
	if len(results) > 0 {
		result := results[0]
		if result.SearchType != SearchTypeFullText {
			t.Errorf("Expected SearchType to be FullText, got %v", result.SearchType)
		}

		if result.Score < 0 || result.Score > 1 {
			t.Errorf("Expected score to be between 0-1, got %f", result.Score)
		}
	}
}

func TestFTSIndex_SearchWithFileTypes(t *testing.T) {
	tempDB, err := os.CreateTemp("", "fts_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	chunker := &mockSearchChunker{}
	index, err := NewFTSIndex(tempDB.Name(), chunker)
	if err != nil {
		t.Fatalf("NewFTSIndex failed: %v", err)
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

	results, err := index.Search(ctx, "function", opts)
	if err != nil {
		t.Errorf("Search with file types failed: %v", err)
	}

	t.Logf("Found %d results with file type filter", len(results))
}

func TestFTSIndex_GetStats(t *testing.T) {
	tempDB, err := os.CreateTemp("", "fts_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	chunker := NewDefaultChunker()
	index, err := NewFTSIndex(tempDB.Name(), chunker)
	if err != nil {
		t.Fatalf("NewFTSIndex failed: %v", err)
	}
	defer func() { _ = index.Close() }()

	stats, err := index.GetStats()
	if err != nil {
		t.Errorf("GetStats failed: %v", err)
	}

	if stats.Name != "fts_index" {
		t.Errorf("Expected stats name 'fts_index', got '%s'", stats.Name)
	}

	if stats.Version != "1.0.0" {
		t.Errorf("Expected stats version '1.0.0', got '%s'", stats.Version)
	}

	if stats.TotalDocuments < 0 {
		t.Errorf("Expected non-negative total documents, got %d", stats.TotalDocuments)
	}
}

func TestFTSIndex_PrepareFTSQuery(t *testing.T) {
	tempDB, err := os.CreateTemp("", "fts_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	chunker := NewDefaultChunker()
	index, err := NewFTSIndex(tempDB.Name(), chunker)
	if err != nil {
		t.Fatalf("NewFTSIndex failed: %v", err)
	}
	defer func() { _ = index.Close() }()

	testCases := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"hello world", "hello AND world"},
		{"multiple word query", "multiple AND word AND query"},
		{"", ""},
		{"\"quoted string\"", "\"\"quoted AND string\"\""},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := index.prepareFTSQuery(tc.input)
			if result != tc.expected {
				t.Errorf("Expected prepareFTSQuery(%s) = %s, got %s", tc.input, tc.expected, result)
			}
		})
	}
}

func TestFTSIndex_NormalizeRank(t *testing.T) {
	tempDB, err := os.CreateTemp("", "fts_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	chunker := NewDefaultChunker()
	index, err := NewFTSIndex(tempDB.Name(), chunker)
	if err != nil {
		t.Fatalf("NewFTSIndex failed: %v", err)
	}
	defer func() { _ = index.Close() }()

	testCases := []struct {
		rank     float64
		expected bool // whether result should be > 0
	}{
		{-1.0, true},  // Good rank should give positive score
		{-0.1, true},  // Better rank should give positive score
		{0.0, false},  // Zero rank should give zero score
		{1.0, false},  // Positive rank should give zero score
	}

	for _, tc := range testCases {
		score := index.normalizeRank(tc.rank)
		if tc.expected && score <= 0 {
			t.Errorf("Expected positive score for rank %f, got %f", tc.rank, score)
		}
		if !tc.expected && score > 0 {
			t.Errorf("Expected zero score for rank %f, got %f", tc.rank, score)
		}

		// Score should always be between 0 and 1
		if score < 0 || score > 1 {
			t.Errorf("Score should be between 0 and 1, got %f for rank %f", score, tc.rank)
		}
	}
}

func TestFTSIndex_DetectLanguage(t *testing.T) {
	tempDB, err := os.CreateTemp("", "fts_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	chunker := NewDefaultChunker()
	index, err := NewFTSIndex(tempDB.Name(), chunker)
	if err != nil {
		t.Fatalf("NewFTSIndex failed: %v", err)
	}
	defer func() { _ = index.Close() }()

	testCases := []struct {
		filename string
		expected string
	}{
		{"main.go", "go"},
		{"script.js", "javascript"},
		{"app.ts", "typescript"},
		{"main.py", "python"},
		{"App.java", "java"},
		{"main.cpp", "cpp"},
		{"lib.c", "c"},
		{"main.rs", "rust"},
		{"README.md", "markdown"},
		{"data.txt", "text"},
		{"unknown.xyz", "text"},
		{"", "text"},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			result := index.detectLanguage(tc.filename)
			if result != tc.expected {
				t.Errorf("Expected detectLanguage(%s) = %s, got %s", tc.filename, tc.expected, result)
			}
		})
	}
}

func TestFTSIndex_Close(t *testing.T) {
	tempDB, err := os.CreateTemp("", "fts_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	chunker := NewDefaultChunker()
	index, err := NewFTSIndex(tempDB.Name(), chunker)
	if err != nil {
		t.Fatalf("NewFTSIndex failed: %v", err)
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

func TestFTSIndex_SearchWithOffset(t *testing.T) {
	tempDB, err := os.CreateTemp("", "fts_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	chunker := &mockSearchChunker{}
	index, err := NewFTSIndex(tempDB.Name(), chunker)
	if err != nil {
		t.Fatalf("NewFTSIndex failed: %v", err)
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

	// Test search with offset
	opts := &SearchOptions{
		MaxResults: 5,
		Offset:     10,
	}

	results, err := index.Search(ctx, "function", opts)
	if err != nil {
		t.Errorf("Search with offset failed: %v", err)
	}

	// Should execute without error (results depend on data)
	t.Logf("Found %d results with offset", len(results))
}

// Mock chunker for search testing that returns content suitable for FTS
type mockSearchChunker struct{}

func (m *mockSearchChunker) ChunkFile(ctx context.Context, file *FileContent) ([]*Chunk, error) {
	// Use file path to create unique chunk IDs
	baseName := file.Info.Path
	if baseName == "" {
		baseName = "unknown"
	}
	
	// Return chunks with searchable content
	return []*Chunk{
		{
			ID:        baseName + "_search_chunk1",
			FilePath:  file.Info.Path,
			Content:   "function testSearch() { return 'searchable content'; }",
			StartLine: 1,
			EndLine:   3,
			Language:  file.Info.Language,
			ChunkType: ChunkTypeCode,
			Metadata: map[string]interface{}{
				"hash": baseName + "_search_hash1",
			},
		},
		{
			ID:        baseName + "_search_chunk2", 
			FilePath:  file.Info.Path,
			Content:   "// Another function for testing search functionality",
			StartLine: 4,
			EndLine:   6,
			Language:  file.Info.Language,
			ChunkType: ChunkTypeComment,
			Metadata: map[string]interface{}{
				"hash": baseName + "_search_hash2",
			},
		},
	}, nil
}

func (m *mockSearchChunker) ChunkText(text string, language string) ([]*Chunk, error) {
	return []*Chunk{
		{
			ID:        "text_search_chunk1",
			Content:   text,
			StartLine: 1,
			EndLine:   1,
			Language:  language,
			ChunkType: ChunkTypeCode,
			Metadata: map[string]interface{}{
				"hash": "text_search_hash1",
			},
		},
	}, nil
}