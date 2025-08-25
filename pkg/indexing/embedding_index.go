package indexing

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// EmbeddingIndex implements Index interface for semantic search using vector embeddings
type EmbeddingIndex struct {
	name              string
	version           string
	dbPath            string
	db                *sql.DB
	embeddingProvider EmbeddingProvider
	chunker           Chunker
	dimensions        int
}

// NewEmbeddingIndex creates a new embedding index
func NewEmbeddingIndex(dbPath string, provider EmbeddingProvider, chunker Chunker) (*EmbeddingIndex, error) {
	index := &EmbeddingIndex{
		name:              "embedding_index",
		version:           "1.0.0",
		dbPath:            dbPath,
		embeddingProvider: provider,
		chunker:           chunker,
		dimensions:        provider.GetDimensions(),
	}

	if err := index.initDB(); err != nil {
		return nil, fmt.Errorf("failed to initialize embedding index database: %w", err)
	}

	return index, nil
}

// initDB initializes the SQLite database for embedding storage
func (ei *EmbeddingIndex) initDB() error {
	db, err := sql.Open("sqlite", ei.dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	ei.db = db

	// Create embeddings table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS embeddings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			chunk_id TEXT UNIQUE NOT NULL,
			file_path TEXT NOT NULL,
			content TEXT NOT NULL,
			start_line INTEGER NOT NULL,
			end_line INTEGER NOT NULL,
			chunk_type TEXT NOT NULL,
			language TEXT NOT NULL,
			embedding BLOB NOT NULL,
			working_dir TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(chunk_id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create embeddings table: %w", err)
	}

	// Create indexes for performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_embeddings_file_path ON embeddings(file_path)",
		"CREATE INDEX IF NOT EXISTS idx_embeddings_working_dir ON embeddings(working_dir)",
		"CREATE INDEX IF NOT EXISTS idx_embeddings_language ON embeddings(language)",
		"CREATE INDEX IF NOT EXISTS idx_embeddings_chunk_type ON embeddings(chunk_type)",
	}

	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// Name returns the index name
func (ei *EmbeddingIndex) Name() string {
	return ei.name
}

// Version returns the index version
func (ei *EmbeddingIndex) Version() string {
	return ei.version
}

// Update processes file changes and updates the embedding index
func (ei *EmbeddingIndex) Update(ctx context.Context, tag *IndexTag, changes *IndexChanges) error {
	// Handle deleted files
	for _, filePath := range changes.Deleted {
		if err := ei.removeFileEmbeddings(ctx, tag.WorkingDirectory, filePath); err != nil {
			return fmt.Errorf("failed to remove embeddings for %s: %w", filePath, err)
		}
	}

	// Handle added and modified files
	allFiles := append(changes.Added, changes.Modified...)
	for _, filePath := range allFiles {
		if err := ei.indexFile(ctx, tag.WorkingDirectory, filePath); err != nil {
			// Log error but continue with other files
			fmt.Printf("Warning: Failed to index embeddings for %s: %v\n", filePath, err)
			continue
		}
	}

	return nil
}

// Remove removes files from the index
func (ei *EmbeddingIndex) Remove(ctx context.Context, tag *IndexTag, files []string) error {
	for _, filePath := range files {
		if err := ei.removeFileEmbeddings(ctx, tag.WorkingDirectory, filePath); err != nil {
			return fmt.Errorf("failed to remove embeddings for %s: %w", filePath, err)
		}
	}
	return nil
}

// Search performs semantic search using vector similarity
func (ei *EmbeddingIndex) Search(ctx context.Context, query string, opts *SearchOptions) ([]*SearchResult, error) {
	if opts == nil {
		opts = &SearchOptions{MaxResults: 50}
	}

	// Generate query embedding
	queryEmbedding, err := ei.embeddingProvider.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Get all embeddings for similarity calculation
	sqlQuery := `
		SELECT chunk_id, file_path, content, start_line, end_line, 
		       chunk_type, language, embedding
		FROM embeddings 
		WHERE 1=1
	`
	args := []any{}

	// Add file type filtering if specified
	if len(opts.FileTypes) > 0 {
		placeholders := strings.Repeat("?,", len(opts.FileTypes))
		placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma
		sqlQuery += fmt.Sprintf(" AND language IN (%s)", placeholders)
		
		for _, fileType := range opts.FileTypes {
			args = append(args, fileType)
		}
	}

	rows, err := ei.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute embedding search: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var candidates []embeddingCandidate
	for rows.Next() {
		var candidate embeddingCandidate
		var embeddingBlob []byte

		err := rows.Scan(
			&candidate.chunkID, &candidate.filePath, &candidate.content,
			&candidate.startLine, &candidate.endLine,
			&candidate.chunkType, &candidate.language, &embeddingBlob,
		)
		if err != nil {
			continue // Skip invalid rows
		}

		// Deserialize embedding
		if err := json.Unmarshal(embeddingBlob, &candidate.embedding); err != nil {
			continue // Skip invalid embeddings
		}

		// Calculate cosine similarity
		candidate.similarity = ei.cosineSimilarity(queryEmbedding, candidate.embedding)
		candidates = append(candidates, candidate)
	}

	// Sort by similarity (descending)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].similarity > candidates[j].similarity
	})

	// Convert to search results
	var results []*SearchResult
	maxResults := opts.MaxResults
	if maxResults > len(candidates) {
		maxResults = len(candidates)
	}

	for i := 0; i < maxResults; i++ {
		candidate := candidates[i]
		
		result := &SearchResult{
			FilePath:     candidate.filePath,
			LineNumber:   candidate.startLine,
			Content:      candidate.content,
			Snippet:      ei.createSnippet(candidate.content),
			Score:        candidate.similarity * 100, // Convert to percentage
			SearchType:   SearchTypeSemantic,
		}

		// Include full content if requested
		if opts.IncludeContent {
			result.Content = candidate.content
		}

		results = append(results, result)
	}

	return results, nil
}

