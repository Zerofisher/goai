package search

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// FileInfo represents cached information about a file.
type FileInfo struct {
	Path         string
	Size         int64
	ModifiedTime time.Time
	Language     string
}

// IndexCache provides a simple caching mechanism for search operations.
type IndexCache struct {
	mu          sync.RWMutex
	files       map[string]*FileInfo
	lastUpdated time.Time
	ttl         time.Duration
}

// NewIndexCache creates a new index cache.
func NewIndexCache(ttl time.Duration) *IndexCache {
	if ttl <= 0 {
		ttl = 15 * time.Minute // Default TTL
	}
	return &IndexCache{
		files: make(map[string]*FileInfo),
		ttl:   ttl,
	}
}

// IsExpired checks if the cache has expired.
func (c *IndexCache) IsExpired() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Since(c.lastUpdated) > c.ttl
}

// Clear clears the cache.
func (c *IndexCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.files = make(map[string]*FileInfo)
	c.lastUpdated = time.Time{}
}

// Get retrieves file info from cache.
func (c *IndexCache) Get(path string) (*FileInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.IsExpired() {
		return nil, false
	}

	info, exists := c.files[path]
	return info, exists
}

// Set adds or updates file info in cache.
func (c *IndexCache) Set(path string, info *FileInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.files[path] = info
	c.lastUpdated = time.Now()
}

// Indexer provides indexing and caching functionality for the search tool.
type Indexer struct {
	workDir     string
	cache       *IndexCache
	ignorePatterns []string
	mu          sync.RWMutex
}

// NewIndexer creates a new indexer.
func NewIndexer(workDir string) *Indexer {
	return &Indexer{
		workDir: workDir,
		cache:   NewIndexCache(15 * time.Minute),
		ignorePatterns: []string{
			".git",
			".svn",
			".hg",
			"node_modules",
			"vendor",
			"__pycache__",
			"*.pyc",
			"*.pyo",
			"*.so",
			"*.dylib",
			"*.dll",
			"*.exe",
			"*.o",
			"*.a",
			"*.class",
			"*.jar",
		},
	}
}

// SetIgnorePatterns sets patterns for files/directories to ignore.
func (idx *Indexer) SetIgnorePatterns(patterns []string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	idx.ignorePatterns = patterns
}

// ShouldIgnore checks if a path should be ignored.
func (idx *Indexer) ShouldIgnore(path string) bool {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	base := filepath.Base(path)

	for _, pattern := range idx.ignorePatterns {
		// Check exact match
		if base == pattern {
			return true
		}

		// Check glob pattern
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}

		// Check if path contains the pattern (for directories)
		// Handle both forward and backward slashes
		if strings.Contains(path, "/"+pattern+"/") ||
		   strings.Contains(path, "/"+pattern) ||
		   strings.Contains(path, pattern+"/") {
			return true
		}
	}

	return false
}

// RefreshIndex updates the file cache by scanning the work directory.
func (idx *Indexer) RefreshIndex() error {
	// Clear expired cache
	if idx.cache.IsExpired() {
		idx.cache.Clear()
	}

	return filepath.Walk(idx.workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		// Skip ignored paths
		if idx.ShouldIgnore(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories themselves (we only index files)
		if info.IsDir() {
			return nil
		}

		// Skip very large files (> 10MB)
		if info.Size() > 10*1024*1024 {
			return nil
		}

		// Create file info
		fileInfo := &FileInfo{
			Path:         path,
			Size:         info.Size(),
			ModifiedTime: info.ModTime(),
			Language:     detectLanguage(path),
		}

		// Cache the file info
		relPath, _ := filepath.Rel(idx.workDir, path)
		idx.cache.Set(relPath, fileInfo)

		return nil
	})
}

// GetRelevantFiles returns files relevant to a query.
func (idx *Indexer) GetRelevantFiles(query string, filePattern string) ([]string, error) {
	// Refresh index if expired
	if idx.cache.IsExpired() {
		if err := idx.RefreshIndex(); err != nil {
			return nil, err
		}
	}

	var relevantFiles []string

	idx.cache.mu.RLock()
	defer idx.cache.mu.RUnlock()

	for relPath, fileInfo := range idx.cache.files {
		// Check file pattern if specified
		if filePattern != "" {
			matched, _ := filepath.Match(filePattern, filepath.Base(fileInfo.Path))
			if !matched {
				continue
			}
		}

		// Simple relevance check - file name contains query terms
		if query != "" {
			queryLower := strings.ToLower(query)
			pathLower := strings.ToLower(relPath)

			// Check if file path contains any query terms
			relevant := false
			for _, term := range strings.Fields(queryLower) {
				if strings.Contains(pathLower, term) {
					relevant = true
					break
				}
			}

			if !relevant {
				continue
			}
		}

		relevantFiles = append(relevantFiles, fileInfo.Path)
	}

	return relevantFiles, nil
}

