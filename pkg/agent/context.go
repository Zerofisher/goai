package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Context manages the agent's operating context
type Context struct {
	workDir      string
	systemPrompt string
	projectInfo  *ProjectInfo
	startTime    time.Time
}

// ProjectInfo contains information about the current project
type ProjectInfo struct {
	Name        string
	Path        string
	Language    string
	HasGit_     bool
	GitBranch   string
	FileCount   int
	TotalSize   int64
	LastModified time.Time
}

// GetName returns the project name.
func (p *ProjectInfo) GetName() string {
	return p.Name
}

// GetLanguage returns the primary language.
func (p *ProjectInfo) GetLanguage() string {
	return p.Language
}

// HasGit returns whether the project has git.
func (p *ProjectInfo) HasGit() bool {
	return p.HasGit_
}

// GetGitBranch returns the current git branch.
func (p *ProjectInfo) GetGitBranch() string {
	return p.GitBranch
}

// NewContext creates a new context manager
func NewContext(workDir string) *Context {
	if workDir == "" {
		workDir, _ = os.Getwd()
	}

	// Ensure absolute path
	absPath, err := filepath.Abs(workDir)
	if err == nil {
		workDir = absPath
	}

	ctx := &Context{
		workDir:   workDir,
		startTime: time.Now(),
	}

	// Initialize project info
	ctx.projectInfo = ctx.analyzeProject()

	// Generate system prompt
	ctx.systemPrompt = ctx.generateSystemPrompt()

	return ctx
}

// GetSystemPrompt returns the system prompt for the LLM
func (c *Context) GetSystemPrompt() string {
	return c.systemPrompt
}

// GetWorkDir returns the working directory
func (c *Context) GetWorkDir() string {
	return c.workDir
}

// GetProjectName returns the project name.
func (c *Context) GetProjectName() string {
	if c.projectInfo != nil {
		return c.projectInfo.Name
	}
	return "unknown"
}

// GetProjectLanguage returns the project's primary language.
func (c *Context) GetProjectLanguage() string {
	if c.projectInfo != nil {
		return c.projectInfo.Language
	}
	return "unknown"
}

// ProjectHasGit returns whether the project has git.
func (c *Context) ProjectHasGit() bool {
	if c.projectInfo != nil {
		return c.projectInfo.HasGit_
	}
	return false
}

// GetProjectGitBranch returns the current git branch.
func (c *Context) GetProjectGitBranch() string {
	if c.projectInfo != nil {
		return c.projectInfo.GitBranch
	}
	return ""
}

// GetProjectInfo returns project information
func (c *Context) GetProjectInfo() *ProjectInfo {
	return c.projectInfo
}

// GetUptime returns how long the context has been active
func (c *Context) GetUptime() time.Duration {
	return time.Since(c.startTime)
}

// generateSystemPrompt creates the system prompt based on context
func (c *Context) generateSystemPrompt() string {
	var parts []string

	// Base prompt
	parts = append(parts, "You are GoAI Coder, an intelligent programming assistant.")
	parts = append(parts, "You help users with coding tasks, debugging, and software development.")
	parts = append(parts, "")

	// Context information
	parts = append(parts, "## Operating Context")
	parts = append(parts, fmt.Sprintf("- Working Directory: %s", c.workDir))
	parts = append(parts, fmt.Sprintf("- Operating System: %s/%s", runtime.GOOS, runtime.GOARCH))
	parts = append(parts, fmt.Sprintf("- Go Version: %s", runtime.Version()))

	if c.projectInfo != nil {
		parts = append(parts, "")
		parts = append(parts, "## Project Information")
		parts = append(parts, fmt.Sprintf("- Project Name: %s", c.projectInfo.Name))
		parts = append(parts, fmt.Sprintf("- Language: %s", c.projectInfo.Language))
		if c.projectInfo.HasGit_ {
			parts = append(parts, fmt.Sprintf("- Git Branch: %s", c.projectInfo.GitBranch))
		}
		parts = append(parts, fmt.Sprintf("- Files: %d", c.projectInfo.FileCount))
	}

	parts = append(parts, "")
	parts = append(parts, "## Guidelines")
	parts = append(parts, "1. Always validate file paths and ensure they are within the working directory")
	parts = append(parts, "2. Provide clear explanations for your actions")
	parts = append(parts, "3. Handle errors gracefully and suggest alternatives")
	parts = append(parts, "4. Use appropriate tools for each task")
	parts = append(parts, "5. Follow best practices and coding standards")
	parts = append(parts, "")
	parts = append(parts, "## Available Tools")
	parts = append(parts, "You have access to the following tools:")
	parts = append(parts, "- bash: Execute shell commands")
	parts = append(parts, "- read_file: Read file contents")
	parts = append(parts, "- write_file: Write content to files")
	parts = append(parts, "- list_files: List directory contents")
	parts = append(parts, "- edit_file: Edit specific parts of files")
	parts = append(parts, "- search: Search for patterns in files")
	parts = append(parts, "")
	parts = append(parts, "Use tools when needed to accomplish tasks. Always check the results and handle errors appropriately.")

	return strings.Join(parts, "\n")
}

