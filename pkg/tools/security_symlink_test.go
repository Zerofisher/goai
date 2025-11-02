package tools

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSymlinkEscape tests that symlink-based path traversal is blocked
func TestSymlinkEscape(t *testing.T) {
	// Create test environment
	tempDir := t.TempDir()
	workDir := filepath.Join(tempDir, "workspace")
	sensitiveDir := filepath.Join(tempDir, "sensitive")

	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sensitiveDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create sensitive file
	secretFile := filepath.Join(sensitiveDir, "secret.txt")
	if err := os.WriteFile(secretFile, []byte("SECRET_DATA"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create symlink inside workspace pointing to sensitive dir
	symlinkPath := filepath.Join(workDir, "escape")
	if err := os.Symlink(sensitiveDir, symlinkPath); err != nil {
		t.Skip("Symlink creation failed (may need permissions)")
	}

	validator := NewSecurityValidator(workDir)

	// Test 1: Direct access to symlink target should be blocked
	escapePath := filepath.Join(symlinkPath, "secret.txt")
	err := validator.ValidatePath(escapePath)
	if err == nil {
		t.Error("Expected error for symlink escape, but got nil")
	} else {
		t.Logf("✓ Symlink escape blocked: %v", err)
	}

	// Test 2: Symlink itself should be allowed if it's inside workspace
	err = validator.ValidatePath(symlinkPath)
	if err != nil {
		// This might fail because the symlink points outside
		t.Logf("Symlink itself blocked (acceptable): %v", err)
	}
}

// TestAllowedDirsSymlinkEscape tests allowedDirs branch
func TestAllowedDirsSymlinkEscape(t *testing.T) {
	// Create test environment
	tempDir := t.TempDir()
	workDir := filepath.Join(tempDir, "workspace")
	allowedDir := filepath.Join(tempDir, "allowed")
	sensitiveDir := filepath.Join(tempDir, "sensitive")

	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(allowedDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sensitiveDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create sensitive file
	secretFile := filepath.Join(sensitiveDir, "secret.txt")
	if err := os.WriteFile(secretFile, []byte("SECRET_DATA"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create symlink inside allowed dir pointing to sensitive dir
	symlinkPath := filepath.Join(allowedDir, "escape")
	if err := os.Symlink(sensitiveDir, symlinkPath); err != nil {
		t.Skip("Symlink creation failed (may need permissions)")
	}

	validator := NewSecurityValidator(workDir)
	validator.SetAllowedDirs([]string{allowedDir})

	// Test: Access through symlink should be blocked
	escapePath := filepath.Join(symlinkPath, "secret.txt")
	err := validator.ValidatePath(escapePath)
	if err == nil {
		t.Error("Expected error for symlink escape in allowedDirs, but got nil")
	} else {
		t.Logf("✓ Symlink escape in allowedDirs blocked: %v", err)
	}
}

// TestPathTraversalWithDots tests classic .. traversal
func TestPathTraversalWithDots(t *testing.T) {
	tempDir := t.TempDir()
	workDir := filepath.Join(tempDir, "workspace")

	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatal(err)
	}

	validator := NewSecurityValidator(workDir)

	tests := []struct {
		name string
		path string
		want bool // true if should be allowed
	}{
		{
			name: "direct parent escape",
			path: filepath.Join(workDir, "..", "secret.txt"),
			want: false,
		},
		{
			name: "multiple parent escape",
			path: filepath.Join(workDir, "..", "..", "secret.txt"),
			want: false,
		},
		{
			name: "valid path with dots in name",
			path: filepath.Join(workDir, "file..txt"),
			want: true,
		},
		{
			name: "valid subdirectory",
			path: filepath.Join(workDir, "subdir", "file.txt"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePath(tt.path)
			if tt.want && err != nil {
				t.Errorf("Expected path to be allowed, but got error: %v", err)
			}
			if !tt.want && err == nil {
				t.Errorf("Expected path to be blocked, but it was allowed")
			}
		})
	}
}
