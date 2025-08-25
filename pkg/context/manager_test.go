package context

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Zerofisher/goai/pkg/types"
)

// TestNewContextManager tests creation of context manager
func TestNewContextManager(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "goai_manager_test_*")
	if err != nil {
		t.Fatal("Failed to create temp dir:", err)
	}
	defer os.RemoveAll(tmpDir)

	cm, err := NewContextManager(tmpDir)
	if err != nil {
		t.Fatal("Failed to create context manager:", err)
	}

	if cm == nil {
		t.Fatal("Expected context manager to not be nil")
	}

	if cm.workdir != tmpDir {
		t.Errorf("Expected workdir '%s', got '%s'", tmpDir, cm.workdir)
	}

	expectedConfigPath := filepath.Join(tmpDir, "GOAI.md")
	if cm.configPath != expectedConfigPath {
		t.Errorf("Expected configPath '%s', got '%s'", expectedConfigPath, cm.configPath)
	}

	if cm.structAnalyzer == nil {
		t.Error("Expected structAnalyzer to be initialized")
	}

	if cm.depAnalyzer == nil {
		t.Error("Expected depAnalyzer to be initialized")
	}
}

// TestBuildProjectContext tests building project context
func TestBuildProjectContext(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "goai_context_test_*")
	if err != nil {
		t.Fatal("Failed to create temp dir:", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some test files
	testFiles := []string{
		"main.go",
		"pkg/types/models.go",
		"pkg/utils/helpers.go",
		"cmd/app/main.go",
		"README.md",
		"go.mod",
	}

	for _, file := range testFiles {
		dir := filepath.Dir(filepath.Join(tmpDir, file))
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal("Failed to create dir:", err)
		}

		content := "// Test file content"
		if filepath.Ext(file) == ".md" {
			content = "# Test README"
		} else if filepath.Base(file) == "go.mod" {
			content = "module test-project\n\ngo 1.21\n"
		}

		if err := os.WriteFile(filepath.Join(tmpDir, file), []byte(content), 0644); err != nil {
			t.Fatal("Failed to create file:", err)
		}
	}

	// Create GOAI.md configuration
	goaiContent := `# GOAI Configuration

## Project
Project Name: test-project
Language: go
Description: Test project for context building
`

	if err := os.WriteFile(filepath.Join(tmpDir, "GOAI.md"), []byte(goaiContent), 0644); err != nil {
		t.Fatal("Failed to create GOAI.md:", err)
	}

	cm, err := NewContextManager(tmpDir)
	if err != nil {
		t.Fatal("Failed to create context manager:", err)
	}

	context, err := cm.BuildProjectContext("")
	if err != nil {
		t.Fatal("Failed to build project context:", err)
	}

	// Test basic context properties
	if context.WorkingDirectory != tmpDir {
		t.Errorf("Expected WorkingDirectory '%s', got '%s'", tmpDir, context.WorkingDirectory)
	}

	if context.LoadedAt.IsZero() {
		t.Error("Expected LoadedAt to be set")
	}

	// Test project config
	if context.ProjectConfig == nil {
		t.Fatal("Expected ProjectConfig to not be nil")
	}

	if context.ProjectConfig.ProjectName != "test-project" {
		t.Errorf("Expected ProjectName 'test-project', got '%s'", context.ProjectConfig.ProjectName)
	}

	if context.ProjectConfig.Language != "go" {
		t.Errorf("Expected Language 'go', got '%s'", context.ProjectConfig.Language)
	}

	// Test project structure
	if context.ProjectStructure == nil {
		t.Fatal("Expected ProjectStructure to not be nil")
	}

	if context.ProjectStructure.RootPath != tmpDir {
		t.Errorf("Expected RootPath '%s', got '%s'", tmpDir, context.ProjectStructure.RootPath)
	}

	// Should have found our test files
	if len(context.ProjectStructure.Files) == 0 {
		t.Error("Expected to find some files in project structure")
	}

	// Test dependencies
	if context.Dependencies == nil {
		t.Fatal("Expected Dependencies to not be nil")
	}

	// Dependencies should be an empty slice since the test go.mod has no external dependencies
	if context.Dependencies == nil {
		t.Error("Dependencies should not be nil")
	}
}

