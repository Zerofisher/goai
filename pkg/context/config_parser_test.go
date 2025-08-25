package context

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Zerofisher/goai/pkg/types"
)

// TestNewGOAIConfigParser tests creation of parser
func TestNewGOAIConfigParser(t *testing.T) {
	parser := NewGOAIConfigParser()

	if parser == nil {
		t.Fatal("Expected parser to not be nil")
	}

	if !parser.AllowUnknownFields {
		t.Error("Expected AllowUnknownFields to be true by default")
	}

	if parser.StrictMode {
		t.Error("Expected StrictMode to be false by default")
	}
}

// TestParseBasicConfiguration tests parsing a basic GOAI.md file
func TestParseBasicConfiguration(t *testing.T) {
	content := `# GOAI Configuration

## Project
Project Name: test-project
Language: go
Description: A test project
Version: 1.0.0
Author: Test Author
Repository: https://github.com/test/test-project
License: MIT

## Dependencies
- github.com/spf13/cobra
- github.com/go-git/go-git/v5

## Exclusions
- vendor/
- node_modules/

## Coding Style
Indent Size: 4
Use Spaces: true
Max Line Length: 120
Format On Save: true

## Testing
Framework: testing
Coverage Goal: 85.0
Require Tests: true
Test Timeout: 30s

- *_test.go
- **/*_test.go
`

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "goai_test_*.md")
	if err != nil {
		t.Fatal("Failed to create temp file:", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatal("Failed to write temp file:", err)
	}
	tmpFile.Close()

	parser := NewGOAIConfigParser()
	config, err := parser.ParseFile(tmpFile.Name())

	if err != nil {
		t.Fatal("Failed to parse config:", err)
	}

	// Test basic project info
	if config.ProjectName != "test-project" {
		t.Errorf("Expected ProjectName 'test-project', got '%s'", config.ProjectName)
	}

	if config.Language != "go" {
		t.Errorf("Expected Language 'go', got '%s'", config.Language)
	}

	if config.Description != "A test project" {
		t.Errorf("Expected Description 'A test project', got '%s'", config.Description)
	}

	if config.Version != "1.0.0" {
		t.Errorf("Expected Version '1.0.0', got '%s'", config.Version)
	}

	if config.Author != "Test Author" {
		t.Errorf("Expected Author 'Test Author', got '%s'", config.Author)
	}

	if config.Repository != "https://github.com/test/test-project" {
		t.Errorf("Expected Repository 'https://github.com/test/test-project', got '%s'", config.Repository)
	}

	if config.License != "MIT" {
		t.Errorf("Expected License 'MIT', got '%s'", config.License)
	}

	// Test dependencies
	if len(config.Dependencies) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(config.Dependencies))
	}

	expectedDeps := []string{"github.com/spf13/cobra", "github.com/go-git/go-git/v5"}
	for i, expectedDep := range expectedDeps {
		if i < len(config.Dependencies) && config.Dependencies[i] != expectedDep {
			t.Errorf("Expected dependency[%d] '%s', got '%s'", i, expectedDep, config.Dependencies[i])
		}
	}

	// Test exclusions
	if len(config.Exclusions) != 2 {
		t.Errorf("Expected 2 exclusions, got %d", len(config.Exclusions))
	}

	expectedExclusions := []string{"vendor/", "node_modules/"}
	for i, expectedExclusion := range expectedExclusions {
		if i < len(config.Exclusions) && config.Exclusions[i] != expectedExclusion {
			t.Errorf("Expected exclusion[%d] '%s', got '%s'", i, expectedExclusion, config.Exclusions[i])
		}
	}

	// Test coding style
	if config.CodingStyle == nil {
		t.Fatal("Expected CodingStyle to not be nil")
	}

	if config.CodingStyle.IndentSize != 4 {
		t.Errorf("Expected IndentSize 4, got %d", config.CodingStyle.IndentSize)
	}

	if !config.CodingStyle.UseSpaces {
		t.Error("Expected UseSpaces to be true")
	}

	if config.CodingStyle.MaxLineLength != 120 {
		t.Errorf("Expected MaxLineLength 120, got %d", config.CodingStyle.MaxLineLength)
	}

	if !config.CodingStyle.FormatOnSave {
		t.Error("Expected FormatOnSave to be true")
	}

	// Test test config
	if config.TestConfig == nil {
		t.Fatal("Expected TestConfig to not be nil")
	}

	if config.TestConfig.Framework != "testing" {
		t.Errorf("Expected Framework 'testing', got '%s'", config.TestConfig.Framework)
	}

	if config.TestConfig.CoverageGoal != 85.0 {
		t.Errorf("Expected CoverageGoal 85.0, got %f", config.TestConfig.CoverageGoal)
	}

	if !config.TestConfig.RequireTests {
		t.Error("Expected RequireTests to be true")
	}

	if config.TestConfig.TestTimeout != "30s" {
		t.Errorf("Expected TestTimeout '30s', got '%s'", config.TestConfig.TestTimeout)
	}

	if len(config.TestConfig.TestPatterns) != 2 {
		t.Errorf("Expected 2 test patterns, got %d: %v", len(config.TestConfig.TestPatterns), config.TestConfig.TestPatterns)
	}
}

