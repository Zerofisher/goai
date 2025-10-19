package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test model defaults
	if cfg.Model.Provider != "openai" {
		t.Errorf("Default provider = %s, want openai", cfg.Model.Provider)
	}

	if cfg.Model.Name != "gpt-4.1-mini" {
		t.Errorf("Default model name = %s, want gpt-4.1-mini", cfg.Model.Name)
	}

	if cfg.Model.MaxTokens != 16000 {
		t.Errorf("Default max tokens = %d, want 16000", cfg.Model.MaxTokens)
	}

	// Test tools defaults
	if len(cfg.Tools.Enabled) != 5 {
		t.Errorf("Default enabled tools count = %d, want 5", len(cfg.Tools.Enabled))
	}

	if cfg.Tools.Bash.TimeoutMs != 30000 {
		t.Errorf("Default bash timeout = %d, want 30000", cfg.Tools.Bash.TimeoutMs)
	}

	// Test todo defaults
	if cfg.Todo.MaxItems != 20 {
		t.Errorf("Default max todo items = %d, want 20", cfg.Todo.MaxItems)
	}

	// Test output defaults
	if cfg.Output.Format != "markdown" {
		t.Errorf("Default output format = %s, want markdown", cfg.Output.Format)
	}

	if !cfg.Output.Colors {
		t.Error("Default colors should be enabled")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "missing provider",
			config: &Config{
				Model: ModelConfig{
					Name:   "gpt-4",
					APIKey: "test-key",
				},
				Tools: ToolsConfig{
					Enabled: []string{"bash"},
				},
				WorkDir: ".",
			},
			wantErr: true,
			errMsg:  "model provider is required",
		},
		{
			name: "missing model name",
			config: &Config{
				Model: ModelConfig{
					Provider: "openai",
					APIKey:   "test-key",
				},
				Tools: ToolsConfig{
					Enabled: []string{"bash"},
				},
				WorkDir: ".",
			},
			wantErr: true,
			errMsg:  "model name is required",
		},
		{
			name: "no tools enabled",
			config: &Config{
				Model: ModelConfig{
					Provider: "openai",
					Name:     "gpt-4",
					APIKey:   "test-key",
				},
				Tools:   ToolsConfig{},
				WorkDir: ".",
			},
			wantErr: true,
			errMsg:  "at least one tool must be enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg {
					t.Errorf("Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()

	// Test YAML format
	t.Run("YAML", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Model.Provider = "claude"
		cfg.Model.Name = "claude-3-opus"

		yamlPath := filepath.Join(tmpDir, "config.yaml")
		if err := cfg.SaveToFile(yamlPath); err != nil {
			t.Fatalf("SaveToFile failed: %v", err)
		}

		loaded, err := LoadFromFile(yamlPath)
		if err != nil {
			t.Fatalf("LoadFromFile failed: %v", err)
		}

		if loaded.Model.Provider != "claude" {
			t.Errorf("Loaded provider = %s, want claude", loaded.Model.Provider)
		}

		if loaded.Model.Name != "claude-3-opus" {
			t.Errorf("Loaded model name = %s, want claude-3-opus", loaded.Model.Name)
		}
	})

	// Test JSON format
	t.Run("JSON", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.Model.Provider = "moonshot"
		cfg.Todo.MaxItems = 50

		jsonPath := filepath.Join(tmpDir, "config.json")
		if err := cfg.SaveToFile(jsonPath); err != nil {
			t.Fatalf("SaveToFile failed: %v", err)
		}

		loaded, err := LoadFromFile(jsonPath)
		if err != nil {
			t.Fatalf("LoadFromFile failed: %v", err)
		}

		if loaded.Model.Provider != "moonshot" {
			t.Errorf("Loaded provider = %s, want moonshot", loaded.Model.Provider)
		}

		if loaded.Todo.MaxItems != 50 {
			t.Errorf("Loaded max items = %d, want 50", loaded.Todo.MaxItems)
		}
	})
}

func TestExpandEnvVar(t *testing.T) {
	// Set test environment variable
	_ = os.Setenv("TEST_API_KEY", "secret-key-123")
	defer func() { _ = os.Unsetenv("TEST_API_KEY") }()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "expand ${VAR} format",
			input: "${TEST_API_KEY}",
			want:  "secret-key-123",
		},
		{
			name:  "expand $VAR format",
			input: "$TEST_API_KEY",
			want:  "secret-key-123",
		},
		{
			name:  "non-existent var keeps original",
			input: "${NON_EXISTENT_VAR}",
			want:  "${NON_EXISTENT_VAR}",
		},
		{
			name:  "regular string unchanged",
			input: "regular-string",
			want:  "regular-string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandEnvVar(tt.input)
			if got != tt.want {
				t.Errorf("expandEnvVar(%s) = %s, want %s", tt.input, got, tt.want)
			}
		})
	}
}

func TestConfig_IsToolEnabled(t *testing.T) {
	cfg := &Config{
		Tools: ToolsConfig{
			Enabled: []string{"bash", "file", "edit"},
		},
	}

	tests := []struct {
		tool    string
		enabled bool
	}{
		{"bash", true},
		{"file", true},
		{"edit", true},
		{"todo", false},
		{"search", false},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			if got := cfg.IsToolEnabled(tt.tool); got != tt.enabled {
				t.Errorf("IsToolEnabled(%s) = %v, want %v", tt.tool, got, tt.enabled)
			}
		})
	}
}