// TestLoadConfiguration tests loading GOAI configuration
func TestLoadConfiguration(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "goai_config_test_*")
	if err != nil {
		t.Fatal("Failed to create temp dir:", err)
	}
	defer os.RemoveAll(tmpDir)

	configContent := `# GOAI Configuration

## Project
Project Name: config-test
Language: go
Description: Configuration loading test
Version: 2.0.0

## Coding Style
Indent Size: 2
Use Spaces: true
Max Line Length: 100
`

	configPath := filepath.Join(tmpDir, "GOAI.md")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal("Failed to create config file:", err)
	}

	cm, err := NewContextManager(tmpDir)
	if err != nil {
		t.Fatal("Failed to create context manager:", err)
	}

	config, err := cm.LoadConfiguration(configPath)
	if err != nil {
		t.Fatal("Failed to load configuration:", err)
	}

	if config.ProjectName != "config-test" {
		t.Errorf("Expected ProjectName 'config-test', got '%s'", config.ProjectName)
	}

	if config.Language != "go" {
		t.Errorf("Expected Language 'go', got '%s'", config.Language)
	}

	if config.Description != "Configuration loading test" {
		t.Errorf("Expected Description 'Configuration loading test', got '%s'", config.Description)
	}

	if config.Version != "2.0.0" {
		t.Errorf("Expected Version '2.0.0', got '%s'", config.Version)
	}

	if config.CodingStyle == nil {
		t.Fatal("Expected CodingStyle to not be nil")
	}

	if config.CodingStyle.IndentSize != 2 {
		t.Errorf("Expected IndentSize 2, got %d", config.CodingStyle.IndentSize)
	}

	if config.CodingStyle.MaxLineLength != 100 {
		t.Errorf("Expected MaxLineLength 100, got %d", config.CodingStyle.MaxLineLength)
	}
}

// TestLoadConfigurationMissing tests loading missing configuration
func TestLoadConfigurationMissing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "goai_missing_config_test_*")
	if err != nil {
		t.Fatal("Failed to create temp dir:", err)
	}
	defer os.RemoveAll(tmpDir)

	cm, err := NewContextManager(tmpDir)
	if err != nil {
		t.Fatal("Failed to create context manager:", err)
	}

	// Should create default config when file is missing
	context, err := cm.BuildProjectContext("")
	if err != nil {
		t.Fatal("Failed to build project context:", err)
	}

	if context.ProjectConfig == nil {
		t.Fatal("Expected default ProjectConfig to be created")
	}

	expectedProjectName := filepath.Base(tmpDir)
	if context.ProjectConfig.ProjectName != expectedProjectName {
		t.Errorf("Expected ProjectName '%s', got '%s'", expectedProjectName, context.ProjectConfig.ProjectName)
	}
}

// TestDetectLanguage tests language detection
func TestDetectLanguage(t *testing.T) {
	testCases := []struct {
		name     string
		files    map[string]string
		expected string
	}{
		{
			name: "Go project with go.mod",
			files: map[string]string{
				"go.mod": "module test\n\ngo 1.21\n",
				"main.go": "package main\n",
			},
			expected: "go",
		},
		{
			name: "Go project with .go files",
			files: map[string]string{
				"main.go": "package main\n",
				"utils.go": "package utils\n",
			},
			expected: "go",
		},
		{
			name: "Node.js project",
			files: map[string]string{
				"package.json": `{"name": "test"}`,
				"index.js": "console.log('hello');",
			},
			expected: "javascript",
		},
		{
			name: "TypeScript project",
			files: map[string]string{
				"package.json": `{"name": "test"}`,
				"index.ts": "const x: string = 'hello';",
			},
			expected: "typescript",
		},
		{
			name: "Python project with requirements.txt",
			files: map[string]string{
				"requirements.txt": "requests==2.25.1\n",
				"main.py": "print('hello')",
			},
			expected: "python",
		},
		{
			name: "Python project with setup.py",
			files: map[string]string{
				"setup.py": "from setuptools import setup\n",
				"main.py": "print('hello')",
			},
			expected: "python",
		},
		{
			name: "Rust project",
			files: map[string]string{
				"Cargo.toml": "[package]\nname = \"test\"\n",
				"src/main.rs": "fn main() {}",
			},
			expected: "rust",
		},
		{
			name: "Unknown project",
			files: map[string]string{
				"README.md": "# Test Project\n",
			},
			expected: "unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "goai_lang_test_*")
			if err != nil {
				t.Fatal("Failed to create temp dir:", err)
			}
			defer os.RemoveAll(tmpDir)

			// Create test files
			for filename, content := range tc.files {
				dir := filepath.Dir(filepath.Join(tmpDir, filename))
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal("Failed to create dir:", err)
				}

				if err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644); err != nil {
					t.Fatal("Failed to create file:", err)
				}
			}

			cm, err := NewContextManager(tmpDir)
			if err != nil {
				t.Fatal("Failed to create context manager:", err)
			}

			detected := cm.detectLanguage()
			if detected != tc.expected {
				t.Errorf("Expected language '%s', got '%s'", tc.expected, detected)
			}
		})
	}
}

// TestGetWorkingDirectory tests getting working directory
func TestGetWorkingDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "goai_workdir_test_*")
	if err != nil {
		t.Fatal("Failed to create temp dir:", err)
	}
	defer os.RemoveAll(tmpDir)

	cm, err := NewContextManager(tmpDir)
	if err != nil {
		t.Fatal("Failed to create context manager:", err)
	}

	if cm.GetWorkingDirectory() != tmpDir {
		t.Errorf("Expected working directory '%s', got '%s'", tmpDir, cm.GetWorkingDirectory())
	}
}

