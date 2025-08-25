package context

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Zerofisher/goai/pkg/types"
)

// GOAIConfigParser handles parsing and writing GOAI.md configuration files
type GOAIConfigParser struct {
	// Configuration for parsing behavior
	AllowUnknownFields bool
	StrictMode         bool
}

// NewGOAIConfigParser creates a new configuration parser
func NewGOAIConfigParser() *GOAIConfigParser {
	return &GOAIConfigParser{
		AllowUnknownFields: true,
		StrictMode:         false,
	}
}

// ParseFile parses a GOAI.md configuration file
func (p *GOAIConfigParser) ParseFile(filePath string) (*types.GOAIConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	return p.Parse(file)
}

// Parse parses GOAI.md configuration from a reader
func (p *GOAIConfigParser) Parse(file *os.File) (*types.GOAIConfig, error) {
	config := &types.GOAIConfig{
		CodingStyle: &types.CodingStyle{
			IndentSize:    4,
			UseSpaces:     true,
			MaxLineLength: 120,
		},
		TestConfig: &types.TestConfig{
			Framework:     "testing",
			TestPatterns:  []string{},
			CoverageGoal:  80.0,
			RequireTests:  true,
		},
	}

	scanner := bufio.NewScanner(file)
	currentSection := ""
	inCodeBlock := false
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines
		if line == "" {
			continue
		}

		// Handle code blocks
		if strings.HasPrefix(line, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		
		// Skip content inside code blocks
		if inCodeBlock {
			continue
		}

		// Parse headers
		if strings.HasPrefix(line, "#") {
			currentSection = p.parseHeader(line)
			continue
		}


		// Parse configuration based on current section
		if err := p.parseConfigLine(config, currentSection, line); err != nil && p.StrictMode {
			return nil, fmt.Errorf("failed to parse line '%s': %w", line, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan file: %w", err)
	}

	// Add default test patterns if none were specified
	if config.TestConfig != nil && len(config.TestConfig.TestPatterns) == 0 {
		config.TestConfig.TestPatterns = []string{"*_test.go"}
	}

	return config, nil
}

// parseHeader extracts section name from markdown header
func (p *GOAIConfigParser) parseHeader(line string) string {
	// Remove # symbols and convert to lowercase
	header := strings.TrimSpace(strings.TrimLeft(line, "#"))
	return strings.ToLower(strings.ReplaceAll(header, " ", "_"))
}

// parseConfigLine parses a configuration line based on the current section
func (p *GOAIConfigParser) parseConfigLine(config *types.GOAIConfig, section, line string) error {
	// Skip markdown list items and comments
	if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") || strings.HasPrefix(line, "<!--") {
		return p.parseListItem(config, section, line)
	}

	// Parse key-value pairs
	if strings.Contains(line, ":") {
		// Handle special cases like "Test Patterns:" which introduce a following list
		if strings.Contains(strings.ToLower(line), "test patterns:") && strings.TrimSpace(strings.Split(line, ":")[1]) == "" {
			// This is just a section header for patterns, don't parse as key-value
			return nil
		}
		return p.parseKeyValue(config, section, line)
	}

	return nil
}

// parseListItem parses markdown list items
func (p *GOAIConfigParser) parseListItem(config *types.GOAIConfig, section, line string) error {
	// Remove list markers
	content := strings.TrimSpace(strings.TrimLeft(line, "-*"))
	
	// Skip empty content
	if content == "" {
		return nil
	}
	
	switch section {
	case "dependencies", "requirements":
		if config.Dependencies == nil {
			config.Dependencies = []string{}
		}
		config.Dependencies = append(config.Dependencies, content)
		
	case "exclusions", "ignore":
		if config.Exclusions == nil {
			config.Exclusions = []string{}
		}
		config.Exclusions = append(config.Exclusions, content)
		
	case "patterns", "test_patterns", "testing":
		// Handle Test Patterns section - only add if content doesn't look like a header
		if !strings.Contains(strings.ToLower(content), "test patterns") {
			if config.TestConfig == nil {
				config.TestConfig = &types.TestConfig{TestPatterns: []string{}}
			}
			config.TestConfig.TestPatterns = append(config.TestConfig.TestPatterns, content)
		}
	}

	return nil
}

// parseKeyValue parses key-value configuration lines
func (p *GOAIConfigParser) parseKeyValue(config *types.GOAIConfig, section, line string) error {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return nil
	}

	key := strings.TrimSpace(strings.ToLower(strings.ReplaceAll(parts[0], " ", "_")))
	value := strings.TrimSpace(parts[1])

	switch section {
	case "project", "general", "":
		return p.parseProjectConfig(config, key, value)
	case "coding_style", "style":
		return p.parseCodingStyle(config, key, value)
	case "testing", "test":
		return p.parseTestConfig(config, key, value)
	default:
		if !p.AllowUnknownFields {
			return fmt.Errorf("unknown section: %s", section)
		}
	}

	return nil
}

