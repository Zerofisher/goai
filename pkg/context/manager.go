package context

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/Zerofisher/goai/pkg/indexing"
	"github.com/Zerofisher/goai/pkg/types"
)

// ContextManager implements the ContextManager interface
type ContextManager struct {
	workdir        string
	gitRepo        *git.Repository
	configPath     string
	structAnalyzer *ProjectStructureAnalyzer
	depAnalyzer    *DependencyAnalyzer
	watcher        *FileWatcher
	indexManager   indexing.CodebaseIndexer
}

// NewContextManager creates a new context manager
func NewContextManager(workdir string) (*ContextManager, error) {
	cm := &ContextManager{
		workdir:    workdir,
		configPath: filepath.Join(workdir, "GOAI.md"),
	}

	// Initialize Git repository if it exists
	if err := cm.initGitRepo(); err != nil {
		// Git repo is optional, so we log the error but continue
		fmt.Printf("Warning: Could not initialize Git repo: %v\n", err)
	}

	// Initialize analyzers
	cm.structAnalyzer = NewProjectStructureAnalyzer(workdir)
	cm.depAnalyzer = NewDependencyAnalyzer(workdir)

	// Initialize index manager (optional - may fail in test environments)
	dbPath := filepath.Join(workdir, ".goai", "index.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err == nil {
		indexManager, err := indexing.NewIndexManagerWithDefaults(dbPath)
		if err != nil {
			fmt.Printf("Warning: Failed to initialize index manager: %v\n", err)
		} else {
			cm.indexManager = indexManager
		}
	} else {
		fmt.Printf("Warning: Failed to create index directory: %v\n", err)
	}

	return cm, nil
}

// initGitRepo initializes the Git repository
func (cm *ContextManager) initGitRepo() error {
	repo, err := git.PlainOpen(cm.workdir)
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	cm.gitRepo = repo
	return nil
}

// BuildProjectContext builds a complete project context
func (cm *ContextManager) BuildProjectContext(workdir string) (*types.ProjectContext, error) {
	if workdir != "" {
		cm.workdir = workdir
	}

	context := &types.ProjectContext{
		WorkingDirectory: cm.workdir,
		LoadedAt:         time.Now(),
	}

	// Load configuration
	config, err := cm.LoadConfiguration(cm.configPath)
	if err != nil {
		// Config is optional, create a default one
		config = &types.GOAIConfig{
			ProjectName: filepath.Base(cm.workdir),
			Language:    cm.detectLanguage(),
		}
	}
	context.ProjectConfig = config

	// Analyze project structure
	structure, err := cm.structAnalyzer.AnalyzeStructure()
	if err != nil {
		return nil, fmt.Errorf("failed to analyze project structure: %w", err)
	}
	context.ProjectStructure = structure

	// Analyze dependencies
	deps, err := cm.depAnalyzer.AnalyzeDependencies()
	if err != nil {
		// Log error but don't fail the entire context building
		fmt.Printf("Warning: Could not analyze dependencies: %v\n", err)
		context.Dependencies = []*types.Dependency{}
	} else {
		context.Dependencies = deps
	}

	// Load Git information if available
	if cm.gitRepo != nil {
		gitInfo, err := cm.loadGitInfo()
		if err == nil {
			context.GitInfo = gitInfo
		}
	}

	// Load recent changes
	changes, err := cm.GetRecentChanges(time.Now().Add(-24 * time.Hour))
	if err == nil {
		context.RecentChanges = changes
	}

	return context, nil
}

// LoadConfiguration loads GOAI configuration from file
func (cm *ContextManager) LoadConfiguration(configPath string) (*types.GOAIConfig, error) {
	if configPath == "" {
		configPath = cm.configPath
	}

	parser := NewGOAIConfigParser()
	return parser.ParseFile(configPath)
}

// WatchFileChanges sets up file watching for real-time updates
func (cm *ContextManager) WatchFileChanges(callback func(*types.FileChangeEvent)) error {
	if cm.watcher != nil {
		_ = cm.watcher.Stop()
	}

	watcher, err := NewFileWatcher(cm.workdir)
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	cm.watcher = watcher
	return watcher.Start(callback)
}

