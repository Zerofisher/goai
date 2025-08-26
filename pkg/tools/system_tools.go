package tools

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// RunCommandTool executes system commands
type RunCommandTool struct{}

func NewRunCommandTool() *RunCommandTool {
	return &RunCommandTool{}
}

func (t *RunCommandTool) Name() string {
	return "runCommand"
}

func (t *RunCommandTool) Description() string {
	return "Execute system commands and return their output"
}

func (t *RunCommandTool) Parameters() ParameterSchema {
	return ParameterSchema{
		Required: []string{"command"},
		Properties: map[string]ParameterProperty{
			"command": {
				Type:        "string",
				Description: "Command to execute",
			},
			"args": {
				Type:        "array",
				Description: "Command arguments (optional)",
			},
			"workingDir": {
				Type:        "string",
				Description: "Working directory for command execution",
			},
			"timeout": {
				Type:        "number",
				Description: "Command timeout in seconds",
				Default:     30,
			},
			"captureOutput": {
				Type:        "boolean",
				Description: "Capture command output",
				Default:     true,
			},
		},
	}
}

func (t *RunCommandTool) Execute(ctx context.Context, params map[string]any) (*ToolResult, error) {
	command, ok := params["command"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "command parameter must be a string",
		}, nil
	}
	
	// Parse arguments
	var args []string
	if argsRaw, exists := params["args"]; exists {
		if argsArray, ok := argsRaw.([]any); ok {
			for _, arg := range argsArray {
				if argStr, ok := arg.(string); ok {
					args = append(args, argStr)
				}
			}
		}
	}
	
	timeout := 30.0
	if t, ok := params["timeout"].(float64); ok {
		timeout = t
	}
	
	workingDir, _ := params["workingDir"].(string)
	captureOutput, exists := params["captureOutput"].(bool)
	if !exists {
		captureOutput = true
	}
	
	// Create command with timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(timeoutCtx, command, args...)
	
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	
	var stdout, stderr bytes.Buffer
	if captureOutput {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}
	
	// Execute command
	err := cmd.Run()
	
	result := &ToolResult{
		Success: err == nil,
		Metadata: map[string]any{
			"command":     command,
			"args":        args,
			"working_dir": workingDir,
			"timeout":     timeout,
		},
	}
	
	if captureOutput {
		result.Data = map[string]any{
			"stdout": stdout.String(),
			"stderr": stderr.String(),
		}
		result.Output = stdout.String()
		if stderr.Len() > 0 {
			result.Output += "\nSTDERR:\n" + stderr.String()
		}
	}
	
	if err != nil {
		result.Error = fmt.Sprintf("command failed: %v", err)
	}
	
	return result, nil
}

func (t *RunCommandTool) RequiresConfirmation() bool {
	return true
}

func (t *RunCommandTool) Category() string {
	return "system"
}

func (t *RunCommandTool) Preview(ctx context.Context, params map[string]any) (*ToolPreview, error) {
	command, ok := params["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command parameter must be a string")
	}
	
	var args []string
	if argsRaw, exists := params["args"]; exists {
		if argsArray, ok := argsRaw.([]any); ok {
			for _, arg := range argsArray {
				if argStr, ok := arg.(string); ok {
					args = append(args, argStr)
				}
			}
		}
	}
	
	workingDir, _ := params["workingDir"].(string)
	
	fullCommand := command
	if len(args) > 0 {
		fullCommand += " " + strings.Join(args, " ")
	}
	
	description := fmt.Sprintf("Execute command: %s", fullCommand)
	if workingDir != "" {
		description += fmt.Sprintf(" (in %s)", workingDir)
	}
	
	return &ToolPreview{
		ToolName:   t.Name(),
		Parameters: params,
		Description: description,
		ExpectedChanges: []ExpectedChange{
			{
				Type:        "execute",
				Target:      fullCommand,
				Description: "Execute system command",
				Preview:     fullCommand,
			},
		},
		RequiresConfirmation: true,
	}, nil
}

// FetchTool fetches content from URLs
type FetchTool struct {
	httpClient *http.Client
}