// parseProjectConfig parses general project configuration
func (p *GOAIConfigParser) parseProjectConfig(config *types.GOAIConfig, key, value string) error {
	switch key {
	case "project_name", "name":
		config.ProjectName = value
	case "language", "lang":
		config.Language = strings.ToLower(value)
	case "description", "desc":
		config.Description = value
	case "version", "ver":
		config.Version = value
	case "author", "maintainer":
		config.Author = value
	case "repository", "repo", "url":
		config.Repository = value
	case "license":
		config.License = value
	default:
		if !p.AllowUnknownFields {
			return fmt.Errorf("unknown project config key: %s", key)
		}
	}
	return nil
}

// parseCodingStyle parses coding style configuration
func (p *GOAIConfigParser) parseCodingStyle(config *types.GOAIConfig, key, value string) error {
	if config.CodingStyle == nil {
		config.CodingStyle = &types.CodingStyle{}
	}

	switch key {
	case "indent_size", "indent":
		if size, err := strconv.Atoi(value); err == nil {
			config.CodingStyle.IndentSize = size
		}
	case "use_spaces", "spaces":
		config.CodingStyle.UseSpaces = p.parseBool(value)
	case "max_line_length", "line_length", "max_length":
		if length, err := strconv.Atoi(value); err == nil {
			config.CodingStyle.MaxLineLength = length
		}
	case "format_on_save", "auto_format":
		config.CodingStyle.FormatOnSave = p.parseBool(value)
	default:
		if !p.AllowUnknownFields {
			return fmt.Errorf("unknown coding style key: %s", key)
		}
	}
	return nil
}

// parseTestConfig parses testing configuration
func (p *GOAIConfigParser) parseTestConfig(config *types.GOAIConfig, key, value string) error {
	if config.TestConfig == nil {
		config.TestConfig = &types.TestConfig{}
	}

	switch key {
	case "framework", "test_framework":
		config.TestConfig.Framework = value
	case "coverage_goal", "coverage", "target_coverage":
		if coverage, err := strconv.ParseFloat(value, 64); err == nil {
			config.TestConfig.CoverageGoal = coverage
		}
	case "require_tests", "mandatory_tests":
		config.TestConfig.RequireTests = p.parseBool(value)
	case "test_timeout", "timeout":
		config.TestConfig.TestTimeout = value
	default:
		if !p.AllowUnknownFields {
			return fmt.Errorf("unknown test config key: %s", key)
		}
	}
	return nil
}

// parseBool parses boolean values from strings
func (p *GOAIConfigParser) parseBool(value string) bool {
	switch strings.ToLower(value) {
	case "true", "yes", "1", "on", "enabled":
		return true
	case "false", "no", "0", "off", "disabled":
		return false
	default:
		return false
	}
}

// WriteFile writes a GOAI configuration to a file
func (p *GOAIConfigParser) WriteFile(filePath string, config *types.GOAIConfig) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	content := p.GenerateMarkdown(config)
	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GenerateMarkdown generates GOAI.md content from configuration
