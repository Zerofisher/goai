package indexing

import (
	"context"
	"testing"
)

func TestSimpleReranker(t *testing.T) {
	reranker := NewSimpleReranker()
	
	if reranker == nil {
		t.Fatal("NewSimpleReranker returned nil")
	}
	
	// Check default weights
	if reranker.exactMatchWeight != 10.0 {
		t.Errorf("Expected exactMatchWeight to be 10.0, got %f", reranker.exactMatchWeight)
	}
	
	if reranker.prefixMatchWeight != 5.0 {
		t.Errorf("Expected prefixMatchWeight to be 5.0, got %f", reranker.prefixMatchWeight)
	}
	
	if reranker.contentMatchWeight != 2.0 {
		t.Errorf("Expected contentMatchWeight to be 2.0, got %f", reranker.contentMatchWeight)
	}
	
	if reranker.fileTypeWeight != 1.0 {
		t.Errorf("Expected fileTypeWeight to be 1.0, got %f", reranker.fileTypeWeight)
	}
}

func TestSimpleReranker_Rerank_EmptyResults(t *testing.T) {
	reranker := NewSimpleReranker()
	ctx := context.Background()
	
	results, err := reranker.Rerank(ctx, "test", []*SearchResult{})
	if err != nil {
		t.Errorf("Rerank with empty results failed: %v", err)
	}
	
	if len(results) != 0 {
		t.Errorf("Expected empty results to remain empty, got %d results", len(results))
	}
}

func TestSimpleReranker_Rerank_ScoreCalculation(t *testing.T) {
	reranker := NewSimpleReranker()
	ctx := context.Background()
	
	// Create test results with different relevance levels
	results := []*SearchResult{
		{
			FilePath: "test.go",
			Content:  "This is a test function",
			Snippet:  "func test() error",
			Score:    50.0,
		},
		{
			FilePath: "main.go", 
			Content:  "Main function without test",
			Snippet:  "func main()",
			Score:    60.0,
		},
		{
			FilePath: "test_helper.go",
			Content:  "Helper functions for testing",
			Snippet:  "func testHelper() string",
			Score:    40.0,
		},
	}
	
	reranked, err := reranker.Rerank(ctx, "test", results)
	if err != nil {
		t.Errorf("Rerank failed: %v", err)
	}
	
	if len(reranked) != 3 {
		t.Errorf("Expected 3 results, got %d", len(reranked))
	}
	
	// Results should be sorted by relevance (highest score first)
	for i := 1; i < len(reranked); i++ {
		if reranked[i-1].Score < reranked[i].Score {
			t.Errorf("Results not properly sorted by score: %f < %f at position %d", 
				reranked[i-1].Score, reranked[i].Score, i)
		}
	}
	
	// File with exact match in name should be ranked higher
	found := false
	for _, result := range reranked {
		if result.FilePath == "test.go" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected test.go to be in results")
	}
}

func TestSimpleReranker_CalculateRelevanceScore(t *testing.T) {
	reranker := NewSimpleReranker()
	
	testCases := []struct {
		name           string
		query          string
		result         *SearchResult
		expectPositive bool
		description    string
	}{
		{
			name:  "exact file name match",
			query: "test",
			result: &SearchResult{
				FilePath: "test.go",
				Content:  "some content",
				Snippet:  "snippet",
			},
			expectPositive: true,
			description:    "Exact match in filename should get high score",
		},
		{
			name:  "prefix file name match",
			query: "test",
			result: &SearchResult{
				FilePath: "test_helper.go",
				Content:  "some content",
				Snippet:  "snippet",
			},
			expectPositive: true,
			description:    "Prefix match in filename should get medium score",
		},
		{
			name:  "content match",
			query: "function",
			result: &SearchResult{
				FilePath: "main.go",
				Content:  "This is a test function implementation",
				Snippet:  "snippet",
			},
			expectPositive: true,
			description:    "Content match should get positive score",
		},
		{
			name:  "snippet match",
			query: "error",
			result: &SearchResult{
				FilePath: "main.go",
				Content:  "some content",
				Snippet:  "func test() error",
			},
			expectPositive: true,
			description:    "Snippet match should get positive score",
		},
		{
			name:  "symbol exact match",
			query: "testfunc",
			result: &SearchResult{
				FilePath: "main.go",
				Content:  "some content",
				Snippet:  "snippet",
				SymbolInfo: &SymbolInfo{
					Name: "testFunc",
				},
			},
			expectPositive: true,
			description:    "Symbol name match should get high score",
		},
		{
			name:  "no match",
			query: "nonexistent",
			result: &SearchResult{
				FilePath: "main.go",
				Content:  "some content",
				Snippet:  "snippet",
			},
			expectPositive: true, // Still gets file type bonus
			description:    "No match should still get file type score",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			score := reranker.calculateRelevanceScore(tc.query, tc.result)
			if tc.expectPositive && score <= 0 {
				t.Errorf("%s: Expected positive score, got %f", tc.description, score)
			}
		})
	}
}

