package indexing

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

// IndexManager implements CodebaseIndexer and manages multiple index types
type IndexManager struct {
	mu             sync.RWMutex
	indexes        map[string]Index // index name -> index implementation
	indexStatus    map[string]*IndexStatus // working directory -> status
	walker         DirectoryWalker
	filter         FileFilter
	chunker        Chunker
	buildingMutex  sync.Mutex // Prevents concurrent index building
}

// NewIndexManager creates a new index manager
func NewIndexManager() *IndexManager {
	return &IndexManager{
		indexes:     make(map[string]Index),
		indexStatus: make(map[string]*IndexStatus),
		walker:      NewFSDirectoryWalker(),
		filter:      NewDefaultFileFilter(),
		chunker:     NewDefaultChunker(),
	}
}

// RegisterIndex registers an index implementation
func (im *IndexManager) RegisterIndex(index Index) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.indexes[index.Name()] = index
}

// SetFileFilter sets the file filter to use
func (im *IndexManager) SetFileFilter(filter FileFilter) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.filter = filter
}

// SetChunker sets the chunker to use
func (im *IndexManager) SetChunker(chunker Chunker) {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.chunker = chunker
}

// BuildIndex builds indexes for the specified working directory
func (im *IndexManager) BuildIndex(ctx context.Context, workdir string) error {
	im.buildingMutex.Lock()
	defer im.buildingMutex.Unlock()
	
	// Initialize status
	status := &IndexStatus{
		WorkingDirectory: workdir,
		IsBuilding:       true,
		IsReady:         false,
		LastUpdated:     time.Now(),
		IndexVersions:   make(map[string]string),
	}
	
	im.mu.Lock()
	im.indexStatus[workdir] = status
	im.mu.Unlock()
	
	startTime := time.Now()
	
	// Create index tag
	tag := &IndexTag{
		WorkingDirectory: workdir,
		Timestamp:       startTime,
	}
	
	// Walk directory and discover files
	fileCh, err := im.walker.Walk(ctx, workdir, im.filter)
	if err != nil {
		im.updateStatus(workdir, func(s *IndexStatus) {
			s.IsBuilding = false
			s.Errors = append(s.Errors, fmt.Sprintf("Failed to walk directory: %v", err))
		})
		return fmt.Errorf("failed to walk directory: %w", err)
	}
	
	var totalFiles int
	var processedFiles int
	var totalSize int64
	
	// Process files and build changes
	changes := &IndexChanges{
		Added: make([]string, 0),
	}
	
	for fileInfo := range fileCh {
		select {
		case <-ctx.Done():
			im.updateStatus(workdir, func(s *IndexStatus) {
				s.IsBuilding = false
				s.Errors = append(s.Errors, "Index building was cancelled")
			})
			return ctx.Err()
		default:
		}
		
		totalFiles++
		totalSize += fileInfo.Size
		changes.Added = append(changes.Added, fileInfo.Path)
		
		// Update progress
		if totalFiles%100 == 0 {
			im.updateStatus(workdir, func(s *IndexStatus) {
				s.TotalFiles = totalFiles
				s.TotalSize = totalSize
			})
		}
	}
	
	// Update all registered indexes
	im.mu.RLock()
	indexes := make([]Index, 0, len(im.indexes))
	for _, index := range im.indexes {
		indexes = append(indexes, index)
	}
	im.mu.RUnlock()
	
	// Update indexes in parallel for better performance
	var wg sync.WaitGroup
	errCh := make(chan error, len(indexes))
	
	for _, index := range indexes {
		wg.Add(1)
		go func(idx Index) {
			defer wg.Done()
			if err := idx.Update(ctx, tag, changes); err != nil {
				errCh <- fmt.Errorf("failed to update %s index: %w", idx.Name(), err)
			} else {
				// Update version info
				im.updateStatus(workdir, func(s *IndexStatus) {
					s.IndexVersions[idx.Name()] = idx.Version()
				})
				processedFiles++
			}
		}(index)
	}
	
	wg.Wait()
	close(errCh)
	
	// Collect any errors
	var buildErrors []string
	for err := range errCh {
		buildErrors = append(buildErrors, err.Error())
	}
	
	// Update final status
	im.updateStatus(workdir, func(s *IndexStatus) {
		s.IsBuilding = false
		s.IsReady = len(buildErrors) == 0
		s.TotalFiles = totalFiles
		s.IndexedFiles = processedFiles
		s.TotalSize = totalSize
		s.LastUpdated = time.Now()
		if len(buildErrors) > 0 {
			s.Errors = append(s.Errors, buildErrors...)
		}
	})
	
	if len(buildErrors) > 0 {
		return fmt.Errorf("index building completed with errors: %v", buildErrors)
	}
	
	return nil
}

