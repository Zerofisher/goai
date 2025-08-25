package indexing

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// EnhancedIndexManager extends IndexManager with all index types and retrievers
type EnhancedIndexManager struct {
	*IndexManager
	symbolIndex     *SymbolIndex
	embeddingIndex  *EmbeddingIndex
	retrievers      []Retriever
	hybridRetriever *HybridRetriever
	reranker        Reranker
	mu              sync.RWMutex
}

// NewEnhancedIndexManager creates a new enhanced index manager with all components
func NewEnhancedIndexManager(workingDir string) (*EnhancedIndexManager, error) {
	// Create base directory structure
	indexDir := filepath.Join(workingDir, ".goai")
	if err := os.MkdirAll(indexDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create index directory: %w", err)
	}

	// Create base index manager
	baseManager, err := NewIndexManagerWithDefaults(filepath.Join(indexDir, "fts_index.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to create base index manager: %w", err)
	}

	enhanced := &EnhancedIndexManager{
		IndexManager: baseManager,
		retrievers:   make([]Retriever, 0),
	}

	// Initialize additional indexes
	if err := enhanced.initializeIndexes(indexDir); err != nil {
		return nil, fmt.Errorf("failed to initialize enhanced indexes: %w", err)
	}

	// Initialize retrievers and reranker
	enhanced.initializeRetrievers()

	return enhanced, nil
}

// initializeIndexes sets up symbol and embedding indexes
func (em *EnhancedIndexManager) initializeIndexes(indexDir string) error {
	// Initialize symbol index
	symbolIndexPath := filepath.Join(indexDir, "symbol_index.db")
	parser := NewGoTreeSitterParser()
	chunker := NewDefaultChunker()
	
	symbolIndex, err := NewSymbolIndex(symbolIndexPath, parser, chunker)
	if err != nil {
		return fmt.Errorf("failed to create symbol index: %w", err)
	}
	em.symbolIndex = symbolIndex
	em.RegisterIndex(symbolIndex)

	// Initialize embedding index (try real provider first, fallback to mock)
	embeddingIndexPath := filepath.Join(indexDir, "embedding_index.db")
	
	var embeddingProvider EmbeddingProvider
	if ctx := context.Background(); true {
		// Try to create real embedding provider
		if realProvider, err := NewEinoEmbeddingProvider(ctx); err == nil {
			embeddingProvider = realProvider
		} else {
			fmt.Printf("Warning: Failed to create OpenAI embedding provider, using mock: %v\n", err)
			embeddingProvider = NewMockEmbeddingProvider(384)
		}
	}
	
	embeddingIndex, err := NewEmbeddingIndex(embeddingIndexPath, embeddingProvider, chunker)
	if err != nil {
		return fmt.Errorf("failed to create embedding index: %w", err)
	}
	em.embeddingIndex = embeddingIndex
	em.RegisterIndex(embeddingIndex)

	return nil
}

// initializeRetrievers sets up all retrievers and hybrid pipeline
func (em *EnhancedIndexManager) initializeRetrievers() {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Create individual retrievers
	var retrievers []Retriever

	// FTS retriever - get the first registered index (should be FTS)
	em.IndexManager.mu.RLock()
	for _, index := range em.indexes {
		if index.Name() == "fts_index" {
			retrievers = append(retrievers, NewFTSRetriever(index))
			break
		}
	}
	em.IndexManager.mu.RUnlock()

	// Symbol retriever
	if em.symbolIndex != nil {
		retrievers = append(retrievers, NewSymbolRetriever(em.symbolIndex))
	}

	// Semantic retriever
	if em.embeddingIndex != nil {
		retrievers = append(retrievers, NewSemanticRetriever(em.embeddingIndex))
	}

	// Recent files retriever
	retrievers = append(retrievers, NewRecentFilesRetriever("", 0)) // Will be set per request

	em.retrievers = retrievers

	// Create reranker
	em.reranker = NewSimpleReranker()

	// Create hybrid retriever
	em.hybridRetriever = NewHybridRetriever(retrievers, em.reranker)
}