func TestSimpleReranker_GetFileTypeScore(t *testing.T) {
	reranker := NewSimpleReranker()
	
	testCases := []struct {
		filePath     string
		expectedMin  float64
		description  string
	}{
		{"main.go", 3.0, "Go files should get highest score"},
		{"script.js", 2.0, "JavaScript files should get medium score"},
		{"app.ts", 2.0, "TypeScript files should get medium score"},
		{"main.py", 2.0, "Python files should get medium score"},
		{"README.md", 1.0, "Markdown files should get lower score"},
		{"data.txt", 0.5, "Text files should get lowest score"},
		{"unknown.xyz", 1.0, "Unknown files should get default score"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.filePath, func(t *testing.T) {
			score := reranker.getFileTypeScore(tc.filePath)
			if score < tc.expectedMin {
				t.Errorf("%s: Expected score >= %f, got %f", tc.description, tc.expectedMin, score)
			}
		})
	}
}

func TestGetFileName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"main.go", "main.go"},
		{"src/main.go", "main.go"},
		{"pkg/indexing/test.go", "test.go"},
		{"/absolute/path/to/file.go", "file.go"},
		{"", ""},
	}
	
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := getFileName(tc.input)
			if result != tc.expected {
				t.Errorf("Expected getFileName(%s) = %s, got %s", tc.input, tc.expected, result)
			}
		})
	}
}

func TestAdvancedReranker(t *testing.T) {
	reranker := NewAdvancedReranker()
	
	if reranker == nil {
		t.Fatal("NewAdvancedReranker returned nil")
	}
	
	if reranker.baseReranker == nil {
		t.Fatal("AdvancedReranker should have a base reranker")
	}
}

func TestAdvancedReranker_Rerank(t *testing.T) {
	reranker := NewAdvancedReranker()
	ctx := context.Background()
	
	// Test that it properly delegates to base reranker
	results := []*SearchResult{
		{
			FilePath: "test.go",
			Content:  "test content",
			Snippet:  "test snippet",
			Score:    50.0,
		},
		{
			FilePath: "main.go",
			Content:  "main content", 
			Snippet:  "main snippet",
			Score:    60.0,
		},
	}
	
	reranked, err := reranker.Rerank(ctx, "test", results)
	if err != nil {
		t.Errorf("AdvancedReranker.Rerank failed: %v", err)
	}
	
	if len(reranked) != 2 {
		t.Errorf("Expected 2 results, got %d", len(reranked))
	}
	
	// Should be sorted by relevance
	for i := 1; i < len(reranked); i++ {
		if reranked[i-1].Score < reranked[i].Score {
			t.Errorf("Results not properly sorted by score: %f < %f at position %d", 
				reranked[i-1].Score, reranked[i].Score, i)
		}
	}
}

func TestSimpleReranker_SymbolScoring(t *testing.T) {
	reranker := NewSimpleReranker()
	
	// Test symbol exact match
	result := &SearchResult{
		FilePath: "main.go",
		Content:  "some content",
		Snippet:  "snippet",
		SymbolInfo: &SymbolInfo{
			Name: "testFunction",
		},
	}
	
	exactScore := reranker.calculateRelevanceScore("testfunction", result)
	if exactScore <= 0 {
		t.Errorf("Expected positive score for symbol exact match, got %f", exactScore)
	}
	
	// Test symbol prefix match
	prefixScore := reranker.calculateRelevanceScore("test", result)
	if prefixScore <= 0 {
		t.Errorf("Expected positive score for symbol prefix match, got %f", prefixScore)
	}
	
	// Exact match should have higher score than prefix
	if exactScore <= prefixScore {
		t.Errorf("Expected exact match score (%f) to be higher than prefix match (%f)", exactScore, prefixScore)
	}
}

func TestSimpleReranker_MultipleCriteria(t *testing.T) {
	reranker := NewSimpleReranker()
	
	// Result that matches multiple criteria
	result := &SearchResult{
		FilePath: "test_utils.go", // Matches filename
		Content:  "This contains test functions", // Matches content
		Snippet:  "func testHelper() error", // Matches snippet
		Score:    50.0,
		SymbolInfo: &SymbolInfo{
			Name: "testHelper", // Matches symbol
		},
	}
	
	score := reranker.calculateRelevanceScore("test", result)
	
	// Should get bonuses from multiple matching criteria
	expectedMinScore := reranker.prefixMatchWeight + // filename prefix match
		reranker.contentMatchWeight + // content match
		reranker.contentMatchWeight*0.5 + // snippet match  
		reranker.prefixMatchWeight + // symbol prefix match
		reranker.getFileTypeScore("test_utils.go") // file type bonus
		
	if score < expectedMinScore {
		t.Errorf("Expected score >= %f for multiple criteria match, got %f", expectedMinScore, score)
	}
}