// GetStats returns index statistics
func (ei *EmbeddingIndex) GetStats() (*IndexStats, error) {
	var totalEmbeddings int64
	err := ei.db.QueryRow("SELECT COUNT(*) FROM embeddings").Scan(&totalEmbeddings)
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding count: %w", err)
	}

	return &IndexStats{
		Name:           ei.name,
		Version:        ei.version,
		TotalDocuments: totalEmbeddings,
		LastUpdated:    time.Now(),
	}, nil
}

// Close closes the database connection
func (ei *EmbeddingIndex) Close() error {
	if ei.db != nil {
		return ei.db.Close()
	}
	return nil
}

// indexFile chunks a file and creates embeddings
func (ei *EmbeddingIndex) indexFile(ctx context.Context, workingDir, filePath string) error {
	// Read file content
	content, err := ei.readFileContent(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Create file content structure
	fileContent := &FileContent{
		Info: &FileInfo{
			Path:     filePath,
			Language: ei.detectLanguage(filePath),
		},
		Content: content,
	}

	// Chunk file into segments
	chunks, err := ei.chunker.ChunkFile(ctx, fileContent)
	if err != nil {
		return fmt.Errorf("failed to chunk file: %w", err)
	}

	// Remove existing embeddings for this file
	if err := ei.removeFileEmbeddings(ctx, workingDir, filePath); err != nil {
		return fmt.Errorf("failed to remove existing embeddings: %w", err)
	}

	// Process chunks in batches
	batchSize := 10
	for i := 0; i < len(chunks); i += batchSize {
		end := i + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}

		batch := chunks[i:end]
		if err := ei.processBatch(ctx, workingDir, batch); err != nil {
			return fmt.Errorf("failed to process batch: %w", err)
		}
	}

	return nil
}

// processBatch processes a batch of chunks
func (ei *EmbeddingIndex) processBatch(ctx context.Context, workingDir string, chunks []*Chunk) error {
	// Extract text content from chunks
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}

	// Generate embeddings in batch
	embeddings, err := ei.embeddingProvider.GenerateBatchEmbeddings(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate batch embeddings: %w", err)
	}

	// Store embeddings
	tx, err := ei.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO embeddings 
		(chunk_id, file_path, content, start_line, end_line, chunk_type, 
		 language, embedding, working_dir)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for i, chunk := range chunks {
		if i >= len(embeddings) {
			break // Safety check
		}

		embeddingBlob, err := json.Marshal(embeddings[i])
		if err != nil {
			fmt.Printf("Warning: Failed to marshal embedding for chunk %s: %v\n", chunk.ID, err)
			continue
		}

		_, err = stmt.ExecContext(ctx,
			chunk.ID, chunk.FilePath, chunk.Content,
			chunk.StartLine, chunk.EndLine, chunk.ChunkType,
			chunk.Language, embeddingBlob, workingDir,
		)
		if err != nil {
			fmt.Printf("Warning: Failed to insert embedding for chunk %s: %v\n", chunk.ID, err)
			continue
		}
	}

	return tx.Commit()
}

// removeFileEmbeddings removes all embeddings for a specific file
func (ei *EmbeddingIndex) removeFileEmbeddings(ctx context.Context, workingDir, filePath string) error {
	_, err := ei.db.ExecContext(ctx,
		"DELETE FROM embeddings WHERE working_dir = ? AND file_path = ?",
		workingDir, filePath,
	)
	return err
}

// readFileContent reads file content from disk
func (ei *EmbeddingIndex) readFileContent(filePath string) ([]byte, error) {
	// This would normally read from filesystem
	// For now, return placeholder to avoid import cycles
	return []byte{}, fmt.Errorf("file reading not implemented - would read %s", filePath)
}

// detectLanguage detects file language from extension
func (ei *EmbeddingIndex) detectLanguage(filePath string) string {
	switch {
	case strings.HasSuffix(filePath, ".go"):
		return "go"
	case strings.HasSuffix(filePath, ".js"):
		return "javascript"
	case strings.HasSuffix(filePath, ".ts"):
		return "typescript"
	case strings.HasSuffix(filePath, ".py"):
		return "python"
	case strings.HasSuffix(filePath, ".java"):
		return "java"
	case strings.HasSuffix(filePath, ".cpp"), strings.HasSuffix(filePath, ".cc"):
		return "cpp"
	case strings.HasSuffix(filePath, ".c"):
		return "c"
	case strings.HasSuffix(filePath, ".rs"):
		return "rust"
	default:
		return "text"
	}
}

// cosineSimilarity calculates cosine similarity between two vectors
func (ei *EmbeddingIndex) cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0.0 || normB == 0.0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// createSnippet creates a shortened snippet from content
func (ei *EmbeddingIndex) createSnippet(content string) string {
	const maxLength = 200
	if len(content) <= maxLength {
		return content
	}

	// Try to find a good breaking point
	snippet := content[:maxLength]
	if lastSpace := strings.LastIndex(snippet, " "); lastSpace > maxLength/2 {
		snippet = snippet[:lastSpace]
	}

	return snippet + "..."
}

// embeddingCandidate represents a candidate for similarity search
type embeddingCandidate struct {
	chunkID    string
	filePath   string
	content    string
	startLine  int
	endLine    int
	chunkType  string
	language   string
	embedding  []float64
	similarity float64
}