package indexing

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// SymbolIndex implements Index interface for code symbol indexing
type SymbolIndex struct {
	name     string
	version  string
	dbPath   string
	db       *sql.DB
	parser   TreeSitterParser
	chunker  Chunker
}

// NewSymbolIndex creates a new symbol index
func NewSymbolIndex(dbPath string, parser TreeSitterParser, chunker Chunker) (*SymbolIndex, error) {
	index := &SymbolIndex{
		name:    "symbol_index",
		version: "1.0.0",
		dbPath:  dbPath,
		parser:  parser,
		chunker: chunker,
	}

	if err := index.initDB(); err != nil {
		return nil, fmt.Errorf("failed to initialize symbol index database: %w", err)
	}

	return index, nil
}

// initDB initializes the SQLite database for symbol storage
func (si *SymbolIndex) initDB() error {
	db, err := sql.Open("sqlite", si.dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	si.db = db

	// Create symbols table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS symbols (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			kind TEXT NOT NULL,
			file_path TEXT NOT NULL,
			start_line INTEGER NOT NULL,
			end_line INTEGER NOT NULL,
			start_column INTEGER NOT NULL,
			end_column INTEGER NOT NULL,
			language TEXT NOT NULL,
			signature TEXT,
			doc_string TEXT,
			parent TEXT,
			children TEXT,
			working_dir TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(name, kind, file_path, start_line)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create symbols table: %w", err)
	}

	// Create indexes for performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_symbols_name ON symbols(name)",
		"CREATE INDEX IF NOT EXISTS idx_symbols_kind ON symbols(kind)",
		"CREATE INDEX IF NOT EXISTS idx_symbols_file_path ON symbols(file_path)",
		"CREATE INDEX IF NOT EXISTS idx_symbols_working_dir ON symbols(working_dir)",
		"CREATE INDEX IF NOT EXISTS idx_symbols_signature ON symbols(signature)",
	}

	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// Name returns the index name
func (si *SymbolIndex) Name() string {
	return si.name
}

// Version returns the index version
func (si *SymbolIndex) Version() string {
	return si.version
}

// Update processes file changes and updates the symbol index
func (si *SymbolIndex) Update(ctx context.Context, tag *IndexTag, changes *IndexChanges) error {
	// Handle deleted files
	for _, filePath := range changes.Deleted {
		if err := si.removeFileSymbols(ctx, tag.WorkingDirectory, filePath); err != nil {
			return fmt.Errorf("failed to remove symbols for %s: %w", filePath, err)
		}
	}

	// Handle added and modified files
	allFiles := append(changes.Added, changes.Modified...)
	for _, filePath := range allFiles {
		if err := si.indexFile(ctx, tag.WorkingDirectory, filePath); err != nil {
			// Log error but continue with other files
			fmt.Printf("Warning: Failed to index symbols for %s: %v\n", filePath, err)
			continue
		}
	}

	return nil
}

// Remove removes files from the index
func (si *SymbolIndex) Remove(ctx context.Context, tag *IndexTag, files []string) error {
	for _, filePath := range files {
		if err := si.removeFileSymbols(ctx, tag.WorkingDirectory, filePath); err != nil {
			return fmt.Errorf("failed to remove symbols for %s: %w", filePath, err)
		}
	}
	return nil
}

// Search searches for symbols matching the query
func (si *SymbolIndex) Search(ctx context.Context, query string, opts *SearchOptions) ([]*SearchResult, error) {
	if opts == nil {
		opts = &SearchOptions{MaxResults: 50}
	}

	// Build SQL query with pattern matching
	sqlQuery := `
		SELECT name, kind, file_path, start_line, end_line, start_column, end_column,
		       language, signature, doc_string, parent, children
		FROM symbols 
		WHERE (name LIKE ? OR signature LIKE ?)
	`
	args := []interface{}{
		"%" + query + "%",
		"%" + query + "%",
	}

	// Add file type filtering if specified
	if len(opts.FileTypes) > 0 {
		placeholders := strings.Repeat("?,", len(opts.FileTypes))
		placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma
		sqlQuery += fmt.Sprintf(" AND language IN (%s)", placeholders)
		
		for _, fileType := range opts.FileTypes {
			args = append(args, fileType)
		}
	}

	sqlQuery += " ORDER BY name LIMIT ?"
	args = append(args, opts.MaxResults)

	rows, err := si.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute symbol search: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var results []*SearchResult
	for rows.Next() {
		var symbol SymbolInfo
		var children sql.NullString

		err := rows.Scan(
			&symbol.Name, &symbol.Kind, &symbol.FilePath,
			&symbol.StartLine, &symbol.EndLine,
			&symbol.StartColumn, &symbol.EndColumn,
			&symbol.Language, &symbol.Signature,
			&symbol.DocString, &symbol.Parent, &children,
		)
		if err != nil {
			continue // Skip invalid rows
		}

		// Parse children JSON if present
		if children.Valid && children.String != "" {
			if err := json.Unmarshal([]byte(children.String), &symbol.Children); err != nil {
				symbol.Children = []string{} // Default to empty if parse fails
			}
		}

		// Create search result
		result := &SearchResult{
			FilePath:   symbol.FilePath,
			LineNumber: symbol.StartLine,
			ColumnNumber: symbol.StartColumn,
			Snippet:    fmt.Sprintf("%s %s", symbol.Kind, symbol.Name),
			Score:      si.calculateRelevanceScore(query, &symbol),
			SearchType: SearchTypeSymbol,
			SymbolInfo: &symbol,
		}

		// Include content if requested
		if opts.IncludeContent && symbol.Signature != "" {
			result.Content = symbol.Signature
		}

		results = append(results, result)
	}

	return results, nil
}