// TestUpdateConfig tests updating configuration
func TestUpdateConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "goai_update_config_test_*")
	if err != nil {
		t.Fatal("Failed to create temp dir:", err)
	}
	defer os.RemoveAll(tmpDir)

	cm, err := NewContextManager(tmpDir)
	if err != nil {
		t.Fatal("Failed to create context manager:", err)
	}

	config := &types.GOAIConfig{
		ProjectName: "updated-project",
		Language:    "go",
		Description: "Updated project description",
		Version:     "2.1.0",
		CodingStyle: &types.CodingStyle{
			IndentSize:    4,
			UseSpaces:     true,
			MaxLineLength: 120,
		},
		TestConfig: &types.TestConfig{
			Framework:    "testing",
			CoverageGoal: 85.0,
			RequireTests: true,
		},
	}

	err = cm.UpdateConfig(config)
	if err != nil {
		t.Fatal("Failed to update config:", err)
	}

	// Verify file was written
	configPath := filepath.Join(tmpDir, "GOAI.md")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Expected config file to be created")
	}

	// Load and verify the updated config
	loadedConfig, err := cm.LoadConfiguration("")
	if err != nil {
		t.Fatal("Failed to load updated config:", err)
	}

	if loadedConfig.ProjectName != config.ProjectName {
		t.Errorf("Expected ProjectName '%s', got '%s'", config.ProjectName, loadedConfig.ProjectName)
	}

	if loadedConfig.Description != config.Description {
		t.Errorf("Expected Description '%s', got '%s'", config.Description, loadedConfig.Description)
	}
}

// TestRefreshContext tests context refreshing
func TestRefreshContext(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "goai_refresh_test_*")
	if err != nil {
		t.Fatal("Failed to create temp dir:", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create initial file
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal("Failed to create file:", err)
	}

	cm, err := NewContextManager(tmpDir)
	if err != nil {
		t.Fatal("Failed to create context manager:", err)
	}

	// Build initial context
	context1, err := cm.BuildProjectContext("")
	if err != nil {
		t.Fatal("Failed to build initial context:", err)
	}

	initialLoadTime := context1.LoadedAt

	// Add a small delay to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Refresh context
	context2, err := cm.RefreshContext()
	if err != nil {
		t.Fatal("Failed to refresh context:", err)
	}

	// Should have different load times
	if !context2.LoadedAt.After(initialLoadTime) {
		t.Error("Expected refreshed context to have later LoadedAt time")
	}

	// Should have same working directory
	if context2.WorkingDirectory != context1.WorkingDirectory {
		t.Error("Expected same working directory after refresh")
	}
}

// TestStop tests stopping the context manager
func TestStop(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "goai_stop_test_*")
	if err != nil {
		t.Fatal("Failed to create temp dir:", err)
	}
	defer os.RemoveAll(tmpDir)

	cm, err := NewContextManager(tmpDir)
	if err != nil {
		t.Fatal("Failed to create context manager:", err)
	}

	// Stop should not error even if watcher is not started
	err = cm.Stop()
	if err != nil {
		t.Error("Stop should not error when watcher is not running:", err)
	}
}

// BenchmarkBuildProjectContext benchmarks context building performance
func BenchmarkBuildProjectContext(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "goai_benchmark_*")
	if err != nil {
		b.Fatal("Failed to create temp dir:", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some test files
	testFiles := []string{
		"main.go",
		"pkg/types/models.go",
		"pkg/utils/helpers.go",
		"cmd/app/main.go",
		"internal/service/service.go",
		"api/handlers/user.go",
		"api/handlers/auth.go",
		"tests/unit/user_test.go",
		"tests/integration/api_test.go",
		"README.md",
		"go.mod",
		"go.sum",
	}

	for _, file := range testFiles {
		dir := filepath.Dir(filepath.Join(tmpDir, file))
		if err := os.MkdirAll(dir, 0755); err != nil {
			b.Fatal("Failed to create dir:", err)
		}

		content := "// Test file content\npackage main\n"
		if filepath.Ext(file) == ".md" {
			content = "# Test README\n"
		} else if filepath.Base(file) == "go.mod" {
			content = "module benchmark-test\n\ngo 1.21\n"
		}

		if err := os.WriteFile(filepath.Join(tmpDir, file), []byte(content), 0644); err != nil {
			b.Fatal("Failed to create file:", err)
		}
	}

	cm, err := NewContextManager(tmpDir)
	if err != nil {
		b.Fatal("Failed to create context manager:", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := cm.BuildProjectContext("")
		if err != nil {
			b.Fatal("Failed to build project context:", err)
		}
	}
}