// QuickSearch performs a quick search using the cache.
func (idx *Indexer) QuickSearch(pattern string, filePattern string) ([]string, error) {
	// Refresh index if expired
	if idx.cache.IsExpired() {
		if err := idx.RefreshIndex(); err != nil {
			return nil, err
		}
	}

	// Get relevant files from cache
	files, err := idx.GetRelevantFiles("", filePattern)  // Empty query to get all files matching the pattern
	if err != nil {
		return nil, err
	}

	var results []string

	// Search in each file
	for _, file := range files {
		matches, err := idx.searchInFile(file, pattern)
		if err != nil {
			continue // Skip files with errors
		}

		if len(matches) > 0 {
			relPath, _ := filepath.Rel(idx.workDir, file)
			results = append(results, fmt.Sprintf("%s: %d matches", relPath, len(matches)))
		}
	}

	return results, nil
}

// searchInFile searches for a pattern in a single file.
func (idx *Indexer) searchInFile(filepath string, pattern string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close() // Non-critical error, can be ignored
	}()

	var matches []string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Support both case-sensitive and case-insensitive search
	patternLower := strings.ToLower(pattern)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check for pattern match (case-insensitive for now)
		if strings.Contains(strings.ToLower(line), patternLower) ||
		   strings.Contains(line, pattern) {  // Also check exact match
			matches = append(matches, fmt.Sprintf("%d: %s", lineNum, line))

			// Limit matches per file
			if len(matches) >= 10 {
				break
			}
		}
	}

	return matches, scanner.Err()
}

// GetFileStats returns statistics about indexed files.
func (idx *Indexer) GetFileStats() map[string]int {
	stats := make(map[string]int)

	idx.cache.mu.RLock()
	defer idx.cache.mu.RUnlock()

	for _, fileInfo := range idx.cache.files {
		lang := fileInfo.Language
		if lang == "" {
			lang = "unknown"
		}
		stats[lang]++
	}

	stats["total"] = len(idx.cache.files)

	return stats
}

// detectLanguage detects the programming language based on file extension.
func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".go":
		return "Go"
	case ".py":
		return "Python"
	case ".js", ".jsx":
		return "JavaScript"
	case ".ts", ".tsx":
		return "TypeScript"
	case ".java":
		return "Java"
	case ".c":
		return "C"
	case ".cpp", ".cc", ".cxx":
		return "C++"
	case ".cs":
		return "C#"
	case ".rb":
		return "Ruby"
	case ".php":
		return "PHP"
	case ".rs":
		return "Rust"
	case ".swift":
		return "Swift"
	case ".kt":
		return "Kotlin"
	case ".scala":
		return "Scala"
	case ".sh", ".bash":
		return "Shell"
	case ".sql":
		return "SQL"
	case ".html", ".htm":
		return "HTML"
	case ".css", ".scss", ".sass":
		return "CSS"
	case ".json":
		return "JSON"
	case ".xml":
		return "XML"
	case ".yaml", ".yml":
		return "YAML"
	case ".md", ".markdown":
		return "Markdown"
	case ".txt":
		return "Text"
	default:
		return ""
	}
}

// FileTypeFilter returns files of specific programming languages.
func (idx *Indexer) FileTypeFilter(languages []string) ([]string, error) {
	// Refresh index if expired
	if idx.cache.IsExpired() {
		if err := idx.RefreshIndex(); err != nil {
			return nil, err
		}
	}

	var files []string
	langSet := make(map[string]bool)

	for _, lang := range languages {
		langSet[strings.ToLower(lang)] = true
	}

	idx.cache.mu.RLock()
	defer idx.cache.mu.RUnlock()

	for _, fileInfo := range idx.cache.files {
		if langSet[strings.ToLower(fileInfo.Language)] {
			files = append(files, fileInfo.Path)
		}
	}

	return files, nil
}

// GetRecentFiles returns files modified within the specified duration.
func (idx *Indexer) GetRecentFiles(since time.Duration) ([]string, error) {
	// Refresh index if expired
	if idx.cache.IsExpired() {
		if err := idx.RefreshIndex(); err != nil {
			return nil, err
		}
	}

	var files []string
	cutoff := time.Now().Add(-since)

	idx.cache.mu.RLock()
	defer idx.cache.mu.RUnlock()

	for _, fileInfo := range idx.cache.files {
		if fileInfo.ModifiedTime.After(cutoff) {
			files = append(files, fileInfo.Path)
		}
	}

	return files, nil
}