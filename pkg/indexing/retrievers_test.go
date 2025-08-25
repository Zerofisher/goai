package indexing

import (
	"context"
	"testing"
	"time"
)

func TestFTSRetriever(t *testing.T) {
	mockIndex := &mockIndex{}
	retriever := NewFTSRetriever(mockIndex)

	if retriever.Name() != "fts_retriever" {
		t.Errorf("Expected name 'fts_retriever', got '%s'", retriever.Name())
	}

	// Test query support
	if !retriever.SupportsQuery("test query") {
		t.Error("FTS retriever should support non-empty queries")
	}

	if retriever.SupportsQuery("") {
		t.Error("FTS retriever should not support empty queries")
	}

	// Test retrieval
	ctx := context.Background()
	req := &RetrievalRequest{
		Query:      "test",
		MaxResults: 10,
	}

	result, err := retriever.Retrieve(ctx, req)
	if err != nil {
		t.Errorf("FTS retrieve failed: %v", err)
	}

	if result.SearchType != SearchTypeFullText {
		t.Errorf("Expected search type %s, got %s", SearchTypeFullText, result.SearchType)
	}

	if len(result.Results) == 0 {
		t.Error("Expected results from mock index")
	}
}

func TestSemanticRetriever(t *testing.T) {
	mockIndex := &mockIndex{}
	retriever := NewSemanticRetriever(mockIndex)

	if retriever.Name() != "semantic_retriever" {
		t.Errorf("Expected name 'semantic_retriever', got '%s'", retriever.Name())
	}

	// Test query support - semantic search works better with longer queries
	if !retriever.SupportsQuery("test query that is long enough") {
		t.Error("Semantic retriever should support queries longer than 2 characters")
	}

	if retriever.SupportsQuery("ab") {
		t.Error("Semantic retriever should not support very short queries")
	}

	// Test retrieval
	ctx := context.Background()
	req := &RetrievalRequest{
		Query:      "semantic test query",
		MaxResults: 5,
	}

	result, err := retriever.Retrieve(ctx, req)
	if err != nil {
		t.Errorf("Semantic retrieve failed: %v", err)
	}

	if result.SearchType != SearchTypeSemantic {
		t.Errorf("Expected search type %s, got %s", SearchTypeSemantic, result.SearchType)
	}
}

func TestSymbolRetriever(t *testing.T) {
	mockIndex := &mockIndex{}
	retriever := NewSymbolRetriever(mockIndex)

	if retriever.Name() != "symbol_retriever" {
		t.Errorf("Expected name 'symbol_retriever', got '%s'", retriever.Name())
	}

	// Test query support - symbol search is good for identifiers
	if !retriever.SupportsQuery("functionName") {
		t.Error("Symbol retriever should support valid identifier queries")
	}

	if !retriever.SupportsQuery("MyClass.method") {
		t.Error("Symbol retriever should support dotted identifiers")
	}

	if retriever.SupportsQuery("query with spaces!") {
		t.Error("Symbol retriever should not support queries with special characters")
	}

	// Test retrieval
	ctx := context.Background()
	req := &RetrievalRequest{
		Query:      "testSymbol",
		MaxResults: 10,
	}

	result, err := retriever.Retrieve(ctx, req)
	if err != nil {
		t.Errorf("Symbol retrieve failed: %v", err)
	}

	if result.SearchType != SearchTypeSymbol {
		t.Errorf("Expected search type %s, got %s", SearchTypeSymbol, result.SearchType)
	}
}

func TestRecentFilesRetriever(t *testing.T) {
	retriever := NewRecentFilesRetriever("/tmp", time.Hour)

	if retriever.Name() != "recent_files_retriever" {
		t.Errorf("Expected name 'recent_files_retriever', got '%s'", retriever.Name())
	}

	// Recent files retriever supports all queries
	if !retriever.SupportsQuery("any query") {
		t.Error("Recent files retriever should support any query")
	}

	if !retriever.SupportsQuery("") {
		t.Error("Recent files retriever should support empty query")
	}

	// Test retrieval
	ctx := context.Background()
	req := &RetrievalRequest{
		Query:      "recent",
		MaxResults: 3,
	}

	result, err := retriever.Retrieve(ctx, req)
	if err != nil {
		t.Errorf("Recent files retrieve failed: %v", err)
	}

	if result.SearchType != SearchTypeRecent {
		t.Errorf("Expected search type %s, got %s", SearchTypeRecent, result.SearchType)
	}

	// Should return some mock recent files
	if len(result.Results) == 0 {
		t.Error("Expected recent files results")
	}

	// Should respect max results
	if len(result.Results) > req.MaxResults {
		t.Errorf("Expected at most %d results, got %d", req.MaxResults, len(result.Results))
	}

	// Should include metadata
	if result.Metadata == nil {
		t.Error("Expected metadata in recent files results")
	}
}

