package dispatcher

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Zerofisher/goai/pkg/types"
)

// LoggingMiddleware adds logging to tool execution.
func LoggingMiddleware(logger *log.Logger) Middleware {
	return func(ctx context.Context, toolUse types.ToolUse, next ExecuteFunc) types.ToolResult {
		start := time.Now()

		// Log the start
		if logger != nil {
			logger.Printf("[TOOL] Starting execution: %s (ID: %s)", toolUse.Name, toolUse.ID)
		}

		// Execute
		result := next(ctx, toolUse)

		// Log the result
		if logger != nil {
			duration := time.Since(start)
			if result.IsError {
				logger.Printf("[TOOL] Failed: %s (ID: %s) - Duration: %v - Error: %s",
					toolUse.Name, toolUse.ID, duration, result.Content)
			} else {
				logger.Printf("[TOOL] Completed: %s (ID: %s) - Duration: %v",
					toolUse.Name, toolUse.ID, duration)
			}
		}

		return result
	}
}

// PerformanceMiddleware monitors performance of tool execution.
func PerformanceMiddleware(threshold time.Duration) Middleware {
	return func(ctx context.Context, toolUse types.ToolUse, next ExecuteFunc) types.ToolResult {
		start := time.Now()

		// Execute
		result := next(ctx, toolUse)

		// Check performance
		duration := time.Since(start)
		if duration > threshold {
			// Add performance warning to the result
			if !result.IsError {
				result.Content = fmt.Sprintf("%s\n[WARNING: Tool execution took %v, exceeded threshold of %v]",
					result.Content, duration, threshold)
			}
		}

		return result
	}
}

// RecoveryMiddleware recovers from panics during tool execution.
func RecoveryMiddleware() Middleware {
	return func(ctx context.Context, toolUse types.ToolUse, next ExecuteFunc) (result types.ToolResult) {
		defer func() {
			if r := recover(); r != nil {
				// Return an error result instead of crashing
				errorResult := toolUse.Error(fmt.Errorf("tool execution panic: %v", r))
				// Ensure result is not nil
				if errorResult == nil {
					result = types.ToolResult{
						ToolUseID: toolUse.ID,
						Content:   fmt.Sprintf("Tool execution panic: %v", r),
						IsError:   true,
					}
				} else {
					result = *errorResult
				}
			}
		}()

		result = next(ctx, toolUse)
		return result
	}
}

// RetryMiddleware adds retry logic for failed tool executions.
func RetryMiddleware(maxRetries int, retryDelay time.Duration) Middleware {
	return func(ctx context.Context, toolUse types.ToolUse, next ExecuteFunc) types.ToolResult {
		var result types.ToolResult
		var lastErr error

		for i := 0; i <= maxRetries; i++ {
			// Set retry attempt in context for EventsMiddleware
			ctxWithAttempt := context.WithValue(ctx, retryAttemptKey, i+1)

			// Execute
			result = next(ctxWithAttempt, toolUse)

			// Success or non-retryable error
			if !result.IsError || i == maxRetries {
				return result
			}

			// Parse error for retry decision
			if isRetryableError(result.Content) {
				lastErr = fmt.Errorf("tool error: %s", result.Content)

				// Wait before retry (except for the last attempt)
				if i < maxRetries {
					select {
					case <-time.After(retryDelay):
						// Continue to retry
					case <-ctx.Done():
						// Context cancelled
						return *toolUse.Error(fmt.Errorf("retry cancelled: %w", ctx.Err()))
					}
				}
			} else {
				// Non-retryable error, return immediately
				return result
			}
		}

		// All retries failed
		return *toolUse.Error(fmt.Errorf("all %d retries failed: %v", maxRetries, lastErr))
	}
}

// isRetryableError determines if an error is retryable.
func isRetryableError(errorContent string) bool {
	// List of retryable error patterns
	retryablePatterns := []string{
		"timeout",
		"temporary failure",
		"connection refused",
		"network unreachable",
		"resource temporarily unavailable",
	}

	for _, pattern := range retryablePatterns {
		if contains(errorContent, pattern) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		   (s == substr ||
		    len(s) > len(substr) &&
		    (containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// RateLimitMiddleware adds rate limiting to tool execution.
type RateLimiter struct {
	tokens    chan struct{}
	rate      time.Duration
	maxTokens int
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(maxTokens int, rate time.Duration) *RateLimiter {
	rl := &RateLimiter{
		tokens:    make(chan struct{}, maxTokens),
		rate:      rate,
		maxTokens: maxTokens,
	}

	// Fill tokens initially
	for i := 0; i < maxTokens; i++ {
		rl.tokens <- struct{}{}
	}

	// Start token refill goroutine
	go rl.refill()

	return rl
}

// refill adds tokens at the specified rate.
func (rl *RateLimiter) refill() {
	ticker := time.NewTicker(rl.rate)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case rl.tokens <- struct{}{}:
			// Token added
		default:
			// Bucket is full
		}
	}
}

// RateLimitMiddleware creates a rate limiting middleware.
func RateLimitMiddleware(rateLimiter *RateLimiter) Middleware {
	return func(ctx context.Context, toolUse types.ToolUse, next ExecuteFunc) types.ToolResult {
		// Wait for a token
		select {
		case <-rateLimiter.tokens:
			// Got a token, proceed
			return next(ctx, toolUse)
		case <-ctx.Done():
			// Context cancelled while waiting
			return *toolUse.Error(fmt.Errorf("rate limit wait cancelled: %w", ctx.Err()))
		}
	}
}

// CacheMiddleware adds caching for tool results.
type Cache interface {
	Get(key string) (string, bool)
	Set(key string, value string, ttl time.Duration)
}

// CacheMiddleware creates a caching middleware.
func CacheMiddleware(cache Cache, ttl time.Duration) Middleware {
	return func(ctx context.Context, toolUse types.ToolUse, next ExecuteFunc) types.ToolResult {
		// Generate cache key
		cacheKey := fmt.Sprintf("%s:%v", toolUse.Name, toolUse.Input)

		// Check cache
		if cached, found := cache.Get(cacheKey); found {
			return types.ToolResult{
				ToolUseID: toolUse.ID,
				Content:   cached,
				IsError:   false,
			}
		}

		// Execute
		result := next(ctx, toolUse)

		// Cache successful results
		if !result.IsError {
			cache.Set(cacheKey, result.Content, ttl)
		}

		return result
	}
}