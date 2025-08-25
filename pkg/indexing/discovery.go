package indexing

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DefaultFileFilter implements FileFilter with standard programming file patterns
type DefaultFileFilter struct {
	ignorePatterns []string
	maxFileSize    int64 // in bytes
	textExtensions map[string]bool
	codeExtensions map[string]string // extension -> language
}

// NewDefaultFileFilter creates a new file filter with sensible defaults
func NewDefaultFileFilter() *DefaultFileFilter {
	return &DefaultFileFilter{
		ignorePatterns: []string{
			// Version control
			".git", ".svn", ".hg", ".bzr",
			// Dependencies and build artifacts
			"node_modules", "vendor", "target", "build", "dist", "out",
			".gradle", ".mvn", "__pycache__", ".pytest_cache",
			// IDE files
			".vscode", ".idea", "*.swp", "*.swo", "*~",
			// OS files
			".DS_Store", "Thumbs.db",
			// Large files and binaries
			"*.exe", "*.dll", "*.so", "*.dylib", "*.a", "*.lib",
			"*.jpg", "*.jpeg", "*.png", "*.gif", "*.bmp", "*.ico",
			"*.mp3", "*.mp4", "*.avi", "*.mkv", "*.mov",
			"*.zip", "*.tar", "*.gz", "*.rar", "*.7z",
			"*.pdf", "*.doc", "*.docx", "*.xls", "*.xlsx",
		},
		maxFileSize: 10 * 1024 * 1024, // 10MB max file size
		textExtensions: map[string]bool{
			".txt": true, ".md": true, ".rst": true, ".org": true,
			".json": true, ".xml": true, ".yaml": true, ".yml": true,
			".toml": true, ".ini": true, ".cfg": true, ".conf": true,
			".log": true, ".sql": true, ".csv": true, ".tsv": true,
		},
		codeExtensions: map[string]string{
			// Go
			".go": "go",
			// JavaScript/TypeScript
			".js": "javascript", ".jsx": "javascript", ".mjs": "javascript",
			".ts": "typescript", ".tsx": "typescript",
			// Python
			".py": "python", ".pyx": "python", ".pyi": "python",
			// Rust
			".rs": "rust",
			// C/C++
			".c": "c", ".h": "c", ".cpp": "cpp", ".hpp": "cpp", ".cc": "cpp", ".cxx": "cpp",
			// Java/Kotlin
			".java": "java", ".kt": "kotlin", ".kts": "kotlin",
			// C#
			".cs": "csharp",
			// PHP
			".php": "php", ".phtml": "php",
			// Ruby
			".rb": "ruby", ".rake": "ruby",
			// Shell
			".sh": "bash", ".bash": "bash", ".zsh": "zsh", ".fish": "fish",
			// Web
			".html": "html", ".htm": "html", ".css": "css", ".scss": "scss", ".sass": "sass",
			// Other languages
			".lua": "lua", ".vim": "vim", ".r": "r", ".R": "r",
			".swift": "swift", ".scala": "scala", ".clj": "clojure",
			".ex": "elixir", ".exs": "elixir", ".erl": "erlang",
			".hs": "haskell", ".elm": "elm", ".ml": "ocaml",
		},
	}
}

// ShouldIndex determines if a file should be indexed
func (f *DefaultFileFilter) ShouldIndex(path string, info FileInfo) bool {
	// Check file size
	if info.Size > f.maxFileSize {
		return false
	}

	// Check if it's a text or code file
	ext := strings.ToLower(filepath.Ext(path))
	if f.textExtensions[ext] || f.codeExtensions[ext] != "" {
		return true
	}

	// Check if it looks like a text file (no extension but might be code)
	if ext == "" {
		filename := strings.ToLower(filepath.Base(path))
		textFiles := []string{
			"readme", "license", "changelog", "authors", "contributors",
			"makefile", "dockerfile", "jenkinsfile", "vagrantfile",
		}
		for _, textFile := range textFiles {
			if strings.HasPrefix(filename, textFile) {
				return true
			}
		}
	}

	return false
}

// ShouldIgnore checks if a path should be ignored
func (f *DefaultFileFilter) ShouldIgnore(path string) bool {
	// Convert to forward slashes for consistent matching
	normalizedPath := filepath.ToSlash(path)
	
	for _, pattern := range f.ignorePatterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(normalizedPath)); matched {
			return true
		}
		
		// Check if any part of the path matches the pattern
		pathParts := strings.Split(normalizedPath, "/")
		for _, part := range pathParts {
			if matched, _ := filepath.Match(pattern, part); matched {
				return true
			}
		}
		
		// Check for directory patterns (patterns ending with /)
		if strings.HasSuffix(pattern, "/") {
			dirPattern := strings.TrimSuffix(pattern, "/")
			for _, part := range pathParts {
				if part == dirPattern {
					return true
				}
			}
		}
	}
	
	return false
}

