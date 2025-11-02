package search

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Zerofisher/goai/pkg/tools"
)

// Result represents a single search result.
type Result struct {
	File       string
	Line       int
	Column     int
	Content    string
	MatchStart int
	MatchEnd   int
}

// Location represents the location of a symbol.
type Location struct {
	File   string
	Line   int
	Column int
	Type   string // "function", "struct", "interface", "variable", etc.
}

// SearchOptions configures search behavior.
type SearchOptions struct {
	CaseSensitive bool
	WholeWord     bool
	FilePattern   string // Glob pattern for files to search
	MaxResults    int
	Context       int // Number of context lines
}

// SearchTool implements code search functionality using grep.
type SearchTool struct {
	workDir   string
	validator tools.SecurityValidator
}

// NewSearchTool creates a new search tool.
func NewSearchTool(workDir string, validator tools.SecurityValidator) *SearchTool {
	if validator == nil {
		validator = tools.NewSecurityValidator(workDir)
	}
	return &SearchTool{
		workDir:   workDir,
		validator: validator,
	}
}

// Name returns the name of the tool.
func (t *SearchTool) Name() string {
	return "search"
}

// Description returns the description of the tool.
func (t *SearchTool) Description() string {
	return "Search for code patterns, symbols, and text within the project"
}

// InputSchema returns the JSON schema for the input.
func (t *SearchTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Search pattern (regex or plain text)",
			},
			"type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"code", "symbol"},
				"description": "Type of search: 'code' for text search, 'symbol' for symbol search",
			},
			"file_pattern": map[string]interface{}{
				"type":        "string",
				"description": "File pattern to search in (e.g., '*.go', '*.py')",
			},
			"case_sensitive": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the search is case-sensitive",
				"default":     true,
			},
			"whole_word": map[string]interface{}{
				"type":        "boolean",
				"description": "Match whole words only",
				"default":     false,
			},
			"max_results": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of results to return",
				"default":     50,
			},
			"context": map[string]interface{}{
				"type":        "integer",
				"description": "Number of context lines to include",
				"default":     0,
			},
		},
		"required": []string{"pattern", "type"},
	}
}

// Execute performs the search operation.
func (t *SearchTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
	// Extract parameters
	pattern, err := getStringParam(input, "pattern")
	if err != nil {
		return "", err
	}

	searchType, err := getStringParam(input, "type")
	if err != nil {
		return "", err
	}

	// Build search options
	options := t.buildOptions(input)

	// Perform search based on type
	switch searchType {
	case "code":
		results, err := t.SearchCode(ctx, pattern, options)
		if err != nil {
			return "", err
		}
		return t.formatCodeResults(results), nil

	case "symbol":
		locations, err := t.SearchSymbol(ctx, pattern)
		if err != nil {
			return "", err
		}
		return t.formatSymbolResults(locations), nil

	default:
		return "", fmt.Errorf("invalid search type: %s", searchType)
	}
}

// Validate checks if the input is valid.
func (t *SearchTool) Validate(input map[string]interface{}) error {
	// Check required fields
	if _, ok := input["pattern"]; !ok {
		return fmt.Errorf("missing required field: pattern")
	}

	if _, ok := input["type"]; !ok {
		return fmt.Errorf("missing required field: type")
	}

	// Validate search type
	searchType, _ := input["type"].(string)
	if searchType != "code" && searchType != "symbol" {
		return fmt.Errorf("invalid search type: %s (must be 'code' or 'symbol')", searchType)
	}

	// Validate numeric fields
	if maxResults, ok := input["max_results"]; ok {
		if val, ok := maxResults.(float64); ok {
			if val < 1 || val > 1000 {
				return fmt.Errorf("max_results must be between 1 and 1000")
			}
		}
	}

	if context, ok := input["context"]; ok {
		if val, ok := context.(float64); ok {
			if val < 0 || val > 10 {
				return fmt.Errorf("context must be between 0 and 10")
			}
		}
	}

	return nil
}