// TestParseMinimalConfiguration tests parsing a minimal configuration
func TestParseMinimalConfiguration(t *testing.T) {
	content := `# GOAI Configuration

## Project
Project Name: minimal-project
Language: go
`

	tmpFile, err := os.CreateTemp("", "goai_minimal_test_*.md")
	if err != nil {
		t.Fatal("Failed to create temp file:", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatal("Failed to write temp file:", err)
	}
	tmpFile.Close()

	parser := NewGOAIConfigParser()
	config, err := parser.ParseFile(tmpFile.Name())

	if err != nil {
		t.Fatal("Failed to parse minimal config:", err)
	}

	if config.ProjectName != "minimal-project" {
		t.Errorf("Expected ProjectName 'minimal-project', got '%s'", config.ProjectName)
	}

	if config.Language != "go" {
		t.Errorf("Expected Language 'go', got '%s'", config.Language)
	}

	// Check defaults are set
	if config.CodingStyle == nil {
		t.Error("Expected default CodingStyle to be set")
	}

	if config.TestConfig == nil {
		t.Error("Expected default TestConfig to be set")
	}

	// For minimal config, should have default test pattern
	if len(config.TestConfig.TestPatterns) != 1 || config.TestConfig.TestPatterns[0] != "*_test.go" {
		t.Errorf("Expected default test pattern '*_test.go', got %d patterns: %v", len(config.TestConfig.TestPatterns), config.TestConfig.TestPatterns)
	}
}

// TestParseInvalidFile tests parsing invalid files
func TestParseInvalidFile(t *testing.T) {
	parser := NewGOAIConfigParser()

	// Test non-existent file
	_, err := parser.ParseFile("/non/existent/file.md")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test empty file
	tmpFile, err := os.CreateTemp("", "goai_empty_test_*.md")
	if err != nil {
		t.Fatal("Failed to create temp file:", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	config, err := parser.ParseFile(tmpFile.Name())
	if err != nil {
		t.Fatal("Failed to parse empty config:", err)
	}

	// Should still have defaults
	if config.CodingStyle == nil {
		t.Error("Expected default CodingStyle for empty file")
	}
}

// TestGenerateMarkdown tests markdown generation
func TestGenerateMarkdown(t *testing.T) {
	config := &types.GOAIConfig{
		ProjectName:  "test-project",
		Language:     "go",
		Description:  "Test project",
		Version:      "1.0.0",
		Author:       "Test Author",
		Repository:   "https://github.com/test/test",
		License:      "MIT",
		Dependencies: []string{"dep1", "dep2"},
		Exclusions:   []string{"vendor/", "*.tmp"},
		CodingStyle: &types.CodingStyle{
			IndentSize:    2,
			UseSpaces:     true,
			MaxLineLength: 100,
			FormatOnSave:  true,
		},
		TestConfig: &types.TestConfig{
			Framework:    "testing",
			CoverageGoal: 90.0,
			RequireTests: true,
			TestTimeout:  "60s",
			TestPatterns: []string{"*_test.go"},
		},
	}

	parser := NewGOAIConfigParser()
	markdown := parser.GenerateMarkdown(config)

	// Check that key sections are present
	expectedSections := []string{
		"# GOAI Configuration",
		"## Project",
		"Project Name: test-project",
		"Language: go",
		"Description: Test project",
		"Version: 1.0.0",
		"Author: Test Author",
		"Repository: https://github.com/test/test",
		"License: MIT",
		"## Dependencies",
		"- dep1",
		"- dep2",
		"## Exclusions",
		"- vendor/",
		"- *.tmp",
		"## Coding Style",
		"Indent Size: 2",
		"Use Spaces: true",
		"Max Line Length: 100",
		"Format On Save: true",
		"## Testing",
		"Framework: testing",
		"Coverage Goal: 90.0%",
		"Require Tests: true",
		"Test Timeout: 60s",
		"Test Patterns:",
		"- *_test.go",
	}

	for _, expected := range expectedSections {
		if !strings.Contains(markdown, expected) {
			t.Errorf("Expected markdown to contain '%s'", expected)
		}
	}
}

// TestWriteFile tests writing configuration to file
func TestWriteFile(t *testing.T) {
	config := &types.GOAIConfig{
		ProjectName: "write-test",
		Language:    "go",
		CodingStyle: &types.CodingStyle{
			IndentSize:    4,
			UseSpaces:     true,
			MaxLineLength: 120,
		},
		TestConfig: &types.TestConfig{
			Framework:    "testing",
			CoverageGoal: 80.0,
			RequireTests: true,
		},
	}

	tmpDir, err := os.MkdirTemp("", "goai_write_test_*")
	if err != nil {
		t.Fatal("Failed to create temp dir:", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "GOAI.md")
	parser := NewGOAIConfigParser()

	err = parser.WriteFile(filePath, config)
	if err != nil {
		t.Fatal("Failed to write file:", err)
	}

	// Verify file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Expected file to be created")
	}

	// Parse it back and verify
	parsedConfig, err := parser.ParseFile(filePath)
	if err != nil {
		t.Fatal("Failed to parse written file:", err)
	}

	if parsedConfig.ProjectName != config.ProjectName {
		t.Errorf("Expected ProjectName '%s', got '%s'", config.ProjectName, parsedConfig.ProjectName)
	}

	if parsedConfig.Language != config.Language {
		t.Errorf("Expected Language '%s', got '%s'", config.Language, parsedConfig.Language)
	}
}

// TestValidateConfig tests configuration validation
func TestValidateConfig(t *testing.T) {
	parser := NewGOAIConfigParser()

	// Test valid config
	validConfig := &types.GOAIConfig{
		ProjectName: "valid-project",
		Language:    "go",
		CodingStyle: &types.CodingStyle{
			IndentSize:    4,
			MaxLineLength: 120,
		},
		TestConfig: &types.TestConfig{
			CoverageGoal: 85.0,
		},
	}

	errors := parser.ValidateConfig(validConfig)
	if len(errors) > 0 {
		t.Errorf("Expected no validation errors for valid config, got %d", len(errors))
	}

	// Test invalid config
	invalidConfig := &types.GOAIConfig{
		ProjectName: "", // Empty project name
		Language:    "invalid-language",
		CodingStyle: &types.CodingStyle{
			IndentSize:    0,  // Invalid indent size
			MaxLineLength: 50, // Invalid line length
		},
		TestConfig: &types.TestConfig{
			CoverageGoal: 150.0, // Invalid coverage goal
		},
	}

	errors = parser.ValidateConfig(invalidConfig)
	if len(errors) == 0 {
		t.Error("Expected validation errors for invalid config")
	}

	// Check specific error types
	errorStrings := make([]string, len(errors))
	for i, err := range errors {
		errorStrings[i] = err.Error()
	}

	expectedErrors := []string{
		"project name is required",
		"unsupported language",
		"indent size must be between 1 and 8",
		"max line length must be between 80 and 200",
		"coverage goal must be between 0 and 100",
	}

	for _, expected := range expectedErrors {
		found := false
		for _, actual := range errorStrings {
			if strings.Contains(actual, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected validation error containing '%s'", expected)
		}
	}
}

// TestMergeConfigs tests configuration merging
func TestMergeConfigs(t *testing.T) {
	baseConfig := &types.GOAIConfig{
		ProjectName: "base-project",
		Language:    "go",
		Version:     "1.0.0",
		CodingStyle: &types.CodingStyle{
			IndentSize:    2,
			UseSpaces:     true,
			MaxLineLength: 100,
		},
		TestConfig: &types.TestConfig{
			Framework:    "testing",
			CoverageGoal: 80.0,
		},
	}

	overrideConfig := &types.GOAIConfig{
		ProjectName: "override-project",
		Description: "Override description",
		CodingStyle: &types.CodingStyle{
			IndentSize: 4, // Override indent size only
		},
		TestConfig: &types.TestConfig{
			CoverageGoal: 90.0, // Override coverage goal
		},
	}

	parser := NewGOAIConfigParser()
	merged := parser.MergeConfigs(baseConfig, overrideConfig)

	// Test overridden values
	if merged.ProjectName != "override-project" {
		t.Errorf("Expected ProjectName 'override-project', got '%s'", merged.ProjectName)
	}

	if merged.Description != "Override description" {
		t.Errorf("Expected Description 'Override description', got '%s'", merged.Description)
	}

	// Test preserved base values
	if merged.Language != "go" {
		t.Errorf("Expected Language 'go' to be preserved, got '%s'", merged.Language)
	}

	if merged.Version != "1.0.0" {
		t.Errorf("Expected Version '1.0.0' to be preserved, got '%s'", merged.Version)
	}

	// Test merged coding style
	if merged.CodingStyle.IndentSize != 4 {
		t.Errorf("Expected IndentSize 4 (overridden), got %d", merged.CodingStyle.IndentSize)
	}

	if !merged.CodingStyle.UseSpaces {
		t.Error("Expected UseSpaces true to be preserved")
	}

	if merged.CodingStyle.MaxLineLength != 100 {
		t.Errorf("Expected MaxLineLength 100 to be preserved, got %d", merged.CodingStyle.MaxLineLength)
	}

	// Test merged test config
	if merged.TestConfig.Framework != "testing" {
		t.Errorf("Expected Framework 'testing' to be preserved, got '%s'", merged.TestConfig.Framework)
	}

	if merged.TestConfig.CoverageGoal != 90.0 {
		t.Errorf("Expected CoverageGoal 90.0 (overridden), got %f", merged.TestConfig.CoverageGoal)
	}
}

// TestParseBoolValues tests boolean parsing
func TestParseBoolValues(t *testing.T) {
	parser := NewGOAIConfigParser()

	testCases := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"True", true},
		{"TRUE", true},
		{"yes", true},
		{"YES", true},
		{"1", true},
		{"on", true},
		{"ON", true},
		{"enabled", true},
		{"ENABLED", true},
		{"false", false},
		{"False", false},
		{"FALSE", false},
		{"no", false},
		{"NO", false},
		{"0", false},
		{"off", false},
		{"OFF", false},
		{"disabled", false},
		{"DISABLED", false},
		{"invalid", false},
		{"", false},
	}

	for _, tc := range testCases {
		result := parser.parseBool(tc.input)
		if result != tc.expected {
			t.Errorf("parseBool('%s') = %v, expected %v", tc.input, result, tc.expected)
		}
	}
}

// TestParseCodeBlocks tests that code blocks are properly ignored
func TestParseCodeBlocks(t *testing.T) {
	content := `# GOAI Configuration

## Project
Project Name: test-project
Language: go

## Example

` + "```go" + `
// This should be ignored
Project Name: ignored-project
Language: javascript
` + "```" + `

## Project
Description: This should be parsed
`

	tmpFile, err := os.CreateTemp("", "goai_codeblock_test_*.md")
	if err != nil {
		t.Fatal("Failed to create temp file:", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatal("Failed to write temp file:", err)
	}
	tmpFile.Close()

	parser := NewGOAIConfigParser()
	config, err := parser.ParseFile(tmpFile.Name())

	if err != nil {
		t.Fatal("Failed to parse config with code blocks:", err)
	}

	// Should parse the real config, not the code block content
	if config.ProjectName != "test-project" {
		t.Errorf("Expected ProjectName 'test-project', got '%s'", config.ProjectName)
	}

	if config.Language != "go" {
		t.Errorf("Expected Language 'go', got '%s'", config.Language)
	}

	if config.Description != "This should be parsed" {
		t.Errorf("Expected Description 'This should be parsed', got '%s'", config.Description)
	}
}