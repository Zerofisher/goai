package errors

import (
	"context"
	"testing"
	"time"
)

// TestDefaultErrorRecovery tests the basic recovery functionality
func TestDefaultErrorRecovery(t *testing.T) {
	recovery := NewDefaultErrorRecovery()
	
	if recovery == nil {
		t.Fatalf("Expected recovery instance to be created")
	}
	
	if recovery.maxRetries != 3 {
		t.Errorf("Expected maxRetries to be 3, got %d", recovery.maxRetries)
	}
	
	if recovery.retryDelay != time.Second*2 {
		t.Errorf("Expected retryDelay to be 2s, got %v", recovery.retryDelay)
	}
}

// TestCanRecover tests the CanRecover method
func TestCanRecover(t *testing.T) {
	recovery := NewDefaultErrorRecovery()
	
	tests := []struct {
		err      *GoAIError
		expected bool
	}{
		{
			err:      NewGoAIError(ErrInvalidInput, "invalid input"),
			expected: true,
		},
		{
			err:      NewGoAIError(ErrContextLoad, "context load failed"),
			expected: true,
		},
		{
			err:      NewGoAIError(ErrSystemFailure, "system failure"),
			expected: false,
		},
		{
			err:      NewGoAIError(ErrFileRead, "file read failed"),
			expected: false,
		},
	}
	
	for _, test := range tests {
		result := recovery.CanRecover(test.err)
		if result != test.expected {
			t.Errorf("Expected CanRecover(%v) to be %v, got %v", test.err.Code, test.expected, result)
		}
	}
}

// TestRecoverTimeout tests recovery for timeout errors
func TestRecoverTimeout(t *testing.T) {
	recovery := NewDefaultErrorRecovery()
	ctx := context.Background()
	
	err := NewGoAIError(ErrTimeout, "operation timed out")
	result, recoverErr := recovery.Recover(ctx, err)
	
	if recoverErr != nil {
		t.Errorf("Expected no error during recovery, got %v", recoverErr)
	}
	
	if !result.Success {
		t.Errorf("Expected recovery to be successful")
	}
	
	if result.RetryAfter != recovery.retryDelay {
		t.Errorf("Expected RetryAfter to be %v, got %v", recovery.retryDelay, result.RetryAfter)
	}
	
	if result.Message == "" {
		t.Errorf("Expected recovery message to be provided")
	}
}

// TestRecoverContextLoad tests recovery for context loading errors
func TestRecoverContextLoad(t *testing.T) {
	recovery := NewDefaultErrorRecovery()
	ctx := context.Background()
	
	err := NewGoAIError(ErrContextLoad, "failed to load context")
	result, recoverErr := recovery.Recover(ctx, err)
	
	if recoverErr != nil {
		t.Errorf("Expected no error during recovery, got %v", recoverErr)
	}
	
	if !result.Success {
		t.Errorf("Expected recovery to be successful")
	}
	
	if result.Alternative == nil {
		t.Errorf("Expected alternative to be provided")
	}
}

// TestRecoverAnalysis tests recovery for analysis errors
func TestRecoverAnalysis(t *testing.T) {
	recovery := NewDefaultErrorRecovery()
	ctx := context.Background()
	
	err := NewGoAIError(ErrAnalysisFailed, "analysis failed")
	result, recoverErr := recovery.Recover(ctx, err)
	
	if recoverErr != nil {
		t.Errorf("Expected no error during recovery, got %v", recoverErr)
	}
	
	if !result.Success {
		t.Errorf("Expected recovery to be successful")
	}
	
	if result.RetryAfter != recovery.retryDelay {
		t.Errorf("Expected RetryAfter to be %v, got %v", recovery.retryDelay, result.RetryAfter)
	}
}

// TestRecoverCodeGeneration tests recovery for code generation errors
func TestRecoverCodeGeneration(t *testing.T) {
	recovery := NewDefaultErrorRecovery()
	ctx := context.Background()
	
	err := NewGoAIError(ErrCodeGeneration, "code generation failed")
	result, recoverErr := recovery.Recover(ctx, err)
	
	if recoverErr != nil {
		t.Errorf("Expected no error during recovery, got %v", recoverErr)
	}
	
	if !result.Success {
		t.Errorf("Expected recovery to be successful")
	}
	
	if result.Alternative == nil {
		t.Errorf("Expected alternative to be provided")
	}
}

// TestRecoverNonRecoverable tests recovery for non-recoverable errors
func TestRecoverNonRecoverable(t *testing.T) {
	recovery := NewDefaultErrorRecovery()
	ctx := context.Background()
	
	err := NewGoAIError(ErrSystemFailure, "system failure")
	result, recoverErr := recovery.Recover(ctx, err)
	
	if recoverErr != nil {
		t.Errorf("Expected no error during recovery attempt, got %v", recoverErr)
	}
	
	if result.Success {
		t.Errorf("Expected recovery to fail for non-recoverable error")
	}
}

