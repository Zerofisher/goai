package context

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Zerofisher/goai/pkg/types"
)

// ProjectStructureAnalyzer analyzes project directory structure and patterns
type ProjectStructureAnalyzer struct {
	rootPath string
	patterns map[string]*regexp.Regexp
}

// NewProjectStructureAnalyzer creates a new project structure analyzer
func NewProjectStructureAnalyzer(rootPath string) *ProjectStructureAnalyzer {
	return &ProjectStructureAnalyzer{
		rootPath: rootPath,
		patterns: map[string]*regexp.Regexp{
			"test":        regexp.MustCompile(`.*_test\.go$|.*\.test\..*$|test.*\.go$`),
			"config":      regexp.MustCompile(`.*config.*|.*\.conf$|.*\.toml$|.*\.yaml$|.*\.yml$|.*\.json$|.*\.env$`),
			"docs":        regexp.MustCompile(`.*\.md$|.*\.rst$|.*\.txt$|doc.*|readme.*`),
			"build":       regexp.MustCompile(`Makefile|Dockerfile|.*\.mk$|build\..*|.*\.sh$`),
			"source":      regexp.MustCompile(`.*\.go$|.*\.js$|.*\.ts$|.*\.py$|.*\.java$|.*\.rs$`),
			"resource":    regexp.MustCompile(`.*\.sql$|.*\.proto$|.*\.yaml$|.*\.json$|.*\.xml$`),
			"generated":   regexp.MustCompile(`.*\.pb\.go$|.*_gen\.go$|.*generated.*`),
			"vendor":      regexp.MustCompile(`vendor/.*|node_modules/.*`),
			"binary":      regexp.MustCompile(`.*\.so$|.*\.dll$|.*\.exe$|.*\.bin$`),
		},
	}
}

