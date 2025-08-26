package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Zerofisher/goai/pkg/indexing"
)

// SearchCodeTool searches through code using the indexing system
type SearchCodeTool struct {
	indexManager *indexing.EnhancedIndexManager
}

func NewSearchCodeTool(indexManager *indexing.EnhancedIndexManager) *SearchCodeTool {
	return &SearchCodeTool{
		indexManager: indexManager,
	}
}

func (t *SearchCodeTool) Name() string {
	return "searchCode"
}

func (t *SearchCodeTool) Description() string {
	return "Search through codebase using full-text, semantic, or symbol-based search"
}

func (t *SearchCodeTool) Parameters() ParameterSchema {
	return ParameterSchema{
		Required: []string{"query"},
		Properties: map[string]ParameterProperty{
			"query": {
				Type:        "string",
				Description: "Search query string",
			},
			"type": {
				Type:        "string",
				Description: "Search type: 'text', 'semantic', 'symbol', or 'hybrid'",
				Default:     "hybrid",
				Enum:        []string{"text", "semantic", "symbol", "hybrid"},
			},
			"maxResults": {
				Type:        "number",
				Description: "Maximum number of results to return",
				Default:     10,
			},
			"includeContext": {
				Type:        "boolean",
				Description: "Include surrounding context for matches",
				Default:     true,
			},
		},
	}
}

func (t *SearchCodeTool) Execute(ctx context.Context, params map[string]any) (*ToolResult, error) {
	query, ok := params["query"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "query parameter must be a string",
		}, nil
	}
	
	searchType, _ := params["type"].(string)
	if searchType == "" {
		searchType = "hybrid"
	}
	
	maxResults := 10
	if mr, ok := params["maxResults"].(float64); ok {
		maxResults = int(mr)
	}
	
	// TODO: Implement context inclusion functionality for search results
	// includeContext, _ := params["includeContext"].(bool)
	
	opts := &indexing.SearchOptions{
		MaxResults: maxResults,
	}
	
	var results []*indexing.SearchResult
	var err error
	
	switch searchType {
	case "text":
		// Use the SearchFullText method to get multiple results
		results, err = t.indexManager.SearchFullText(ctx, query, opts)
	case "semantic":
		results, err = t.indexManager.SearchSemantic(ctx, query, opts)
	case "symbol":
		results, err = t.indexManager.SearchSymbols(ctx, query, opts)
	case "hybrid":
		// Use hybrid retrieval
		req := &indexing.SearchRequest{
			Query:      query,
			MaxResults: maxResults,
			SearchTypes: []indexing.SearchType{
				indexing.SearchTypeFullText,
				indexing.SearchTypeSemantic,
				indexing.SearchTypeSymbol,
			},
		}
		hybridResult, hybridErr := t.indexManager.SearchWithHybridRetrieval(ctx, req)
		if hybridErr != nil {
			err = hybridErr
		} else if hybridResult != nil {
			// Convert retrieval results to search results
			for _, retrievedResult := range hybridResult.Results {
				results = append(results, &indexing.SearchResult{
					FilePath:   retrievedResult.FilePath,
					Content:    retrievedResult.Content,
					Score:      retrievedResult.Score,
					SearchType: indexing.SearchTypeHybrid,
				})
			}
		}
	default:
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("unsupported search type: %s", searchType),
		}, nil
	}
	
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("search failed: %v", err),
		}, nil
	}
	
	// Format results for display
	formattedResults := make([]map[string]any, 0, len(results))
	for _, result := range results {
		formatted := map[string]any{
			"file_path":     result.FilePath,
			"score":         result.Score,
			"search_type":   result.SearchType,
			"line_number":   result.LineNumber,
			"column_number": result.ColumnNumber,
			"content":       result.Content,
			"snippet":       result.Snippet,
		}
		
		if result.SymbolInfo != nil {
			formatted["symbol_info"] = map[string]any{
				"name":        result.SymbolInfo.Name,
				"kind":        result.SymbolInfo.Kind,
				"start_line":  result.SymbolInfo.StartLine,
				"end_line":    result.SymbolInfo.EndLine,
				"signature":   result.SymbolInfo.Signature,
				"doc_string":  result.SymbolInfo.DocString,
			}
		}
		
		if result.Metadata != nil {
			formatted["metadata"] = result.Metadata
		}
		
		formattedResults = append(formattedResults, formatted)
	}
	
	return &ToolResult{
		Success: true,
		Data:    formattedResults,
		Output:  fmt.Sprintf("Found %d results for '%s' using %s search", len(results), query, searchType),
		Metadata: map[string]any{
			"query":       query,
			"search_type": searchType,
			"result_count": len(results),
			"max_results": maxResults,
		},
	}, nil
}

func (t *SearchCodeTool) RequiresConfirmation() bool {
	return false
}

func (t *SearchCodeTool) Category() string {
	return "search"
}

// ListFilesTool lists files in a directory with optional filtering
type ListFilesTool struct{}