func (p *GOAIConfigParser) GenerateMarkdown(config *types.GOAIConfig) string {
	var sb strings.Builder

	// Header
	sb.WriteString("# GOAI Configuration\n\n")
	sb.WriteString("This file contains configuration settings for the GOAI coder.\n\n")

	// Project Information
	sb.WriteString("## Project\n\n")
	if config.ProjectName != "" {
		sb.WriteString(fmt.Sprintf("Project Name: %s\n", config.ProjectName))
	}
	if config.Language != "" {
		sb.WriteString(fmt.Sprintf("Language: %s\n", config.Language))
	}
	if config.Description != "" {
		sb.WriteString(fmt.Sprintf("Description: %s\n", config.Description))
	}
	if config.Version != "" {
		sb.WriteString(fmt.Sprintf("Version: %s\n", config.Version))
	}
	if config.Author != "" {
		sb.WriteString(fmt.Sprintf("Author: %s\n", config.Author))
	}
	if config.Repository != "" {
		sb.WriteString(fmt.Sprintf("Repository: %s\n", config.Repository))
	}
	if config.License != "" {
		sb.WriteString(fmt.Sprintf("License: %s\n", config.License))
	}
	sb.WriteString("\n")

	// Dependencies
	if len(config.Dependencies) > 0 {
		sb.WriteString("## Dependencies\n\n")
		for _, dep := range config.Dependencies {
			sb.WriteString(fmt.Sprintf("- %s\n", dep))
		}
		sb.WriteString("\n")
	}

	// Exclusions
	if len(config.Exclusions) > 0 {
		sb.WriteString("## Exclusions\n\n")
		for _, exclusion := range config.Exclusions {
			sb.WriteString(fmt.Sprintf("- %s\n", exclusion))
		}
		sb.WriteString("\n")
	}

	// Coding Style
	if config.CodingStyle != nil {
		sb.WriteString("## Coding Style\n\n")
		sb.WriteString(fmt.Sprintf("Indent Size: %d\n", config.CodingStyle.IndentSize))
		sb.WriteString(fmt.Sprintf("Use Spaces: %t\n", config.CodingStyle.UseSpaces))
		sb.WriteString(fmt.Sprintf("Max Line Length: %d\n", config.CodingStyle.MaxLineLength))
		sb.WriteString(fmt.Sprintf("Format On Save: %t\n", config.CodingStyle.FormatOnSave))
		sb.WriteString("\n")
	}

	// Test Configuration
	if config.TestConfig != nil {
		sb.WriteString("## Testing\n\n")
		sb.WriteString(fmt.Sprintf("Framework: %s\n", config.TestConfig.Framework))
		sb.WriteString(fmt.Sprintf("Coverage Goal: %.1f%%\n", config.TestConfig.CoverageGoal))
		sb.WriteString(fmt.Sprintf("Require Tests: %t\n", config.TestConfig.RequireTests))
		if config.TestConfig.TestTimeout != "" {
			sb.WriteString(fmt.Sprintf("Test Timeout: %s\n", config.TestConfig.TestTimeout))
		}
		
		if len(config.TestConfig.TestPatterns) > 0 {
			sb.WriteString("\nTest Patterns:\n")
			for _, pattern := range config.TestConfig.TestPatterns {
				sb.WriteString(fmt.Sprintf("- %s\n", pattern))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// ValidateConfig validates a parsed configuration
func (p *GOAIConfigParser) ValidateConfig(config *types.GOAIConfig) []error {
	var errors []error

	// Validate project name
	if config.ProjectName == "" {
		errors = append(errors, fmt.Errorf("project name is required"))
	} else {
		// Check project name format
		projectNameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
		if !projectNameRegex.MatchString(config.ProjectName) {
			errors = append(errors, fmt.Errorf("project name contains invalid characters"))
		}
	}

	// Validate language
	if config.Language != "" {
		supportedLangs := []string{"go", "javascript", "typescript", "python", "java", "rust", "cpp", "c"}
		supported := false
		for _, lang := range supportedLangs {
			if strings.ToLower(config.Language) == lang {
				supported = true
				break
			}
		}
		if !supported {
			errors = append(errors, fmt.Errorf("unsupported language: %s", config.Language))
		}
	}

	// Validate coding style
	if config.CodingStyle != nil {
		if config.CodingStyle.IndentSize < 1 || config.CodingStyle.IndentSize > 8 {
			errors = append(errors, fmt.Errorf("indent size must be between 1 and 8"))
		}
		if config.CodingStyle.MaxLineLength < 80 || config.CodingStyle.MaxLineLength > 200 {
			errors = append(errors, fmt.Errorf("max line length must be between 80 and 200"))
		}
	}

	// Validate test configuration
	if config.TestConfig != nil {
		if config.TestConfig.CoverageGoal < 0 || config.TestConfig.CoverageGoal > 100 {
			errors = append(errors, fmt.Errorf("coverage goal must be between 0 and 100"))
		}
	}

	return errors
}

// MergeConfigs merges two configurations, with the second one taking precedence
func (p *GOAIConfigParser) MergeConfigs(base, override *types.GOAIConfig) *types.GOAIConfig {
	result := *base // Copy base config

	if override.ProjectName != "" {
		result.ProjectName = override.ProjectName
	}
	if override.Language != "" {
		result.Language = override.Language
	}
	if override.Description != "" {
		result.Description = override.Description
	}
	if override.Version != "" {
		result.Version = override.Version
	}
	if override.Author != "" {
		result.Author = override.Author
	}
	if override.Repository != "" {
		result.Repository = override.Repository
	}
	if override.License != "" {
		result.License = override.License
	}

	if len(override.Dependencies) > 0 {
		result.Dependencies = override.Dependencies
	}
	if len(override.Exclusions) > 0 {
		result.Exclusions = override.Exclusions
	}

	// Merge coding style
	if override.CodingStyle != nil {
		if result.CodingStyle == nil {
			result.CodingStyle = &types.CodingStyle{}
		}
		if override.CodingStyle.IndentSize > 0 {
			result.CodingStyle.IndentSize = override.CodingStyle.IndentSize
		}
		if override.CodingStyle.MaxLineLength > 0 {
			result.CodingStyle.MaxLineLength = override.CodingStyle.MaxLineLength
		}
		// Copy string fields if they're not empty
		if override.CodingStyle.NamingStyle != "" {
			result.CodingStyle.NamingStyle = override.CodingStyle.NamingStyle
		}
		if override.CodingStyle.CommentStyle != "" {
			result.CodingStyle.CommentStyle = override.CodingStyle.CommentStyle
		}
		// Copy map if it has entries
		if len(override.CodingStyle.CustomRules) > 0 {
			result.CodingStyle.CustomRules = override.CodingStyle.CustomRules
		}
		// For boolean fields, only copy if they're explicitly true
		// This preserves the base value when override is false (the zero value)
		if override.CodingStyle.UseSpaces {
			result.CodingStyle.UseSpaces = override.CodingStyle.UseSpaces
		}
		if override.CodingStyle.FormatOnSave {
			result.CodingStyle.FormatOnSave = override.CodingStyle.FormatOnSave
		}
	}

	// Merge test config
	if override.TestConfig != nil {
		if result.TestConfig == nil {
			result.TestConfig = &types.TestConfig{}
		}
		if override.TestConfig.Framework != "" {
			result.TestConfig.Framework = override.TestConfig.Framework
		}
		if override.TestConfig.CoverageGoal > 0 {
			result.TestConfig.CoverageGoal = override.TestConfig.CoverageGoal
		}
		if override.TestConfig.TestTimeout != "" {
			result.TestConfig.TestTimeout = override.TestConfig.TestTimeout
		}
		if len(override.TestConfig.TestPatterns) > 0 {
			result.TestConfig.TestPatterns = override.TestConfig.TestPatterns
		}
		result.TestConfig.RequireTests = override.TestConfig.RequireTests
	}

	return &result
}