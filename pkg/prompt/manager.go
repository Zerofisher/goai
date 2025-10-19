package prompt

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/Zerofisher/goai/pkg/config"
	"gopkg.in/yaml.v3"
)

// Context represents the interface for agent context.
// This interface breaks the circular dependency between prompt and agent packages.
type Context interface {
	GetSystemPrompt() string
	GetWorkDir() string
	GetProjectName() string
	GetProjectLanguage() string
	ProjectHasGit() bool
	GetProjectGitBranch() string
}

// Manager manages system prompts with support for:
// - Multiple sources (file, env, config, default)
// - Template variable interpolation
// - Hot-reload on file changes
// - Profile switching
type Manager struct {
	// Core prompt content
	base     string            // Base prompt content
	partials map[string]string // Named partial templates
	vars     map[string]any    // Template variables
	override string            // Runtime override (highest priority)

	// Metadata
	profile string    // Active profile name
	path    string    // Path to prompt file (if loaded from file)
	mtime   time.Time // Last modification time of file

	// Configuration sources
	cfg     *config.Config
	context Context
}

// PromptConfig represents the YAML configuration for system prompts.
type PromptConfig struct {
	Profile  string            `yaml:"profile"`
	Base     string            `yaml:"base"`      // Path to base markdown file
	Partials map[string]string `yaml:"partials"`  // Map of partial name to file path
	Vars     map[string]any    `yaml:"vars"`      // Custom template variables
}

// NewManager creates a new prompt manager.
func NewManager(cfg *config.Config, ctx Context) *Manager {
	return &Manager{
		partials: make(map[string]string),
		vars:     make(map[string]any),
		cfg:      cfg,
		context:  ctx,
	}
}

// Load loads the system prompt from multiple sources with priority:
// 1. Runtime override (if set)
// 2. .goai/system.yaml or .goai/system.md in work directory
// 3. Environment variable GOAI_SYSTEM_PROMPT
// 4. Config file setting
// 5. Default generator from context
func (m *Manager) Load() error {
	// Check if runtime override is set (highest priority)
	if m.override != "" {
		return nil // Override is already set, no need to load
	}

	// Try to load from .goai/system.yaml or .goai/system.md
	if err := m.loadFromWorkDir(); err == nil {
		return nil // Successfully loaded from work directory
	}

	// Try environment variable
	if envPrompt := os.Getenv("GOAI_SYSTEM_PROMPT"); envPrompt != "" {
		m.base = envPrompt
		return nil
	}

	// Try config file setting
	if m.cfg != nil && m.cfg.Model.SystemPrompt != "" {
		m.base = m.cfg.Model.SystemPrompt
		return nil
	}

	// Fall back to default generator
	if m.context != nil {
		m.base = m.context.GetSystemPrompt()
		return nil
	}

	return fmt.Errorf("no system prompt source available")
}

// loadFromWorkDir attempts to load prompt from .goai directory.
func (m *Manager) loadFromWorkDir() error {
	workDir := m.context.GetWorkDir()
	goaiDir := filepath.Join(workDir, ".goai")

	// Try YAML config first
	yamlPath := filepath.Join(goaiDir, "system.yaml")
	if info, err := os.Stat(yamlPath); err == nil && !info.IsDir() {
		return m.loadFromYAML(yamlPath)
	}

	// Try Markdown file
	mdPath := filepath.Join(goaiDir, "system.md")
	if info, err := os.Stat(mdPath); err == nil && !info.IsDir() {
		return m.loadFromMarkdown(mdPath)
	}

	return fmt.Errorf("no system prompt file found in .goai directory")
}

// loadFromYAML loads prompt configuration from YAML file.
func (m *Manager) loadFromYAML(path string) error {
	// Read YAML file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}

	// Parse YAML
	var cfg PromptConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Store metadata
	info, _ := os.Stat(path)
	m.path = path
	m.mtime = info.ModTime()
	m.profile = cfg.Profile

	// Load base content from file if specified
	if cfg.Base != "" {
		basePath := filepath.Join(filepath.Dir(path), cfg.Base)
		baseContent, err := os.ReadFile(basePath)
		if err != nil {
			return fmt.Errorf("failed to read base file %s: %w", cfg.Base, err)
		}
		m.base = string(baseContent)
	}

	// Load partial templates
	m.partials = make(map[string]string)
	for name, partialPath := range cfg.Partials {
		fullPath := filepath.Join(filepath.Dir(path), partialPath)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read partial %s from %s: %w", name, partialPath, err)
		}
		m.partials[name] = string(content)
	}

	// Store custom variables
	if cfg.Vars != nil {
		m.vars = cfg.Vars
	}

	return nil
}

// loadFromMarkdown loads prompt directly from a Markdown file.
func (m *Manager) loadFromMarkdown(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	// Store metadata
	info, _ := os.Stat(path)
	m.path = path
	m.mtime = info.ModTime()
	m.base = string(content)

	return nil
}

