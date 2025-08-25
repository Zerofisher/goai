package context

import (
	"bufio"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Zerofisher/goai/pkg/types"
)

// DependencyAnalyzer analyzes project dependencies
type DependencyAnalyzer struct {
	rootPath string
	fileSet  *token.FileSet
}

// NewDependencyAnalyzer creates a new dependency analyzer
func NewDependencyAnalyzer(rootPath string) *DependencyAnalyzer {
	return &DependencyAnalyzer{
		rootPath: rootPath,
		fileSet:  token.NewFileSet(),
	}
}

// AnalyzeDependencies analyzes all project dependencies
func (da *DependencyAnalyzer) AnalyzeDependencies() ([]*types.Dependency, error) {
	dependencies := []*types.Dependency{} // Initialize as empty slice, not nil

	// Analyze Go module dependencies
	modDeps, err := da.analyzeGoModDependencies()
	if err == nil {
		dependencies = append(dependencies, modDeps...)
	}

	// Analyze import dependencies
	importDeps, err := da.analyzeImportDependencies()
	if err == nil {
		dependencies = append(dependencies, importDeps...)
	}

	// Analyze other dependency files
	otherDeps, err := da.analyzeOtherDependencies()
	if err == nil {
		dependencies = append(dependencies, otherDeps...)
	}

	return da.deduplicateDependencies(dependencies), nil
}