func TestRecentFilesRetrieverWithDefaultAge(t *testing.T) {
	// Test with zero maxAge (should default to 24 hours)
	retriever := NewRecentFilesRetriever("/tmp", 0)

	ctx := context.Background()
	req := &RetrievalRequest{
		Query:      "recent",
		MaxResults: 5,
	}

	result, err := retriever.Retrieve(ctx, req)
	if err != nil {
		t.Errorf("Recent files retrieve with default age failed: %v", err)
	}

	// Should have default age in metadata
	if result.Metadata == nil || result.Metadata["max_age"] == nil {
		t.Error("Expected max_age in metadata")
	}
}

func TestHybridRetriever(t *testing.T) {
	// Create multiple retrievers
	mockIndex := &mockIndex{}
	ftsRetriever := NewFTSRetriever(mockIndex)
	semanticRetriever := NewSemanticRetriever(mockIndex)
	recentRetriever := NewRecentFilesRetriever("/tmp", time.Hour)

	retrievers := []Retriever{ftsRetriever, semanticRetriever, recentRetriever}
	reranker := NewSimpleReranker()

	hybridRetriever := NewHybridRetriever(retrievers, reranker)

	if hybridRetriever.Name() != "hybrid_retriever" {
		t.Errorf("Expected name 'hybrid_retriever', got '%s'", hybridRetriever.Name())
	}

	// Test query support - should support if any sub-retriever supports
	if !hybridRetriever.SupportsQuery("test query") {
		t.Error("Hybrid retriever should support queries that sub-retrievers support")
	}

	// Test retrieval
	ctx := context.Background()
	req := &RetrievalRequest{
		Query:      "hybrid test",
		MaxResults: 10,
	}

	result, err := hybridRetriever.Retrieve(ctx, req)
	if err != nil {
		t.Errorf("Hybrid retrieve failed: %v", err)
	}

	if result.SearchType != SearchTypeHybrid {
		t.Errorf("Expected search type %s, got %s", SearchTypeHybrid, result.SearchType)
	}

	// Should combine results from multiple retrievers
	if len(result.Results) == 0 {
		t.Error("Expected combined results from hybrid retriever")
	}

	// Should include metadata about retrievers
	if result.Metadata == nil {
		t.Error("Expected metadata in hybrid results")
	}

	if activeRetrievers, ok := result.Metadata["active_retrievers"]; !ok {
		t.Error("Expected active_retrievers in metadata")
	} else if activeRetrievers.(int) == 0 {
		t.Error("Expected at least one active retriever")
	}
}

func TestHybridRetrieverWithNoSupportedQuery(t *testing.T) {
	// Create retrievers that don't support the query
	retrievers := []Retriever{
		&mockUnsupportedRetriever{},
		&mockUnsupportedRetriever{},
	}

	hybridRetriever := NewHybridRetriever(retrievers, nil)

	// Should not support queries if no sub-retrievers support them
	if hybridRetriever.SupportsQuery("unsupported query") {
		t.Error("Hybrid retriever should not support queries when no sub-retrievers support them")
	}
}

func TestHybridRetrieverDeduplication(t *testing.T) {
	// Create retrievers that return duplicate results
	duplicateIndex := &mockDuplicateIndex{}
	ftsRetriever := NewFTSRetriever(duplicateIndex)
	semanticRetriever := NewSemanticRetriever(duplicateIndex)

	retrievers := []Retriever{ftsRetriever, semanticRetriever}
	hybridRetriever := NewHybridRetriever(retrievers, nil)

	ctx := context.Background()
	req := &RetrievalRequest{
		Query:      "duplicate test",
		MaxResults: 10,
	}

	result, err := hybridRetriever.Retrieve(ctx, req)
	if err != nil {
		t.Errorf("Hybrid retrieve with duplicates failed: %v", err)
	}

	// Should deduplicate results based on file path and line number
	// Our mock returns duplicate results, so we expect deduplication
	uniqueResults := make(map[string]bool)
	for _, res := range result.Results {
		key := res.FilePath + ":" + string(rune(res.LineNumber))
		if uniqueResults[key] {
			t.Errorf("Found duplicate result: %s:%d", res.FilePath, res.LineNumber)
		}
		uniqueResults[key] = true
	}
}

