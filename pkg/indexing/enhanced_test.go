package indexing

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestEnhancedIndexManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "enhanced_index_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create enhanced index manager
	manager, err := NewEnhancedIndexManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create enhanced index manager: %v", err)
	}
	defer func() { _ = manager.Close() }()

	// Test available search types
	searchTypes := manager.GetAvailableSearchTypes()
	if len(searchTypes) == 0 {
		t.Error("No search types available")
	}

	expectedTypes := []SearchType{
		SearchTypeFullText,
		SearchTypeSymbol,
		SearchTypeSemantic,
		SearchTypeRecent,
		SearchTypeHybrid,
	}

	for _, expected := range expectedTypes {
		found := false
		for _, actual := range searchTypes {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected search type %s not found in available types", expected)
		}
	}

	// Test getting stats
	stats, err := manager.GetIndexStats()
	if err != nil {
		t.Errorf("Failed to get index stats: %v", err)
	}
	if len(stats) == 0 {
		t.Error("No index stats available")
	}

	t.Logf("Enhanced index manager test completed successfully with %d search types and %d indexes", 
		len(searchTypes), len(stats))
}

func TestEmbeddingProvider(t *testing.T) {
	// Test mock embedding provider
	provider := NewMockEmbeddingProvider(128)
	
	ctx := context.Background()
	
	// Test single embedding
	embedding, err := provider.GenerateEmbedding(ctx, "test text")
	if err != nil {
		t.Fatalf("Failed to generate embedding: %v", err)
	}
	
	if len(embedding) != 128 {
		t.Errorf("Expected embedding dimension 128, got %d", len(embedding))
	}
	
	// Test batch embeddings
	texts := []string{"text1", "text2", "text3"}
	embeddings, err := provider.GenerateBatchEmbeddings(ctx, texts)
	if err != nil {
		t.Fatalf("Failed to generate batch embeddings: %v", err)
	}
	
	if len(embeddings) != 3 {
		t.Errorf("Expected 3 embeddings, got %d", len(embeddings))
	}
	
	// Test consistency - same text should produce same embedding
	embedding2, err := provider.GenerateEmbedding(ctx, "test text")
	if err != nil {
		t.Fatalf("Failed to generate second embedding: %v", err)
	}
	
	for i, v := range embedding {
		if v != embedding2[i] {
			t.Error("Embedding should be consistent for same input")
			break
		}
	}
	
	t.Log("Embedding provider test completed successfully")
}

func TestRetrievers(t *testing.T) {
	// Test FTS retriever
	mockIndex := &mockIndex{}
	ftsRetriever := NewFTSRetriever(mockIndex)
	
	if !ftsRetriever.SupportsQuery("test query") {
		t.Error("FTS retriever should support non-empty queries")
	}
	
	if ftsRetriever.SupportsQuery("") {
		t.Error("FTS retriever should not support empty queries")
	}
	
	// Test recent files retriever
	recentRetriever := NewRecentFilesRetriever("/tmp", time.Hour)
	
	ctx := context.Background()
	req := &RetrievalRequest{
		Query:      "test",
		MaxResults: 10,
	}
	
	result, err := recentRetriever.Retrieve(ctx, req)
	if err != nil {
		t.Errorf("Recent files retriever failed: %v", err)
	}
	
	if result.SearchType != SearchTypeRecent {
		t.Errorf("Expected search type %s, got %s", SearchTypeRecent, result.SearchType)
	}
	
	t.Log("Retrievers test completed successfully")
}

// mockIndex is a simple mock implementation for testing
type mockIndex struct{}

func (m *mockIndex) Name() string { return "mock_index" }
func (m *mockIndex) Version() string { return "1.0.0" }
func (m *mockIndex) Update(ctx context.Context, tag *IndexTag, changes *IndexChanges) error { return nil }
func (m *mockIndex) Remove(ctx context.Context, tag *IndexTag, files []string) error { return nil }
func (m *mockIndex) Search(ctx context.Context, query string, opts *SearchOptions) ([]*SearchResult, error) {
	return []*SearchResult{
		{
			FilePath:   "test.go",
			Snippet:    "mock result",
			Score:      95.0,
			SearchType: SearchTypeFullText,
		},
	}, nil
}
func (m *mockIndex) GetStats() (*IndexStats, error) {
	return &IndexStats{
		Name:           "mock_index",
		Version:        "1.0.0",
		TotalDocuments: 1,
		LastUpdated:    time.Now(),
	}, nil
}
func (m *mockIndex) Close() error { return nil }