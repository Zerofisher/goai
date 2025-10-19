package prompt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Zerofisher/goai/pkg/config"
)

// mockContext is a test implementation of the Context interface.
type mockContext struct {
	workDir    string
	sysPrompt  string
	projName   string
	projLang   string
	hasGit     bool
	gitBranch  string
}

func newMockContext(workDir string) *mockContext {
	return &mockContext{
		workDir:   workDir,
		sysPrompt: "Default system prompt",
		projName:  filepath.Base(workDir),
		projLang:  "Go",
		hasGit:    false,
		gitBranch: "",
	}
}

func (m *mockContext) GetSystemPrompt() string {
	return m.sysPrompt
}

func (m *mockContext) GetWorkDir() string {
	return m.workDir
}

func (m *mockContext) GetProjectName() string {
	return m.projName
}

func (m *mockContext) GetProjectLanguage() string {
	return m.projLang
}

func (m *mockContext) ProjectHasGit() bool {
	return m.hasGit
}

func (m *mockContext) GetProjectGitBranch() string {
	return m.gitBranch
}

// TestNewManager tests creating a new prompt manager.
func TestNewManager(t *testing.T) {
	cfg := &config.Config{}
	ctx := newMockContext("/test/dir")

	manager := NewManager(cfg, ctx)

	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}

	if manager.partials == nil {
		t.Error("partials map not initialized")
	}

	if manager.vars == nil {
		t.Error("vars map not initialized")
	}
}

// TestLoad_Default tests loading default prompt when no sources exist.
func TestLoad_Default(t *testing.T) {
	cfg := &config.Config{}
	ctx := newMockContext("/test/dir")

	manager := NewManager(cfg, ctx)
	err := manager.Load()

	if err != nil {
		t.Errorf("Load() failed: %v", err)
	}

	// Should have loaded from default generator
	if manager.base == "" {
		t.Error("base prompt is empty")
	}
}

// TestLoad_FromConfig tests loading from config file setting.
func TestLoad_FromConfig(t *testing.T) {
	cfg := &config.Config{
		Model: config.ModelConfig{
			SystemPrompt: "Test prompt from config",
		},
	}
	ctx := newMockContext("/test/dir")

	manager := NewManager(cfg, ctx)
	err := manager.Load()

	if err != nil {
		t.Errorf("Load() failed: %v", err)
	}

	if manager.base != "Test prompt from config" {
		t.Errorf("base = %q, want %q", manager.base, "Test prompt from config")
	}
}

// TestLoad_FromEnv tests loading from environment variable.
func TestLoad_FromEnv(t *testing.T) {
	// Set environment variable
	envPrompt := "Test prompt from environment"
	if err := os.Setenv("GOAI_SYSTEM_PROMPT", envPrompt); err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("GOAI_SYSTEM_PROMPT")
	}()

	cfg := &config.Config{}
	ctx := newMockContext("/test/dir")

	manager := NewManager(cfg, ctx)
	err := manager.Load()

	if err != nil {
		t.Errorf("Load() failed: %v", err)
	}

	if manager.base != envPrompt {
		t.Errorf("base = %q, want %q", manager.base, envPrompt)
	}
}

