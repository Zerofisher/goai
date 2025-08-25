package errors

import (
	"errors"
	"testing"
)

// TestGoAIError tests the basic GoAIError functionality
func TestGoAIError(t *testing.T) {
	err := NewGoAIError(ErrInvalidInput, "test error message")
	
	if err.Code != ErrInvalidInput {
		t.Errorf("Expected code %v, got %v", ErrInvalidInput, err.Code)
	}
	
	if err.Message != "test error message" {
		t.Errorf("Expected message 'test error message', got '%v'", err.Message)
	}
	
	if !err.Recoverable {
		t.Errorf("Expected error to be recoverable")
	}
	
	expectedError := "[INVALID_INPUT] test error message"
	if err.Error() != expectedError {
		t.Errorf("Expected error string '%v', got '%v'", expectedError, err.Error())
	}
}

// TestWrapError tests wrapping existing errors
func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := WrapError(ErrContextLoad, "context loading failed", originalErr)
	
	if wrappedErr.Code != ErrContextLoad {
		t.Errorf("Expected code %v, got %v", ErrContextLoad, wrappedErr.Code)
	}
	
	if wrappedErr.Cause != originalErr {
		t.Errorf("Expected cause to be original error")
	}
	
	if wrappedErr.Unwrap() != originalErr {
		t.Errorf("Expected Unwrap to return original error")
	}
	
	expectedError := "[CONTEXT_LOAD] context loading failed: original error"
	if wrappedErr.Error() != expectedError {
		t.Errorf("Expected error string '%v', got '%v'", expectedError, wrappedErr.Error())
	}
}

// TestErrorWithContext tests adding context to errors
func TestErrorWithContext(t *testing.T) {
	err := NewGoAIError(ErrAnalysisFailed, "analysis failed")
	context := map[string]string{"stage": "syntax"}
	
	err = err.WithContext(context)
	
	if err.Context == nil {
		t.Errorf("Expected context to be set")
	}
	
	contextMap, ok := err.Context.(map[string]string)
	if !ok {
		t.Errorf("Expected context to be map[string]string")
	}
	
	if contextMap["stage"] != "syntax" {
		t.Errorf("Expected context stage 'syntax', got '%v'", contextMap["stage"])
	}
}

// TestErrorWithSuggestions tests adding suggestions to errors
func TestErrorWithSuggestions(t *testing.T) {
	err := NewGoAIError(ErrCodeGeneration, "code generation failed")
	suggestions := []string{"Check the execution plan", "Verify analysis results"}
	
	err = err.WithSuggestions(suggestions...)
	
	if len(err.Suggestions) != 2 {
		t.Errorf("Expected 2 suggestions, got %d", len(err.Suggestions))
	}
	
	if err.Suggestions[0] != "Check the execution plan" {
		t.Errorf("Expected first suggestion 'Check the execution plan', got '%v'", err.Suggestions[0])
	}
}

// TestIsRecoverable tests the recoverability logic
func TestIsRecoverable(t *testing.T) {
	tests := []struct {
		code        ErrorCode
		recoverable bool
	}{
		{ErrInvalidInput, true},
		{ErrContextLoad, true},
		{ErrAnalysisFailed, true},
		{ErrCodeGeneration, true},
		{ErrSystemFailure, false},
		{ErrFileRead, false},
		{ErrTestExecution, false},
	}
	
	for _, test := range tests {
		result := isRecoverable(test.code)
		if result != test.recoverable {
			t.Errorf("Expected %v to be recoverable=%v, got %v", test.code, test.recoverable, result)
		}
	}
}

// TestNewValidationError tests the validation error constructor
func TestNewValidationError(t *testing.T) {
	err := NewValidationError("username", "cannot be empty")
	
	if err.Code != ErrInvalidInput {
		t.Errorf("Expected code %v, got %v", ErrInvalidInput, err.Code)
	}
	
	expectedMessage := "validation failed for username: cannot be empty"
	if err.Message != expectedMessage {
		t.Errorf("Expected message '%v', got '%v'", expectedMessage, err.Message)
	}
	
	if len(err.Suggestions) == 0 {
		t.Errorf("Expected suggestions to be provided")
	}
}

// TestNewContextLoadError tests the context loading error constructor
func TestNewContextLoadError(t *testing.T) {
	originalErr := errors.New("file not found")
	err := NewContextLoadError("/path/to/config", originalErr)
	
	if err.Code != ErrContextLoad {
		t.Errorf("Expected code %v, got %v", ErrContextLoad, err.Code)
	}
	
	if err.Cause != originalErr {
		t.Errorf("Expected cause to be original error")
	}
	
	if len(err.Suggestions) == 0 {
		t.Errorf("Expected suggestions to be provided")
	}
}

// TestNewAnalysisError tests the analysis error constructor
func TestNewAnalysisError(t *testing.T) {
	originalErr := errors.New("prompt failed")
	err := NewAnalysisError("problem understanding", originalErr)
	
	if err.Code != ErrAnalysisFailed {
		t.Errorf("Expected code %v, got %v", ErrAnalysisFailed, err.Code)
	}
	
	expectedMessage := "analysis failed at stage: problem understanding"
	if err.Message != expectedMessage {
		t.Errorf("Expected message '%v', got '%v'", expectedMessage, err.Message)
	}
}

// TestNewGenerationError tests the generation error constructor
func TestNewGenerationError(t *testing.T) {
	originalErr := errors.New("template not found")
	err := NewGenerationError("main function", originalErr)
	
	if err.Code != ErrCodeGeneration {
		t.Errorf("Expected code %v, got %v", ErrCodeGeneration, err.Code)
	}
	
	expectedMessage := "code generation failed for: main function"
	if err.Message != expectedMessage {
		t.Errorf("Expected message '%v', got '%v'", expectedMessage, err.Message)
	}
}

// TestNewSystemError tests the system error constructor
func TestNewSystemError(t *testing.T) {
	originalErr := errors.New("out of memory")
	err := NewSystemError("reasoning chain execution", originalErr)
	
	if err.Code != ErrSystemFailure {
		t.Errorf("Expected code %v, got %v", ErrSystemFailure, err.Code)
	}
	
	expectedMessage := "system failure during: reasoning chain execution"
	if err.Message != expectedMessage {
		t.Errorf("Expected message '%v', got '%v'", expectedMessage, err.Message)
	}
	
	if err.Recoverable {
		t.Errorf("Expected system error to not be recoverable")
	}
}

// TestErrorCodes tests that all error codes are defined correctly
func TestErrorCodes(t *testing.T) {
	codes := []ErrorCode{
		ErrInvalidInput,
		ErrMalformedRequest,
		ErrMissingRequired,
		ErrContextLoad,
		ErrConfigLoad,
		ErrFileRead,
		ErrGitIntegration,
		ErrAnalysisFailed,
		ErrPlanningFailed,
		ErrReasoningFailed,
		ErrCodeGeneration,
		ErrTestGeneration,
		ErrDocGeneration,
		ErrTemplateError,
		ErrStaticAnalysis,
		ErrTestExecution,
		ErrCompliance,
		ErrValidationFailed,
		ErrSystemFailure,
		ErrDependencyError,
		ErrTimeout,
		ErrResourceLimit,
	}
	
	for _, code := range codes {
		if code == "" {
			t.Errorf("Error code should not be empty")
		}
		
		// Test that we can create errors with each code
		err := NewGoAIError(code, "test message")
		if err.Code != code {
			t.Errorf("Failed to create error with code %v", code)
		}
	}
}