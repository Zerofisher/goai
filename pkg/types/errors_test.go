package types

import (
	"errors"
	"strings"
	"testing"
)

func TestNewAgentError(t *testing.T) {
	err := NewAgentError("TEST_CODE", "Test message")

	if err.Code != "TEST_CODE" {
		t.Errorf("expected code 'TEST_CODE', got '%s'", err.Code)
	}

	if err.Message != "Test message" {
		t.Errorf("expected message 'Test message', got '%s'", err.Message)
	}

	if err.Details == nil {
		t.Error("expected details to be initialized")
	}
}

func TestAgentError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AgentError
		contains string
	}{
		{
			name: "error without details",
			err:  NewAgentError("CODE", "message"),
			contains: "[CODE] message",
		},
		{
			name: "error with details",
			err:  NewAgentError("CODE", "message").WithDetail("key", "value"),
			contains: "[CODE] message - Details:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.err.Error()
			if !strings.Contains(errStr, tt.contains) {
				t.Errorf("Error() = %v, want to contain %v", errStr, tt.contains)
			}
		})
	}
}

func TestAgentError_WithDetail(t *testing.T) {
	err := NewAgentError("CODE", "message")
	result := err.WithDetail("key1", "value1")
	result = result.WithDetail("key2", 123)

	if result.Details["key1"] != "value1" {
		t.Errorf("expected details['key1'] = 'value1', got %v", result.Details["key1"])
	}

	if result.Details["key2"] != 123 {
		t.Errorf("expected details['key2'] = 123, got %v", result.Details["key2"])
	}
}

func TestAgentError_WithDetails(t *testing.T) {
	err := NewAgentError("CODE", "message")
	details := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}
	result := err.WithDetails(details)

	if len(result.Details) != 3 {
		t.Errorf("expected 3 details, got %d", len(result.Details))
	}

	if result.Details["key1"] != "value1" {
		t.Errorf("expected details['key1'] = 'value1', got %v", result.Details["key1"])
	}

	if result.Details["key2"] != 123 {
		t.Errorf("expected details['key2'] = 123, got %v", result.Details["key2"])
	}

	if result.Details["key3"] != true {
		t.Errorf("expected details['key3'] = true, got %v", result.Details["key3"])
	}
}

func TestAgentError_IsCode(t *testing.T) {
	err := NewAgentError(ErrCodeToolNotFound, "Tool not found")

	if !err.IsCode(ErrCodeToolNotFound) {
		t.Error("expected IsCode(ErrCodeToolNotFound) to return true")
	}

	if err.IsCode(ErrCodeFileNotFound) {
		t.Error("expected IsCode(ErrCodeFileNotFound) to return false")
	}
}

func TestIsAgentError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "agent error",
			err:      NewAgentError("CODE", "message"),
			expected: true,
		},
		{
			name:     "standard error",
			err:      errors.New("standard error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAgentError(tt.err); got != tt.expected {
				t.Errorf("IsAgentError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetAgentError(t *testing.T) {
	agentErr := NewAgentError("CODE", "message")
	standardErr := errors.New("standard error")

	// Test with AgentError
	err, ok := GetAgentError(agentErr)
	if !ok {
		t.Error("expected GetAgentError to return true for AgentError")
	}
	if err != nil && err.Code != "CODE" {
		t.Errorf("expected code 'CODE', got '%s'", err.Code)
	}

	// Test with standard error
	_, ok = GetAgentError(standardErr)
	if ok {
		t.Error("expected GetAgentError to return false for standard error")
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	wrapped := WrapError("WRAP_CODE", "Wrapped message", originalErr)

	if wrapped.Code != "WRAP_CODE" {
		t.Errorf("expected code 'WRAP_CODE', got '%s'", wrapped.Code)
	}

	if wrapped.Message != "Wrapped message" {
		t.Errorf("expected message 'Wrapped message', got '%s'", wrapped.Message)
	}

	originalErrStr, ok := wrapped.Details["original_error"].(string)
	if !ok {
		t.Fatal("expected original_error in details")
	}

	if originalErrStr != "original error" {
		t.Errorf("expected original error 'original error', got '%s'", originalErrStr)
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  *AgentError
		code string
	}{
		{"ErrToolNotFound", ErrToolNotFound, ErrCodeToolNotFound},
		{"ErrToolTimeout", ErrToolTimeout, ErrCodeToolTimeout},
		{"ErrFileNotFound", ErrFileNotFound, ErrCodeFileNotFound},
		{"ErrFileAccessDenied", ErrFileAccessDenied, ErrCodeFileAccessDenied},
		{"ErrPathEscape", ErrPathEscape, ErrCodePathEscape},
		{"ErrFileTooLarge", ErrFileTooLarge, ErrCodeFileTooLarge},
		{"ErrTodoLimitExceeded", ErrTodoLimitExceeded, ErrCodeTodoLimit},
		{"ErrTodoInvalidState", ErrTodoInvalidState, ErrCodeTodoInvalid},
		{"ErrTodoNotFound", ErrTodoNotFound, ErrCodeTodoNotFound},
		{"ErrInvalidInput", ErrInvalidInput, ErrCodeInvalidInput},
		{"ErrInternalError", ErrInternalError, ErrCodeInternalError},
		{"ErrNotImplemented", ErrNotImplemented, ErrCodeNotImplemented},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("expected code %s, got %s", tt.code, tt.err.Code)
			}
			if tt.err.Message == "" {
				t.Error("expected non-empty message")
			}
		})
	}
}

func TestErrorCodes(t *testing.T) {
	// Just verify that all error codes are defined
	codes := []string{
		ErrCodeToolNotFound,
		ErrCodeToolExecution,
		ErrCodeToolValidation,
		ErrCodeToolTimeout,
		ErrCodeToolPermission,
		ErrCodeFileNotFound,
		ErrCodeFileAccessDenied,
		ErrCodePathEscape,
		ErrCodeFileTooLarge,
		ErrCodeLLMConnection,
		ErrCodeLLMRateLimit,
		ErrCodeLLMInvalidKey,
		ErrCodeLLMTokenLimit,
		ErrCodeLLMParsing,
		ErrCodeConfigInvalid,
		ErrCodeConfigMissing,
		ErrCodeConfigLoad,
		ErrCodeTodoLimit,
		ErrCodeTodoInvalid,
		ErrCodeTodoNotFound,
		ErrCodeInvalidInput,
		ErrCodeInternalError,
		ErrCodeNotImplemented,
	}

	for _, code := range codes {
		if code == "" {
			t.Error("found empty error code")
		}
	}
}