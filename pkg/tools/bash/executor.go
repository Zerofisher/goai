package bash

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// BashTool implements bash command execution with security checks.
type BashTool struct {
	workDir   string
	timeout   time.Duration
	validator *Validator
	processor *OutputProcessor
}

// NewBashTool creates a new bash execution tool.
func NewBashTool(workDir string, timeout time.Duration) *BashTool {
	if timeout <= 0 {
		timeout = 30 * time.Second // Default 30 seconds
	}
	return &BashTool{
		workDir:   workDir,
		timeout:   timeout,
		validator: NewValidator(),
		processor: NewOutputProcessor(),
	}
}

// Name returns the name of the tool.
func (t *BashTool) Name() string {
	return "bash"
}

// Description returns the description of the tool.
func (t *BashTool) Description() string {
	return "Execute bash commands within the work directory"
}

// InputSchema returns the JSON schema for the input.
func (t *BashTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "Bash command to execute",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Command timeout in seconds (optional)",
			},
			"env": map[string]interface{}{
				"type":        "object",
				"description": "Additional environment variables (optional)",
			},
		},
		"required": []string{"command"},
	}
}

// Execute runs the bash command and returns its output.
func (t *BashTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
	// Extract command
	cmdRaw, ok := input["command"]
	if !ok {
		return "", fmt.Errorf("missing required parameter: command")
	}

	command, ok := cmdRaw.(string)
	if !ok {
		return "", fmt.Errorf("command must be a string")
	}

	// Extract timeout if provided
	timeout := t.timeout
	if timeoutRaw, ok := input["timeout"]; ok {
		if timeoutFloat, ok := timeoutRaw.(float64); ok {
			timeout = time.Duration(timeoutFloat) * time.Second
		}
	}

	// Extract environment variables
	env := make(map[string]string)
	if envRaw, ok := input["env"]; ok {
		if envMap, ok := envRaw.(map[string]interface{}); ok {
			for k, v := range envMap {
				if vStr, ok := v.(string); ok {
					env[k] = vStr
				}
			}
		}
	}

	// Validate the command
	if err := t.validator.ValidateCommand(command); err != nil {
		return "", fmt.Errorf("command validation failed: %w", err)
	}

	// Execute the command with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	output, err := t.executeCommand(ctx, command, env)
	if err != nil {
		return "", err
	}

	// Process the output
	processedOutput := t.processor.ProcessOutput(output, 5000) // Limit to 5000 lines

	return processedOutput, nil
}

// Validate validates the input parameters.
func (t *BashTool) Validate(input map[string]interface{}) error {
	cmdRaw, ok := input["command"]
	if !ok {
		return fmt.Errorf("missing required parameter: command")
	}

	command, ok := cmdRaw.(string)
	if !ok {
		return fmt.Errorf("command must be a string")
	}

	if command == "" {
		return fmt.Errorf("command cannot be empty")
	}

	// Validate timeout if provided
	if timeoutRaw, ok := input["timeout"]; ok {
		if timeoutFloat, ok := timeoutRaw.(float64); ok {
			if timeoutFloat <= 0 {
				return fmt.Errorf("timeout must be positive")
			}
			if timeoutFloat > 600 { // Max 10 minutes
				return fmt.Errorf("timeout cannot exceed 600 seconds")
			}
		} else {
			return fmt.Errorf("timeout must be a number")
		}
	}

	// Validate environment variables if provided
	if envRaw, ok := input["env"]; ok {
		if _, ok := envRaw.(map[string]interface{}); !ok {
			return fmt.Errorf("env must be an object")
		}
	}

	return nil
}

// executeCommand executes the bash command with the given context and environment.
func (t *BashTool) executeCommand(ctx context.Context, command string, env map[string]string) (string, error) {
	// Create command with bash
	cmd := exec.CommandContext(ctx, "bash", "-c", command)

	// Set working directory
	cmd.Dir = t.workDir

	// Set environment
	cmd.Env = os.Environ()
	for k, v := range env {
		// Filter out potentially dangerous environment variables
		if t.validator.IsSafeEnvVar(k) {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Capture output
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()

	// Combine stdout and stderr
	output := stdout.String()
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n"
		}
		output += "--- stderr ---\n" + stderr.String()
	}

	// Check for context cancellation (timeout)
	if ctx.Err() == context.DeadlineExceeded {
		return output, fmt.Errorf("command timed out after %v", t.timeout)
	}

	// Check for command errors
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return output, fmt.Errorf("command failed with exit code %d", exitErr.ExitCode())
		}
		return output, fmt.Errorf("command execution failed: %w", err)
	}

	return output, nil
}

// SetForbiddenCommands sets custom forbidden commands for the validator.
func (t *BashTool) SetForbiddenCommands(commands []string) {
	t.validator.SetForbiddenCommands(commands)
}

// ExecuteResult represents the result of a command execution.
type ExecuteResult struct {
	Output   string        `json:"output"`
	ExitCode int           `json:"exit_code"`
	Duration time.Duration `json:"duration"`
	Error    string        `json:"error,omitempty"`
}

// ExecuteWithResult executes a command and returns detailed result.
func (t *BashTool) ExecuteWithResult(ctx context.Context, command string, env map[string]string) (*ExecuteResult, error) {
	startTime := time.Now()

	// Validate the command
	if err := t.validator.ValidateCommand(command); err != nil {
		return &ExecuteResult{
			Error:    err.Error(),
			Duration: time.Since(startTime),
		}, err
	}

	// Create command with bash
	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd.Dir = t.workDir
	cmd.Env = os.Environ()

	for k, v := range env {
		if t.validator.IsSafeEnvVar(k) {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Capture output
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()
	duration := time.Since(startTime)

	// Build result
	result := &ExecuteResult{
		Output:   t.combineOutput(stdout.String(), stderr.String()),
		Duration: duration,
	}

	// Set exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Error = fmt.Sprintf("command failed with exit code %d", exitErr.ExitCode())
		} else {
			result.ExitCode = -1
			result.Error = err.Error()
		}
	} else {
		result.ExitCode = 0
	}

	return result, err
}

// combineOutput combines stdout and stderr into a single string.
func (t *BashTool) combineOutput(stdout, stderr string) string {
	output := strings.TrimSpace(stdout)
	stderrTrimmed := strings.TrimSpace(stderr)

	if stderrTrimmed != "" {
		if output != "" {
			output += "\n\n"
		}
		output += "--- stderr ---\n" + stderrTrimmed
	}

	return output
}