// GetIgnorePatterns returns the current ignore patterns
func (f *DefaultFileFilter) GetIgnorePatterns() []string {
	return f.ignorePatterns
}

// AddIgnorePattern adds a new ignore pattern
func (f *DefaultFileFilter) AddIgnorePattern(pattern string) {
	f.ignorePatterns = append(f.ignorePatterns, pattern)
}

// GetLanguage returns the language for a given file extension
func (f *DefaultFileFilter) GetLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if lang, exists := f.codeExtensions[ext]; exists {
		return lang
	}
	return "text"
}

// FSDirectoryWalker implements DirectoryWalker using the file system
type FSDirectoryWalker struct{}

// NewFSDirectoryWalker creates a new filesystem directory walker
func NewFSDirectoryWalker() *FSDirectoryWalker {
	return &FSDirectoryWalker{}
}

// Walk recursively walks through a directory and sends file info through a channel
func (w *FSDirectoryWalker) Walk(ctx context.Context, rootPath string, filter FileFilter) (<-chan FileInfo, error) {
	fileCh := make(chan FileInfo, 100) // Buffered channel for better performance
	
	go func() {
		defer close(fileCh)
		
		err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
			// Check for context cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			
			if err != nil {
				// Log error but continue walking
				fmt.Printf("Error walking path %s: %v\n", path, err)
				return nil
			}
			
			// Skip directories (we only index files)
			if info.IsDir() {
				// Check if this directory should be ignored
				if filter.ShouldIgnore(path) {
					return filepath.SkipDir
				}
				return nil
			}
			
			// Check if file should be ignored
			if filter.ShouldIgnore(path) {
				return nil
			}
			
			// Create FileInfo
			fileInfo := FileInfo{
				Path:     path,
				Size:     info.Size(),
				ModTime:  info.ModTime(),
				IsText:   true, // We'll determine this more accurately later
				Encoding: "utf-8",
			}
			
			// Determine language
			if defaultFilter, ok := filter.(*DefaultFileFilter); ok {
				fileInfo.Language = defaultFilter.GetLanguage(path)
			} else {
				fileInfo.Language = "text"
			}
			
			// Check if file should be indexed
			if filter.ShouldIndex(path, fileInfo) {
				select {
				case fileCh <- fileInfo:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			
			return nil
		})
		
		if err != nil && err != ctx.Err() {
			fmt.Printf("Error during directory walk: %v\n", err)
		}
	}()
	
	return fileCh, nil
}

// GitignoreFilter extends DefaultFileFilter with .gitignore support
type GitignoreFilter struct {
	*DefaultFileFilter
	gitignorePatterns []string
}

// NewGitignoreFilter creates a filter that respects .gitignore files
func NewGitignoreFilter(rootPath string) (*GitignoreFilter, error) {
	filter := &GitignoreFilter{
		DefaultFileFilter: NewDefaultFileFilter(),
	}
	
	// Read .gitignore file if it exists
	gitignorePath := filepath.Join(rootPath, ".gitignore")
	if _, err := os.Stat(gitignorePath); err == nil {
		patterns, err := readGitignoreFile(gitignorePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read .gitignore: %w", err)
		}
		filter.gitignorePatterns = patterns
	}
	
	return filter, nil
}

// ShouldIgnore checks both default patterns and .gitignore patterns
func (f *GitignoreFilter) ShouldIgnore(path string) bool {
	// Check default patterns first
	if f.DefaultFileFilter.ShouldIgnore(path) {
		return true
	}
	
	// Check .gitignore patterns
	normalizedPath := filepath.ToSlash(path)
	for _, pattern := range f.gitignorePatterns {
		if matchesGitignorePattern(normalizedPath, pattern) {
			return true
		}
	}
	
	return false
}

// readGitignoreFile reads and parses a .gitignore file
func readGitignoreFile(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(string(content), "\n")
	var patterns []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	
	return patterns, nil
}

// matchesGitignorePattern checks if a path matches a .gitignore pattern
func matchesGitignorePattern(path, pattern string) bool {
	// Simple pattern matching - this is a basic implementation
	// A full implementation would need to handle all .gitignore rules
	
	// Remove leading slash
	pattern = strings.TrimPrefix(pattern, "/")
	
	// Handle directory patterns
	if strings.HasSuffix(pattern, "/") {
		pattern = strings.TrimSuffix(pattern, "/")
		return strings.Contains(path, pattern+"/")
	}
	
	// Handle wildcards
	if strings.Contains(pattern, "*") {
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
		// Check against full path
		if matched, _ := filepath.Match("*/"+pattern, path); matched {
			return true
		}
	}
	
	// Exact match
	return strings.Contains(path, pattern)
}