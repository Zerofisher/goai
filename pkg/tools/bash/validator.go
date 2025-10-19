package bash

import (
	"fmt"
	"regexp"
	"strings"
)

// Validator validates bash commands for security.
type Validator struct {
	forbiddenCommands  []string
	forbiddenPatterns  []*regexp.Regexp
	dangerousFlags     map[string][]string
	allowedCommands    map[string]bool
	maxCommandLength   int
}

// NewValidator creates a new command validator.
func NewValidator() *Validator {
	v := &Validator{
		maxCommandLength: 10000, // Default max command length
		allowedCommands:  make(map[string]bool),
		dangerousFlags:   make(map[string][]string),
	}

	// Set default forbidden commands
	v.forbiddenCommands = []string{
		"rm -rf /",
		"rm -rf /*",
		"dd if=/dev/zero",
		"mkfs",
		"format",
		":(){:|:&};:",     // Fork bomb
		"chmod -R 777 /",
		"chown -R",
		"shutdown",
		"reboot",
		"halt",
		"poweroff",
		"init 0",
		"init 6",
		"systemctl poweroff",
		"systemctl reboot",
	}

	// Set dangerous command patterns
	v.forbiddenPatterns = []*regexp.Regexp{
		regexp.MustCompile(`rm\s+(-rf|-fr)\s+/`),              // Dangerous rm
		regexp.MustCompile(`>\s*/dev/sd[a-z]`),                // Direct disk write
		regexp.MustCompile(`dd\s+.*of=/dev/[^/]+$`),           // dd to device
		regexp.MustCompile(`/etc/passwd`),                     // Password file access
		regexp.MustCompile(`/etc/shadow`),                     // Shadow file access
		regexp.MustCompile(`curl.*\|\s*sh`),                   // Curl pipe to shell
		regexp.MustCompile(`wget.*\|\s*sh`),                   // Wget pipe to shell
		regexp.MustCompile(`\$\(.*\)`),                        // Command substitution
		regexp.MustCompile("`.+`"),                            // Backtick command substitution
		regexp.MustCompile(`nc\s+-l`),                         // Netcat listener
		regexp.MustCompile(`/dev/tcp/`),                       // Network redirection
		regexp.MustCompile(`iptables\s+-F`),                   // Firewall flush
		regexp.MustCompile(`base64\s+-d.*\|\s*sh`),            // Base64 decode to shell
		regexp.MustCompile(`eval\s+`),                         // Eval command
		regexp.MustCompile(`exec\s+[^>]+$`),                   // Exec replacement
	}

	// Set dangerous flags for common commands
	v.dangerousFlags = map[string][]string{
		"rm":     {"-rf", "-fr", "--no-preserve-root"},
		"chmod":  {"-R 777", "777 -R", "a+rwx"},
		"chown":  {"-R root", "root:root -R"},
		"kill":   {"-9 1", "-KILL 1", "-TERM 1"},
		"pkill":  {"-9", "-KILL"},
		"killall": {"-9", "-KILL"},
	}

	return v
}

// ValidateCommand validates a command for safety.
func (v *Validator) ValidateCommand(command string) error {
	// Check command length
	if len(command) > v.maxCommandLength {
		return fmt.Errorf("command exceeds maximum length of %d characters", v.maxCommandLength)
	}

	// Check for empty command
	command = strings.TrimSpace(command)
	if command == "" {
		return fmt.Errorf("command cannot be empty")
	}

	// Check for null bytes
	if strings.Contains(command, "\x00") {
		return fmt.Errorf("command contains null bytes")
	}

	// Convert to lowercase for checking
	lowerCmd := strings.ToLower(command)

	// Check for exact forbidden commands
	for _, forbidden := range v.forbiddenCommands {
		if strings.Contains(lowerCmd, strings.ToLower(forbidden)) {
			return fmt.Errorf("forbidden command detected: %s", forbidden)
		}
	}

	// Check against forbidden patterns
	for _, pattern := range v.forbiddenPatterns {
		if pattern.MatchString(command) {
			return fmt.Errorf("dangerous command pattern detected")
		}
	}

	// Check for dangerous command and flag combinations
	for cmd, flags := range v.dangerousFlags {
		if strings.Contains(lowerCmd, cmd) {
			for _, flag := range flags {
				if strings.Contains(lowerCmd, strings.ToLower(flag)) {
					return fmt.Errorf("dangerous %s with flag %s detected", cmd, flag)
				}
			}
		}
	}

	// Check for multiple command chaining with dangerous operators
	if err := v.validateCommandChaining(command); err != nil {
		return err
	}

	// Check for shell injection attempts
	if err := v.validateShellInjection(command); err != nil {
		return err
	}

	return nil
}