// GetStats returns index statistics
func (si *SymbolIndex) GetStats() (*IndexStats, error) {
	var totalSymbols int64
	err := si.db.QueryRow("SELECT COUNT(*) FROM symbols").Scan(&totalSymbols)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbol count: %w", err)
	}

	return &IndexStats{
		Name:           si.name,
		Version:        si.version,
		TotalDocuments: totalSymbols,
		LastUpdated:    time.Now(),
	}, nil
}

// Close closes the database connection
func (si *SymbolIndex) Close() error {
	if si.db != nil {
		return si.db.Close()
	}
	return nil
}

// indexFile parses a file and extracts symbols
func (si *SymbolIndex) indexFile(ctx context.Context, workingDir, filePath string) error {
	// Read file content
	content, err := si.readFileContent(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse file with tree-sitter
	parsedFile, err := si.parser.ParseFile(ctx, filePath, content)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// Remove existing symbols for this file
	if err := si.removeFileSymbols(ctx, workingDir, filePath); err != nil {
		return fmt.Errorf("failed to remove existing symbols: %w", err)
	}

	// Insert new symbols
	tx, err := si.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO symbols 
		(name, kind, file_path, start_line, end_line, start_column, end_column,
		 language, signature, doc_string, parent, children, working_dir)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	for _, symbol := range parsedFile.Symbols {
		childrenJSON, _ := json.Marshal(symbol.Children)
		
		_, err := stmt.ExecContext(ctx,
			symbol.Name, symbol.Kind, symbol.FilePath,
			symbol.StartLine, symbol.EndLine,
			symbol.StartColumn, symbol.EndColumn,
			symbol.Language, symbol.Signature,
			symbol.DocString, symbol.Parent,
			string(childrenJSON), workingDir,
		)
		if err != nil {
			fmt.Printf("Warning: Failed to insert symbol %s: %v\n", symbol.Name, err)
			continue
		}
	}

	return tx.Commit()
}

// removeFileSymbols removes all symbols for a specific file
func (si *SymbolIndex) removeFileSymbols(ctx context.Context, workingDir, filePath string) error {
	_, err := si.db.ExecContext(ctx,
		"DELETE FROM symbols WHERE working_dir = ? AND file_path = ?",
		workingDir, filePath,
	)
	return err
}

// readFileContent reads file content from disk
func (si *SymbolIndex) readFileContent(filePath string) ([]byte, error) {
	// This would normally read from filesystem
	// For now, return placeholder to avoid import cycles
	return []byte{}, fmt.Errorf("file reading not implemented - would read %s", filePath)
}

// calculateRelevanceScore calculates relevance score for a symbol match
func (si *SymbolIndex) calculateRelevanceScore(query string, symbol *SymbolInfo) float64 {
	score := 0.0
	queryLower := strings.ToLower(query)
	nameLower := strings.ToLower(symbol.Name)
	
	// Exact match gets highest score
	if nameLower == queryLower {
		score += 100.0
	} else if strings.HasPrefix(nameLower, queryLower) {
		score += 80.0
	} else if strings.Contains(nameLower, queryLower) {
		score += 60.0
	}
	
	// Signature match
	if symbol.Signature != "" {
		sigLower := strings.ToLower(symbol.Signature)
		if strings.Contains(sigLower, queryLower) {
			score += 40.0
		}
	}
	
	// Symbol kind weighting
	switch symbol.Kind {
	case SymbolKindFunction, SymbolKindMethod:
		score += 10.0
	case SymbolKindClass, SymbolKindStruct, SymbolKindInterface:
		score += 8.0
	case SymbolKindVariable, SymbolKindConstant:
		score += 5.0
	}
	
	return score
}