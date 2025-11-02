package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration for GoAI Coder.
type Config struct {
	Model   ModelConfig  `yaml:"model" json:"model"`
	Tools   ToolsConfig  `yaml:"tools" json:"tools"`
	Todo    TodoConfig   `yaml:"todo" json:"todo"`
	Output  OutputConfig `yaml:"output" json:"output"`
	WorkDir string       `yaml:"work_dir" json:"work_dir"`
	Debug   bool         `yaml:"debug" json:"debug"`
}

// ModelConfig contains LLM model configuration.
type ModelConfig struct {
	Provider     string `yaml:"provider" json:"provider"`           // "openai", "claude", "moonshot", etc.
	Name         string `yaml:"name" json:"name"`                   // Model name e.g., "gpt-4", "claude-3-opus"
	APIKey       string `yaml:"api_key" json:"api_key"`             // API key (can use ${ENV_VAR} syntax)
	BaseURL      string `yaml:"base_url" json:"base_url"`           // Optional custom base URL
	MaxTokens    int    `yaml:"max_tokens" json:"max_tokens"`       // Maximum tokens for LLM response
	Timeout      int    `yaml:"timeout" json:"timeout"`             // Request timeout in seconds
	SystemPrompt string `yaml:"system_prompt" json:"system_prompt"` // Optional system prompt override
}

// ToolsConfig contains tools configuration.
type ToolsConfig struct {
	Enabled []string     `yaml:"enabled" json:"enabled"` // List of enabled tools
	Bash    BashConfig   `yaml:"bash" json:"bash"`
	File    FileConfig   `yaml:"file" json:"file"`
	Edit    EditConfig   `yaml:"edit" json:"edit"`
	Search  SearchConfig `yaml:"search" json:"search"`
}

// BashConfig contains bash tool configuration.
type BashConfig struct {
	TimeoutMs          int      `yaml:"timeout_ms" json:"timeout_ms"`                   // Command timeout in milliseconds
	ForbiddenCommands  []string `yaml:"forbidden_commands" json:"forbidden_commands"`   // Commands to block
	AllowedDirectories []string `yaml:"allowed_directories" json:"allowed_directories"` // Directories where commands can run
	MaxOutputChars     int      `yaml:"max_output_chars" json:"max_output_chars"`       // Maximum output characters
	EnableSudo         bool     `yaml:"enable_sudo" json:"enable_sudo"`                 // Whether to allow sudo (dangerous!)
}

// FileConfig contains file operation configuration.
type FileConfig struct {
	MaxFileSize       int      `yaml:"max_file_size" json:"max_file_size"`           // Maximum file size in bytes
	MaxListFiles      int      `yaml:"max_list_files" json:"max_list_files"`         // Maximum files to list
	AllowedExtensions []string `yaml:"allowed_extensions" json:"allowed_extensions"` // If set, only these extensions allowed
	BlockedPaths      []string `yaml:"blocked_paths" json:"blocked_paths"`           // Paths to block access to
}

// EditConfig contains edit tool configuration.
type EditConfig struct {
	CreateBackup    bool `yaml:"create_backup" json:"create_backup"`       // Create backup before editing
	MaxEditSize     int  `yaml:"max_edit_size" json:"max_edit_size"`       // Maximum size for files to edit
	PreserveMode    bool `yaml:"preserve_mode" json:"preserve_mode"`       // Preserve file permissions
	ValidateContent bool `yaml:"validate_content" json:"validate_content"` // Validate content before writing
}

// SearchConfig contains search tool configuration.
type SearchConfig struct {
	MaxResults      int      `yaml:"max_results" json:"max_results"`           // Maximum search results
	IncludeHidden   bool     `yaml:"include_hidden" json:"include_hidden"`     // Include hidden files
	ExcludePatterns []string `yaml:"exclude_patterns" json:"exclude_patterns"` // Patterns to exclude
	SearchTypes     []string `yaml:"search_types" json:"search_types"`         // Types of search: code, text, symbol
	CaseSensitive   bool     `yaml:"case_sensitive" json:"case_sensitive"`     // Case sensitive search
}

// TodoConfig contains todo management configuration.
type TodoConfig struct {
	MaxItems           int  `yaml:"max_items" json:"max_items"`                       // Maximum todo items
	ReminderInterval   int  `yaml:"reminder_interval" json:"reminder_interval"`       // Rounds between reminders
	AutoCleanCompleted bool `yaml:"auto_clean_completed" json:"auto_clean_completed"` // Auto remove completed items
	ShowProgress       bool `yaml:"show_progress" json:"show_progress"`               // Show progress statistics
}