func NewListFilesTool() *ListFilesTool {
	return &ListFilesTool{}
}

func (t *ListFilesTool) Name() string {
	return "listFiles"
}

func (t *ListFilesTool) Description() string {
	return "List files and directories with optional filtering and sorting"
}

func (t *ListFilesTool) Parameters() ParameterSchema {
	return ParameterSchema{
		Properties: map[string]ParameterProperty{
			"path": {
				Type:        "string",
				Description: "Directory path to list (default: current directory)",
				Default:     ".",
			},
			"recursive": {
				Type:        "boolean",
				Description: "List files recursively",
				Default:     false,
			},
			"pattern": {
				Type:        "string",
				Description: "File pattern to filter (glob pattern)",
			},
			"includeHidden": {
				Type:        "boolean",
				Description: "Include hidden files and directories",
				Default:     false,
			},
			"sortBy": {
				Type:        "string",
				Description: "Sort files by: 'name', 'size', 'modified'",
				Default:     "name",
				Enum:        []string{"name", "size", "modified"},
			},
			"maxResults": {
				Type:        "number",
				Description: "Maximum number of files to return",
				Default:     100,
			},
		},
	}
}

func (t *ListFilesTool) Execute(ctx context.Context, params map[string]any) (*ToolResult, error) {
	path, _ := params["path"].(string)
	if path == "" {
		path = "."
	}
	
	recursive, _ := params["recursive"].(bool)
	pattern, _ := params["pattern"].(string)
	includeHidden, _ := params["includeHidden"].(bool)
	sortBy, _ := params["sortBy"].(string)
	if sortBy == "" {
		sortBy = "name"
	}
	
	maxResults := 100
	if mr, ok := params["maxResults"].(float64); ok {
		maxResults = int(mr)
	}
	
	cleanPath := filepath.Clean(path)
	
	// Check if path exists
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("path does not exist: %s", cleanPath),
		}, nil
	}
	
	var files []FileInfo
	
	if recursive {
		err := filepath.Walk(cleanPath, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip files with errors
			}
			
			// Skip hidden files if not requested
			if !includeHidden && strings.HasPrefix(info.Name(), ".") {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			
			// Apply pattern filter if specified
			if pattern != "" {
				matched, matchErr := filepath.Match(pattern, info.Name())
				if matchErr != nil || !matched {
					return nil
				}
			}
			
			files = append(files, FileInfo{
				Path:    filePath,
				Name:    info.Name(),
				Size:    info.Size(),
				IsDir:   info.IsDir(),
				ModTime: info.ModTime(),
			})
			
			// Stop if we've reached the limit
			if len(files) >= maxResults {
				return fmt.Errorf("limit reached") // Use error to stop walking
			}
			
			return nil
		})
		if err != nil && err.Error() != "limit reached" {
			return &ToolResult{
				Success: false,
				Error:   fmt.Sprintf("failed to walk directory: %v", err),
			}, nil
		}
	} else {
		entries, err := os.ReadDir(cleanPath)
		if err != nil {
			return &ToolResult{
				Success: false,
				Error:   fmt.Sprintf("failed to read directory: %v", err),
			}, nil
		}
		
		for _, entry := range entries {
			// Skip hidden files if not requested
			if !includeHidden && strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			
			// Apply pattern filter if specified
			if pattern != "" {
				matched, err := filepath.Match(pattern, entry.Name())
				if err != nil || !matched {
					continue
				}
			}
			
			info, err := entry.Info()
			if err != nil {
				continue
			}
			
			files = append(files, FileInfo{
				Path:    filepath.Join(cleanPath, entry.Name()),
				Name:    entry.Name(),
				Size:    info.Size(),
				IsDir:   entry.IsDir(),
				ModTime: info.ModTime(),
			})
			
			if len(files) >= maxResults {
				break
			}
		}
	}
	
	// Sort files
	switch sortBy {
	case "size":
		sort.Slice(files, func(i, j int) bool {
			return files[i].Size > files[j].Size
		})
	case "modified":
		sort.Slice(files, func(i, j int) bool {
			return files[i].ModTime.After(files[j].ModTime)
		})
	default: // name
		sort.Slice(files, func(i, j int) bool {
			return files[i].Name < files[j].Name
		})
	}
	
	// Convert to map format for JSON serialization
	fileList := make([]map[string]any, len(files))
	for i, file := range files {
		fileList[i] = map[string]any{
			"path":         file.Path,
			"name":         file.Name,
			"size":         file.Size,
			"is_directory": file.IsDir,
			"modified":     file.ModTime,
		}
	}
	
	return &ToolResult{
		Success: true,
		Data:    fileList,
		Output:  fmt.Sprintf("Listed %d files in %s", len(files), cleanPath),
		Metadata: map[string]any{
			"directory":     cleanPath,
			"file_count":    len(files),
			"recursive":     recursive,
			"pattern":       pattern,
			"sort_by":       sortBy,
		},
	}, nil
}

