package indexing

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// FTSIndex implements full-text search using SQLite FTS5
type FTSIndex struct {
	db       *sql.DB
	dbPath   string
	chunker  Chunker
	name     string
	version  string
}

// NewFTSIndex creates a new FTS index
func NewFTSIndex(dbPath string, chunker Chunker) (*FTSIndex, error) {
	index := &FTSIndex{
		dbPath:  dbPath,
		chunker: chunker,
		name:    "fts_index",
		version: "1.0.0",
	}
	
	if err := index.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize FTS index: %w", err)
	}
	
	return index, nil
}

// initialize creates the database and tables
func (f *FTSIndex) initialize() error {
	var err error
	f.db, err = sql.Open("sqlite3", f.dbPath+"?_journal=WAL&_synchronous=NORMAL")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	
	// Create FTS5 table for full-text search
	createFTSTable := `
		CREATE VIRTUAL TABLE IF NOT EXISTS fts_chunks USING fts5(
			chunk_id,
			file_path,
			content,
			language,
			chunk_type,
			start_line,
			end_line,
			tokenize = 'porter unicode61'
		);
	`
	
	// Create metadata table for storing chunk information
	createMetaTable := `
		CREATE TABLE IF NOT EXISTS chunk_metadata (
			chunk_id TEXT PRIMARY KEY,
			file_path TEXT NOT NULL,
			content_hash TEXT NOT NULL,
			start_line INTEGER NOT NULL,
			end_line INTEGER NOT NULL,
			start_byte INTEGER NOT NULL,
			end_byte INTEGER NOT NULL,
			language TEXT NOT NULL,
			chunk_type TEXT NOT NULL,
			file_size INTEGER NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);
	`
	
	// Create indexes for better query performance
	createIndexes := `
		CREATE INDEX IF NOT EXISTS idx_chunk_file_path ON chunk_metadata(file_path);
		CREATE INDEX IF NOT EXISTS idx_chunk_language ON chunk_metadata(language);
		CREATE INDEX IF NOT EXISTS idx_chunk_type ON chunk_metadata(chunk_type);
		CREATE INDEX IF NOT EXISTS idx_chunk_hash ON chunk_metadata(content_hash);
		CREATE INDEX IF NOT EXISTS idx_updated_at ON chunk_metadata(updated_at);
	`
	
	// Execute table creation
	if _, err := f.db.Exec(createFTSTable); err != nil {
		return fmt.Errorf("failed to create FTS table: %w", err)
	}
	
	if _, err := f.db.Exec(createMetaTable); err != nil {
		return fmt.Errorf("failed to create metadata table: %w", err)
	}
	
	if _, err := f.db.Exec(createIndexes); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}
	
	// Configure FTS5 for better search
	configureFTS := `
		INSERT OR REPLACE INTO fts_chunks(fts_chunks, rank) VALUES('automerge', 8);
	`
	if _, err := f.db.Exec(configureFTS); err != nil {
		// This might fail on older SQLite versions, so we'll log but not fail
		fmt.Printf("Warning: Failed to configure FTS5 automerge: %v\n", err)
	}
	
	return nil
}

// Name returns the index name
func (f *FTSIndex) Name() string {
	return f.name
}

// Version returns the index version
func (f *FTSIndex) Version() string {
	return f.version
}

// Update adds or updates chunks in the index
func (f *FTSIndex) Update(ctx context.Context, tag *IndexTag, changes *IndexChanges) error {
	tx, err := f.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Process added files
	for _, filePath := range changes.Added {
		if err := f.indexFile(ctx, tx, filePath); err != nil {
			return fmt.Errorf("failed to index file %s: %w", filePath, err)
		}
	}
	
	// Process modified files
	for _, filePath := range changes.Modified {
		// Remove existing chunks for this file
		if err := f.removeFileChunks(ctx, tx, filePath); err != nil {
			return fmt.Errorf("failed to remove existing chunks for %s: %w", filePath, err)
		}
		
		// Re-index the file
		if err := f.indexFile(ctx, tx, filePath); err != nil {
			return fmt.Errorf("failed to re-index file %s: %w", filePath, err)
		}
	}
	
	return tx.Commit()
}

// Remove removes files from the index
func (f *FTSIndex) Remove(ctx context.Context, tag *IndexTag, files []string) error {
	tx, err := f.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	for _, filePath := range files {
		if filePath == "*" {
			// Remove all chunks
			if err := f.removeAllChunks(ctx, tx); err != nil {
				return fmt.Errorf("failed to remove all chunks: %w", err)
			}
		} else {
			if err := f.removeFileChunks(ctx, tx, filePath); err != nil {
				return fmt.Errorf("failed to remove chunks for %s: %w", filePath, err)
			}
		}
	}
	
	return tx.Commit()
}