// Compose renders the final system prompt with template variables.
func (m *Manager) Compose() (string, error) {
	// Use override if set
	content := m.base
	if m.override != "" {
		content = m.override
	}

	// If content is empty, return error
	if content == "" {
		return "", fmt.Errorf("no prompt content available")
	}

	// Build template variables
	templateVars := m.buildTemplateVars()

	// Parse and execute template
	tmpl, err := template.New("system").Parse(content)
	if err != nil {
		// If template parsing fails, return content as-is
		// This allows non-template prompts to work
		return content, nil
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateVars); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}

	result := buf.String()

	// Validate result
	if strings.TrimSpace(result) == "" {
		return "", fmt.Errorf("rendered prompt is empty")
	}

	// Check length limit (warn if too long, but don't fail)
	const maxLength = 50000 // 50k characters
	if len(result) > maxLength {
		// Truncate if too long
		result = result[:maxLength] + "\n\n[... prompt truncated due to length ...]"
	}

	return result, nil
}

// buildTemplateVars constructs the variables available for template interpolation.
func (m *Manager) buildTemplateVars() map[string]any {
	vars := make(map[string]any)

	// System information
	vars["WorkDir"] = m.context.GetWorkDir()
	vars["OS"] = runtime.GOOS
	vars["Arch"] = runtime.GOARCH
	vars["GoVersion"] = runtime.Version()

	// Project information
	vars["Project"] = map[string]any{
		"Name":     m.context.GetProjectName(),
		"Language": m.context.GetProjectLanguage(),
		"HasGit":   m.context.ProjectHasGit(),
		"Branch":   m.context.GetProjectGitBranch(),
	}

	// Tools information (TODO: integrate with dispatcher to get actual tool list)
	vars["Tools"] = []string{
		"bash", "read_file", "write_file", "list_files", "edit_file", "search",
	}

	// Policies/guidelines
	vars["Policies"] = map[string]any{
		"MaxFileSize":      "10MB",
		"AllowedCommands":  "read-only operations preferred",
		"SecurityLevel":    "standard",
	}

	// Merge custom variables (they can override defaults)
	for k, v := range m.vars {
		vars[k] = v
	}

	// Add partials as template variables
	for name, content := range m.partials {
		vars[name] = content
	}

	return vars
}

// Override sets a runtime override for the system prompt (highest priority).
func (m *Manager) Override(text string) {
	m.override = text
}

// ClearOverride removes the runtime override.
func (m *Manager) ClearOverride() {
	m.override = ""
}

// UseProfile switches to a different profile (if YAML config supports multiple profiles).
// For now, this is a placeholder for future multi-profile support.
func (m *Manager) UseProfile(name string) error {
	if m.path == "" {
		return fmt.Errorf("no config file loaded, cannot switch profile")
	}

	// TODO: Implement multi-profile support
	// For now, just store the profile name
	m.profile = name

	return fmt.Errorf("multi-profile support not yet implemented")
}

// ReloadIfChanged checks if the source file has been modified and reloads if necessary.
// Returns true if reloaded, false otherwise.
func (m *Manager) ReloadIfChanged() (bool, error) {
	// Only reload if we have a file path
	if m.path == "" {
		return false, nil
	}

	// Check if file still exists
	info, err := os.Stat(m.path)
	if err != nil {
		return false, fmt.Errorf("failed to stat file: %w", err)
	}

	// Check if modification time has changed
	if info.ModTime().Equal(m.mtime) || info.ModTime().Before(m.mtime) {
		return false, nil // No change
	}

	// File has been modified, reload it
	if strings.HasSuffix(m.path, ".yaml") {
		if err := m.loadFromYAML(m.path); err != nil {
			return false, fmt.Errorf("failed to reload YAML: %w", err)
		}
	} else {
		if err := m.loadFromMarkdown(m.path); err != nil {
			return false, fmt.Errorf("failed to reload markdown: %w", err)
		}
	}

	return true, nil
}

// GetSummary returns a summary of the current prompt configuration.
func (m *Manager) GetSummary() string {
	var parts []string

	// Source
	if m.override != "" {
		parts = append(parts, "Source: Runtime Override")
	} else if m.path != "" {
		parts = append(parts, fmt.Sprintf("Source: %s", m.path))
		if m.profile != "" {
			parts = append(parts, fmt.Sprintf("Profile: %s", m.profile))
		}
	} else if os.Getenv("GOAI_SYSTEM_PROMPT") != "" {
		parts = append(parts, "Source: Environment Variable")
	} else if m.cfg != nil && m.cfg.Model.SystemPrompt != "" {
		parts = append(parts, "Source: Config File")
	} else {
		parts = append(parts, "Source: Default Generator")
	}

	// Content preview
	content, err := m.Compose()
	if err != nil {
		parts = append(parts, fmt.Sprintf("Error: %v", err))
	} else {
		lines := strings.Split(content, "\n")
		if len(lines) > 0 {
			preview := strings.TrimSpace(lines[0])
			if len(preview) > 60 {
				preview = preview[:60] + "..."
			}
			parts = append(parts, fmt.Sprintf("Preview: %s", preview))
		}
		parts = append(parts, fmt.Sprintf("Length: %d chars", len(content)))
	}

	// Partials
	if len(m.partials) > 0 {
		partialNames := make([]string, 0, len(m.partials))
		for name := range m.partials {
			partialNames = append(partialNames, name)
		}
		parts = append(parts, fmt.Sprintf("Partials: %s", strings.Join(partialNames, ", ")))
	}

	return strings.Join(parts, "\n")
}

// GetPath returns the file path if loaded from a file, empty string otherwise.
func (m *Manager) GetPath() string {
	return m.path
}

// GetProfile returns the current profile name.
func (m *Manager) GetProfile() string {
	return m.profile
}

// HasOverride returns true if a runtime override is set.
func (m *Manager) HasOverride() bool {
	return m.override != ""
}
