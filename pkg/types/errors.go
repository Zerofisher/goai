package types

import (
	"encoding/json"
	"fmt"
)

// AgentError represents a structured error with code and details
type AgentError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *AgentError) Error() string {
	if len(e.Details) > 0 {
		detailsJSON, _ := json.Marshal(e.Details)
		return fmt.Sprintf("[%s] %s - Details: %s", e.Code, e.Message, string(detailsJSON))
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewAgentError creates a new AgentError
func NewAgentError(code, message string) *AgentError {
	return &AgentError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// WithDetail adds a detail to the error
func (e *AgentError) WithDetail(key string, value interface{}) *AgentError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithDetails adds multiple details to the error
func (e *AgentError) WithDetails(details map[string]interface{}) *AgentError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// IsCode checks if the error has a specific code
func (e *AgentError) IsCode(code string) bool {
	return e.Code == code
}

// Common error codes
const (
	// Tool errors
	ErrCodeToolNotFound     = "TOOL_NOT_FOUND"
	ErrCodeToolExecution    = "TOOL_EXECUTION_ERROR"
	ErrCodeToolValidation   = "TOOL_VALIDATION_ERROR"
	ErrCodeToolTimeout      = "TOOL_TIMEOUT"
	ErrCodeToolPermission   = "TOOL_PERMISSION_DENIED"

	// File operation errors
	ErrCodeFileNotFound     = "FILE_NOT_FOUND"
	ErrCodeFileAccessDenied = "FILE_ACCESS_DENIED"
	ErrCodePathEscape       = "PATH_ESCAPE"
	ErrCodeFileTooLarge     = "FILE_TOO_LARGE"

	// LLM errors
	ErrCodeLLMConnection   = "LLM_CONNECTION_ERROR"
	ErrCodeLLMRateLimit    = "LLM_RATE_LIMIT"
	ErrCodeLLMInvalidKey   = "LLM_INVALID_KEY"
	ErrCodeLLMTokenLimit   = "LLM_TOKEN_LIMIT"
	ErrCodeLLMParsing      = "LLM_PARSING_ERROR"

	// Configuration errors
	ErrCodeConfigInvalid   = "CONFIG_INVALID"
	ErrCodeConfigMissing   = "CONFIG_MISSING"
	ErrCodeConfigLoad      = "CONFIG_LOAD_ERROR"

	// Todo errors
	ErrCodeTodoLimit       = "TODO_LIMIT_EXCEEDED"
	ErrCodeTodoInvalid     = "TODO_INVALID_STATE"
	ErrCodeTodoNotFound    = "TODO_NOT_FOUND"

	// General errors
	ErrCodeInvalidInput    = "INVALID_INPUT"
	ErrCodeInternalError   = "INTERNAL_ERROR"
	ErrCodeNotImplemented  = "NOT_IMPLEMENTED"
)

// Common errors
var (
	// Tool errors
	ErrToolNotFound = NewAgentError(ErrCodeToolNotFound, "Tool not found")
	ErrToolTimeout  = NewAgentError(ErrCodeToolTimeout, "Tool execution timeout")

	// File errors
	ErrFileNotFound     = NewAgentError(ErrCodeFileNotFound, "File not found")
	ErrFileAccessDenied = NewAgentError(ErrCodeFileAccessDenied, "File access denied")
	ErrPathEscape       = NewAgentError(ErrCodePathEscape, "Path escapes workspace")
	ErrFileTooLarge     = NewAgentError(ErrCodeFileTooLarge, "File too large")

	// Todo errors
	ErrTodoLimitExceeded = NewAgentError(ErrCodeTodoLimit, "Todo limit exceeded")
	ErrTodoInvalidState  = NewAgentError(ErrCodeTodoInvalid, "Invalid todo state transition")
	ErrTodoNotFound      = NewAgentError(ErrCodeTodoNotFound, "Todo item not found")

	// General errors
	ErrInvalidInput   = NewAgentError(ErrCodeInvalidInput, "Invalid input")
	ErrInternalError  = NewAgentError(ErrCodeInternalError, "Internal error")
	ErrNotImplemented = NewAgentError(ErrCodeNotImplemented, "Not implemented")
)

// IsAgentError checks if an error is an AgentError
func IsAgentError(err error) bool {
	_, ok := err.(*AgentError)
	return ok
}

// GetAgentError converts an error to AgentError if possible
func GetAgentError(err error) (*AgentError, bool) {
	agentErr, ok := err.(*AgentError)
	return agentErr, ok
}

// WrapError wraps a standard error into an AgentError
func WrapError(code, message string, err error) *AgentError {
	return NewAgentError(code, message).WithDetail("original_error", err.Error())
}