// TestLoad_FromMarkdownFile tests loading from .goai/system.md file.
func TestLoad_FromMarkdownFile(t *testing.T) {
	// Create temporary directory structure
	tempDir, err := os.MkdirTemp("", "prompt_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir) // Non-critical error, can be ignored in cleanup
	}()

	// Create .goai directory
	goaiDir := filepath.Join(tempDir, ".goai")
	if err := os.Mkdir(goaiDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create system.md file
	promptContent := "# Test Prompt\n\nThis is a test prompt from markdown file."
	mdPath := filepath.Join(goaiDir, "system.md")
	if err := os.WriteFile(mdPath, []byte(promptContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create manager with temp directory as work dir
	cfg := &config.Config{WorkDir: tempDir}
	ctx := newMockContext(tempDir)

	manager := NewManager(cfg, ctx)
	err = manager.Load()

	if err != nil {
		t.Errorf("Load() failed: %v", err)
	}

	if manager.base != promptContent {
		t.Errorf("base = %q, want %q", manager.base, promptContent)
	}

	if manager.path != mdPath {
		t.Errorf("path = %q, want %q", manager.path, mdPath)
	}
}

// TestLoad_FromYAMLFile tests loading from .goai/system.yaml file.
func TestLoad_FromYAMLFile(t *testing.T) {
	// Create temporary directory structure
	tempDir, err := os.MkdirTemp("", "prompt_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir) // Non-critical error, can be ignored in cleanup
	}()

	// Create .goai directory
	goaiDir := filepath.Join(tempDir, ".goai")
	if err := os.Mkdir(goaiDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create partials directory
	partialsDir := filepath.Join(goaiDir, "partials")
	if err := os.Mkdir(partialsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create base markdown file
	baseContent := "# System Prompt\n\nYou are {{ .Project.Name }} assistant.\n\nTools: {{ .Tools }}"
	basePath := filepath.Join(goaiDir, "base.md")
	if err := os.WriteFile(basePath, []byte(baseContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create partial file
	partialContent := "## Coding Guidelines\n\n- Write clean code\n- Add tests"
	partialPath := filepath.Join(partialsDir, "guidelines.md")
	if err := os.WriteFile(partialPath, []byte(partialContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create system.yaml file
	yamlContent := `profile: coding
base: base.md
partials:
  guidelines: partials/guidelines.md
vars:
  max_tokens: 16000
  prefer_read_first: true
`
	yamlPath := filepath.Join(goaiDir, "system.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create manager
	cfg := &config.Config{WorkDir: tempDir}
	ctx := newMockContext(tempDir)

	manager := NewManager(cfg, ctx)
	err = manager.Load()

	if err != nil {
		t.Errorf("Load() failed: %v", err)
	}

	if manager.base != baseContent {
		t.Errorf("base content mismatch")
	}

	if manager.profile != "coding" {
		t.Errorf("profile = %q, want 'coding'", manager.profile)
	}

	if len(manager.partials) != 1 {
		t.Errorf("partials count = %d, want 1", len(manager.partials))
	}

	if manager.partials["guidelines"] != partialContent {
		t.Error("partial content mismatch")
	}

	if len(manager.vars) != 2 {
		t.Errorf("vars count = %d, want 2", len(manager.vars))
	}
}

// TestCompose tests template composition.
func TestCompose(t *testing.T) {
	cfg := &config.Config{WorkDir: "/test/dir"}
	ctx := newMockContext("/test/dir")

	manager := NewManager(cfg, ctx)
	manager.base = "Work dir: {{ .WorkDir }}\nOS: {{ .OS }}"

	result, err := manager.Compose()
	if err != nil {
		t.Errorf("Compose() failed: %v", err)
	}

	if !strings.Contains(result, "/test/dir") {
		t.Errorf("result does not contain work dir: %s", result)
	}

	if !strings.Contains(result, "OS:") {
		t.Errorf("result does not contain OS: %s", result)
	}
}

// TestCompose_WithPartials tests composition with partials.
func TestCompose_WithPartials(t *testing.T) {
	cfg := &config.Config{WorkDir: "/test"}
	ctx := newMockContext("/test")

	manager := NewManager(cfg, ctx)
	manager.base = "# Main Prompt\n\n{{ .guidelines }}"
	manager.partials = map[string]string{
		"guidelines": "## Guidelines\n\n- Be helpful",
	}

	result, err := manager.Compose()
	if err != nil {
		t.Errorf("Compose() failed: %v", err)
	}

	if !strings.Contains(result, "# Main Prompt") {
		t.Error("result does not contain main prompt header")
	}

	if !strings.Contains(result, "## Guidelines") {
		t.Error("result does not contain guidelines from partial")
	}

	if !strings.Contains(result, "Be helpful") {
		t.Error("result does not contain guideline content")
	}
}

// TestCompose_WithProjectInfo tests template with project information.
func TestCompose_WithProjectInfo(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "project_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir) // Non-critical error, can be ignored in cleanup
	}()

	cfg := &config.Config{WorkDir: tempDir}
	ctx := newMockContext(tempDir)

	manager := NewManager(cfg, ctx)
	manager.base = "Project: {{ .Project.Name }}\nLanguage: {{ .Project.Language }}"

	result, err := manager.Compose()
	if err != nil {
		t.Errorf("Compose() failed: %v", err)
	}

	// Should contain project name (directory name)
	if !strings.Contains(result, "Project:") {
		t.Error("result does not contain project name")
	}
}

// TestCompose_EmptyContent tests composition with empty content.
func TestCompose_EmptyContent(t *testing.T) {
	cfg := &config.Config{}
	ctx := newMockContext("")

	manager := NewManager(cfg, ctx)
	manager.base = ""
	manager.override = ""

	_, err := manager.Compose()
	if err == nil {
		t.Error("Compose() should fail with empty content")
	}
}

// TestCompose_TooLong tests truncation of very long prompts.
func TestCompose_TooLong(t *testing.T) {
	cfg := &config.Config{}
	ctx := newMockContext("")

	manager := NewManager(cfg, ctx)
	// Create a prompt longer than 50k characters
	manager.base = strings.Repeat("This is a very long prompt. ", 10000)

	result, err := manager.Compose()
	if err != nil {
		t.Errorf("Compose() failed: %v", err)
	}

	// Should be truncated
	if len(result) > 50100 { // Allow some buffer for truncation message
		t.Errorf("result too long: %d characters", len(result))
	}

	if !strings.Contains(result, "truncated") {
		t.Error("truncation message not found")
	}
}

// TestOverride tests runtime override functionality.
func TestOverride(t *testing.T) {
	cfg := &config.Config{}
	ctx := newMockContext("")

	manager := NewManager(cfg, ctx)
	manager.base = "Original prompt"

	// Set override
	overrideText := "Override prompt"
	manager.Override(overrideText)

	result, err := manager.Compose()
	if err != nil {
		t.Errorf("Compose() failed: %v", err)
	}

	if result != overrideText {
		t.Errorf("result = %q, want %q", result, overrideText)
	}

	// Clear override
	manager.ClearOverride()
	result, err = manager.Compose()
	if err != nil {
		t.Errorf("Compose() failed after clear: %v", err)
	}

	if result != "Original prompt" {
		t.Errorf("result after clear = %q, want 'Original prompt'", result)
	}
}

// TestReloadIfChanged tests hot-reload functionality.
func TestReloadIfChanged(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "reload_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir) // Non-critical error, can be ignored in cleanup
	}()

	// Create .goai directory
	goaiDir := filepath.Join(tempDir, ".goai")
	if err := os.Mkdir(goaiDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create initial prompt file
	mdPath := filepath.Join(goaiDir, "system.md")
	initialContent := "Initial prompt"
	if err := os.WriteFile(mdPath, []byte(initialContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Load initial prompt
	cfg := &config.Config{WorkDir: tempDir}
	ctx := newMockContext(tempDir)
	manager := NewManager(cfg, ctx)
	if err := manager.Load(); err != nil {
		t.Fatal(err)
	}

	// Check no reload needed initially
	reloaded, err := manager.ReloadIfChanged()
	if err != nil {
		t.Errorf("ReloadIfChanged() error: %v", err)
	}
	if reloaded {
		t.Error("should not reload when file hasn't changed")
	}

	// Wait a bit to ensure different mtime
	time.Sleep(10 * time.Millisecond)

	// Modify the file
	updatedContent := "Updated prompt"
	if err := os.WriteFile(mdPath, []byte(updatedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Check reload happens
	reloaded, err = manager.ReloadIfChanged()
	if err != nil {
		t.Errorf("ReloadIfChanged() error: %v", err)
	}
	if !reloaded {
		t.Error("should reload when file has changed")
	}

	// Verify content was reloaded
	if manager.base != updatedContent {
		t.Errorf("base = %q, want %q", manager.base, updatedContent)
	}
}

// TestGetSummary tests summary generation.
func TestGetSummary(t *testing.T) {
	cfg := &config.Config{}
	ctx := newMockContext("")

	manager := NewManager(cfg, ctx)
	if err := manager.Load(); err != nil {
		t.Fatal(err)
	}

	summary := manager.GetSummary()

	if summary == "" {
		t.Error("summary is empty")
	}

	// Should contain source information
	if !strings.Contains(summary, "Source:") {
		t.Error("summary does not contain source information")
	}

	// Should contain length information
	if !strings.Contains(summary, "Length:") {
		t.Error("summary does not contain length information")
	}
}

// TestGetSummary_WithOverride tests summary with runtime override.
func TestGetSummary_WithOverride(t *testing.T) {
	cfg := &config.Config{}
	ctx := newMockContext("")

	manager := NewManager(cfg, ctx)
	manager.Override("Override text")

	summary := manager.GetSummary()

	if !strings.Contains(summary, "Runtime Override") {
		t.Errorf("summary does not indicate override: %s", summary)
	}
}

// TestHasOverride tests checking for override.
func TestHasOverride(t *testing.T) {
	cfg := &config.Config{}
	ctx := newMockContext("")

	manager := NewManager(cfg, ctx)

	if manager.HasOverride() {
		t.Error("HasOverride() = true initially, want false")
	}

	manager.Override("test")

	if !manager.HasOverride() {
		t.Error("HasOverride() = false after override, want true")
	}

	manager.ClearOverride()

	if manager.HasOverride() {
		t.Error("HasOverride() = true after clear, want false")
	}
}

// TestLoadPriority tests the priority order of prompt sources.
func TestLoadPriority(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "priority_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir) // Non-critical error, can be ignored in cleanup
	}()

	// Set environment variable (lower priority)
	if err := os.Setenv("GOAI_SYSTEM_PROMPT", "From env"); err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("GOAI_SYSTEM_PROMPT")
	}()

	// Set config (lower priority)
	cfg := &config.Config{
		WorkDir: tempDir,
		Model: config.ModelConfig{
			SystemPrompt: "From config",
		},
	}

	// Create .goai/system.md (higher priority than env and config)
	goaiDir := filepath.Join(tempDir, ".goai")
	if err := os.Mkdir(goaiDir, 0755); err != nil {
		t.Fatal(err)
	}
	mdPath := filepath.Join(goaiDir, "system.md")
	if err := os.WriteFile(mdPath, []byte("From file"), 0644); err != nil {
		t.Fatal(err)
	}

	// Load should prefer file over env and config
	ctx := newMockContext(tempDir)
	manager := NewManager(cfg, ctx)
	if err := manager.Load(); err != nil {
		t.Fatal(err)
	}

	if manager.base != "From file" {
		t.Errorf("base = %q, want 'From file' (file should have higher priority)", manager.base)
	}

	// Override should be highest priority
	manager.Override("From override")
	result, _ := manager.Compose()
	if result != "From override" {
		t.Errorf("result = %q, want 'From override' (override should have highest priority)", result)
	}
}