func (t *ListFilesTool) RequiresConfirmation() bool {
	return false
}

func (t *ListFilesTool) Category() string {
	return "search"
}

type FileInfo struct {
	Path    string
	Name    string
	Size    int64
	IsDir   bool
	ModTime time.Time
}

// ViewDiffTool shows differences between two files or between a file and its Git version
type ViewDiffTool struct{}

func NewViewDiffTool() *ViewDiffTool {
	return &ViewDiffTool{}
}

func (t *ViewDiffTool) Name() string {
	return "viewDiff"
}

func (t *ViewDiffTool) Description() string {
	return "View differences between two files or between a file and its Git version"
}

func (t *ViewDiffTool) Parameters() ParameterSchema {
	return ParameterSchema{
		Required: []string{"path"},
		Properties: map[string]ParameterProperty{
			"path": {
				Type:        "string",
				Description: "Path to the file to diff",
				Format:      "file-path",
			},
			"compareTo": {
				Type:        "string",
				Description: "Path to compare against (if empty, compares with Git)",
			},
			"contextLines": {
				Type:        "number",
				Description: "Number of context lines to show around changes",
				Default:     3,
			},
			"unified": {
				Type:        "boolean",
				Description: "Use unified diff format",
				Default:     true,
			},
		},
	}
}

func (t *ViewDiffTool) Execute(ctx context.Context, params map[string]any) (*ToolResult, error) {
	path, ok := params["path"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "path parameter must be a string",
		}, nil
	}
	
	compareTo, _ := params["compareTo"].(string)
	contextLines := 3
	if cl, ok := params["contextLines"].(float64); ok {
		contextLines = int(cl)
	}
	
	cleanPath := filepath.Clean(path)
	
	// Check if the file exists
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("file does not exist: %s", cleanPath),
		}, nil
	}
	
	var diffOutput string
	var err error
	
	if compareTo != "" {
		// Compare two files
		diffOutput, err = t.diffTwoFiles(cleanPath, compareTo, contextLines)
	} else {
		// Compare with Git version
		diffOutput, err = t.diffWithGit(cleanPath, contextLines)
	}
	
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to generate diff: %v", err),
		}, nil
	}
	
	return &ToolResult{
		Success: true,
		Data:    diffOutput,
		Output:  fmt.Sprintf("Generated diff for %s", cleanPath),
		Metadata: map[string]any{
			"file_path":     cleanPath,
			"compare_to":    compareTo,
			"context_lines": contextLines,
			"has_changes":   len(diffOutput) > 0,
		},
	}, nil
}

func (t *ViewDiffTool) diffTwoFiles(path1, path2 string, contextLines int) (string, error) {
	content1, err := os.ReadFile(path1)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %v", path1, err)
	}
	
	content2, err := os.ReadFile(path2)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %v", path2, err)
	}
	
	// Simple line-by-line diff implementation
	lines1 := strings.Split(string(content1), "\n")
	lines2 := strings.Split(string(content2), "\n")
	
	return t.generateUnifiedDiff(path1, path2, lines1, lines2), nil
}

func (t *ViewDiffTool) diffWithGit(path string, contextLines int) (string, error) {
	// Use git diff to compare the file with the repository version
	cmd := exec.Command("git", "diff", fmt.Sprintf("-U%d", contextLines), "--", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// If git command fails, return a more informative error
		return "", fmt.Errorf("failed to run git diff (ensure file is in a git repository): %v\nOutput: %s", err, string(out))
	}
	
	// If no changes, git diff returns empty output
	if len(out) == 0 {
		return "No changes detected compared to repository version", nil
	}
	
	return string(out), nil
}

func (t *ViewDiffTool) generateUnifiedDiff(path1, path2 string, lines1, lines2 []string) string {
	var diff strings.Builder
	
	diff.WriteString(fmt.Sprintf("--- %s\n", path1))
	diff.WriteString(fmt.Sprintf("+++ %s\n", path2))
	
	// TODO: Improve diff algorithm - current implementation is naive (simple line-by-line comparison)
	// Consider implementing proper LCS-based diff algorithm or using existing library for better
	// performance and accuracy with large files
	// Suggested libraries: github.com/sergi/go-diff or github.com/pmezard/go-difflib
	maxLen := len(lines1)
	if len(lines2) > maxLen {
		maxLen = len(lines2)
	}
	
	for i := 0; i < maxLen; i++ {
		line1 := ""
		line2 := ""
		
		if i < len(lines1) {
			line1 = lines1[i]
		}
		if i < len(lines2) {
			line2 = lines2[i]
		}
		
		if line1 != line2 {
			if line1 != "" {
				diff.WriteString(fmt.Sprintf("-%s\n", line1))
			}
			if line2 != "" {
				diff.WriteString(fmt.Sprintf("+%s\n", line2))
			}
		}
	}
	
	return diff.String()
}

func (t *ViewDiffTool) RequiresConfirmation() bool {
	return false
}

func (t *ViewDiffTool) Category() string {
	return "search"
}