// analyzeGoModDependencies analyzes go.mod file for module dependencies
func (da *DependencyAnalyzer) analyzeGoModDependencies() ([]*types.Dependency, error) {
	goModPath := filepath.Join(da.rootPath, "go.mod")
	file, err := os.Open(goModPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open go.mod: %w", err)
	}
	defer func() { _ = file.Close() }()

	var dependencies []*types.Dependency
	scanner := bufio.NewScanner(file)
	inRequireBlock := false
	depRegex := regexp.MustCompile(`^\s*([^\s]+)\s+([^\s]+)(?:\s+//\s*(.*))?$`)
	blockRequireRegex := regexp.MustCompile(`^\s*require\s*\($`)
	blockEndRegex := regexp.MustCompile(`^\s*\)$`)
	singleRequireRegex := regexp.MustCompile(`^\s*require\s+([^\s]+)\s+([^\s]+)(?:\s+//\s*(.*))?$`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for single-line require
		if matches := singleRequireRegex.FindStringSubmatch(line); matches != nil {
			dep := &types.Dependency{
				Name:        matches[1],
				Version:     matches[2],
				Type:        "go-module",
				Source:      "go.mod",
				Required:    !strings.Contains(matches[2], "indirect"),
				Description: strings.TrimSpace(matches[3]),
			}
			dependencies = append(dependencies, dep)
			continue
		}

		// Check for require block start
		if blockRequireRegex.MatchString(line) {
			inRequireBlock = true
			continue
		}

		// Check for require block end
		if inRequireBlock && blockEndRegex.MatchString(line) {
			inRequireBlock = false
			continue
		}

		// Parse dependencies in require block
		if inRequireBlock {
			if matches := depRegex.FindStringSubmatch(line); matches != nil {
				dep := &types.Dependency{
					Name:        matches[1],
					Version:     matches[2],
					Type:        "go-module",
					Source:      "go.mod",
					Required:    !strings.Contains(matches[2], "indirect"),
					Description: strings.TrimSpace(matches[3]),
				}
				dependencies = append(dependencies, dep)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan go.mod: %w", err)
	}

	return dependencies, nil
}

// analyzeImportDependencies analyzes Go source files for import dependencies
func (da *DependencyAnalyzer) analyzeImportDependencies() ([]*types.Dependency, error) {
	var dependencies []*types.Dependency
	imports := make(map[string]int) // track usage count

	err := filepath.Walk(da.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process .go files, skip test files for now
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Skip vendor directories
		if strings.Contains(path, "vendor/") {
			return nil
		}

		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Parse the Go file
		file, err := parser.ParseFile(da.fileSet, path, src, parser.ParseComments)
		if err != nil {
			// Skip files that can't be parsed
			return nil
		}

		// Extract imports
		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if importPath != "" {
				imports[importPath]++
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to analyze imports: %w", err)
	}

	// Convert imports to dependencies
	for importPath, count := range imports {
		dep := &types.Dependency{
			Name:        importPath,
			Version:     "unknown",
			Type:        da.classifyImport(importPath),
			Source:      "import",
			Required:    count > 0,
			Description: fmt.Sprintf("Used in %d files", count),
		}
		dependencies = append(dependencies, dep)
	}

	return dependencies, nil
}

// classifyImport classifies an import path
func (da *DependencyAnalyzer) classifyImport(importPath string) string {
	// Standard library
	if !strings.Contains(importPath, ".") || strings.HasPrefix(importPath, "builtin") {
		return "stdlib"
	}

	// Internal/local imports
	if strings.HasPrefix(importPath, "github.com/Zerofisher/goai") {
		return "internal"
	}

	// Third-party imports
	if strings.Contains(importPath, "/") {
		return "external"
	}

	return "unknown"
}

// analyzeOtherDependencies analyzes other dependency files (package.json, requirements.txt, etc.)
func (da *DependencyAnalyzer) analyzeOtherDependencies() ([]*types.Dependency, error) {
	var dependencies []*types.Dependency

	// Check for package.json (Node.js)
	if deps, err := da.analyzePackageJSON(); err == nil {
		dependencies = append(dependencies, deps...)
	}

	// Check for requirements.txt (Python)
	if deps, err := da.analyzeRequirementsTxt(); err == nil {
		dependencies = append(dependencies, deps...)
	}

	// Check for Cargo.toml (Rust)
	if deps, err := da.analyzeCargoToml(); err == nil {
		dependencies = append(dependencies, deps...)
	}

	return dependencies, nil
}

// analyzePackageJSON analyzes package.json for Node.js dependencies
func (da *DependencyAnalyzer) analyzePackageJSON() ([]*types.Dependency, error) {
	packageJSONPath := filepath.Join(da.rootPath, "package.json")
	if _, err := os.Stat(packageJSONPath); os.IsNotExist(err) {
		return nil, err
	}

	// For now, just mark that we found a package.json
	// In a full implementation, we'd parse the JSON
	dep := &types.Dependency{
		Name:        "package.json",
		Version:     "unknown",
		Type:        "nodejs-config",
		Source:      "package.json",
		Required:    true,
		Description: "Node.js package configuration",
	}

	return []*types.Dependency{dep}, nil
}

// analyzeRequirementsTxt analyzes requirements.txt for Python dependencies
func (da *DependencyAnalyzer) analyzeRequirementsTxt() ([]*types.Dependency, error) {
	reqPath := filepath.Join(da.rootPath, "requirements.txt")
	file, err := os.Open(reqPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var dependencies []*types.Dependency
	scanner := bufio.NewScanner(file)
	depRegex := regexp.MustCompile(`^([^=<>!#\s]+)([=<>!]=?.*)?(?:\s*#.*)?$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if matches := depRegex.FindStringSubmatch(line); matches != nil {
			dep := &types.Dependency{
				Name:        matches[1],
				Version:     strings.TrimSpace(matches[2]),
				Type:        "python-package",
				Source:      "requirements.txt",
				Required:    true,
				Description: "Python package dependency",
			}
			dependencies = append(dependencies, dep)
		}
	}

	return dependencies, scanner.Err()
}

// analyzeCargoToml analyzes Cargo.toml for Rust dependencies
func (da *DependencyAnalyzer) analyzeCargoToml() ([]*types.Dependency, error) {
	cargoPath := filepath.Join(da.rootPath, "Cargo.toml")
	if _, err := os.Stat(cargoPath); os.IsNotExist(err) {
		return nil, err
	}

	// For now, just mark that we found a Cargo.toml
	// In a full implementation, we'd parse the TOML
	dep := &types.Dependency{
		Name:        "Cargo.toml",
		Version:     "unknown",
		Type:        "rust-config",
		Source:      "Cargo.toml",
		Required:    true,
		Description: "Rust package configuration",
	}

	return []*types.Dependency{dep}, nil
}

// deduplicateDependencies removes duplicate dependencies
func (da *DependencyAnalyzer) deduplicateDependencies(deps []*types.Dependency) []*types.Dependency {
	seen := make(map[string]bool)
	unique := []*types.Dependency{} // Initialize as empty slice, not nil

	for _, dep := range deps {
		key := fmt.Sprintf("%s:%s:%s", dep.Name, dep.Type, dep.Source)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, dep)
		}
	}

	return unique
}

// GetDependencyGraph builds a dependency graph for visualization
func (da *DependencyAnalyzer) GetDependencyGraph() (map[string][]string, error) {
	graph := make(map[string][]string)

	err := filepath.Walk(da.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process .go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip vendor directories
		if strings.Contains(path, "vendor/") {
			return nil
		}

		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Parse the Go file
		file, err := parser.ParseFile(da.fileSet, path, src, parser.ParseComments)
		if err != nil {
			return nil
		}

		// Get relative path for the file
		relPath, err := filepath.Rel(da.rootPath, path)
		if err != nil {
			return err
		}

		// Extract imports for this file
		var imports []string
		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if importPath != "" {
				imports = append(imports, importPath)
			}
		}

		graph[relPath] = imports
		return nil
	})

	return graph, err
}

// GetUnusedDependencies finds dependencies that might be unused
func (da *DependencyAnalyzer) GetUnusedDependencies() ([]*types.Dependency, error) {
	// Get all dependencies from go.mod
	modDeps, err := da.analyzeGoModDependencies()
	if err != nil {
		return nil, err
	}

	// Get all imports
	importDeps, err := da.analyzeImportDependencies()
	if err != nil {
		return nil, err
	}

	// Create a map of used imports
	usedImports := make(map[string]bool)
	for _, dep := range importDeps {
		if dep.Type == "external" {
			usedImports[dep.Name] = true
		}
	}

	// Find go.mod dependencies that aren't imported
	var unused []*types.Dependency
	for _, dep := range modDeps {
		if dep.Type == "go-module" && !dep.Required {
			continue // Skip indirect dependencies
		}

		// Check if this module path is used in imports
		used := false
		for importPath := range usedImports {
			if strings.HasPrefix(importPath, dep.Name) {
				used = true
				break
			}
		}

		if !used {
			unused = append(unused, dep)
		}
	}

	return unused, nil
}