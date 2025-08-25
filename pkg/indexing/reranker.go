package indexing

import (
	"context"
	"sort"
	"strings"
)

// SimpleReranker implements a basic reranking algorithm
type SimpleReranker struct {
	// Configuration for reranking weights
	exactMatchWeight   float64
	prefixMatchWeight  float64
	contentMatchWeight float64
	fileTypeWeight     float64
}

// NewSimpleReranker creates a new simple reranker
func NewSimpleReranker() *SimpleReranker {
	return &SimpleReranker{
		exactMatchWeight:   10.0,
		prefixMatchWeight:  5.0,
		contentMatchWeight: 2.0,
		fileTypeWeight:     1.0,
	}
}

// Rerank reorders search results based on relevance
func (r *SimpleReranker) Rerank(ctx context.Context, query string, results []*SearchResult) ([]*SearchResult, error) {
	if len(results) == 0 {
		return results, nil
	}

	// Calculate new scores for each result
	queryLower := strings.ToLower(query)
	
	for _, result := range results {
		newScore := r.calculateRelevanceScore(queryLower, result)
		// Combine with original score (weighted average)
		result.Score = (result.Score + newScore*2) / 3
	}

	// Sort by new scores
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// calculateRelevanceScore calculates a relevance score for a result
func (r *SimpleReranker) calculateRelevanceScore(query string, result *SearchResult) float64 {
	score := 0.0
	
	// File name scoring
	fileName := strings.ToLower(getFileName(result.FilePath))
	if strings.Contains(fileName, query) {
		if fileName == query {
			score += r.exactMatchWeight
		} else if strings.HasPrefix(fileName, query) {
			score += r.prefixMatchWeight
		} else {
			score += r.contentMatchWeight
		}
	}

	// Content scoring
	contentLower := strings.ToLower(result.Content)
	snippetLower := strings.ToLower(result.Snippet)
	
	if strings.Contains(contentLower, query) {
		score += r.contentMatchWeight
	}
	if strings.Contains(snippetLower, query) {
		score += r.contentMatchWeight * 0.5
	}

	// File type preferences
	score += r.getFileTypeScore(result.FilePath)

	// Symbol info bonus
	if result.SymbolInfo != nil {
		symbolNameLower := strings.ToLower(result.SymbolInfo.Name)
		if strings.Contains(symbolNameLower, query) {
			if symbolNameLower == query {
				score += r.exactMatchWeight
			} else if strings.HasPrefix(symbolNameLower, query) {
				score += r.prefixMatchWeight
			} else {
				score += r.contentMatchWeight
			}
		}
	}

	return score
}

// getFileTypeScore returns a preference score based on file type
func (r *SimpleReranker) getFileTypeScore(filePath string) float64 {
	switch {
	case strings.HasSuffix(filePath, ".go"):
		return r.fileTypeWeight * 3.0 // High preference for Go files
	case strings.HasSuffix(filePath, ".js"), strings.HasSuffix(filePath, ".ts"):
		return r.fileTypeWeight * 2.0 // Medium preference for JS/TS
	case strings.HasSuffix(filePath, ".py"):
		return r.fileTypeWeight * 2.0 // Medium preference for Python
	case strings.HasSuffix(filePath, ".md"):
		return r.fileTypeWeight * 1.0 // Lower preference for markdown
	case strings.HasSuffix(filePath, ".txt"):
		return r.fileTypeWeight * 0.5 // Lower preference for text files
	default:
		return r.fileTypeWeight * 1.0 // Default preference
	}
}

// getFileName extracts the file name from a path
func getFileName(filePath string) string {
	parts := strings.Split(filePath, "/")
	if len(parts) == 0 {
		return filePath
	}
	return parts[len(parts)-1]
}

// AdvancedReranker would implement more sophisticated reranking
// This could include ML-based reranking, user feedback learning, etc.
type AdvancedReranker struct {
	// Placeholder for future advanced reranking implementation
	baseReranker Reranker
}

// NewAdvancedReranker creates a new advanced reranker
func NewAdvancedReranker() *AdvancedReranker {
	return &AdvancedReranker{
		baseReranker: NewSimpleReranker(),
	}
}

// Rerank performs advanced reranking (currently delegates to simple reranker)
func (r *AdvancedReranker) Rerank(ctx context.Context, query string, results []*SearchResult) ([]*SearchResult, error) {
	// For now, delegate to the simple reranker
	// In the future, this could implement:
	// - ML-based relevance scoring
	// - User behavior learning
	// - Context-aware ranking
	// - Cross-language understanding
	
	return r.baseReranker.Rerank(ctx, query, results)
}