// GetRecentChanges retrieves recent Git changes since a specific time
func (cm *ContextManager) GetRecentChanges(since time.Time) ([]*types.GitChange, error) {
	if cm.gitRepo == nil {
		return []*types.GitChange{}, nil
	}

	ref, err := cm.gitRepo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	commits, err := cm.gitRepo.Log(&git.LogOptions{
		From:  ref.Hash(),
		Since: &since,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	var changes []*types.GitChange
	err = commits.ForEach(func(commit *object.Commit) error {
		// Get file stats for this commit
		stats, err := commit.Stats()
		if err != nil {
			return err
		}

		for _, stat := range stats {
			change := &types.GitChange{
				CommitHash: commit.Hash.String(),
				Author:     commit.Author.Name,
				Message:    commit.Message,
				Timestamp:  commit.Author.When,
				FilePath:   stat.Name,
				Additions:  stat.Addition,
				Deletions:  stat.Deletion,
			}
			changes = append(changes, change)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return changes, nil
}

// loadGitInfo loads basic Git repository information
func (cm *ContextManager) loadGitInfo() (*types.GitInfo, error) {
	ref, err := cm.gitRepo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	commit, err := cm.gitRepo.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit object: %w", err)
	}

	// Get remote URLs
	remotes, err := cm.gitRepo.Remotes()
	if err != nil {
		return nil, fmt.Errorf("failed to get remotes: %w", err)
	}

	var remoteURLs []string
	for _, remote := range remotes {
		remoteURLs = append(remoteURLs, remote.Config().URLs...)
	}

	// Get current branch name
	branchName := "HEAD"
	if ref.Name().IsBranch() {
		branchName = ref.Name().Short()
	}

	return &types.GitInfo{
		CurrentBranch:    branchName,
		CurrentCommit:    commit.Hash.String(),
		RemoteURLs:       remoteURLs,
		LastCommitTime:   commit.Author.When,
		LastCommitMsg:    commit.Message,
		LastCommitAuthor: commit.Author.Name,
	}, nil
}

// detectLanguage attempts to detect the primary language of the project
func (cm *ContextManager) detectLanguage() string {
	// Check for Go files
	goFiles, _ := filepath.Glob(filepath.Join(cm.workdir, "*.go"))
	if len(goFiles) > 0 {
		return "go"
	}

	// Check for go.mod
	if _, err := os.Stat(filepath.Join(cm.workdir, "go.mod")); err == nil {
		return "go"
	}

	// Check for package.json
	if _, err := os.Stat(filepath.Join(cm.workdir, "package.json")); err == nil {
		// Check for TypeScript
		tsFiles, _ := filepath.Glob(filepath.Join(cm.workdir, "*.ts"))
		if len(tsFiles) > 0 {
			return "typescript"
		}
		return "javascript"
	}

	// Check for Python files
	pyFiles, _ := filepath.Glob(filepath.Join(cm.workdir, "*.py"))
	if len(pyFiles) > 0 {
		return "python"
	}

	// Check for requirements.txt or setup.py
	if _, err := os.Stat(filepath.Join(cm.workdir, "requirements.txt")); err == nil {
		return "python"
	}
	if _, err := os.Stat(filepath.Join(cm.workdir, "setup.py")); err == nil {
		return "python"
	}

	// Check for Java files
	javaFiles, _ := filepath.Glob(filepath.Join(cm.workdir, "*.java"))
	if len(javaFiles) > 0 {
		return "java"
	}

	// Check for Rust files
	if _, err := os.Stat(filepath.Join(cm.workdir, "Cargo.toml")); err == nil {
		return "rust"
	}

	return "unknown"
}

// Stop gracefully stops the context manager
func (cm *ContextManager) Stop() error {
	if cm.watcher != nil {
		return cm.watcher.Stop()
	}
	return nil
}

// GetWorkingDirectory returns the current working directory
func (cm *ContextManager) GetWorkingDirectory() string {
	return cm.workdir
}

// IsGitRepository returns true if the working directory is a Git repository
func (cm *ContextManager) IsGitRepository() bool {
	return cm.gitRepo != nil
}

// GetGitStatus returns the current Git status
func (cm *ContextManager) GetGitStatus() (*types.GitStatus, error) {
	if cm.gitRepo == nil {
		return nil, fmt.Errorf("not a git repository")
	}

	worktree, err := cm.gitRepo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	gitStatus := &types.GitStatus{
		IsClean: status.IsClean(),
		Files:   make(map[string]string),
	}

	for file, fileStatus := range status {
		gitStatus.Files[file] = string(fileStatus.Staging) + string(fileStatus.Worktree)
	}

	return gitStatus, nil
}

// RefreshContext rebuilds the project context
func (cm *ContextManager) RefreshContext() (*types.ProjectContext, error) {
	return cm.BuildProjectContext(cm.workdir)
}

// UpdateConfig updates the GOAI configuration
func (cm *ContextManager) UpdateConfig(config *types.GOAIConfig) error {
	parser := NewGOAIConfigParser()
	return parser.WriteFile(cm.configPath, config)
}

// BuildIndex builds the codebase index for enhanced search and reasoning
func (cm *ContextManager) BuildIndex(ctx context.Context) error {
	if cm.indexManager == nil {
		return fmt.Errorf("index manager not initialized")
	}
	
	return cm.indexManager.BuildIndex(ctx, cm.workdir)
}

// RefreshIndex refreshes the index with specified file changes
func (cm *ContextManager) RefreshIndex(ctx context.Context, paths []string) error {
	if cm.indexManager == nil {
		return fmt.Errorf("index manager not initialized")
	}
	
	return cm.indexManager.RefreshIndex(ctx, paths)
}

// SearchCode performs code search using the index
func (cm *ContextManager) SearchCode(ctx context.Context, query string, maxResults int) ([]*indexing.SearchResult, error) {
	if cm.indexManager == nil {
		return nil, fmt.Errorf("index manager not initialized")
	}
	
	searchReq := &indexing.SearchRequest{
		Query:          query,
		WorkingDir:     cm.workdir,
		MaxResults:     maxResults,
		IncludeContent: true,
		SearchTypes:    []indexing.SearchType{indexing.SearchTypeFullText},
	}
	
	// For now, return a single result - this will be enhanced when we implement multiple results
	result, err := cm.indexManager.Search(ctx, searchReq)
	if err != nil {
		return nil, err
	}
	
	return []*indexing.SearchResult{result}, nil
}

// GetRelevantContext returns relevant code context for reasoning tasks
func (cm *ContextManager) GetRelevantContext(ctx context.Context, query string, maxItems int) ([]*indexing.ContextItem, error) {
	if cm.indexManager == nil {
		return nil, fmt.Errorf("index manager not initialized")
	}
	
	opts := &indexing.ContextOptions{
		MaxItems:           maxItems,
		MaxCharsPerItem:    1000,
		ContextTypes:       []indexing.ContextType{indexing.ContextTypeFunction, indexing.ContextTypeClass, indexing.ContextTypeInterface},
		RelevanceThreshold: 0.5,
		WorkingDir:         cm.workdir,
		IncludeSource:      true,
	}
	
	return cm.indexManager.GetContextItems(ctx, query, opts)
}

// GetIndexStatus returns the current index status
func (cm *ContextManager) GetIndexStatus() (*indexing.IndexStatus, error) {
	if cm.indexManager == nil {
		return nil, fmt.Errorf("index manager not initialized")
	}
	
	return cm.indexManager.GetIndexStatus(cm.workdir)
}

// IsIndexReady checks if the index is ready for queries
func (cm *ContextManager) IsIndexReady() bool {
	if cm.indexManager == nil {
		return false
	}
	
	return cm.indexManager.IsIndexReady(cm.workdir)
}

// WaitForIndex waits for indexing to complete
func (cm *ContextManager) WaitForIndex(ctx context.Context) error {
	if cm.indexManager == nil {
		return fmt.Errorf("index manager not initialized")
	}
	
	return cm.indexManager.WaitForIndex(ctx, cm.workdir)
}

// Close closes the context manager and all its resources
func (cm *ContextManager) Close() error {
	// Stop file watcher
	if err := cm.Stop(); err != nil {
		fmt.Printf("Warning: Failed to stop file watcher: %v\n", err)
	}
	
	// Close index manager
	if cm.indexManager != nil {
		if indexManager, ok := cm.indexManager.(*indexing.IndexManager); ok {
			if err := indexManager.Close(); err != nil {
				fmt.Printf("Warning: Failed to close index manager: %v\n", err)
			}
		}
	}
	
	return nil
}