// OutputConfig contains output formatting configuration.
type OutputConfig struct {
	MaxChars           int    `yaml:"max_chars" json:"max_chars"`                         // Maximum output characters
	Format             string `yaml:"format" json:"format"`                               // Output format: "markdown", "plain"
	Colors             bool   `yaml:"colors" json:"colors"`                               // Enable colored output
	ShowTimestamp      bool   `yaml:"show_timestamp" json:"show_timestamp"`               // Show timestamps
	CodeTheme          string `yaml:"code_theme" json:"code_theme"`                       // Syntax highlighting theme
	WrapLines          bool   `yaml:"wrap_lines" json:"wrap_lines"`                       // Wrap long lines
	ShowSpinner        bool   `yaml:"show_spinner" json:"show_spinner"`                   // Show spinner during processing
	ShowToolEvents     bool   `yaml:"show_tool_events" json:"show_tool_events"`           // Show tool execution events
	ToolEventDetail    string `yaml:"tool_event_detail" json:"tool_event_detail"`         // Tool event detail level: "compact", "full"
	ToolOutputMaxLines int    `yaml:"tool_output_max_lines" json:"tool_output_max_lines"` // Maximum tool output lines
	ToolOutputMaxChars int    `yaml:"tool_output_max_chars" json:"tool_output_max_chars"` // Maximum tool output characters
	ToolOutputFormat   string `yaml:"tool_output_format" json:"tool_output_format"`       // Tool output format: "auto", "plain", "json"
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{
		Model: ModelConfig{
			Provider:  "openai",
			Name:      "gpt-4.1-mini",
			APIKey:    "${OPENAI_API_KEY}",
			MaxTokens: 16000,
			Timeout:   60,
		},
		Tools: ToolsConfig{
			Enabled: []string{"bash", "file", "edit", "todo", "search"},
			Bash: BashConfig{
				TimeoutMs: 30000,
				ForbiddenCommands: []string{
					"rm -rf /",
					"shutdown",
					"reboot",
					"mkfs",
					"dd if=/dev/zero",
					":(){ :|:& };:", // Fork bomb
				},
				MaxOutputChars: 100000,
				EnableSudo:     false,
			},
			File: FileConfig{
				MaxFileSize:  10 * 1024 * 1024, // 10MB
				MaxListFiles: 1000,
			},
			Edit: EditConfig{
				CreateBackup:    true,
				MaxEditSize:     5 * 1024 * 1024, // 5MB
				PreserveMode:    true,
				ValidateContent: true,
			},
			Search: SearchConfig{
				MaxResults:    100,
				IncludeHidden: false,
				ExcludePatterns: []string{
					"*.pyc",
					"__pycache__",
					"node_modules",
					".git",
					"*.min.js",
					"*.min.css",
				},
				SearchTypes:   []string{"code", "text"},
				CaseSensitive: false,
			},
		},
		Todo: TodoConfig{
			MaxItems:           20,
			ReminderInterval:   10,
			AutoCleanCompleted: false,
			ShowProgress:       true,
		},
		Output: OutputConfig{
			MaxChars:           100000,
			Format:             "markdown",
			Colors:             true,
			ShowTimestamp:      false,
			CodeTheme:          "monokai",
			WrapLines:          true,
			ShowSpinner:        true,
			ShowToolEvents:     true,
			ToolEventDetail:    "compact",
			ToolOutputMaxLines: 200,
			ToolOutputMaxChars: 20000,
			ToolOutputFormat:   "auto",
		},
		WorkDir: ".",
		Debug:   false,
	}
}

