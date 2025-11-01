package dispatcher

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Zerofisher/goai/pkg/types"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// retryAttemptKey is the context key for retry attempt number
const retryAttemptKey contextKey = "retry_attempt"

// ToolObserver defines the interface for observing tool execution events
type ToolObserver interface {
	// OnToolEvent is called when a tool event occurs
	// Implementations should be non-blocking and handle errors gracefully
	OnToolEvent(ctx context.Context, e types.ToolEvent)
}

// EventsOptions configures the events middleware behavior
type EventsOptions struct {
	// MaxOutputChars is the maximum number of characters to include in output
	// Default: 20000
	MaxOutputChars int

	// MaskKeys is a list of sensitive key patterns to mask in input
	// Keys are matched case-insensitively with substring matching
	// Default: API_KEY, TOKEN, PASSWORD, SECRET, AUTH, KEY, etc.
	MaskKeys []string
}

// DefaultEventsOptions returns default options for events middleware
func DefaultEventsOptions() EventsOptions {
	return EventsOptions{
		MaxOutputChars: 20000,
		MaskKeys: []string{
			"api_key", "apikey", "token", "password", "passwd", "pwd",
			"secret", "auth", "key", "access_key", "private_key",
			"authorization", "credential", "credentials",
		},
	}
}

// EventsMiddleware creates a middleware that emits tool execution events
// It wraps tool execution and emits started/succeeded/failed events to the observer
func EventsMiddleware(obs ToolObserver, opts EventsOptions) Middleware {
	if obs == nil {
		// No observer, return no-op middleware
		return func(ctx context.Context, tu types.ToolUse, next ExecuteFunc) types.ToolResult {
			return next(ctx, tu)
		}
	}

	// Apply defaults
	if opts.MaxOutputChars <= 0 {
		opts.MaxOutputChars = 20000
	}
	if len(opts.MaskKeys) == 0 {
		opts.MaskKeys = DefaultEventsOptions().MaskKeys
	}

	return func(ctx context.Context, tu types.ToolUse, next ExecuteFunc) types.ToolResult {
		startedAt := time.Now()

		// Sanitize input before emitting
		safeInput := sanitizeInput(tu.Input, opts.MaskKeys)

		// Get attempt number from context (if retry middleware is present)
		attempt := attemptFromContext(ctx)

		// Emit started event
		emitEvent(ctx, obs, types.ToolEvent{
			ID:        tu.ID,
			Name:      tu.Name,
			Input:     safeInput,
			StartedAt: startedAt,
			Attempt:   attempt,
			Type:      types.ToolEventStarted,
		})

		// Execute the tool
		result := next(ctx, tu)

		// Calculate duration
		endedAt := time.Now()
		duration := endedAt.Sub(startedAt)

		// Build completed event
		event := types.ToolEvent{
			ID:        tu.ID,
			Name:      tu.Name,
			Input:     safeInput,
			StartedAt: startedAt,
			EndedAt:   &endedAt,
			Duration:  &duration,
			Attempt:   attempt,
			Metadata:  make(map[string]interface{}),
		}

		if result.IsError {
			// Failed event
			event.Type = types.ToolEventFailed
			event.Error = result.Content
		} else {
			// Succeeded event
			event.Type = types.ToolEventSucceeded
			event.Output, event.Metadata = truncateOutput(result.Content, opts.MaxOutputChars)
		}

		// Emit completed event
		emitEvent(ctx, obs, event)

		return result
	}
}

// emitEvent safely emits an event to the observer
// It wraps the call with recover() to prevent observer panics from affecting tool execution
func emitEvent(ctx context.Context, obs ToolObserver, event types.ToolEvent) {
	defer func() {
		if r := recover(); r != nil {
			// Log the panic but don't propagate it
			// In production, this should use proper logging
			fmt.Printf("WARNING: ToolObserver.OnToolEvent panicked: %v\n", r)
		}
	}()

	obs.OnToolEvent(ctx, event)
}

// sanitizeInput creates a sanitized copy of the input map
// It masks sensitive values based on key patterns
func sanitizeInput(input map[string]interface{}, maskKeys []string) map[string]interface{} {
	if input == nil {
		return nil
	}

	result := make(map[string]interface{}, len(input))

	for key, value := range input {
		if shouldMaskKey(key, maskKeys) {
			result[key] = "****"
		} else {
			// For nested maps, recursively sanitize
			if nestedMap, ok := value.(map[string]interface{}); ok {
				result[key] = sanitizeInput(nestedMap, maskKeys)
			} else if strValue, ok := value.(string); ok {
				// Truncate very long string values
				if len(strValue) > 500 {
					result[key] = strValue[:500] + "... [truncated]"
				} else {
					result[key] = strValue
				}
			} else {
				result[key] = value
			}
		}
	}

	return result
}

// shouldMaskKey checks if a key should be masked based on patterns
func shouldMaskKey(key string, maskKeys []string) bool {
	lowerKey := strings.ToLower(key)
	for _, pattern := range maskKeys {
		if strings.Contains(lowerKey, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// truncateOutput truncates output to maxChars and returns metadata about truncation
func truncateOutput(output string, maxChars int) (string, map[string]interface{}) {
	metadata := make(map[string]interface{})

	// Count lines
	lines := strings.Split(output, "\n")
	metadata["output_lines"] = len(lines)

	// Check if truncation is needed
	if len(output) <= maxChars {
		metadata["truncated"] = false
		return output, metadata
	}

	// Truncate and add marker
	truncated := output[:maxChars]
	metadata["truncated"] = true
	metadata["original_length"] = len(output)
	metadata["shown_length"] = maxChars

	return truncated + "\n... [output truncated]", metadata
}

// attemptFromContext extracts the retry attempt number from context
// Returns 1 if not present (first attempt)
func attemptFromContext(ctx context.Context) int {
	if attempt, ok := ctx.Value(retryAttemptKey).(int); ok {
		return attempt
	}
	return 1
}