func NewFetchTool() *FetchTool {
	return &FetchTool{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (t *FetchTool) Name() string {
	return "fetch"
}

func (t *FetchTool) Description() string {
	return "Fetch content from HTTP/HTTPS URLs"
}

func (t *FetchTool) Parameters() ParameterSchema {
	return ParameterSchema{
		Required: []string{"url"},
		Properties: map[string]ParameterProperty{
			"url": {
				Type:        "string",
				Description: "URL to fetch content from",
				Format:      "url",
			},
			"method": {
				Type:        "string",
				Description: "HTTP method (GET, POST, etc.)",
				Default:     "GET",
				Enum:        []string{"GET", "POST", "PUT", "DELETE", "HEAD"},
			},
			"headers": {
				Type:        "object",
				Description: "HTTP headers to include",
			},
			"body": {
				Type:        "string",
				Description: "Request body (for POST, PUT, etc.)",
			},
			"timeout": {
				Type:        "number",
				Description: "Request timeout in seconds",
				Default:     30,
			},
			"followRedirects": {
				Type:        "boolean",
				Description: "Follow HTTP redirects",
				Default:     true,
			},
		},
	}
}

func (t *FetchTool) Execute(ctx context.Context, params map[string]any) (*ToolResult, error) {
	url, ok := params["url"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "url parameter must be a string",
		}, nil
	}
	
	method, _ := params["method"].(string)
	if method == "" {
		method = "GET"
	}
	
	timeout := 30.0
	if t, ok := params["timeout"].(float64); ok {
		timeout = t
	}
	
	followRedirects, exists := params["followRedirects"].(bool)
	if !exists {
		followRedirects = true
	}
	
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	
	if !followRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	
	// Prepare request body
	var body io.Reader
	if bodyStr, exists := params["body"].(string); exists && bodyStr != "" {
		body = strings.NewReader(bodyStr)
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to create request: %v", err),
		}, nil
	}
	
	// Add headers
	if headersRaw, exists := params["headers"]; exists {
		if headers, ok := headersRaw.(map[string]any); ok {
			for key, value := range headers {
				if valueStr, ok := value.(string); ok {
					req.Header.Set(key, valueStr)
				}
			}
		}
	}
	
	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("request failed: %v", err),
		}, nil
	}
	defer func() { _ = resp.Body.Close() }()
	
	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read response body: %v", err),
		}, nil
	}
	
	// Prepare response headers
	respHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			respHeaders[key] = values[0]
		}
	}
	
	return &ToolResult{
		Success: resp.StatusCode >= 200 && resp.StatusCode < 300,
		Data: map[string]any{
			"status_code": resp.StatusCode,
			"headers":     respHeaders,
			"body":        string(respBody),
		},
		Output: fmt.Sprintf("Fetched %d bytes from %s (status: %d)", len(respBody), url, resp.StatusCode),
		Metadata: map[string]any{
			"url":           url,
			"method":        method,
			"status_code":   resp.StatusCode,
			"content_length": len(respBody),
		},
	}, nil
}

func (t *FetchTool) RequiresConfirmation() bool {
	return false
}

func (t *FetchTool) Category() string {
	return "system"
}

// GitTool provides Git operations
type GitTool struct{}

func NewGitTool() *GitTool {
	return &GitTool{}
}

func (t *GitTool) Name() string {
	return "git"
}

func (t *GitTool) Description() string {
	return "Perform Git operations (status, diff, log, etc.)"
}

func (t *GitTool) Parameters() ParameterSchema {
	return ParameterSchema{
		Required: []string{"operation"},
		Properties: map[string]ParameterProperty{
			"operation": {
				Type:        "string",
				Description: "Git operation to perform",
				Enum:        []string{"status", "diff", "log", "branch", "show"},
			},
			"args": {
				Type:        "array",
				Description: "Additional arguments for the Git command",
			},
			"workingDir": {
				Type:        "string",
				Description: "Working directory (Git repository root)",
			},
		},
	}
}

func (t *GitTool) Execute(ctx context.Context, params map[string]any) (*ToolResult, error) {
	operation, ok := params["operation"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "operation parameter must be a string",
		}, nil
	}
	
	// Parse additional arguments
	var args []string
	if argsRaw, exists := params["args"]; exists {
		if argsArray, ok := argsRaw.([]any); ok {
			for _, arg := range argsArray {
				if argStr, ok := arg.(string); ok {
					args = append(args, argStr)
				}
			}
		}
	}
	
	workingDir, _ := params["workingDir"].(string)
	
	// Build Git command
	gitArgs := []string{operation}
	gitArgs = append(gitArgs, args...)
	
	cmd := exec.CommandContext(ctx, "git", gitArgs...)
	if workingDir != "" {
		cmd.Dir = workingDir
	}
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	
	result := &ToolResult{
		Success: err == nil,
		Data: map[string]any{
			"stdout": stdout.String(),
			"stderr": stderr.String(),
		},
		Output: stdout.String(),
		Metadata: map[string]any{
			"operation":    operation,
			"args":         args,
			"working_dir":  workingDir,
		},
	}
	
	if err != nil {
		result.Error = fmt.Sprintf("git command failed: %v", err)
		if stderr.Len() > 0 {
			result.Error += "\nSTDERR: " + stderr.String()
		}
	}
	
	return result, nil
}

func (t *GitTool) RequiresConfirmation() bool {
	return false
}

func (t *GitTool) Category() string {
	return "system"
}