// Search performs full-text search using FTS5
func (f *FTSIndex) Search(ctx context.Context, query string, opts *SearchOptions) ([]*SearchResult, error) {
	if opts == nil {
		opts = &SearchOptions{MaxResults: 50}
	}
	
	// Sanitize and prepare FTS query
	ftsQuery := f.prepareFTSQuery(query)
	
	// Build SQL query
	sqlQuery := `
		SELECT 
			f.chunk_id,
			f.file_path,
			snippet(fts_chunks, 2, '<mark>', '</mark>', '...', 32) as snippet,
			rank,
			m.start_line,
			m.end_line,
			m.language,
			m.chunk_type
		FROM fts_chunks f
		JOIN chunk_metadata m ON f.chunk_id = m.chunk_id
		WHERE fts_chunks MATCH ?
	`
	
	// Add file type filtering if specified
	args := []interface{}{ftsQuery}
	if len(opts.FileTypes) > 0 {
		placeholders := make([]string, len(opts.FileTypes))
		for i, fileType := range opts.FileTypes {
			placeholders[i] = "?"
			args = append(args, fileType)
		}
		sqlQuery += fmt.Sprintf(" AND m.language IN (%s)", strings.Join(placeholders, ","))
	}
	
	sqlQuery += " ORDER BY rank LIMIT ?"
	args = append(args, opts.MaxResults)
	
	if opts.Offset > 0 {
		sqlQuery += " OFFSET ?"
		args = append(args, opts.Offset)
	}
	
	rows, err := f.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute FTS query: %w", err)
	}
	defer rows.Close()
	
	var results []*SearchResult
	for rows.Next() {
		var result SearchResult
		var rank float64
		
		err := rows.Scan(
			&result.FilePath,
			&result.FilePath,
			&result.Snippet,
			&rank,
			&result.LineNumber,
			&result.ColumnNumber,
			&result.Metadata,
			&result.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}
		
		// Convert FTS5 rank to a normalized score (0-1)
		result.Score = f.normalizeRank(rank)
		result.SearchType = SearchTypeFullText
		
		if opts.IncludeContent {
			content, err := f.getChunkContent(ctx, result.FilePath)
			if err == nil {
				result.Content = content
			}
		}
		
		results = append(results, &result)
	}
	
	return results, rows.Err()
}