// LoadFromFile loads configuration from a YAML or JSON file.
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := DefaultConfig() // Start with defaults

	// Determine format by extension
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, cfg); err != nil {
			if err := json.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config as YAML or JSON: %w", err)
			}
		}
	}

	// Process environment variables
	cfg.expandEnvVars()

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// SaveToFile saves the configuration to a file.
func (c *Config) SaveToFile(path string) error {
	// Determine format by extension
	ext := strings.ToLower(filepath.Ext(path))

	var data []byte
	var err error

	switch ext {
	case ".json":
		data, err = json.MarshalIndent(c, "", "  ")
	default: // Default to YAML
		data, err = yaml.Marshal(c)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// expandEnvVars expands environment variables in string fields.
func (c *Config) expandEnvVars() {
	// Expand API key
	c.Model.APIKey = expandEnvVar(c.Model.APIKey)
	c.Model.BaseURL = expandEnvVar(c.Model.BaseURL)

	// Expand work directory
	c.WorkDir = expandEnvVar(c.WorkDir)

	// Expand allowed directories for bash
	for i, dir := range c.Tools.Bash.AllowedDirectories {
		c.Tools.Bash.AllowedDirectories[i] = expandEnvVar(dir)
	}

	// Expand blocked paths for file operations
	for i, path := range c.Tools.File.BlockedPaths {
		c.Tools.File.BlockedPaths[i] = expandEnvVar(path)
	}
}

// expandEnvVar expands a single environment variable reference.
func expandEnvVar(s string) string {
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		envVar := s[2 : len(s)-1]
		if val := os.Getenv(envVar); val != "" {
			return val
		}
		// Keep original if env var not found
		return s
	}
	// Also expand $VAR format
	return os.ExpandEnv(s)
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	// Validate model configuration
	if c.Model.Provider == "" {
		return fmt.Errorf("model provider is required")
	}

	if c.Model.Name == "" {
		return fmt.Errorf("model name is required")
	}

	if c.Model.APIKey == "" && !strings.HasPrefix(c.Model.APIKey, "${") {
		return fmt.Errorf("model API key is required")
	}

	if c.Model.MaxTokens <= 0 {
		c.Model.MaxTokens = 16000 // Set default if invalid
	}

	if c.Model.Timeout <= 0 {
		c.Model.Timeout = 60 // Set default timeout
	}

	// Validate tools configuration
	if len(c.Tools.Enabled) == 0 {
		return fmt.Errorf("at least one tool must be enabled")
	}

	// Validate bash configuration
	if c.Tools.Bash.TimeoutMs <= 0 {
		c.Tools.Bash.TimeoutMs = 30000
	}

	if c.Tools.Bash.MaxOutputChars <= 0 {
		c.Tools.Bash.MaxOutputChars = 100000
	}

	// Validate file configuration
	if c.Tools.File.MaxFileSize <= 0 {
		c.Tools.File.MaxFileSize = 10 * 1024 * 1024
	}

	if c.Tools.File.MaxListFiles <= 0 {
		c.Tools.File.MaxListFiles = 1000
	}

	// Validate todo configuration
	if c.Todo.MaxItems <= 0 {
		c.Todo.MaxItems = 20
	}

	if c.Todo.ReminderInterval <= 0 {
		c.Todo.ReminderInterval = 10
	}

	// Validate output configuration
	if c.Output.MaxChars <= 0 {
		c.Output.MaxChars = 100000
	}

	if c.Output.Format == "" {
		c.Output.Format = "markdown"
	}

	// Validate tool event configuration
	if c.Output.ToolEventDetail == "" {
		c.Output.ToolEventDetail = "compact"
	} else if c.Output.ToolEventDetail != "compact" && c.Output.ToolEventDetail != "full" {
		c.Output.ToolEventDetail = "compact" // Default to compact if invalid
	}

	if c.Output.ToolOutputMaxLines <= 0 {
		c.Output.ToolOutputMaxLines = 200
	}

	if c.Output.ToolOutputMaxChars <= 0 {
		c.Output.ToolOutputMaxChars = 20000
	}

	if c.Output.ToolOutputFormat == "" {
		c.Output.ToolOutputFormat = "auto"
	} else if c.Output.ToolOutputFormat != "auto" && c.Output.ToolOutputFormat != "plain" && c.Output.ToolOutputFormat != "json" {
		c.Output.ToolOutputFormat = "auto" // Default to auto if invalid
	}

	// Validate work directory
	if c.WorkDir == "" {
		c.WorkDir = "."
	}

	// Resolve work directory to absolute path
	absPath, err := filepath.Abs(c.WorkDir)
	if err != nil {
		return fmt.Errorf("invalid work directory: %w", err)
	}
	c.WorkDir = absPath

	// Check if work directory exists
	if info, err := os.Stat(c.WorkDir); err != nil {
		return fmt.Errorf("work directory does not exist: %w", err)
	} else if !info.IsDir() {
		return fmt.Errorf("work directory is not a directory: %s", c.WorkDir)
	}

	return nil
}

// IsToolEnabled checks if a tool is enabled.
func (c *Config) IsToolEnabled(toolName string) bool {
	for _, enabled := range c.Tools.Enabled {
		if enabled == toolName {
			return true
		}
	}
	return false
}

// GetTimeout returns the timeout duration for the model.
func (c *Config) GetTimeout() time.Duration {
	return time.Duration(c.Model.Timeout) * time.Second
}

// GetBashTimeout returns the timeout duration for bash commands.
func (c *Config) GetBashTimeout() time.Duration {
	return time.Duration(c.Tools.Bash.TimeoutMs) * time.Millisecond
}

// IsForbiddenCommand checks if a command is forbidden.
func (c *Config) IsForbiddenCommand(command string) bool {
	for _, forbidden := range c.Tools.Bash.ForbiddenCommands {
		if strings.Contains(command, forbidden) {
			return true
		}
	}
	return false
}

// IsBlockedPath checks if a path is blocked.
func (c *Config) IsBlockedPath(path string) bool {
	for _, blocked := range c.Tools.File.BlockedPaths {
		// Check exact match or if path starts with blocked path
		if path == blocked || strings.HasPrefix(path, blocked+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

// IsAllowedExtension checks if a file extension is allowed.
func (c *Config) IsAllowedExtension(filename string) bool {
	// If no restrictions, allow all
	if len(c.Tools.File.AllowedExtensions) == 0 {
		return true
	}

	ext := filepath.Ext(filename)
	for _, allowed := range c.Tools.File.AllowedExtensions {
		if ext == allowed || ext == "."+allowed {
			return true
		}
	}
	return false
}

// ShouldExcludePattern checks if a pattern should be excluded from search.
func (c *Config) ShouldExcludePattern(path string) bool {
	for _, pattern := range c.Tools.Search.ExcludePatterns {
		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if matched {
			return true
		}
		// Also check if path contains the pattern (for directories)
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}
