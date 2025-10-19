package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SecurityValidator validates tool inputs for security concerns
type SecurityValidator interface {
	// ValidatePath checks if a file path is safe to access
	ValidatePath(path string) error

	// ValidateCommand checks if a command is safe to execute
	ValidateCommand(cmd string) error

	// CheckPermission checks if a tool operation is permitted
	CheckPermission(tool string, input map[string]interface{}) error
}

// DefaultSecurityValidator provides default security validation
type DefaultSecurityValidator struct {
	workDir           string
	forbiddenCommands []string
	forbiddenPaths    []string
	allowedDirs       []string
}

// NewSecurityValidator creates a new security validator
func NewSecurityValidator(workDir string) *DefaultSecurityValidator {
	return &DefaultSecurityValidator{
		workDir: workDir,
		forbiddenCommands: []string{
			"rm -rf /",
			"rm -rf /*",
			"shutdown",
			"reboot",
			"halt",
			"poweroff",
			"init 0",
			"init 6",
			"mkfs",
			"dd if=/dev/zero",
			"dd if=/dev/random",
			":(){ :|:& };:", // Fork bomb
			"systemctl poweroff",
			"systemctl reboot",
			"systemctl halt",
		},
		forbiddenPaths: []string{
			"/etc/passwd",
			"/etc/shadow",
			"/etc/sudoers",
			"~/.ssh/",
			"~/.gnupg/",
			"/root/",
			"/boot/",
			"/sys/",
			"/proc/",
		},
		allowedDirs: []string{}, // Empty means only workDir is allowed
	}
}

// SetForbiddenCommands sets the list of forbidden commands
func (v *DefaultSecurityValidator) SetForbiddenCommands(commands []string) {
	v.forbiddenCommands = commands
}

// SetForbiddenPaths sets the list of forbidden paths
func (v *DefaultSecurityValidator) SetForbiddenPaths(paths []string) {
	v.forbiddenPaths = paths
}

// SetAllowedDirs sets the list of allowed directories
func (v *DefaultSecurityValidator) SetAllowedDirs(dirs []string) {
	v.allowedDirs = dirs
}

// ValidatePath checks if a file path is safe to access
func (v *DefaultSecurityValidator) ValidatePath(path string) error {
	// Expand home directory first if needed
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot expand home directory: %w", err)
		}
		path = filepath.Join(homeDir, path[2:])
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check forbidden paths
	for _, forbidden := range v.forbiddenPaths {
		// Expand ~ in forbidden path
		forbiddenPath := forbidden
		if strings.HasPrefix(forbiddenPath, "~/") {
			homeDir, _ := os.UserHomeDir()
			forbiddenPath = filepath.Join(homeDir, forbiddenPath[2:])
		}

		// Check if path matches or is under forbidden path
		if absPath == forbiddenPath || strings.HasPrefix(absPath, forbiddenPath) {
			return fmt.Errorf("access to path %s is forbidden", path)
		}
	}

	// Check if path is within allowed directories
	if len(v.allowedDirs) > 0 {
		allowed := false
		for _, allowedDir := range v.allowedDirs {
			allowedAbs, err := filepath.Abs(allowedDir)
			if err != nil {
				continue
			}
			if strings.HasPrefix(absPath, allowedAbs) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("path %s is outside allowed directories", path)
		}
	} else {
		// If no specific allowed dirs, check if within workDir
		workAbs, err := filepath.Abs(v.workDir)
		if err != nil {
			return fmt.Errorf("invalid work directory: %w", err)
		}

		// Resolve any symlinks in both paths to ensure correct comparison
		workAbs, err = filepath.EvalSymlinks(workAbs)
		if err != nil {
			// If workDir doesn't exist yet, that's okay, just use the absolute path
			workAbs, _ = filepath.Abs(v.workDir)
		}

		// For the target path, only evaluate symlinks if it exists
		evalPath := absPath
		if _, err := os.Stat(absPath); err == nil {
			if evaluated, err := filepath.EvalSymlinks(absPath); err == nil {
				evalPath = evaluated
			}
		}

		// Check if path is within work directory
		relPath, err := filepath.Rel(workAbs, evalPath)
		if err != nil {
			return fmt.Errorf("path %s is not within work directory", path)
		}

		// Check for path traversal attempts
		// Only check if the path starts with ".." to escape the work directory
		if strings.HasPrefix(relPath, "..") {
			return fmt.Errorf("path traversal detected in %s", path)
		}
	}

	// Check for symbolic links that might escape
	info, err := os.Lstat(absPath)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		// Resolve the symlink
		resolved, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return fmt.Errorf("cannot resolve symlink %s: %w", path, err)
		}
		// Recursively validate the resolved path
		return v.ValidatePath(resolved)
	}

	return nil
}