// SearchCode searches for code patterns using grep.
func (t *SearchTool) SearchCode(ctx context.Context, pattern string, options SearchOptions) ([]Result, error) {
	// Build grep command
	args := []string{"-n", "-H"} // Line numbers and filenames

	if !options.CaseSensitive {
		args = append(args, "-i")
	}

	if options.WholeWord {
		args = append(args, "-w")
	}

	if options.Context > 0 {
		args = append(args, fmt.Sprintf("-C%d", options.Context))
	}

	// Limit matches at grep level for efficiency
	if options.MaxResults > 0 {
		args = append(args, "-m", fmt.Sprintf("%d", options.MaxResults))
	}

	// Add pattern
	args = append(args, pattern)

	// Add file pattern if specified
	if options.FilePattern != "" {
		// Use find to get matching files first
		files, err := t.findMatchingFiles(ctx, options.FilePattern)
		if err != nil {
			return nil, err
		}
		if len(files) == 0 {
			return []Result{}, nil // No matching files
		}
		args = append(args, files...)
	} else {
		// Search recursively in workdir
		args = append(args, "-r", t.workDir)
	}

	// Execute grep with context binding
	cmd := exec.CommandContext(ctx, "grep", args...)
	cmd.Dir = t.workDir

	// Use a buffer with size limit to prevent excessive memory usage
	var outBuf bytes.Buffer
	const maxOutputSize = 10 * 1024 * 1024 // 10MB limit
	cmd.Stdout = &limitedWriter{w: &outBuf, limit: maxOutputSize}

	err := cmd.Run()
	if err != nil {
		// Check if context was canceled
		if ctx.Err() != nil {
			return nil, fmt.Errorf("search canceled: %w", ctx.Err())
		}
		// grep returns exit code 1 when no matches found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []Result{}, nil
		}
		return nil, fmt.Errorf("grep failed: %w", err)
	}

	// Parse results
	results := t.parseGrepOutput(outBuf.String())

	// Additional limit check (belt and suspenders)
	if options.MaxResults > 0 && len(results) > options.MaxResults {
		results = results[:options.MaxResults]
	}

	return results, nil
}

// SearchSymbol searches for symbol definitions.
func (t *SearchTool) SearchSymbol(ctx context.Context, symbol string) ([]Location, error) {
	// For Go files, use simple grep patterns for common declarations
	patterns := []string{
		fmt.Sprintf(`func\s+%s\s*\(`, symbol),           // Function
		fmt.Sprintf(`func\s+\([^)]+\)\s+%s\s*\(`, symbol), // Method
		fmt.Sprintf(`type\s+%s\s+`, symbol),              // Type
		fmt.Sprintf(`var\s+%s\s+`, symbol),               // Variable
		fmt.Sprintf(`const\s+%s\s+`, symbol),             // Constant
		fmt.Sprintf(`^%s\s*:=`, symbol),                  // Short variable declaration
	}

	var locations []Location
	const maxOutputSize = 5 * 1024 * 1024 // 5MB limit per pattern

	for _, pattern := range patterns {
		// Check context before each pattern
		if ctx.Err() != nil {
			return nil, fmt.Errorf("search canceled: %w", ctx.Err())
		}

		cmd := exec.CommandContext(ctx, "grep", "-rn", "-E", "-m", "100", pattern, t.workDir, "--include=*.go")

		var outBuf bytes.Buffer
		cmd.Stdout = &limitedWriter{w: &outBuf, limit: maxOutputSize}

		err := cmd.Run()
		if err != nil {
			// Check if context was canceled
			if ctx.Err() != nil {
				return nil, fmt.Errorf("search canceled: %w", ctx.Err())
			}
			// Continue if no matches for this pattern
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				continue
			}
			return nil, fmt.Errorf("grep failed: %w", err)
		}

		locs := t.parseSymbolOutput(outBuf.String(), symbol)
		locations = append(locations, locs...)
	}

	// Remove duplicates
	locations = t.deduplicateLocations(locations)

	// Sort by file and line
	sort.Slice(locations, func(i, j int) bool {
		if locations[i].File != locations[j].File {
			return locations[i].File < locations[j].File
		}
		return locations[i].Line < locations[j].Line
	})

	return locations, nil
}

// Helper functions

func (t *SearchTool) buildOptions(input map[string]interface{}) SearchOptions {
	options := SearchOptions{
		CaseSensitive: true,
		WholeWord:     false,
		MaxResults:    50,
		Context:       0,
	}

	if val, ok := input["case_sensitive"].(bool); ok {
		options.CaseSensitive = val
	}

	if val, ok := input["whole_word"].(bool); ok {
		options.WholeWord = val
	}

	if val, ok := input["file_pattern"].(string); ok {
		options.FilePattern = val
	}

	if val, ok := input["max_results"].(float64); ok {
		options.MaxResults = int(val)
	}

	if val, ok := input["context"].(float64); ok {
		options.Context = int(val)
	}

	return options
}

func (t *SearchTool) findMatchingFiles(ctx context.Context, pattern string) ([]string, error) {
	// Use find to locate files recursively. This provides a single source of truth
	// and consistent behavior across platforms (within POSIX constraints).
	cmd := exec.CommandContext(ctx, "find", t.workDir, "-name", pattern, "-type", "f")

	var outBuf bytes.Buffer
	const maxOutputSize = 2 * 1024 * 1024 // 2MB limit for file list
	cmd.Stdout = &limitedWriter{w: &outBuf, limit: maxOutputSize}

	err := cmd.Run()
	if err != nil {
		// Check if context was canceled
		if ctx.Err() != nil {
			return nil, fmt.Errorf("find canceled: %w", ctx.Err())
		}
		// No matches or find error: return empty slice to indicate no files
		return []string{}, nil
	}

	var matches []string
	scanner := bufio.NewScanner(&outBuf)
	for scanner.Scan() {
		file := scanner.Text()
		if file != "" {
			matches = append(matches, file)
		}
	}

	return matches, nil
}