// validateCommandChaining checks for dangerous command chaining.
func (v *Validator) validateCommandChaining(command string) error {
	// Check for dangerous pipe combinations
	dangerousPipes := []string{
		"| sh",
		"| bash",
		"| zsh",
		"| python",
		"| perl",
		"| ruby",
		"| php",
	}

	lowerCmd := strings.ToLower(command)
	for _, pipe := range dangerousPipes {
		if strings.Contains(lowerCmd, pipe) {
			return fmt.Errorf("dangerous pipe to interpreter detected: %s", pipe)
		}
	}

	// Check for excessive command chaining
	separators := []string{"&&", "||", ";", "|"}
	chainCount := 0
	for _, sep := range separators {
		chainCount += strings.Count(command, sep)
	}

	if chainCount > 10 {
		return fmt.Errorf("excessive command chaining detected (%d chains)", chainCount)
	}

	return nil
}

// validateShellInjection checks for shell injection attempts.
func (v *Validator) validateShellInjection(command string) error {
	// Check for suspicious character sequences
	suspiciousPatterns := []string{
		"${IFS}",     // Internal Field Separator abuse
		"$'\\x",      // Hex escape sequences
		"$'\\0",      // Octal escape sequences
		"${PATH",     // Path manipulation
		"${LD_",      // Library path manipulation
		"${BASH_ENV", // Bash environment manipulation
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(command, pattern) {
			return fmt.Errorf("suspicious pattern detected: %s", pattern)
		}
	}

	// Check for excessive special characters
	specialChars := []string{"$", "`", "\\", ";", "&", "|", "(", ")", "{", "}", "[", "]", "<", ">"}
	specialCount := 0
	for _, char := range specialChars {
		specialCount += strings.Count(command, char)
	}

	// If special characters make up more than 30% of the command, it's suspicious
	if len(command) > 0 && float64(specialCount)/float64(len(command)) > 0.3 {
		return fmt.Errorf("excessive special characters detected")
	}

	return nil
}

// IsSafeEnvVar checks if an environment variable name is safe to set.
func (v *Validator) IsSafeEnvVar(name string) bool {
	// Dangerous environment variables
	dangerousVars := []string{
		"PATH",
		"LD_LIBRARY_PATH",
		"LD_PRELOAD",
		"BASH_ENV",
		"ENV",
		"SHELL",
		"IFS",
		"PS1",
		"PS2",
		"PS3",
		"PS4",
	}

	upperName := strings.ToUpper(name)
	for _, dangerous := range dangerousVars {
		if upperName == dangerous {
			return false
		}
	}

	// Check for valid environment variable name
	validEnvVar := regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	return validEnvVar.MatchString(name)
}

// SetForbiddenCommands sets custom forbidden commands.
func (v *Validator) SetForbiddenCommands(commands []string) {
	v.forbiddenCommands = commands
}

// AddForbiddenCommand adds a forbidden command to the list.
func (v *Validator) AddForbiddenCommand(command string) {
	v.forbiddenCommands = append(v.forbiddenCommands, command)
}

// AddForbiddenPattern adds a forbidden pattern to check against.
func (v *Validator) AddForbiddenPattern(pattern string) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}
	v.forbiddenPatterns = append(v.forbiddenPatterns, regex)
	return nil
}

// SetMaxCommandLength sets the maximum allowed command length.
func (v *Validator) SetMaxCommandLength(length int) {
	if length > 0 {
		v.maxCommandLength = length
	}
}

// ValidatePath checks if a path is safe to use in commands.
func (v *Validator) ValidatePath(path string) error {
	// Check for path traversal
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal detected")
	}

	// Check for absolute system paths
	dangerousPaths := []string{
		"/etc",
		"/sys",
		"/proc",
		"/dev",
		"/boot",
		"/root",
		"/var/log",
		"/usr/bin",
		"/usr/sbin",
		"/bin",
		"/sbin",
	}

	for _, dangerous := range dangerousPaths {
		if strings.HasPrefix(path, dangerous) {
			return fmt.Errorf("access to system path denied: %s", dangerous)
		}
	}

	return nil
}