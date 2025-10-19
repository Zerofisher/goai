package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewSecurityValidator(t *testing.T) {
	workDir := "/test/workspace"
	validator := NewSecurityValidator(workDir)

	if validator.workDir != workDir {
		t.Errorf("workDir = %s, want %s", validator.workDir, workDir)
	}

	if len(validator.forbiddenCommands) == 0 {
		t.Error("forbiddenCommands should have default values")
	}

	if len(validator.forbiddenPaths) == 0 {
		t.Error("forbiddenPaths should have default values")
	}
}

func TestDefaultSecurityValidator_ValidateCommand(t *testing.T) {
	validator := NewSecurityValidator("/workspace")

	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{
			name:    "valid command",
			command: "ls -la",
			wantErr: false,
		},
		{
			name:    "empty command",
			command: "  ",
			wantErr: true,
		},
		{
			name:    "forbidden rm -rf /",
			command: "rm -rf /",
			wantErr: true,
		},
		{
			name:    "forbidden shutdown",
			command: "shutdown -h now",
			wantErr: true,
		},
		{
			name:    "forbidden reboot",
			command: "sudo reboot",
			wantErr: true,
		},
		{
			name:    "fork bomb",
			command: ":(){ :|:& };:",
			wantErr: true,
		},
		{
			name:    "dangerous sudo rm",
			command: "sudo rm -rf /var",
			wantErr: true,
		},
		{
			name:    "dangerous dd command",
			command: "dd if=/dev/zero of=/dev/sda",
			wantErr: true,
		},
		{
			name:    "dangerous chmod 777",
			command: "chmod 777 /etc/passwd",
			wantErr: true,
		},
		{
			name:    "command substitution",
			command: "echo $(whoami)",
			wantErr: true,
		},
		{
			name:    "backticks substitution",
			command: "echo `whoami`",
			wantErr: true,
		},
		{
			name:    "command chaining with &&",
			command: "ls && rm file",
			wantErr: true,
		},
		{
			name:    "command chaining with ||",
			command: "ls || rm file",
			wantErr: true,
		},
		{
			name:    "command separator ;",
			command: "ls; rm file",
			wantErr: true,
		},
		{
			name:    "simple pipe allowed",
			command: "ls | grep test",
			wantErr: false,
		},
		{
			name:    "pipe with ||",
			command: "ls || echo fail",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCommand(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultSecurityValidator_ValidatePath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test_workspace")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	validator := NewSecurityValidator(tmpDir)

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid file in workspace",
			path:    testFile,
			wantErr: false,
		},
		{
			name:    "relative path in workspace",
			path:    "test.txt",
			wantErr: false,
		},
		{
			name:    "forbidden /etc/passwd",
			path:    "/etc/passwd",
			wantErr: true,
		},
		{
			name:    "forbidden /etc/shadow",
			path:    "/etc/shadow",
			wantErr: true,
		},
		{
			name:    "path with traversal",
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "path outside workspace",
			path:    "/tmp/outside",
			wantErr: true,
		},
	}

	// Change to test directory for relative path tests
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Logf("Test '%s' failed: path=%s, err=%v, wantErr=%v", tt.name, tt.path, err, tt.wantErr)
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultSecurityValidator_CheckPermission(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_workspace")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	validator := NewSecurityValidator(tmpDir)

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a test directory
	testDir := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		tool    string
		input   map[string]interface{}
		wantErr bool
	}{
		{
			name: "bash with valid command",
			tool: "bash",
			input: map[string]interface{}{
				"command": "ls -la",
			},
			wantErr: false,
		},
		{
			name: "bash with forbidden command",
			tool: "bash",
			input: map[string]interface{}{
				"command": "rm -rf /",
			},
			wantErr: true,
		},
		{
			name: "file_read with valid path",
			tool: "file_read",
			input: map[string]interface{}{
				"path": testFile,
			},
			wantErr: false,
		},
		{
			name: "file_read with forbidden path",
			tool: "file_read",
			input: map[string]interface{}{
				"path": "/etc/passwd",
			},
			wantErr: true,
		},
		{
			name: "delete file allowed",
			tool: "delete",
			input: map[string]interface{}{
				"path": testFile,
			},
			wantErr: false,
		},
		{
			name: "delete directory not allowed",
			tool: "delete",
			input: map[string]interface{}{
				"path": testDir,
			},
			wantErr: true,
		},
		{
			name: "unknown tool",
			tool: "unknown",
			input: map[string]interface{}{
				"param": "value",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.CheckPermission(tt.tool, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPermission() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultSecurityValidator_SetMethods(t *testing.T) {
	validator := NewSecurityValidator("/workspace")

	// Test SetForbiddenCommands
	customCommands := []string{"custom1", "custom2"}
	validator.SetForbiddenCommands(customCommands)
	if len(validator.forbiddenCommands) != len(customCommands) {
		t.Error("SetForbiddenCommands did not update correctly")
	}

	// Test SetForbiddenPaths
	customPaths := []string{"/custom/path1", "/custom/path2"}
	validator.SetForbiddenPaths(customPaths)
	if len(validator.forbiddenPaths) != len(customPaths) {
		t.Error("SetForbiddenPaths did not update correctly")
	}

	// Test SetAllowedDirs
	customDirs := []string{"/allowed1", "/allowed2"}
	validator.SetAllowedDirs(customDirs)
	if len(validator.allowedDirs) != len(customDirs) {
		t.Error("SetAllowedDirs did not update correctly")
	}
}

func TestContainsShellInjection(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{
			name:     "clean command",
			command:  "ls -la",
			expected: false,
		},
		{
			name:     "command substitution with $()",
			command:  "echo $(whoami)",
			expected: true,
		},
		{
			name:     "command substitution with backticks",
			command:  "echo `whoami`",
			expected: true,
		},
		{
			name:     "command chaining with &&",
			command:  "ls && echo done",
			expected: true,
		},
		{
			name:     "command chaining with ||",
			command:  "ls || echo fail",
			expected: true,
		},
		{
			name:     "command separator ;",
			command:  "ls; echo done",
			expected: true,
		},
		{
			name:     "simple pipe allowed",
			command:  "ls | grep test",
			expected: false,
		},
		{
			name:     "double pipe ||",
			command:  "ls || echo fail",
			expected: true,
		},
		{
			name:     "IFS manipulation",
			command:  "$(IFS=;echo test)",
			expected: true,
		},
		{
			name:     "newline injection",
			command:  "echo test\nrm -rf /",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsShellInjection(tt.command)
			if result != tt.expected {
				t.Errorf("containsShellInjection(%s) = %v, want %v", tt.command, result, tt.expected)
			}
		})
	}
}

func TestPathSanitizer(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_workspace")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	sanitizer := NewPathSanitizer(tmpDir)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "relative path",
			path:    "test.txt",
			wantErr: false,
		},
		{
			name:    "absolute path in workspace",
			path:    filepath.Join(tmpDir, "test.txt"),
			wantErr: false,
		},
		{
			name:    "path with spaces",
			path:    "  test.txt  ",
			wantErr: false,
		},
		{
			name:    "path traversal attempt",
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "path outside workspace",
			path:    "/etc/passwd",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sanitizer.Sanitize(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sanitize() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && !strings.HasPrefix(result, tmpDir) {
				t.Errorf("Sanitized path %s should be within workspace %s", result, tmpDir)
			}
		})
	}
}

func TestValidatePath_WithSymlinks(t *testing.T) {
	// Skip if we can't create symlinks (e.g., on Windows without admin)
	if os.Getenv("SKIP_SYMLINK_TEST") != "" {
		t.Skip("Skipping symlink test")
	}

	tmpDir, err := os.MkdirTemp("", "test_workspace")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	validator := NewSecurityValidator(tmpDir)

	// Create a file outside workspace
	outsideDir, err := os.MkdirTemp("", "outside")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(outsideDir)

	outsideFile := filepath.Join(outsideDir, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("secret"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a symlink pointing outside
	symlink := filepath.Join(tmpDir, "link")
	if err := os.Symlink(outsideFile, symlink); err != nil {
		t.Skip("Cannot create symlink, skipping test")
	}

	// Symlink should be detected and validated
	err = validator.ValidatePath(symlink)
	if err == nil {
		t.Error("ValidatePath should fail for symlink pointing outside workspace")
	}
}