// AnalyzeStructure analyzes the project structure
func (psa *ProjectStructureAnalyzer) AnalyzeStructure() (*types.ProjectStructure, error) {
	structure := &types.ProjectStructure{
		RootPath:    psa.rootPath,
		Directories: []types.Directory{},
		Files:       []types.FileInfo{},
		Patterns:    []types.Pattern{},
	}

	// Walk through the directory tree
	err := filepath.Walk(psa.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files and directories (except .git for metadata)
		if strings.HasPrefix(info.Name(), ".") && info.Name() != ".git" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		relativePath, err := filepath.Rel(psa.rootPath, path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			dir := types.Directory{
				Path:        relativePath,
				Name:        info.Name(),
				Type:        psa.classifyDirectory(relativePath, info.Name()),
				Description: psa.describeDirectory(relativePath, info.Name()),
			}
			structure.Directories = append(structure.Directories, dir)
		} else {
			fileInfo := types.FileInfo{
				Path:         relativePath,
				Name:         info.Name(),
				Extension:    filepath.Ext(info.Name()),
				Size:         info.Size(),
				ModifiedTime: info.ModTime(),
				IsOpen:       false, // We don't track open files here
			}
			structure.Files = append(structure.Files, fileInfo)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory tree: %w", err)
	}

	// Analyze patterns
	structure.Patterns = psa.analyzePatterns(structure)

	return structure, nil
}

// classifyDirectory classifies a directory based on its path and name
func (psa *ProjectStructureAnalyzer) classifyDirectory(path, name string) string {
	lowName := strings.ToLower(name)
	lowPath := strings.ToLower(path)

	// Go-specific patterns
	if lowName == "cmd" || strings.Contains(lowPath, "/cmd") {
		return "command"
	}
	if lowName == "pkg" || strings.Contains(lowPath, "/pkg") {
		return "package"
	}
	if lowName == "internal" || strings.Contains(lowPath, "/internal") {
		return "internal"
	}
	if lowName == "api" || strings.Contains(lowPath, "/api") {
		return "api"
	}
	if lowName == "web" || strings.Contains(lowPath, "/web") {
		return "web"
	}

	// Common patterns
	if lowName == "test" || lowName == "tests" || strings.Contains(lowName, "test") {
		return "test"
	}
	if lowName == "doc" || lowName == "docs" || lowName == "documentation" {
		return "documentation"
	}
	if lowName == "config" || lowName == "configs" || lowName == "configuration" {
		return "configuration"
	}
	if lowName == "scripts" || lowName == "script" {
		return "script"
	}
	if lowName == "assets" || lowName == "static" || lowName == "resources" {
		return "resource"
	}
	if lowName == "build" || lowName == "dist" || lowName == "target" {
		return "build"
	}
	if lowName == "vendor" || lowName == "node_modules" || lowName == "third_party" {
		return "dependency"
	}
	if lowName == "examples" || lowName == "example" || lowName == "samples" {
		return "example"
	}
	if lowName == "tools" || lowName == "util" || lowName == "utils" || lowName == "utilities" {
		return "utility"
	}

	return "source"
}

// describeDirectory provides a description for a directory
func (psa *ProjectStructureAnalyzer) describeDirectory(path, name string) string {
	dirType := psa.classifyDirectory(path, name)

	descriptions := map[string]string{
		"command":       "Command-line applications and executables",
		"package":       "Reusable library packages",
		"internal":      "Internal/private packages not for external use",
		"api":           "API definitions and handlers",
		"web":           "Web assets and templates",
		"test":          "Test files and testing utilities",
		"documentation": "Project documentation and guides",
		"configuration": "Configuration files and settings",
		"script":        "Build scripts and automation tools",
		"resource":      "Static resources and assets",
		"build":         "Build artifacts and distribution files",
		"dependency":    "Third-party dependencies",
		"example":       "Example code and samples",
		"utility":       "Utility functions and helper tools",
		"source":        "Source code files",
	}

	if desc, exists := descriptions[dirType]; exists {
		return desc
	}
	return "General purpose directory"
}

// analyzePatterns analyzes file patterns in the project
func (psa *ProjectStructureAnalyzer) analyzePatterns(structure *types.ProjectStructure) []types.Pattern {
	patterns := []types.Pattern{}

	// Count files by pattern
	patternCounts := make(map[string]int)
	for _, file := range structure.Files {
		for patternName, pattern := range psa.patterns {
			if pattern.MatchString(file.Path) || pattern.MatchString(file.Name) {
				patternCounts[patternName]++
			}
		}
	}

	// Create pattern descriptions
	for patternName, count := range patternCounts {
		if count > 0 {
			pattern := types.Pattern{
				Name:        patternName,
				Pattern:     psa.patterns[patternName].String(),
				Type:        "file",
				Description: fmt.Sprintf("%s files (%d found)", patternName, count),
			}
			patterns = append(patterns, pattern)
		}
	}

	// Analyze directory patterns
	dirPatterns := make(map[string]int)
	for _, dir := range structure.Directories {
		dirPatterns[dir.Type]++
	}

	for dirType, count := range dirPatterns {
		if count > 0 {
			pattern := types.Pattern{
				Name:        fmt.Sprintf("%s_dirs", dirType),
				Pattern:     dirType,
				Type:        "directory",
				Description: fmt.Sprintf("%s directories (%d found)", dirType, count),
			}
			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

// GetProjectType attempts to determine the project type
func (psa *ProjectStructureAnalyzer) GetProjectType() string {
	// Check for Go project
	if psa.hasFile("go.mod") || psa.hasFile("go.sum") {
		return "go"
	}

	// Check for Node.js project
	if psa.hasFile("package.json") {
		return "nodejs"
	}

	// Check for Python project
	if psa.hasFile("requirements.txt") || psa.hasFile("setup.py") || psa.hasFile("pyproject.toml") {
		return "python"
	}

	// Check for Java project
	if psa.hasFile("pom.xml") || psa.hasFile("build.gradle") {
		return "java"
	}

	// Check for Rust project
	if psa.hasFile("Cargo.toml") {
		return "rust"
	}

	// Check for C/C++ project
	if psa.hasFile("Makefile") || psa.hasFile("CMakeLists.txt") {
		return "c/cpp"
	}

	return "unknown"
}

// hasFile checks if a file exists in the root directory
func (psa *ProjectStructureAnalyzer) hasFile(filename string) bool {
	_, err := os.Stat(filepath.Join(psa.rootPath, filename))
	return err == nil
}

// GetSourceFiles returns all source files
func (psa *ProjectStructureAnalyzer) GetSourceFiles() ([]string, error) {
	var sourceFiles []string

	err := filepath.Walk(psa.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if psa.patterns["source"].MatchString(path) {
				relativePath, err := filepath.Rel(psa.rootPath, path)
				if err != nil {
					return err
				}
				sourceFiles = append(sourceFiles, relativePath)
			}
		}

		return nil
	})

	return sourceFiles, err
}

// GetTestFiles returns all test files
func (psa *ProjectStructureAnalyzer) GetTestFiles() ([]string, error) {
	var testFiles []string

	err := filepath.Walk(psa.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if psa.patterns["test"].MatchString(path) {
				relativePath, err := filepath.Rel(psa.rootPath, path)
				if err != nil {
					return err
				}
				testFiles = append(testFiles, relativePath)
			}
		}

		return nil
	})

	return testFiles, err
}

// GetConfigFiles returns all configuration files
func (psa *ProjectStructureAnalyzer) GetConfigFiles() ([]string, error) {
	var configFiles []string

	err := filepath.Walk(psa.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if psa.patterns["config"].MatchString(path) {
				relativePath, err := filepath.Rel(psa.rootPath, path)
				if err != nil {
					return err
				}
				configFiles = append(configFiles, relativePath)
			}
		}

		return nil
	})

	return configFiles, err
}