// SearchWithHybridRetrieval performs hybrid search using all available retrievers
func (em *EnhancedIndexManager) SearchWithHybridRetrieval(ctx context.Context, req *SearchRequest) (*RetrievalResult, error) {
	if em.hybridRetriever == nil {
		return nil, fmt.Errorf("hybrid retriever not initialized")
	}

	retrievalReq := &RetrievalRequest{
		Query:      req.Query,
		WorkingDir: req.WorkingDir,
		MaxResults: req.MaxResults,
		SearchTypes: req.SearchTypes,
		Context:    req.Context,
	}

	return em.hybridRetriever.Retrieve(ctx, retrievalReq)
}

// SearchSymbols searches specifically for code symbols
func (em *EnhancedIndexManager) SearchSymbols(ctx context.Context, query string, opts *SearchOptions) ([]*SearchResult, error) {
	if em.symbolIndex == nil {
		return nil, fmt.Errorf("symbol index not available")
	}

	return em.symbolIndex.Search(ctx, query, opts)
}

// SearchSemantic performs semantic search using embeddings
func (em *EnhancedIndexManager) SearchSemantic(ctx context.Context, query string, opts *SearchOptions) ([]*SearchResult, error) {
	if em.embeddingIndex == nil {
		return nil, fmt.Errorf("embedding index not available")
	}

	return em.embeddingIndex.Search(ctx, query, opts)
}

// GetAvailableSearchTypes returns the types of search supported by this manager
func (em *EnhancedIndexManager) GetAvailableSearchTypes() []SearchType {
	var types []SearchType
	
	em.mu.RLock()
	defer em.mu.RUnlock()

	// Always have FTS
	types = append(types, SearchTypeFullText)
	
	if em.symbolIndex != nil {
		types = append(types, SearchTypeSymbol)
	}
	
	if em.embeddingIndex != nil {
		types = append(types, SearchTypeSemantic)
	}
	
	// Always have recent and hybrid
	types = append(types, SearchTypeRecent, SearchTypeHybrid)
	
	return types
}

// GetIndexStats returns comprehensive statistics for all indexes
func (em *EnhancedIndexManager) GetIndexStats() (map[string]*IndexStats, error) {
	stats := make(map[string]*IndexStats)
	
	// Get base FTS stats
	em.IndexManager.mu.RLock()
	for name, index := range em.indexes {
		if indexStats, err := index.GetStats(); err == nil {
			stats[name] = indexStats
		}
	}
	em.IndexManager.mu.RUnlock()
	
	// Get symbol stats
	if em.symbolIndex != nil {
		if symbolStats, err := em.symbolIndex.GetStats(); err == nil {
			stats["symbol_index"] = symbolStats
		}
	}
	
	// Get embedding stats
	if em.embeddingIndex != nil {
		if embeddingStats, err := em.embeddingIndex.GetStats(); err == nil {
			stats["embedding_index"] = embeddingStats
		}
	}
	
	return stats, nil
}

// Close closes all indexes and cleans up resources
func (em *EnhancedIndexManager) Close() error {
	var errors []string
	
	// Close additional indexes
	if em.symbolIndex != nil {
		if err := em.symbolIndex.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("failed to close symbol index: %v", err))
		}
	}
	
	if em.embeddingIndex != nil {
		if err := em.embeddingIndex.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("failed to close embedding index: %v", err))
		}
	}
	
	// Close base manager
	if err := em.IndexManager.Close(); err != nil {
		errors = append(errors, fmt.Sprintf("failed to close base manager: %v", err))
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("errors during close: %v", errors)
	}
	
	return nil
}

// BuildEnhancedIndex builds all index types
func (em *EnhancedIndexManager) BuildEnhancedIndex(ctx context.Context, workdir string) error {
	// Use the base manager's build functionality which will trigger all registered indexes
	return em.BuildIndex(ctx, workdir)
}

// RefreshEnhancedIndex refreshes all indexes with specific file changes
func (em *EnhancedIndexManager) RefreshEnhancedIndex(ctx context.Context, paths []string) error {
	// Use the base manager's refresh functionality which will trigger all registered indexes
	return em.RefreshIndex(ctx, paths)
}