func (t *SearchTool) parseGrepOutput(output string) []Result {
	var results []Result
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		// Parse grep output format: file:line:content
		parts := strings.SplitN(line, ":", 3)
		if len(parts) >= 3 {
			lineNum := 0
			if _, err := fmt.Sscanf(parts[1], "%d", &lineNum); err != nil {
				continue // Skip lines with invalid line numbers
			}

			// Make path relative to workdir
			relPath, _ := filepath.Rel(t.workDir, parts[0])

			results = append(results, Result{
				File:    relPath,
				Line:    lineNum,
				Content: parts[2],
			})
		}
	}

	return results
}

func (t *SearchTool) parseSymbolOutput(output string, symbol string) []Location {
	var locations []Location
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		// Parse grep output format: file:line:content
		parts := strings.SplitN(line, ":", 3)
		if len(parts) >= 3 {
			lineNum := 0
			if _, err := fmt.Sscanf(parts[1], "%d", &lineNum); err != nil {
				continue // Skip lines with invalid line numbers
			}

			// Make path relative to workdir
			relPath, _ := filepath.Rel(t.workDir, parts[0])

			// Determine symbol type from content
			content := parts[2]
			symbolType := t.detectSymbolType(content, symbol)

			locations = append(locations, Location{
				File: relPath,
				Line: lineNum,
				Type: symbolType,
			})
		}
	}

	return locations
}

func (t *SearchTool) detectSymbolType(content, symbol string) string {
	content = strings.TrimSpace(content)

	if strings.HasPrefix(content, "func ") {
		if strings.Contains(content, ") "+symbol) {
			return "method"
		}
		return "function"
	}
	if strings.HasPrefix(content, "type ") {
		if strings.Contains(content, "struct") {
			return "struct"
		}
		if strings.Contains(content, "interface") {
			return "interface"
		}
		return "type"
	}
	if strings.HasPrefix(content, "var ") {
		return "variable"
	}
	if strings.HasPrefix(content, "const ") {
		return "constant"
	}
	if strings.Contains(content, ":=") {
		return "variable"
	}

	return "unknown"
}

func (t *SearchTool) deduplicateLocations(locations []Location) []Location {
	seen := make(map[string]bool)
	var unique []Location

	for _, loc := range locations {
		key := fmt.Sprintf("%s:%d", loc.File, loc.Line)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, loc)
		}
	}

	return unique
}

func (t *SearchTool) formatCodeResults(results []Result) string {
	if len(results) == 0 {
		return "No matches found."
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Found %d matches:\n\n", len(results)))

	for _, result := range results {
		output.WriteString(fmt.Sprintf("%s:%d\n", result.File, result.Line))
		output.WriteString(fmt.Sprintf("  %s\n", strings.TrimSpace(result.Content)))
		output.WriteString("\n")
	}

	return output.String()
}

func (t *SearchTool) formatSymbolResults(locations []Location) string {
	if len(locations) == 0 {
		return "No symbol definitions found."
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Found %d symbol definitions:\n\n", len(locations)))

	for _, loc := range locations {
		output.WriteString(fmt.Sprintf("%s:%d [%s]\n", loc.File, loc.Line, loc.Type))
	}

	return output.String()
}

// Helper function to extract string parameters
func getStringParam(input map[string]interface{}, key string) (string, error) {
	val, ok := input[key]
	if !ok {
		return "", fmt.Errorf("missing required parameter: %s", key)
	}

	str, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("parameter %s must be a string", key)
	}

	return str, nil
}

// limitedWriter wraps an io.Writer with a maximum size limit
type limitedWriter struct {
	w       *bytes.Buffer
	limit   int
	written int
}

func (lw *limitedWriter) Write(p []byte) (n int, err error) {
	if lw.written >= lw.limit {
		return 0, fmt.Errorf("output size limit exceeded (%d bytes)", lw.limit)
	}

	remaining := lw.limit - lw.written
	toWrite := len(p)
	if toWrite > remaining {
		toWrite = remaining
	}

	n, writeErr := lw.w.Write(p[:toWrite])
	lw.written += n

	if toWrite < len(p) {
		return n, fmt.Errorf("output size limit exceeded (%d bytes)", lw.limit)
	}

	return n, writeErr
}