func TestConfig_IsForbiddenCommand(t *testing.T) {
	cfg := &Config{
		Tools: ToolsConfig{
			Bash: BashConfig{
				ForbiddenCommands: []string{
					"rm -rf /",
					"shutdown",
					"reboot",
				},
			},
		},
	}

	tests := []struct {
		command   string
		forbidden bool
	}{
		{"rm -rf /", true},
		{"rm -rf /home", true}, // Contains "rm -rf /"
		{"shutdown now", true},
		{"sudo reboot", true},
		{"ls -la", false},
		{"echo hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			if got := cfg.IsForbiddenCommand(tt.command); got != tt.forbidden {
				t.Errorf("IsForbiddenCommand(%s) = %v, want %v", tt.command, got, tt.forbidden)
			}
		})
	}
}

func TestConfig_IsBlockedPath(t *testing.T) {
	cfg := &Config{
		Tools: ToolsConfig{
			File: FileConfig{
				BlockedPaths: []string{
					"/etc",
					"/sys",
					"/proc",
				},
			},
		},
	}

	tests := []struct {
		path    string
		blocked bool
	}{
		{"/etc", true},
		{"/etc/passwd", true},
		{"/sys/kernel", true},
		{"/proc/1/status", true},
		{"/home/user", false},
		{"/tmp/test", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := cfg.IsBlockedPath(tt.path); got != tt.blocked {
				t.Errorf("IsBlockedPath(%s) = %v, want %v", tt.path, got, tt.blocked)
			}
		})
	}
}

func TestConfig_IsAllowedExtension(t *testing.T) {
	tests := []struct {
		name       string
		extensions []string
		filename   string
		allowed    bool
	}{
		{
			name:       "no restrictions allows all",
			extensions: []string{},
			filename:   "any.file",
			allowed:    true,
		},
		{
			name:       "allowed extension",
			extensions: []string{".go", ".txt", ".md"},
			filename:   "test.go",
			allowed:    true,
		},
		{
			name:       "extension without dot",
			extensions: []string{"go", "txt", "md"},
			filename:   "test.go",
			allowed:    true,
		},
		{
			name:       "disallowed extension",
			extensions: []string{".go", ".txt"},
			filename:   "test.py",
			allowed:    false,
		},
		{
			name:       "no extension",
			extensions: []string{".go", ".txt"},
			filename:   "README",
			allowed:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Tools: ToolsConfig{
					File: FileConfig{
						AllowedExtensions: tt.extensions,
					},
				},
			}

			if got := cfg.IsAllowedExtension(tt.filename); got != tt.allowed {
				t.Errorf("IsAllowedExtension(%s) = %v, want %v", tt.filename, got, tt.allowed)
			}
		})
	}
}

func TestConfig_ShouldExcludePattern(t *testing.T) {
	cfg := &Config{
		Tools: ToolsConfig{
			Search: SearchConfig{
				ExcludePatterns: []string{
					"*.pyc",
					"__pycache__",
					"node_modules",
					".git",
				},
			},
		},
	}

	tests := []struct {
		path     string
		excluded bool
	}{
		{"test.pyc", true},
		{"/path/to/file.pyc", true},
		{"__pycache__/test.py", true},
		{"node_modules/package/index.js", true},
		{".git/config", true},
		{"test.py", false},
		{"src/main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := cfg.ShouldExcludePattern(tt.path); got != tt.excluded {
				t.Errorf("ShouldExcludePattern(%s) = %v, want %v", tt.path, got, tt.excluded)
			}
		})
	}
}

func TestLoadFromFile_NonExistent(t *testing.T) {
	// Should return default config for non-existent file
	cfg, err := LoadFromFile("/non/existent/path/config.yaml")
	if err != nil {
		t.Fatalf("LoadFromFile should return default config for non-existent file, got error: %v", err)
	}

	if cfg == nil {
		t.Fatal("LoadFromFile returned nil config")
		return // Early return for static analysis
	}

	// Should have default values
	if cfg.Model.Provider != "openai" {
		t.Errorf("Expected default provider, got %s", cfg.Model.Provider)
	}
}

func TestConfig_EnvironmentExpansion(t *testing.T) {
	// Set test environment variables
	_ = os.Setenv("TEST_API_KEY", "env-api-key")
	_ = os.Setenv("TEST_WORKDIR", "/test/work")
	defer func() { _ = os.Unsetenv("TEST_API_KEY") }()
	defer func() { _ = os.Unsetenv("TEST_WORKDIR") }()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	// Create a config with environment variables
	configContent := `
model:
  provider: openai
  name: gpt-4
  api_key: ${TEST_API_KEY}
work_dir: ${TEST_WORKDIR}
tools:
  enabled:
    - bash
  bash:
    allowed_directories:
      - ${TEST_WORKDIR}/allowed
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := LoadFromFile(configPath)
	if err != nil {
		// If it fails because the work directory doesn't exist, that's expected
		if !strings.Contains(err.Error(), "work directory does not exist") {
			t.Fatalf("LoadFromFile failed: %v", err)
		}
		// Create the directory and try again
		_ = os.MkdirAll("/test/work", 0755)
		cfg, err = LoadFromFile(configPath)
		if err != nil {
			t.Skipf("Skipping test due to directory permissions: %v", err)
		}
	}

	if cfg.Model.APIKey != "env-api-key" {
		t.Errorf("API key not expanded: got %s, want env-api-key", cfg.Model.APIKey)
	}
}