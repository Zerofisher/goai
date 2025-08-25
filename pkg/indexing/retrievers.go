package indexing

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// FTSRetriever performs full-text search retrieval
type FTSRetriever struct {
	index Index
}

// NewFTSRetriever creates a new FTS retriever
func NewFTSRetriever(index Index) *FTSRetriever {
	return &FTSRetriever{index: index}
}

// Name returns the retriever name
func (r *FTSRetriever) Name() string {
	return "fts_retriever"
}

// Retrieve performs FTS retrieval
func (r *FTSRetriever) Retrieve(ctx context.Context, req *RetrievalRequest) (*RetrievalResult, error) {
	start := time.Now()
	
	searchOpts := &SearchOptions{
		MaxResults:     req.MaxResults,
		IncludeContent: true,
	}
	
	results, err := r.index.Search(ctx, req.Query, searchOpts)
	if err != nil {
		return nil, fmt.Errorf("FTS search failed: %w", err)
	}
	
	return &RetrievalResult{
		Results:      results,
		TotalResults: len(results),
		SearchType:   SearchTypeFullText,
		Latency:      time.Since(start),
	}, nil
}

// SupportsQuery checks if the retriever supports the query
func (r *FTSRetriever) SupportsQuery(query string) bool {
	return len(query) > 0 // FTS supports any non-empty query
}

// SemanticRetriever performs semantic search retrieval
type SemanticRetriever struct {
	index Index
}

// NewSemanticRetriever creates a new semantic retriever
func NewSemanticRetriever(index Index) *SemanticRetriever {
	return &SemanticRetriever{index: index}
}

// Name returns the retriever name
func (r *SemanticRetriever) Name() string {
	return "semantic_retriever"
}

// Retrieve performs semantic retrieval
func (r *SemanticRetriever) Retrieve(ctx context.Context, req *RetrievalRequest) (*RetrievalResult, error) {
	start := time.Now()
	
	searchOpts := &SearchOptions{
		MaxResults:     req.MaxResults,
		IncludeContent: true,
	}
	
	results, err := r.index.Search(ctx, req.Query, searchOpts)
	if err != nil {
		return nil, fmt.Errorf("semantic search failed: %w", err)
	}
	
	return &RetrievalResult{
		Results:      results,
		TotalResults: len(results),
		SearchType:   SearchTypeSemantic,
		Latency:      time.Since(start),
	}, nil
}

// SupportsQuery checks if the retriever supports the query
func (r *SemanticRetriever) SupportsQuery(query string) bool {
	return len(query) > 2 // Semantic search works better with longer queries
}

// SymbolRetriever performs symbol search retrieval
type SymbolRetriever struct {
	index Index
}

// NewSymbolRetriever creates a new symbol retriever
func NewSymbolRetriever(index Index) *SymbolRetriever {
	return &SymbolRetriever{index: index}
}

// Name returns the retriever name
func (r *SymbolRetriever) Name() string {
	return "symbol_retriever"
}

// Retrieve performs symbol retrieval
func (r *SymbolRetriever) Retrieve(ctx context.Context, req *RetrievalRequest) (*RetrievalResult, error) {
	start := time.Now()
	
	searchOpts := &SearchOptions{
		MaxResults:     req.MaxResults,
		IncludeContent: true,
	}
	
	results, err := r.index.Search(ctx, req.Query, searchOpts)
	if err != nil {
		return nil, fmt.Errorf("symbol search failed: %w", err)
	}
	
	return &RetrievalResult{
		Results:      results,
		TotalResults: len(results),
		SearchType:   SearchTypeSymbol,
		Latency:      time.Since(start),
	}, nil
}

// SupportsQuery checks if the retriever supports the query
func (r *SymbolRetriever) SupportsQuery(query string) bool {
	// Symbol search is good for identifiers
	return len(query) > 0 && isValidIdentifier(query)
}

// RecentFilesRetriever retrieves recently modified files
type RecentFilesRetriever struct {
	workingDir string
	maxAge     time.Duration
}

// NewRecentFilesRetriever creates a new recent files retriever
func NewRecentFilesRetriever(workingDir string, maxAge time.Duration) *RecentFilesRetriever {
	if maxAge == 0 {
		maxAge = 24 * time.Hour // Default to 24 hours
	}
	
	return &RecentFilesRetriever{
		workingDir: workingDir,
		maxAge:     maxAge,
	}
}

// Name returns the retriever name
func (r *RecentFilesRetriever) Name() string {
	return "recent_files_retriever"
}