// TestSuggestAlternatives tests alternative suggestions
func TestSuggestAlternatives(t *testing.T) {
	recovery := NewDefaultErrorRecovery()
	
	tests := []struct {
		err               *GoAIError
		expectedMinAlts   int
		expectedAltName   string
	}{
		{
			err:               NewGoAIError(ErrAnalysisFailed, "analysis failed"),
			expectedMinAlts:   2,
			expectedAltName:   "Simplified Analysis",
		},
		{
			err:               NewGoAIError(ErrCodeGeneration, "code generation failed"),
			expectedMinAlts:   2,
			expectedAltName:   "Template-based Generation",
		},
		{
			err:               NewGoAIError(ErrContextLoad, "context load failed"),
			expectedMinAlts:   2,
			expectedAltName:   "Minimal Context Mode",
		},
		{
			err:               NewGoAIError(ErrSystemFailure, "system failure"),
			expectedMinAlts:   1,
			expectedAltName:   "Fallback Mode",
		},
	}
	
	for _, test := range tests {
		alternatives := recovery.SuggestAlternatives(test.err)
		
		if len(alternatives) < test.expectedMinAlts {
			t.Errorf("Expected at least %d alternatives for %v, got %d", test.expectedMinAlts, test.err.Code, len(alternatives))
		}
		
		found := false
		for _, alt := range alternatives {
			if alt.Name == test.expectedAltName {
				found = true
				break
			}
		}
		
		if !found {
			t.Errorf("Expected to find alternative '%s' for %v", test.expectedAltName, test.err.Code)
		}
	}
}

// TestRetryStrategy tests the retry strategy functionality
func TestRetryStrategy(t *testing.T) {
	strategy := NewRetryStrategy(3, time.Millisecond*10)
	
	if strategy.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries to be 3, got %d", strategy.MaxRetries)
	}
	
	if strategy.InitialDelay != time.Millisecond*10 {
		t.Errorf("Expected InitialDelay to be 10ms, got %v", strategy.InitialDelay)
	}
}

// TestRetryStrategySuccess tests successful retry execution
func TestRetryStrategySuccess(t *testing.T) {
	strategy := NewRetryStrategy(3, time.Millisecond*1)
	ctx := context.Background()
	
	attemptCount := 0
	operation := func() error {
		attemptCount++
		if attemptCount < 3 {
			return NewGoAIError(ErrTimeout, "temporary failure")
		}
		return nil // Success on third attempt
	}
	
	err := strategy.Execute(ctx, operation)
	if err != nil {
		t.Errorf("Expected operation to succeed after retries, got error: %v", err)
	}
	
	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}
}

// TestRetryStrategyMaxRetriesExceeded tests retry exhaustion
func TestRetryStrategyMaxRetriesExceeded(t *testing.T) {
	strategy := NewRetryStrategy(2, time.Millisecond*1)
	ctx := context.Background()
	
	attemptCount := 0
	operation := func() error {
		attemptCount++
		return NewGoAIError(ErrTimeout, "persistent failure")
	}
	
	err := strategy.Execute(ctx, operation)
	if err == nil {
		t.Errorf("Expected operation to fail after max retries")
	}
	
	if attemptCount != 2 {
		t.Errorf("Expected 2 attempts, got %d", attemptCount)
	}
	
	// Check that the error is wrapped correctly
	goaiErr, ok := err.(*GoAIError)
	if !ok {
		t.Errorf("Expected GoAIError, got %T", err)
	}
	
	if goaiErr.Code != ErrSystemFailure {
		t.Errorf("Expected error code %v, got %v", ErrSystemFailure, goaiErr.Code)
	}
}

// TestRetryStrategyNonRecoverableError tests that non-recoverable errors are not retried
func TestRetryStrategyNonRecoverableError(t *testing.T) {
	strategy := NewRetryStrategy(3, time.Millisecond*1)
	ctx := context.Background()
	
	attemptCount := 0
	operation := func() error {
		attemptCount++
		return NewGoAIError(ErrSystemFailure, "non-recoverable failure")
	}
	
	err := strategy.Execute(ctx, operation)
	if err == nil {
		t.Errorf("Expected operation to fail")
	}
	
	if attemptCount != 1 {
		t.Errorf("Expected only 1 attempt for non-recoverable error, got %d", attemptCount)
	}
}

// TestRetryStrategyContextCancellation tests context cancellation during retries
func TestRetryStrategyContextCancellation(t *testing.T) {
	strategy := NewRetryStrategy(5, time.Millisecond*100)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()
	
	attemptCount := 0
	operation := func() error {
		attemptCount++
		return NewGoAIError(ErrTimeout, "failure")
	}
	
	err := strategy.Execute(ctx, operation)
	if err == nil {
		t.Errorf("Expected operation to fail due to context cancellation")
	}
	
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context deadline exceeded error, got %v", err)
	}
	
	// Should have attempted at least once but not all 5 times
	if attemptCount == 0 || attemptCount >= 5 {
		t.Errorf("Expected 1-4 attempts due to context cancellation, got %d", attemptCount)
	}
}

// TestAlternativeStruct tests the Alternative struct
func TestAlternativeStruct(t *testing.T) {
	alt := Alternative{
		Name:        "Test Alternative",
		Description: "Test description",
		Parameters:  map[string]interface{}{"param1": "value1"},
		Confidence:  0.8,
	}
	
	if alt.Name != "Test Alternative" {
		t.Errorf("Expected name 'Test Alternative', got %s", alt.Name)
	}
	
	if alt.Confidence != 0.8 {
		t.Errorf("Expected confidence 0.8, got %f", alt.Confidence)
	}
	
	params, ok := alt.Parameters.(map[string]interface{})
	if !ok {
		t.Errorf("Expected parameters to be map[string]interface{}")
	}
	
	if params["param1"] != "value1" {
		t.Errorf("Expected param1 to be 'value1', got %v", params["param1"])
	}
}