// GetStats returns index statistics
func (f *FTSIndex) GetStats() (*IndexStats, error) {
	var totalDocs int64
	var totalSize int64
	
	err := f.db.QueryRow("SELECT COUNT(*), SUM(file_size) FROM chunk_metadata").Scan(&totalDocs, &totalSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	
	return &IndexStats{
		Name:           f.name,
		Version:        f.version,
		TotalDocuments: totalDocs,
		TotalSize:      totalSize,
		LastUpdated:    time.Now(), // TODO: Store actual last updated time
	}, nil
}

// Close closes the database connection
func (f *FTSIndex) Close() error {
	if f.db != nil {
		return f.db.Close()
	}
	return nil
}

// indexFile processes and indexes a single file
func (f *FTSIndex) indexFile(ctx context.Context, tx *sql.Tx, filePath string) error {
	// Read file content
	fileContent, err := f.readFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Chunk the file
	chunks, err := f.chunker.ChunkFile(ctx, fileContent)
	if err != nil {
		return fmt.Errorf("failed to chunk file: %w", err)
	}
	
	now := time.Now()
	
	// Insert chunks into FTS and metadata tables
	for _, chunk := range chunks {
		// Insert into FTS table
		_, err := tx.ExecContext(ctx, `
			INSERT INTO fts_chunks (chunk_id, file_path, content, language, chunk_type, start_line, end_line)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, chunk.ID, chunk.FilePath, chunk.Content, chunk.Language, string(chunk.ChunkType), chunk.StartLine, chunk.EndLine)
		
		if err != nil {
			return fmt.Errorf("failed to insert chunk into FTS table: %w", err)
		}
		
		// Insert into metadata table
		contentHash := chunk.Metadata["hash"].(string)
		_, err = tx.ExecContext(ctx, `
			INSERT INTO chunk_metadata (
				chunk_id, file_path, content_hash, start_line, end_line, start_byte, end_byte,
				language, chunk_type, file_size, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, chunk.ID, chunk.FilePath, contentHash, chunk.StartLine, chunk.EndLine, chunk.StartByte, chunk.EndByte,
			chunk.Language, string(chunk.ChunkType), fileContent.Info.Size, now, now)
		
		if err != nil {
			return fmt.Errorf("failed to insert chunk metadata: %w", err)
		}
	}
	
	return nil
}

// removeFileChunks removes all chunks for a specific file
func (f *FTSIndex) removeFileChunks(ctx context.Context, tx *sql.Tx, filePath string) error {
	// Get chunk IDs to remove from FTS table
	rows, err := tx.QueryContext(ctx, "SELECT chunk_id FROM chunk_metadata WHERE file_path = ?", filePath)
	if err != nil {
		return fmt.Errorf("failed to query chunks to remove: %w", err)
	}
	defer rows.Close()
	
	var chunkIDs []string
	for rows.Next() {
		var chunkID string
		if err := rows.Scan(&chunkID); err != nil {
			return fmt.Errorf("failed to scan chunk ID: %w", err)
		}
		chunkIDs = append(chunkIDs, chunkID)
	}
	
	// Remove from FTS table
	for _, chunkID := range chunkIDs {
		_, err := tx.ExecContext(ctx, "DELETE FROM fts_chunks WHERE chunk_id = ?", chunkID)
		if err != nil {
			return fmt.Errorf("failed to remove chunk from FTS table: %w", err)
		}
	}
	
	// Remove from metadata table
	_, err = tx.ExecContext(ctx, "DELETE FROM chunk_metadata WHERE file_path = ?", filePath)
	if err != nil {
		return fmt.Errorf("failed to remove chunk metadata: %w", err)
	}
	
	return nil
}

// removeAllChunks removes all chunks from the index
func (f *FTSIndex) removeAllChunks(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, "DELETE FROM fts_chunks"); err != nil {
		return fmt.Errorf("failed to clear FTS table: %w", err)
	}
	
	if _, err := tx.ExecContext(ctx, "DELETE FROM chunk_metadata"); err != nil {
		return fmt.Errorf("failed to clear metadata table: %w", err)
	}
	
	return nil
}

// prepareFTSQuery sanitizes and prepares a query for FTS5
func (f *FTSIndex) prepareFTSQuery(query string) string {
	// Basic sanitization - escape FTS5 special characters
	query = strings.ReplaceAll(query, "\"", "\"\"")
	
	// Split query into terms
	terms := strings.Fields(query)
	if len(terms) == 0 {
		return query
	}
	
	// For simple queries, just return as-is
	if len(terms) == 1 {
		return terms[0]
	}
	
	// For multi-term queries, use AND by default
	return strings.Join(terms, " AND ")
}

// normalizeRank converts FTS5 rank to a 0-1 score
func (f *FTSIndex) normalizeRank(rank float64) float64 {
	// FTS5 rank is negative (closer to 0 = better match)
	// Convert to positive score between 0 and 1
	if rank >= 0 {
		return 0.0
	}
	
	// Simple normalization - this could be improved with better ranking
	normalizedScore := 1.0 / (1.0 + (-rank))
	if normalizedScore > 1.0 {
		normalizedScore = 1.0
	}
	
	return normalizedScore
}

// getChunkContent retrieves the full content of a chunk
func (f *FTSIndex) getChunkContent(ctx context.Context, chunkID string) (string, error) {
	var content string
	err := f.db.QueryRowContext(ctx, "SELECT content FROM fts_chunks WHERE chunk_id = ?", chunkID).Scan(&content)
	if err != nil {
		return "", fmt.Errorf("failed to get chunk content: %w", err)
	}
	return content, nil
}

// readFile reads a file and returns FileContent
func (f *FTSIndex) readFile(filePath string) (*FileContent, error) {
	// This is a simplified implementation - in reality, you'd want proper file reading
	// with encoding detection, language detection, etc.
	
	info := &FileInfo{
		Path:     filePath,
		Language: f.detectLanguage(filePath),
		IsText:   true,
		Encoding: "utf-8",
	}
	
	// For now, return empty content - this would be implemented with actual file reading
	return &FileContent{
		Info:    info,
		Content: []byte{},
		Hash:    "",
	}, nil
}

// detectLanguage detects the programming language from file extension
func (f *FTSIndex) detectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	
	languageMap := map[string]string{
		".go":   "go",
		".js":   "javascript",
		".ts":   "typescript",
		".py":   "python",
		".java": "java",
		".cpp":  "cpp",
		".c":    "c",
		".rs":   "rust",
		".md":   "markdown",
		".txt":  "text",
	}
	
	if lang, exists := languageMap[ext]; exists {
		return lang
	}
	
	return "text"
}