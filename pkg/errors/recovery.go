package errors

import (
	"context"
	"time"
)

// RecoveryResult represents the result of an error recovery attempt
type RecoveryResult struct {
	Success     bool        `json:"success"`
	Message     string      `json:"message"`
	RetryAfter  time.Duration `json:"retry_after,omitempty"`
	Alternative interface{} `json:"alternative,omitempty"`
}

// Alternative represents an alternative approach when recovery fails
type Alternative struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters,omitempty"`
	Confidence  float64     `json:"confidence"`
}

// ErrorRecovery defines the interface for error recovery mechanisms
type ErrorRecovery interface {
	CanRecover(err *GoAIError) bool
	Recover(ctx context.Context, err *GoAIError) (*RecoveryResult, error)
	SuggestAlternatives(err *GoAIError) []Alternative
}

// DefaultErrorRecovery implements basic error recovery strategies
type DefaultErrorRecovery struct {
	maxRetries    int
	retryDelay    time.Duration
	backoffFactor float64
}

// NewDefaultErrorRecovery creates a new default error recovery instance
func NewDefaultErrorRecovery() *DefaultErrorRecovery {
	return &DefaultErrorRecovery{
		maxRetries:    3,
		retryDelay:    time.Second * 2,
		backoffFactor: 2.0,
	}
}

// CanRecover determines if an error can be recovered from
func (r *DefaultErrorRecovery) CanRecover(err *GoAIError) bool {
	return err.Recoverable
}

// Recover attempts to recover from an error
func (r *DefaultErrorRecovery) Recover(ctx context.Context, err *GoAIError) (*RecoveryResult, error) {
	if !r.CanRecover(err) {
		return &RecoveryResult{
			Success: false,
			Message: "Error is not recoverable",
		}, nil
	}

	switch err.Code {
	case ErrTimeout:
		return &RecoveryResult{
			Success:    true,
			Message:    "Retry with increased timeout",
			RetryAfter: r.retryDelay,
		}, nil

	case ErrContextLoad:
		return r.recoverContextLoad(err)

	case ErrAnalysisFailed:
		return r.recoverAnalysis(err)

	case ErrCodeGeneration:
		return r.recoverCodeGeneration(err)

	default:
		return &RecoveryResult{
			Success: false,
			Message: "No specific recovery strategy available",
		}, nil
	}
}

// SuggestAlternatives provides alternative approaches when recovery fails
func (r *DefaultErrorRecovery) SuggestAlternatives(err *GoAIError) []Alternative {
	var alternatives []Alternative

	switch err.Code {
	case ErrAnalysisFailed:
		alternatives = append(alternatives, Alternative{
			Name:        "Simplified Analysis",
			Description: "Use a simpler analysis approach with reduced complexity",
			Confidence:  0.7,
		})
		alternatives = append(alternatives, Alternative{
			Name:        "Manual Guidance Mode",
			Description: "Request more detailed user input to guide the analysis",
			Confidence:  0.8,
		})

	case ErrCodeGeneration:
		alternatives = append(alternatives, Alternative{
			Name:        "Template-based Generation",
			Description: "Use predefined templates instead of dynamic generation",
			Confidence:  0.6,
		})
		alternatives = append(alternatives, Alternative{
			Name:        "Step-by-step Generation",
			Description: "Generate code in smaller, incremental steps",
			Confidence:  0.8,
		})

	case ErrContextLoad:
		alternatives = append(alternatives, Alternative{
			Name:        "Minimal Context Mode",
			Description: "Proceed with basic project information only",
			Confidence:  0.5,
		})
		alternatives = append(alternatives, Alternative{
			Name:        "Manual Context Input",
			Description: "Request user to provide project context manually",
			Confidence:  0.9,
		})

	default:
		alternatives = append(alternatives, Alternative{
			Name:        "Fallback Mode",
			Description: "Use simplified approach with reduced functionality",
			Confidence:  0.4,
		})
	}

	return alternatives
}

// recoverContextLoad attempts to recover from context loading errors
func (r *DefaultErrorRecovery) recoverContextLoad(err *GoAIError) (*RecoveryResult, error) {
	// Try alternative context loading strategies
	alternatives := []string{
		"Try loading context from parent directory",
		"Use default project structure",
		"Request manual context specification",
	}

	return &RecoveryResult{
		Success:     true,
		Message:     "Use alternative context loading strategy",
		Alternative: alternatives,
	}, nil
}

// recoverAnalysis attempts to recover from analysis errors
func (r *DefaultErrorRecovery) recoverAnalysis(err *GoAIError) (*RecoveryResult, error) {
	return &RecoveryResult{
		Success:    true,
		Message:    "Retry analysis with simplified approach",
		RetryAfter: r.retryDelay,
		Alternative: map[string]interface{}{
			"use_simplified_prompts": true,
			"reduce_complexity":      true,
		},
	}, nil
}

// recoverCodeGeneration attempts to recover from code generation errors
func (r *DefaultErrorRecovery) recoverCodeGeneration(err *GoAIError) (*RecoveryResult, error) {
	return &RecoveryResult{
		Success:    true,
		Message:    "Retry with template-based generation",
		RetryAfter: r.retryDelay,
		Alternative: map[string]interface{}{
			"use_templates": true,
			"incremental":   true,
		},
	}, nil
}

// RetryStrategy defines how to handle retries with exponential backoff
type RetryStrategy struct {
	MaxRetries    int
	InitialDelay  time.Duration
	BackoffFactor float64
	MaxDelay      time.Duration
}

// NewRetryStrategy creates a new retry strategy
func NewRetryStrategy(maxRetries int, initialDelay time.Duration) *RetryStrategy {
	return &RetryStrategy{
		MaxRetries:    maxRetries,
		InitialDelay:  initialDelay,
		BackoffFactor: 2.0,
		MaxDelay:      time.Minute * 5,
	}
}

// Execute runs an operation with retry logic
func (rs *RetryStrategy) Execute(ctx context.Context, operation func() error) error {
	var lastErr error
	delay := rs.InitialDelay

	for attempt := 0; attempt < rs.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// Continue with retry
			}
		}

		err := operation()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is recoverable
		if goaiErr, ok := err.(*GoAIError); ok && !goaiErr.Recoverable {
			return err // Don't retry non-recoverable errors
		}

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * rs.BackoffFactor)
		if delay > rs.MaxDelay {
			delay = rs.MaxDelay
		}
	}

	return WrapError(ErrSystemFailure, "maximum retry attempts exceeded", lastErr)
}