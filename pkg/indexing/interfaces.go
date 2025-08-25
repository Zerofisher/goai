package indexing

import (
	"context"
)

// CodebaseIndexer is the main interface for codebase indexing operations
type CodebaseIndexer interface {
	// Index lifecycle management
	BuildIndex(ctx context.Context, workdir string) error
	RefreshIndex(ctx context.Context, paths []string) error
	ClearIndex(ctx context.Context, workdir string) error
	
	// Index status
	GetIndexStatus(workdir string) (*IndexStatus, error)
	IsIndexReady(workdir string) bool
	WaitForIndex(ctx context.Context, workdir string) error
	
	// Search and retrieval
	Search(ctx context.Context, req *SearchRequest) (*SearchResult, error)
	GetContextItems(ctx context.Context, query string, opts *ContextOptions) ([]*ContextItem, error)
}

// Index represents a specific index implementation
type Index interface {
	Name() string
	Version() string
	
	Update(ctx context.Context, tag *IndexTag, changes *IndexChanges) error
	Remove(ctx context.Context, tag *IndexTag, files []string) error
	Search(ctx context.Context, query string, opts *SearchOptions) ([]*SearchResult, error)
	
	GetStats() (*IndexStats, error)
	Close() error
}

// Retriever represents a search retriever implementation
type Retriever interface {
	Name() string
	Retrieve(ctx context.Context, req *RetrievalRequest) (*RetrievalResult, error)
	SupportsQuery(query string) bool
}

// FileFilter handles file discovery and filtering
type FileFilter interface {
	ShouldIndex(path string, info FileInfo) bool
	ShouldIgnore(path string) bool
	GetIgnorePatterns() []string
}

// DirectoryWalker handles recursive directory traversal
type DirectoryWalker interface {
	Walk(ctx context.Context, rootPath string, filter FileFilter) (<-chan FileInfo, error)
}

// Chunker breaks files into indexable chunks
type Chunker interface {
	ChunkFile(ctx context.Context, file *FileContent) ([]*Chunk, error)
	ChunkText(text string, language string) ([]*Chunk, error)
}

// EmbeddingProvider generates vector embeddings
type EmbeddingProvider interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float64, error)
	GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float64, error)
	GetDimensions() int
}

// VectorDB handles vector storage and similarity search
type VectorDB interface {
	Store(ctx context.Context, id string, vector []float64, metadata map[string]interface{}) error
	Search(ctx context.Context, query []float64, limit int) ([]*VectorSearchResult, error)
	Delete(ctx context.Context, ids []string) error
	Close() error
}

// Reranker reorders search results
type Reranker interface {
	Rerank(ctx context.Context, query string, results []*SearchResult) ([]*SearchResult, error)
}

// TreeSitterParser analyzes code structure
type TreeSitterParser interface {
	ParseFile(ctx context.Context, filePath string, content []byte) (*ParsedFile, error)
	GetSupportedLanguages() []string
}