// ValidateCommand checks if a command is safe to execute
func (v *DefaultSecurityValidator) ValidateCommand(cmd string) error {
	// Check for empty command
	if strings.TrimSpace(cmd) == "" {
		return fmt.Errorf("empty command")
	}

	// Convert to lowercase for case-insensitive comparison
	cmdLower := strings.ToLower(cmd)

	// Check against forbidden commands
	for _, forbidden := range v.forbiddenCommands {
		if strings.Contains(cmdLower, strings.ToLower(forbidden)) {
			return fmt.Errorf("command contains forbidden pattern: %s", forbidden)
		}
	}

	// Check for dangerous patterns
	dangerousPatterns := []string{
		"sudo rm",
		"sudo dd",
		"sudo mkfs",
		"> /dev/",
		"< /dev/zero",
		"< /dev/random",
		"/dev/null >",
		"chmod -R 777",
		"chmod 777",
		"chown -R",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(cmdLower, strings.ToLower(pattern)) {
			return fmt.Errorf("command contains dangerous pattern: %s", pattern)
		}
	}

	// Check for shell injection attempts
	if containsShellInjection(cmd) {
		return fmt.Errorf("potential shell injection detected")
	}

	return nil
}

// CheckPermission checks if a tool operation is permitted
func (v *DefaultSecurityValidator) CheckPermission(tool string, input map[string]interface{}) error {
	// Tool-specific permission checks
	switch tool {
	case "bash", "shell", "execute":
		if cmd, ok := input["command"].(string); ok {
			return v.ValidateCommand(cmd)
		}

	case "file_read", "file_write", "edit":
		if path, ok := input["path"].(string); ok {
			return v.ValidatePath(path)
		}
		if filePath, ok := input["file_path"].(string); ok {
			return v.ValidatePath(filePath)
		}

	case "delete", "remove":
		// Extra strict for delete operations
		if path, ok := input["path"].(string); ok {
			if err := v.ValidatePath(path); err != nil {
				return err
			}
			// Don't allow deletion of directories
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				return fmt.Errorf("deletion of directories is not allowed")
			}
		}
	}

	return nil
}

// containsShellInjection checks for potential shell injection patterns
func containsShellInjection(cmd string) bool {
	// Check for common injection patterns
	injectionPatterns := []string{
		"$(",      // Command substitution
		"`",       // Backticks for command substitution
		"&&",      // Command chaining
		"||",      // Command chaining
		";",       // Command separator
		"|",       // Pipe (could be dangerous)
		"$(IFS",   // IFS manipulation
		"${IFS",   // IFS manipulation
		"\n",      // Newline injection
		"\r",      // Carriage return injection
	}

	for _, pattern := range injectionPatterns {
		if strings.Contains(cmd, pattern) {
			// Allow some safe patterns
			if pattern == "|" {
				// Allow simple pipes like "ls | grep"
				if !strings.Contains(cmd, "||") && !strings.Contains(cmd, "|&") {
					continue
				}
			}
			return true
		}
	}

	return false
}

// PathSanitizer provides path sanitization utilities
type PathSanitizer struct {
	workDir string
}

// NewPathSanitizer creates a new path sanitizer
func NewPathSanitizer(workDir string) *PathSanitizer {
	return &PathSanitizer{
		workDir: workDir,
	}
}

// Sanitize cleans and validates a path
func (s *PathSanitizer) Sanitize(path string) (string, error) {
	// Remove leading/trailing whitespace
	path = strings.TrimSpace(path)

	// Expand home directory
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot expand home directory: %w", err)
		}
		path = filepath.Join(homeDir, path[2:])
	}

	// Convert to absolute path if relative
	if !filepath.IsAbs(path) {
		path = filepath.Join(s.workDir, path)
	}

	// Clean the path
	path = filepath.Clean(path)

	// Verify it's within workDir
	workAbs, err := filepath.Abs(s.workDir)
	if err != nil {
		return "", fmt.Errorf("invalid work directory: %w", err)
	}

	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Check if path is within work directory
	relPath, err := filepath.Rel(workAbs, pathAbs)
	if err != nil {
		return "", fmt.Errorf("path %s is not within work directory", path)
	}

	// Check for path traversal
	if strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("path traversal detected")
	}

	return pathAbs, nil
}