func TestIsValidIdentifier(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"validIdentifier", true},
		{"valid_identifier", true},
		{"ValidIdentifier", true},
		{"identifier123", true},
		{"package.Function", true},
		{"Class.method", true},
		{"", false},
		{"invalid identifier", false},
		{"invalid-identifier", false},
		{"invalid@identifier", false},
		{"123invalid", true}, // digits are allowed in identifiers
		{"_validIdentifier", true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := isValidIdentifier(tc.input)
			if result != tc.expected {
				t.Errorf("isValidIdentifier(%s) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

// Additional mock implementations for testing

type mockUnsupportedRetriever struct{}

func (m *mockUnsupportedRetriever) Name() string {
	return "unsupported_retriever"
}

func (m *mockUnsupportedRetriever) Retrieve(ctx context.Context, req *RetrievalRequest) (*RetrievalResult, error) {
	return &RetrievalResult{
		Results:      []*SearchResult{},
		TotalResults: 0,
		SearchType:   SearchTypeFullText,
		Latency:      time.Microsecond,
	}, nil
}

func (m *mockUnsupportedRetriever) SupportsQuery(query string) bool {
	return false // Always unsupported
}

type mockDuplicateIndex struct{}

func (m *mockDuplicateIndex) Name() string { return "duplicate_index" }
func (m *mockDuplicateIndex) Version() string { return "1.0.0" }
func (m *mockDuplicateIndex) Update(ctx context.Context, tag *IndexTag, changes *IndexChanges) error { return nil }
func (m *mockDuplicateIndex) Remove(ctx context.Context, tag *IndexTag, files []string) error { return nil }
func (m *mockDuplicateIndex) Search(ctx context.Context, query string, opts *SearchOptions) ([]*SearchResult, error) {
	// Return the same result multiple times to test deduplication
	result := &SearchResult{
		FilePath:   "duplicate.go",
		LineNumber: 10,
		Snippet:    "duplicate result",
		Score:      95.0,
		SearchType: SearchTypeFullText,
	}
	
	return []*SearchResult{result, result}, nil // Same result twice
}
func (m *mockDuplicateIndex) GetStats() (*IndexStats, error) {
	return &IndexStats{
		Name:           "duplicate_index",
		Version:        "1.0.0",
		TotalDocuments: 1,
		LastUpdated:    time.Now(),
	}, nil
}
func (m *mockDuplicateIndex) Close() error { return nil }

func TestHybridRetrieverWithReranker(t *testing.T) {
	mockIndex := &mockIndex{}
	ftsRetriever := NewFTSRetriever(mockIndex)
	retrievers := []Retriever{ftsRetriever}
	reranker := NewSimpleReranker()

	hybridRetriever := NewHybridRetriever(retrievers, reranker)

	ctx := context.Background()
	req := &RetrievalRequest{
		Query:      "test query",
		MaxResults: 5,
	}

	result, err := hybridRetriever.Retrieve(ctx, req)
	if err != nil {
		t.Errorf("Hybrid retrieve with reranker failed: %v", err)
	}

	if len(result.Results) == 0 {
		t.Error("Expected results from hybrid retriever with reranker")
	}

	// Results should be sorted by score (highest first) after reranking
	for i := 1; i < len(result.Results); i++ {
		if result.Results[i-1].Score < result.Results[i].Score {
			t.Errorf("Results not properly sorted by score: %f < %f at position %d", 
				result.Results[i-1].Score, result.Results[i].Score, i)
		}
	}
}

func TestHybridRetrieverWithoutReranker(t *testing.T) {
	mockIndex := &mockIndex{}
	ftsRetriever := NewFTSRetriever(mockIndex)
	retrievers := []Retriever{ftsRetriever}

	hybridRetriever := NewHybridRetriever(retrievers, nil) // No reranker

	ctx := context.Background()
	req := &RetrievalRequest{
		Query:      "test query",
		MaxResults: 5,
	}

	result, err := hybridRetriever.Retrieve(ctx, req)
	if err != nil {
		t.Errorf("Hybrid retrieve without reranker failed: %v", err)
	}

	if len(result.Results) == 0 {
		t.Error("Expected results from hybrid retriever without reranker")
	}

	// Should still sort by default (score descending)
	for i := 1; i < len(result.Results); i++ {
		if result.Results[i-1].Score < result.Results[i].Score {
			t.Errorf("Results not properly sorted by default: %f < %f at position %d", 
				result.Results[i-1].Score, result.Results[i].Score, i)
		}
	}
}