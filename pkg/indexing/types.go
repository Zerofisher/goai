package indexing

import (
	"time"
)

// IndexStatus represents the current status of an index
type IndexStatus struct {
	WorkingDirectory string                 `json:"working_directory"`
	IsReady         bool                   `json:"is_ready"`
	IsBuilding      bool                   `json:"is_building"`
	LastUpdated     time.Time              `json:"last_updated"`
	TotalFiles      int                    `json:"total_files"`
	IndexedFiles    int                    `json:"indexed_files"`
	TotalSize       int64                  `json:"total_size"`
	IndexSize       int64                  `json:"index_size"`
	IndexVersions   map[string]string      `json:"index_versions"`
	Errors          []string               `json:"errors,omitempty"`
}

// SearchRequest contains parameters for search operations
type SearchRequest struct {
	Query           string            `json:"query"`
	WorkingDir      string            `json:"working_dir"`
	FileTypes       []string          `json:"file_types,omitempty"`
	MaxResults      int               `json:"max_results,omitempty"`
	IncludeContent  bool              `json:"include_content"`
	SearchTypes     []SearchType      `json:"search_types"`
	Context         map[string]interface{} `json:"context,omitempty"`
}

// SearchResult represents a search result
type SearchResult struct {
	FilePath        string            `json:"file_path"`
	LineNumber      int               `json:"line_number,omitempty"`
	ColumnNumber    int               `json:"column_number,omitempty"`
	Content         string            `json:"content,omitempty"`
	Snippet         string            `json:"snippet"`
	Score           float64           `json:"score"`
	SearchType      SearchType        `json:"search_type"`
	SymbolInfo      *SymbolInfo       `json:"symbol_info,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ContextItem represents a code context item for reasoning
type ContextItem struct {
	FilePath        string            `json:"file_path"`
	Content         string            `json:"content"`
	StartLine       int               `json:"start_line"`
	EndLine         int               `json:"end_line"`
	ContextType     ContextType       `json:"context_type"`
	Relevance       float64           `json:"relevance"`
	SymbolInfo      *SymbolInfo       `json:"symbol_info,omitempty"`
}

// ContextOptions controls how context items are selected
type ContextOptions struct {
	MaxItems        int               `json:"max_items"`
	MaxCharsPerItem int               `json:"max_chars_per_item"`
	ContextTypes    []ContextType     `json:"context_types"`
	RelevanceThreshold float64        `json:"relevance_threshold"`
}

// IndexTag identifies an index operation
type IndexTag struct {
	WorkingDirectory string            `json:"working_directory"`
	Branch          string            `json:"branch,omitempty"`
	Commit          string            `json:"commit,omitempty"`
	Timestamp       time.Time         `json:"timestamp"`
}

// IndexChanges represents files that need indexing
type IndexChanges struct {
	Added           []string          `json:"added"`
	Modified        []string          `json:"modified"`
	Deleted         []string          `json:"deleted"`
}

// SearchOptions controls index search behavior
type SearchOptions struct {
	MaxResults      int               `json:"max_results"`
	Offset          int               `json:"offset"`
	IncludeContent  bool              `json:"include_content"`
	FileTypes       []string          `json:"file_types"`
	Filters         map[string]interface{} `json:"filters"`
}

// IndexStats provides index statistics
type IndexStats struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	TotalDocuments  int64             `json:"total_documents"`
	TotalSize       int64             `json:"total_size"`
	LastUpdated     time.Time         `json:"last_updated"`
	BuildTime       time.Duration     `json:"build_time"`
	QueryStats      *QueryStats       `json:"query_stats,omitempty"`
}

// QueryStats tracks search performance
type QueryStats struct {
	TotalQueries    int64             `json:"total_queries"`
	AverageLatency  time.Duration     `json:"average_latency"`
	CacheHitRate    float64           `json:"cache_hit_rate"`
}

// RetrievalRequest contains parameters for retrieval operations
type RetrievalRequest struct {
	Query           string            `json:"query"`
	WorkingDir      string            `json:"working_dir"`
	MaxResults      int               `json:"max_results"`
	SearchTypes     []SearchType      `json:"search_types"`
	Context         map[string]interface{} `json:"context"`
}

// RetrievalResult contains results from a retriever
type RetrievalResult struct {
	Results         []*SearchResult   `json:"results"`
	TotalResults    int               `json:"total_results"`
	SearchType      SearchType        `json:"search_type"`
	Latency         time.Duration     `json:"latency"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// VectorSearchResult represents a vector similarity search result
type VectorSearchResult struct {
	ID              string            `json:"id"`
	Score           float64           `json:"score"`
	Vector          []float64         `json:"vector,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// FileInfo contains file metadata for indexing
type FileInfo struct {
	Path            string            `json:"path"`
	Size            int64             `json:"size"`
	ModTime         time.Time         `json:"mod_time"`
	Language        string            `json:"language"`
	IsText          bool              `json:"is_text"`
	Encoding        string            `json:"encoding"`
}

// FileContent contains file content and metadata
type FileContent struct {
	Info            *FileInfo         `json:"info"`
	Content         []byte            `json:"content"`
	Hash            string            `json:"hash"`
}

// Chunk represents a segment of a file for indexing
type Chunk struct {
	ID              string            `json:"id"`
	FilePath        string            `json:"file_path"`
	Content         string            `json:"content"`
	StartLine       int               `json:"start_line"`
	EndLine         int               `json:"end_line"`
	StartByte       int               `json:"start_byte"`
	EndByte         int               `json:"end_byte"`
	Language        string            `json:"language"`
	ChunkType       ChunkType         `json:"chunk_type"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// SymbolInfo represents code symbols (functions, classes, etc.)
type SymbolInfo struct {
	Name            string            `json:"name"`
	Kind            SymbolKind        `json:"kind"`
	FilePath        string            `json:"file_path"`
	StartLine       int               `json:"start_line"`
	EndLine         int               `json:"end_line"`
	StartColumn     int               `json:"start_column"`
	EndColumn       int               `json:"end_column"`
	Language        string            `json:"language"`
	Signature       string            `json:"signature,omitempty"`
	DocString       string            `json:"doc_string,omitempty"`
	Parent          string            `json:"parent,omitempty"`
	Children        []string          `json:"children,omitempty"`
}

// ParsedFile represents a tree-sitter parsed file
type ParsedFile struct {
	FilePath        string            `json:"file_path"`
	Language        string            `json:"language"`
	Symbols         []*SymbolInfo     `json:"symbols"`
	Imports         []string          `json:"imports"`
	Dependencies    []string          `json:"dependencies"`
	Errors          []ParseError      `json:"errors,omitempty"`
}

// ParseError represents a parsing error
type ParseError struct {
	Line            int               `json:"line"`
	Column          int               `json:"column"`
	Message         string            `json:"message"`
	Severity        string            `json:"severity"`
}

// Enums

// SearchType represents different types of search
type SearchType string

const (
	SearchTypeFullText   SearchType = "full_text"
	SearchTypeSemantic   SearchType = "semantic"
	SearchTypeSymbol     SearchType = "symbol"
	SearchTypeRecent     SearchType = "recent"
	SearchTypeHybrid     SearchType = "hybrid"
)

// ContextType represents different types of code context
type ContextType string

const (
	ContextTypeFunction     ContextType = "function"
	ContextTypeClass        ContextType = "class"
	ContextTypeInterface    ContextType = "interface"
	ContextTypeVariable     ContextType = "variable"
	ContextTypeImport       ContextType = "import"
	ContextTypeComment      ContextType = "comment"
	ContextTypeSnippet      ContextType = "snippet"
)

// ChunkType represents different types of chunks
type ChunkType string

const (
	ChunkTypeCode           ChunkType = "code"
	ChunkTypeFunction       ChunkType = "function"
	ChunkTypeClass          ChunkType = "class"
	ChunkTypeComment        ChunkType = "comment"
	ChunkTypeDocumentation  ChunkType = "documentation"
	ChunkTypeTest           ChunkType = "test"
)

// SymbolKind represents different kinds of code symbols
type SymbolKind string

const (
	SymbolKindFunction      SymbolKind = "function"
	SymbolKindMethod        SymbolKind = "method"
	SymbolKindClass         SymbolKind = "class"
	SymbolKindInterface     SymbolKind = "interface"
	SymbolKindStruct        SymbolKind = "struct"
	SymbolKindVariable      SymbolKind = "variable"
	SymbolKindConstant      SymbolKind = "constant"
	SymbolKindType          SymbolKind = "type"
	SymbolKindField         SymbolKind = "field"
	SymbolKindProperty      SymbolKind = "property"
	SymbolKindPackage       SymbolKind = "package"
	SymbolKindModule        SymbolKind = "module"
)