// RefreshIndex updates indexes for specific paths
func (im *IndexManager) RefreshIndex(ctx context.Context, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	
	// Determine working directory from first path
	workdir := filepath.Dir(paths[0])
	for len(workdir) > 1 {
		if im.IsIndexReady(workdir) {
			break
		}
		workdir = filepath.Dir(workdir)
	}
	
	// Create index tag
	tag := &IndexTag{
		WorkingDirectory: workdir,
		Timestamp:       time.Now(),
	}
	
	// Create changes
	changes := &IndexChanges{
		Modified: paths,
	}
	
	// Update all indexes
	im.mu.RLock()
	defer im.mu.RUnlock()
	
	var errors []string
	for name, index := range im.indexes {
		if err := index.Update(ctx, tag, changes); err != nil {
			errors = append(errors, fmt.Sprintf("failed to update %s index: %v", name, err))
		}
	}
	
	// Update status
	im.updateStatus(workdir, func(s *IndexStatus) {
		s.LastUpdated = time.Now()
		if len(errors) > 0 {
			s.Errors = append(s.Errors, errors...)
		}
	})
	
	if len(errors) > 0 {
		return fmt.Errorf("refresh completed with errors: %v", errors)
	}
	
	return nil
}

// ClearIndex clears all indexes for a working directory
func (im *IndexManager) ClearIndex(ctx context.Context, workdir string) error {
	im.mu.Lock()
	defer im.mu.Unlock()
	
	var errors []string
	tag := &IndexTag{
		WorkingDirectory: workdir,
		Timestamp:       time.Now(),
	}
	
	for name, index := range im.indexes {
		// Get all files to remove (this is a simplified approach)
		changes := &IndexChanges{
			Deleted: []string{"*"}, // Remove all files
		}
		
		if err := index.Remove(ctx, tag, changes.Deleted); err != nil {
			errors = append(errors, fmt.Sprintf("failed to clear %s index: %v", name, err))
		}
	}
	
	// Remove status
	delete(im.indexStatus, workdir)
	
	if len(errors) > 0 {
		return fmt.Errorf("clear completed with errors: %v", errors)
	}
	
	return nil
}

// GetIndexStatus returns the current index status
func (im *IndexManager) GetIndexStatus(workdir string) (*IndexStatus, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()
	
	status, exists := im.indexStatus[workdir]
	if !exists {
		return nil, fmt.Errorf("no index found for directory: %s", workdir)
	}
	
	// Return a copy to avoid race conditions
	statusCopy := *status
	return &statusCopy, nil
}

// IsIndexReady checks if indexes are ready for a working directory
func (im *IndexManager) IsIndexReady(workdir string) bool {
	im.mu.RLock()
	defer im.mu.RUnlock()
	
	status, exists := im.indexStatus[workdir]
	return exists && status.IsReady && !status.IsBuilding
}

// WaitForIndex waits for indexing to complete
func (im *IndexManager) WaitForIndex(ctx context.Context, workdir string) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if im.IsIndexReady(workdir) {
				return nil
			}
			
			// Check if there was an error
			status, _ := im.GetIndexStatus(workdir)
			if status != nil && !status.IsBuilding && len(status.Errors) > 0 {
				return fmt.Errorf("indexing failed: %v", status.Errors)
			}
		}
	}
}

// Search performs a search across all indexes
func (im *IndexManager) Search(ctx context.Context, req *SearchRequest) (*SearchResult, error) {
	if !im.IsIndexReady(req.WorkingDir) {
		return nil, fmt.Errorf("index not ready for directory: %s", req.WorkingDir)
	}
	
	// For now, return a simple implementation
	// This will be enhanced when we implement specific index types
	return &SearchResult{
		FilePath:   "example.go",
		Content:    "// Example search result",
		Snippet:    "Example search result",
		Score:      0.8,
		SearchType: SearchTypeFullText,
	}, nil
}

// GetContextItems returns context items for reasoning
func (im *IndexManager) GetContextItems(ctx context.Context, query string, opts *ContextOptions) ([]*ContextItem, error) {
	// For now, return a simple implementation
	// This will be enhanced when we implement specific retrievers
	return []*ContextItem{
		{
			FilePath:    "example.go",
			Content:     "// Example context item",
			StartLine:   1,
			EndLine:     5,
			ContextType: ContextTypeFunction,
			Relevance:   0.9,
		},
	}, nil
}

// updateStatus safely updates the index status
func (im *IndexManager) updateStatus(workdir string, updateFunc func(*IndexStatus)) {
	im.mu.Lock()
	defer im.mu.Unlock()
	
	if status, exists := im.indexStatus[workdir]; exists {
		updateFunc(status)
	}
}

// Close closes all indexes and cleans up resources
func (im *IndexManager) Close() error {
	im.mu.Lock()
	defer im.mu.Unlock()
	
	var errors []string
	for name, index := range im.indexes {
		if err := index.Close(); err != nil {
			errors = append(errors, fmt.Sprintf("failed to close %s index: %v", name, err))
		}
	}
	
	// Clear maps
	im.indexes = make(map[string]Index)
	im.indexStatus = make(map[string]*IndexStatus)
	
	if len(errors) > 0 {
		return fmt.Errorf("close completed with errors: %v", errors)
	}
	
	return nil
}