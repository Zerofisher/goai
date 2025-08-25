package context

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Zerofisher/goai/pkg/types"
)

// FileWatcher watches for file system changes
type FileWatcher struct {
	rootPath    string
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	isRunning   bool
	callback    func(*types.FileChangeEvent)
	pollInterval time.Duration
	lastScan    map[string]time.Time
}

// NewFileWatcher creates a new file watcher
func NewFileWatcher(rootPath string) (*FileWatcher, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &FileWatcher{
		rootPath:     rootPath,
		ctx:          ctx,
		cancel:       cancel,
		pollInterval: 2 * time.Second, // Poll every 2 seconds
		lastScan:     make(map[string]time.Time),
	}, nil
}

// Start starts watching for file changes
func (fw *FileWatcher) Start(callback func(*types.FileChangeEvent)) error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.isRunning {
		return fmt.Errorf("file watcher is already running")
	}

	fw.callback = callback
	fw.isRunning = true

	// Perform initial scan to populate lastScan
	fw.performScan(false)

	// Start the polling goroutine
	go fw.pollForChanges()

	return nil
}

// Stop stops the file watcher
func (fw *FileWatcher) Stop() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if !fw.isRunning {
		return nil
	}

	fw.cancel()
	fw.isRunning = false
	return nil
}

// pollForChanges polls the file system for changes
func (fw *FileWatcher) pollForChanges() {
	ticker := time.NewTicker(fw.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-fw.ctx.Done():
			return
		case <-ticker.C:
			fw.performScan(true)
		}
	}
}

// performScan scans the directory for changes
func (fw *FileWatcher) performScan(notifyChanges bool) {
	currentScan := make(map[string]time.Time)

	err := filepath.Walk(fw.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		// Skip hidden files and directories
		if len(filepath.Base(path)) > 0 && filepath.Base(path)[0] == '.' {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip vendor and node_modules directories
		relPath, _ := filepath.Rel(fw.rootPath, path)
		if filepath.Dir(relPath) == "vendor" || filepath.Dir(relPath) == "node_modules" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Only watch files, not directories
		if info.IsDir() {
			return nil
		}

		currentScan[path] = info.ModTime()

		if notifyChanges && fw.callback != nil {
			fw.checkForChange(path, info.ModTime())
		}

		return nil
	})

	if err == nil {
		fw.mu.Lock()
		// Check for deleted files
		if notifyChanges && fw.callback != nil {
			for oldPath := range fw.lastScan {
				if _, exists := currentScan[oldPath]; !exists {
					fw.notifyChange(oldPath, "deleted")
				}
			}
		}
		fw.lastScan = currentScan
		fw.mu.Unlock()
	}
}

// checkForChange checks if a file has changed
func (fw *FileWatcher) checkForChange(path string, modTime time.Time) {
	fw.mu.RLock()
	lastModTime, exists := fw.lastScan[path]
	fw.mu.RUnlock()

	if !exists {
		// New file
		fw.notifyChange(path, "created")
	} else if modTime.After(lastModTime) {
		// Modified file
		fw.notifyChange(path, "modified")
	}
}

// notifyChange notifies about a file change
func (fw *FileWatcher) notifyChange(path, eventType string) {
	if fw.callback == nil {
		return
	}

	relPath, err := filepath.Rel(fw.rootPath, path)
	if err != nil {
		relPath = path
	}

	event := &types.FileChangeEvent{
		FilePath:  relPath,
		EventType: eventType,
		Timestamp: time.Now(),
	}

	// Run callback in goroutine to avoid blocking
	go fw.callback(event)
}

// IsRunning returns true if the watcher is running
func (fw *FileWatcher) IsRunning() bool {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	return fw.isRunning
}

// SetPollInterval sets the polling interval
func (fw *FileWatcher) SetPollInterval(interval time.Duration) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.pollInterval = interval
}

// GetWatchedFiles returns a list of currently watched files
func (fw *FileWatcher) GetWatchedFiles() []string {
	fw.mu.RLock()
	defer fw.mu.RUnlock()

	files := make([]string, 0, len(fw.lastScan))
	for path := range fw.lastScan {
		relPath, err := filepath.Rel(fw.rootPath, path)
		if err != nil {
			relPath = path
		}
		files = append(files, relPath)
	}

	return files
}

// GetFileCount returns the number of watched files
func (fw *FileWatcher) GetFileCount() int {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	return len(fw.lastScan)
}