// Retrieve performs recent files retrieval
func (r *RecentFilesRetriever) Retrieve(ctx context.Context, req *RetrievalRequest) (*RetrievalResult, error) {
	start := time.Now()
	
	// This is a simplified implementation
	// In practice, this would integrate with the file system or Git history
	var results []*SearchResult
	
	// Mock recent files for now
	recentFiles := []string{
		"main.go",
		"pkg/indexing/manager.go",
		"internal/reasoning/engine.go",
	}
	
	cutoff := time.Now().Add(-r.maxAge)
	for _, file := range recentFiles {
		result := &SearchResult{
			FilePath:   file,
			Snippet:    fmt.Sprintf("Recently modified: %s", file),
			Score:      90.0,
			SearchType: SearchTypeRecent,
		}
		results = append(results, result)
		
		if len(results) >= req.MaxResults {
			break
		}
	}
	
	return &RetrievalResult{
		Results:      results,
		TotalResults: len(results),
		SearchType:   SearchTypeRecent,
		Latency:      time.Since(start),
		Metadata: map[string]any{
			"cutoff_time": cutoff,
			"max_age":     r.maxAge.String(),
		},
	}, nil
}

// SupportsQuery checks if the retriever supports the query
func (r *RecentFilesRetriever) SupportsQuery(query string) bool {
	// Recent files retriever works best for general queries
	return true
}

// HybridRetriever combines multiple retrievers
type HybridRetriever struct {
	retrievers []Retriever
	reranker   Reranker
}

// NewHybridRetriever creates a new hybrid retriever
func NewHybridRetriever(retrievers []Retriever, reranker Reranker) *HybridRetriever {
	return &HybridRetriever{
		retrievers: retrievers,
		reranker:   reranker,
	}
}

// Name returns the retriever name
func (r *HybridRetriever) Name() string {
	return "hybrid_retriever"
}

// Retrieve performs hybrid retrieval using multiple retrievers
func (r *HybridRetriever) Retrieve(ctx context.Context, req *RetrievalRequest) (*RetrievalResult, error) {
	start := time.Now()
	
	var allResults []*SearchResult
	var totalResults int
	
	// Run all applicable retrievers in parallel
	type retrieverResult struct {
		results *RetrievalResult
		err     error
		name    string
	}
	
	resultsChan := make(chan retrieverResult, len(r.retrievers))
	
	// Start all retrievers
	activeRetrievers := 0
	for _, retriever := range r.retrievers {
		if retriever.SupportsQuery(req.Query) {
			activeRetrievers++
			go func(ret Retriever) {
				result, err := ret.Retrieve(ctx, req)
				resultsChan <- retrieverResult{
					results: result,
					err:     err,
					name:    ret.Name(),
				}
			}(retriever)
		}
	}
	
	// Collect results
	for i := 0; i < activeRetrievers; i++ {
		result := <-resultsChan
		if result.err != nil {
			fmt.Printf("Warning: Retriever %s failed: %v\n", result.name, result.err)
			continue
		}
		
		allResults = append(allResults, result.results.Results...)
		totalResults += result.results.TotalResults
	}
	
	// Remove duplicates based on file path and line number
	uniqueResults := r.deduplicateResults(allResults)
	
	// Apply reranking if available
	if r.reranker != nil {
		rerankedResults, err := r.reranker.Rerank(ctx, req.Query, uniqueResults)
		if err == nil {
			uniqueResults = rerankedResults
		}
	} else {
		// Default sorting by score
		sort.Slice(uniqueResults, func(i, j int) bool {
			return uniqueResults[i].Score > uniqueResults[j].Score
		})
	}
	
	// Limit results
	if len(uniqueResults) > req.MaxResults {
		uniqueResults = uniqueResults[:req.MaxResults]
	}
	
	return &RetrievalResult{
		Results:      uniqueResults,
		TotalResults: totalResults,
		SearchType:   SearchTypeHybrid,
		Latency:      time.Since(start),
		Metadata: map[string]any{
			"active_retrievers": activeRetrievers,
			"total_retrievers":  len(r.retrievers),
		},
	}, nil
}

// SupportsQuery checks if any retriever supports the query
func (r *HybridRetriever) SupportsQuery(query string) bool {
	for _, retriever := range r.retrievers {
		if retriever.SupportsQuery(query) {
			return true
		}
	}
	return false
}

// deduplicateResults removes duplicate results
func (r *HybridRetriever) deduplicateResults(results []*SearchResult) []*SearchResult {
	seen := make(map[string]bool)
	var unique []*SearchResult
	
	for _, result := range results {
		key := fmt.Sprintf("%s:%d", result.FilePath, result.LineNumber)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, result)
		}
	}
	
	return unique
}

// isValidIdentifier checks if a string looks like a code identifier
func isValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	
	// Simple check for identifier-like strings
	// In practice, this could be more sophisticated
	for _, char := range s {
		if (char < 'a' || char > 'z') &&
			(char < 'A' || char > 'Z') &&
			(char < '0' || char > '9') &&
			char != '_' && char != '.' {
			return false
		}
	}
	
	return true
}