// analyzeProject analyzes the current project directory
func (c *Context) analyzeProject() *ProjectInfo {
	info := &ProjectInfo{
		Name: filepath.Base(c.workDir),
		Path: c.workDir,
	}

	// Check for Git
	gitDir := filepath.Join(c.workDir, ".git")
	if stat, err := os.Stat(gitDir); err == nil && stat.IsDir() {
		info.HasGit_ = true
		info.GitBranch = c.getGitBranch()
	}

	// Detect primary language
	info.Language = c.detectLanguage()

	// Count files and calculate size
	c.walkProject(info)

	return info
}

// getGitBranch gets the current git branch
func (c *Context) getGitBranch() string {
	headFile := filepath.Join(c.workDir, ".git", "HEAD")
	data, err := os.ReadFile(headFile)
	if err != nil {
		return "unknown"
	}

	content := strings.TrimSpace(string(data))
	if strings.HasPrefix(content, "ref: refs/heads/") {
		return strings.TrimPrefix(content, "ref: refs/heads/")
	}

	return "detached"
}

// detectLanguage detects the primary programming language
func (c *Context) detectLanguage() string {
	// Simple language detection based on file extensions
	langCounts := make(map[string]int)

	filepath.Walk(c.workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Skip hidden files and vendor directories
		if strings.Contains(path, "/.") || strings.Contains(path, "/vendor/") {
			return nil
		}

		ext := filepath.Ext(path)
		switch ext {
		case ".go":
			langCounts["Go"]++
		case ".py":
			langCounts["Python"]++
		case ".js", ".jsx", ".ts", ".tsx":
			langCounts["JavaScript/TypeScript"]++
		case ".java":
			langCounts["Java"]++
		case ".cpp", ".cc", ".cxx", ".c", ".h", ".hpp":
			langCounts["C/C++"]++
		case ".rs":
			langCounts["Rust"]++
		case ".rb":
			langCounts["Ruby"]++
		case ".php":
			langCounts["PHP"]++
		case ".cs":
			langCounts["C#"]++
		case ".swift":
			langCounts["Swift"]++
		}

		return nil
	})

	// Find the most common language
	maxCount := 0
	language := "Unknown"
	for lang, count := range langCounts {
		if count > maxCount {
			maxCount = count
			language = lang
		}
	}

	return language
}

// walkProject walks the project directory and collects statistics
func (c *Context) walkProject(info *ProjectInfo) {
	filepath.Walk(c.workDir, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden directories and vendor
		if fileInfo.IsDir() {
			name := fileInfo.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files
		if strings.HasPrefix(fileInfo.Name(), ".") {
			return nil
		}

		info.FileCount++
		info.TotalSize += fileInfo.Size()

		if fileInfo.ModTime().After(info.LastModified) {
			info.LastModified = fileInfo.ModTime()
		}

		return nil
	})
}

// UpdateContext updates the context with new information
func (c *Context) UpdateContext(key string, value interface{}) {
	// This could be extended to update various context parameters
	// For now, it's a placeholder for future functionality
}

// GetContextSummary returns a summary of the current context
func (c *Context) GetContextSummary() string {
	var parts []string

	parts = append(parts, fmt.Sprintf("Work Directory: %s", c.workDir))

	if c.projectInfo != nil {
		parts = append(parts, fmt.Sprintf("Project: %s (%s)", c.projectInfo.Name, c.projectInfo.Language))
		if c.projectInfo.HasGit_ {
			parts = append(parts, fmt.Sprintf("Git Branch: %s", c.projectInfo.GitBranch))
		}
		parts = append(parts, fmt.Sprintf("Files: %d (%.2f MB)",
			c.projectInfo.FileCount,
			float64(c.projectInfo.TotalSize)/(1024*1024)))
	}

	parts = append(parts, fmt.Sprintf("Uptime: %s", c.GetUptime().Round(time.Second)))

	return strings.Join(parts, " | ")
}