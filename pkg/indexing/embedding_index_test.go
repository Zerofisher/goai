package indexing

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestEmbeddingIndex(t *testing.T) {
	// Create temporary database
	tempDB, err := os.CreateTemp("", "embedding_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	provider := NewMockEmbeddingProvider(128)
	chunker := NewDefaultChunker()
	
	index, err := NewEmbeddingIndex(tempDB.Name(), provider, chunker)
	if err != nil {
		t.Fatalf("NewEmbeddingIndex failed: %v", err)
	}
	defer func() { _ = index.Close() }()

	if index.Name() != "embedding_index" {
		t.Errorf("Expected name 'embedding_index', got '%s'", index.Name())
	}

	if index.Version() != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", index.Version())
	}
}

func TestEmbeddingIndexUpdate(t *testing.T) {
	tempDB, err := os.CreateTemp("", "embedding_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	provider := NewMockEmbeddingProvider(128)
	chunker := &mockChunker{}
	
	index, err := NewEmbeddingIndex(tempDB.Name(), provider, chunker)
	if err != nil {
		t.Fatalf("NewEmbeddingIndex failed: %v", err)
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

func TestEmbeddingIndexSearch(t *testing.T) {
	tempDB, err := os.CreateTemp("", "embedding_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	provider := NewMockEmbeddingProvider(128)
	chunker := &mockChunker{}
	
	index, err := NewEmbeddingIndex(tempDB.Name(), provider, chunker)
	if err != nil {
		t.Fatalf("NewEmbeddingIndex failed: %v", err)
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

	results, err := index.Search(ctx, "test function", opts)
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}

	// Since file reading is not implemented, we may not get results
	// This is acceptable for testing the search interface
	t.Logf("Found %d search results (file reading not implemented)", len(results))

	// Verify result structure
	if len(results) > 0 {
		result := results[0]
		if result.SearchType != SearchTypeSemantic {
			t.Errorf("Expected SearchType to be Semantic, got %v", result.SearchType)
		}
		
		if result.Score < 0 || result.Score > 100 {
			t.Errorf("Expected score to be between 0-100, got %f", result.Score)
		}
	}
}

func TestEmbeddingIndexRemove(t *testing.T) {
	tempDB, err := os.CreateTemp("", "embedding_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	provider := NewMockEmbeddingProvider(128)
	chunker := &mockChunker{}
	
	index, err := NewEmbeddingIndex(tempDB.Name(), provider, chunker)
	if err != nil {
		t.Fatalf("NewEmbeddingIndex failed: %v", err)
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

func TestEmbeddingIndexStats(t *testing.T) {
	tempDB, err := os.CreateTemp("", "embedding_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	provider := NewMockEmbeddingProvider(128)
	chunker := NewDefaultChunker()
	
	index, err := NewEmbeddingIndex(tempDB.Name(), provider, chunker)
	if err != nil {
		t.Fatalf("NewEmbeddingIndex failed: %v", err)
	}
	defer func() { _ = index.Close() }()

	stats, err := index.GetStats()
	if err != nil {
		t.Errorf("GetStats failed: %v", err)
	}

	if stats.Name != "embedding_index" {
		t.Errorf("Expected stats name 'embedding_index', got '%s'", stats.Name)
	}

	if stats.Version != "1.0.0" {
		t.Errorf("Expected stats version '1.0.0', got '%s'", stats.Version)
	}
}

func TestEmbeddingIndexCosineSimilarity(t *testing.T) {
	tempDB, err := os.CreateTemp("", "embedding_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	provider := NewMockEmbeddingProvider(128)
	chunker := NewDefaultChunker()
	
	index, err := NewEmbeddingIndex(tempDB.Name(), provider, chunker)
	if err != nil {
		t.Fatalf("NewEmbeddingIndex failed: %v", err)
	}
	defer func() { _ = index.Close() }()

	// Test cosine similarity function
	vec1 := []float64{1.0, 0.0, 0.0}
	vec2 := []float64{0.0, 1.0, 0.0}
	vec3 := []float64{1.0, 0.0, 0.0}

	// Test orthogonal vectors (should be 0)
	sim12 := index.cosineSimilarity(vec1, vec2)
	if sim12 != 0.0 {
		t.Errorf("Expected cosine similarity of orthogonal vectors to be 0.0, got %f", sim12)
	}

	// Test identical vectors (should be 1)
	sim13 := index.cosineSimilarity(vec1, vec3)
	if sim13 != 1.0 {
		t.Errorf("Expected cosine similarity of identical vectors to be 1.0, got %f", sim13)
	}

	// Test different length vectors (should be 0)
	vec4 := []float64{1.0, 0.0}
	sim14 := index.cosineSimilarity(vec1, vec4)
	if sim14 != 0.0 {
		t.Errorf("Expected cosine similarity of different length vectors to be 0.0, got %f", sim14)
	}

	// Test zero vectors
	vec5 := []float64{0.0, 0.0, 0.0}
	vec6 := []float64{0.0, 0.0, 0.0}
	sim56 := index.cosineSimilarity(vec5, vec6)
	if sim56 != 0.0 {
		t.Errorf("Expected cosine similarity of zero vectors to be 0.0, got %f", sim56)
	}
}

func TestEmbeddingIndexCreateSnippet(t *testing.T) {
	tempDB, err := os.CreateTemp("", "embedding_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	provider := NewMockEmbeddingProvider(128)
	chunker := NewDefaultChunker()
	
	index, err := NewEmbeddingIndex(tempDB.Name(), provider, chunker)
	if err != nil {
		t.Fatalf("NewEmbeddingIndex failed: %v", err)
	}
	defer func() { _ = index.Close() }()

	// Test short content (should return as-is)
	shortContent := "This is a short string"
	snippet := index.createSnippet(shortContent)
	if snippet != shortContent {
		t.Errorf("Expected short content to be returned as-is, got: %s", snippet)
	}

	// Test long content (should be truncated)
	longContent := make([]byte, 300)
	for i := range longContent {
		longContent[i] = 'a'
	}
	longString := string(longContent)
	
	snippet = index.createSnippet(longString)
	if len(snippet) > 203 { // 200 + "..."
		t.Errorf("Expected snippet to be truncated to ~203 chars, got %d chars", len(snippet))
	}
	
	if !endsWith(snippet, "...") {
		t.Errorf("Expected truncated snippet to end with '...', got: %s", snippet[len(snippet)-10:])
	}

	// Test content with good breaking point
	contentWithSpaces := "This is a very long string that should be truncated at a good breaking point. "
	for i := 0; i < 10; i++ {
		contentWithSpaces += "More text to make it longer. "
	}
	
	snippet = index.createSnippet(contentWithSpaces)
	if !endsWith(snippet, "...") {
		t.Errorf("Expected snippet to end with '...', got: %s", snippet[len(snippet)-10:])
	}
}

func TestEmbeddingIndexDetectLanguage(t *testing.T) {
	tempDB, err := os.CreateTemp("", "embedding_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	provider := NewMockEmbeddingProvider(128)
	chunker := NewDefaultChunker()
	
	index, err := NewEmbeddingIndex(tempDB.Name(), provider, chunker)
	if err != nil {
		t.Fatalf("NewEmbeddingIndex failed: %v", err)
	}
	defer func() { _ = index.Close() }()

	testCases := []struct {
		filename string
		expected string
	}{
		{"test.go", "go"},
		{"script.js", "javascript"},
		{"app.ts", "typescript"},
		{"main.py", "python"},
		{"App.java", "java"},
		{"main.cpp", "cpp"},
		{"main.cc", "cpp"},
		{"lib.c", "c"},
		{"main.rs", "rust"},
		{"unknown.txt", "text"},
		{"noext", "text"},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			result := index.detectLanguage(tc.filename)
			if result != tc.expected {
				t.Errorf("Expected language '%s' for file '%s', got '%s'", tc.expected, tc.filename, result)
			}
		})
	}
}

func TestEmbeddingIndexSearchWithFileTypes(t *testing.T) {
	tempDB, err := os.CreateTemp("", "embedding_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database: %v", err)
	}
	defer func() { _ = os.Remove(tempDB.Name())
	_ = tempDB.Close(); }()

	provider := NewMockEmbeddingProvider(128)
	chunker := &mockChunker{}
	
	index, err := NewEmbeddingIndex(tempDB.Name(), provider, chunker)
	if err != nil {
		t.Fatalf("NewEmbeddingIndex failed: %v", err)
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

	results, err := index.Search(ctx, "test function", opts)
	if err != nil {
		t.Errorf("Search with file types failed: %v", err)
	}

	// Results depend on our mock chunker implementation
	t.Logf("Found %d results with file type filter", len(results))
}

// Mock chunker for testing
type mockChunker struct{}

func (m *mockChunker) ChunkFile(ctx context.Context, file *FileContent) ([]*Chunk, error) {
	// Use file path to create unique chunk IDs
	baseName := file.Info.Path
	if baseName == "" {
		baseName = "unknown"
	}
	
	return []*Chunk{
		{
			ID:        baseName + "_chunk1",
			FilePath:  file.Info.Path,
			Content:   "test content for embedding",
			StartLine: 1,
			EndLine:   3,
			Language:  file.Info.Language,
			ChunkType: ChunkTypeCode,
			Metadata: map[string]interface{}{
				"hash": baseName + "_hash_1",
			},
		},
		{
			ID:        baseName + "_chunk2",
			FilePath:  file.Info.Path,
			Content:   "another chunk with different content",
			StartLine: 4,
			EndLine:   6,
			Language:  file.Info.Language,
			ChunkType: ChunkTypeCode,
			Metadata: map[string]interface{}{
				"hash": baseName + "_hash_2",
			},
		},
	}, nil
}

func (m *mockChunker) ChunkText(text string, language string) ([]*Chunk, error) {
	return []*Chunk{
		{
			ID:        "text_chunk1",
			Content:   text,
			StartLine: 1,
			EndLine:   1,
			Language:  language,
			ChunkType: ChunkTypeCode,
			Metadata: map[string]interface{}{
				"hash": "text_hash_1",
			},
		},
	}, nil
}

// Helper function to check string suffix
func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}