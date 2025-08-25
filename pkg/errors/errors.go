package errors

import "fmt"

// ErrorCode represents specific error types
type ErrorCode string

const (
	// Input validation errors
	ErrInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrMalformedRequest ErrorCode = "MALFORMED_REQUEST"
	ErrMissingRequired  ErrorCode = "MISSING_REQUIRED"

	// Context loading errors
	ErrContextLoad      ErrorCode = "CONTEXT_LOAD"
	ErrConfigLoad       ErrorCode = "CONFIG_LOAD"
	ErrFileRead         ErrorCode = "FILE_READ"
	ErrGitIntegration   ErrorCode = "GIT_INTEGRATION"

	// Analysis errors
	ErrAnalysisFailed   ErrorCode = "ANALYSIS_FAILED"
	ErrPlanningFailed   ErrorCode = "PLANNING_FAILED"
	ErrReasoningFailed  ErrorCode = "REASONING_FAILED"

	// Generation errors
	ErrCodeGeneration   ErrorCode = "CODE_GENERATION"
	ErrTestGeneration   ErrorCode = "TEST_GENERATION"
	ErrDocGeneration    ErrorCode = "DOC_GENERATION"
	ErrTemplateError    ErrorCode = "TEMPLATE_ERROR"

	// Validation errors
	ErrStaticAnalysis   ErrorCode = "STATIC_ANALYSIS"
	ErrTestExecution    ErrorCode = "TEST_EXECUTION"
	ErrCompliance       ErrorCode = "COMPLIANCE"
	ErrValidationFailed ErrorCode = "VALIDATION_FAILED"

	// System errors
	ErrSystemFailure    ErrorCode = "SYSTEM_FAILURE"
	ErrDependencyError  ErrorCode = "DEPENDENCY_ERROR"
	ErrTimeout          ErrorCode = "TIMEOUT"
	ErrResourceLimit    ErrorCode = "RESOURCE_LIMIT"
)

// GoAIError represents a structured error with recovery information
type GoAIError struct {
	Code        ErrorCode   `json:"code"`
	Message     string      `json:"message"`
	Context     interface{} `json:"context,omitempty"`
	Suggestions []string    `json:"suggestions,omitempty"`
	Recoverable bool        `json:"recoverable"`
	Cause       error       `json:"-"`
}

// Error implements the error interface
func (e *GoAIError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *GoAIError) Unwrap() error {
	return e.Cause
}

// NewGoAIError creates a new GoAI error
func NewGoAIError(code ErrorCode, message string) *GoAIError {
	return &GoAIError{
		Code:        code,
		Message:     message,
		Recoverable: isRecoverable(code),
	}
}

// WrapError wraps an existing error with GoAI error information
func WrapError(code ErrorCode, message string, cause error) *GoAIError {
	return &GoAIError{
		Code:        code,
		Message:     message,
		Cause:       cause,
		Recoverable: isRecoverable(code),
	}
}

// WithContext adds context information to the error
func (e *GoAIError) WithContext(context interface{}) *GoAIError {
	e.Context = context
	return e
}

// WithSuggestions adds recovery suggestions to the error
func (e *GoAIError) WithSuggestions(suggestions ...string) *GoAIError {
	e.Suggestions = append(e.Suggestions, suggestions...)
	return e
}

// isRecoverable determines if an error type is recoverable
func isRecoverable(code ErrorCode) bool {
	recoverableErrors := map[ErrorCode]bool{
		ErrInvalidInput:     true,
		ErrMalformedRequest: true,
		ErrMissingRequired:  true,
		ErrContextLoad:      true,
		ErrConfigLoad:       true,
		ErrAnalysisFailed:   true,
		ErrPlanningFailed:   true,
		ErrCodeGeneration:   true,
		ErrTestGeneration:   true,
		ErrDocGeneration:    true,
		ErrStaticAnalysis:   true,
		ErrCompliance:       true,
		ErrTimeout:          true,
	}
	return recoverableErrors[code]
}

// Common error constructors for convenience

// NewValidationError creates a validation error
func NewValidationError(field, message string) *GoAIError {
	return NewGoAIError(ErrInvalidInput, fmt.Sprintf("validation failed for %s: %s", field, message)).
		WithContext(map[string]string{"field": field}).
		WithSuggestions(fmt.Sprintf("Please check the %s field and try again", field))
}

// NewContextLoadError creates a context loading error
func NewContextLoadError(path string, cause error) *GoAIError {
	return WrapError(ErrContextLoad, fmt.Sprintf("failed to load context from %s", path), cause).
		WithContext(map[string]string{"path": path}).
		WithSuggestions("Check if the path exists and is accessible", "Verify file permissions")
}

// NewAnalysisError creates an analysis error
func NewAnalysisError(stage string, cause error) *GoAIError {
	return WrapError(ErrAnalysisFailed, fmt.Sprintf("analysis failed at stage: %s", stage), cause).
		WithContext(map[string]string{"stage": stage}).
		WithSuggestions("Try simplifying the problem description", "Provide more context about the requirements")
}

// NewGenerationError creates a generation error
func NewGenerationError(component string, cause error) *GoAIError {
	return WrapError(ErrCodeGeneration, fmt.Sprintf("code generation failed for: %s", component), cause).
		WithContext(map[string]string{"component": component}).
		WithSuggestions("Check the execution plan for issues", "Verify the analysis results are complete")
}

// NewSystemError creates a system error
func NewSystemError(operation string, cause error) *GoAIError {
	return WrapError(ErrSystemFailure, fmt.Sprintf("system failure during: %s", operation), cause).
		WithContext(map[string]string{"operation": operation}).
		WithSuggestions("Check system resources", "Retry the operation")
}