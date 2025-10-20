package dispatcher

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/Zerofisher/goai/pkg/types"
)

// mockObserver captures events for testing
type mockObserver struct {
	events []types.ToolEvent
	mu     sync.Mutex
}

func (m *mockObserver) OnToolEvent(_ context.Context, e types.ToolEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, e)
}

func (m *mockObserver) getEvents() []types.ToolEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]types.ToolEvent(nil), m.events...)
}

// TestEventsMiddleware_BasicFlow tests started → succeeded event sequence
func TestEventsMiddleware_BasicFlow(t *testing.T) {
	obs := &mockObserver{}
	opts := DefaultEventsOptions()

	middleware := EventsMiddleware(obs, opts)

	// Mock tool execution
	tu := types.ToolUse{
		ID:   "test-1",
		Name: "test_tool",
		Input: map[string]interface{}{
			"arg": "value",
		},
	}

	next := func(_ context.Context, _ types.ToolUse) types.ToolResult {
		return types.ToolResult{
			ToolUseID: tu.ID,
			Content:   "success output",
			IsError:   false,
		}
	}

	// Execute
	ctx := context.Background()
	result := middleware(ctx, tu, next)

	// Verify result
	if result.IsError {
		t.Errorf("Expected success, got error: %s", result.Content)
	}

	// Verify events
	events := obs.getEvents()
	if len(events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(events))
	}

	// Check started event
	if events[0].Type != types.ToolEventStarted {
		t.Errorf("First event should be started, got %s", events[0].Type)
	}
	if events[0].Name != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got '%s'", events[0].Name)
	}

	// Check succeeded event
	if events[1].Type != types.ToolEventSucceeded {
		t.Errorf("Second event should be succeeded, got %s", events[1].Type)
	}
	if events[1].Output != "success output" {
		t.Errorf("Expected output 'success output', got '%s'", events[1].Output)
	}
	if events[1].Duration == nil {
		t.Error("Duration should be set for completed event")
	}
}

// TestEventsMiddleware_FailedFlow tests started → failed event sequence
func TestEventsMiddleware_FailedFlow(t *testing.T) {
	obs := &mockObserver{}
	opts := DefaultEventsOptions()

	middleware := EventsMiddleware(obs, opts)

	tu := types.ToolUse{
		ID:   "test-2",
		Name: "failing_tool",
		Input: map[string]interface{}{},
	}

	next := func(_ context.Context, _ types.ToolUse) types.ToolResult {
		return types.ToolResult{
			ToolUseID: tu.ID,
			Content:   "Error: something went wrong",
			IsError:   true,
		}
	}

	ctx := context.Background()
	result := middleware(ctx, tu, next)

	if !result.IsError {
		t.Error("Expected error result")
	}

	events := obs.getEvents()
	if len(events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(events))
	}

	// Check failed event
	if events[1].Type != types.ToolEventFailed {
		t.Errorf("Second event should be failed, got %s", events[1].Type)
	}
	if events[1].Error != "Error: something went wrong" {
		t.Errorf("Expected error message, got '%s'", events[1].Error)
	}
}

// TestEventsMiddleware_OutputTruncation tests output truncation logic
func TestEventsMiddleware_OutputTruncation(t *testing.T) {
	obs := &mockObserver{}
	opts := EventsOptions{
		MaxOutputChars: 100,
		MaskKeys:       []string{},
	}

	middleware := EventsMiddleware(obs, opts)

	// Create long output
	longOutput := strings.Repeat("x", 200)

	tu := types.ToolUse{
		ID:   "test-3",
		Name: "long_output_tool",
		Input: map[string]interface{}{},
	}

	next := func(_ context.Context, _ types.ToolUse) types.ToolResult {
		return types.ToolResult{
			ToolUseID: tu.ID,
			Content:   longOutput,
			IsError:   false,
		}
	}

	ctx := context.Background()
	middleware(ctx, tu, next)

	events := obs.getEvents()
	succeededEvent := events[1]

	// Check truncation
	if len(succeededEvent.Output) > 150 { // 100 + truncation marker
		t.Errorf("Output should be truncated, got %d chars", len(succeededEvent.Output))
	}

	if len(succeededEvent.Output) <= 100 {
		t.Error("Output should include truncation marker")
	}

	if !strings.Contains(succeededEvent.Output, "truncated") {
		t.Error("Truncated output should contain 'truncated' marker")
	}

	// Check metadata
	if truncated, ok := succeededEvent.Metadata["truncated"].(bool); !ok || !truncated {
		t.Error("Metadata should indicate truncation")
	}
	if originalLen, ok := succeededEvent.Metadata["original_length"].(int); !ok || originalLen != 200 {
		t.Errorf("Metadata should record original length as 200, got %d", originalLen)
	}
}

// TestSanitizeInput tests sensitive data masking
func TestSanitizeInput(t *testing.T) {
	maskKeys := []string{"api_key", "password", "secret", "token"}

	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "mask api_key",
			input: map[string]interface{}{
				"api_key": "secret123",
				"name":    "test",
			},
			expected: map[string]interface{}{
				"api_key": "****",
				"name":    "test",
			},
		},
		{
			name: "mask password case insensitive",
			input: map[string]interface{}{
				"PASSWORD": "mypass",
				"user":     "alice",
			},
			expected: map[string]interface{}{
				"PASSWORD": "****",
				"user":     "alice",
			},
		},
		{
			name: "mask nested map",
			input: map[string]interface{}{
				"config": map[string]interface{}{
					"api_key": "nested_secret",
					"timeout": 30,
				},
			},
			expected: map[string]interface{}{
				"config": map[string]interface{}{
					"api_key": "****",
					"timeout": 30,
				},
			},
		},
		{
			name: "truncate long strings",
			input: map[string]interface{}{
				"content": strings.Repeat("a", 600),
			},
			expected: map[string]interface{}{
				"content": strings.Repeat("a", 500) + "... [truncated]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeInput(tt.input, maskKeys)

			for key, expectedVal := range tt.expected {
				actualVal, ok := result[key]
				if !ok {
					t.Errorf("Missing key: %s", key)
					continue
				}

				// Handle nested map
				if expectedMap, ok := expectedVal.(map[string]interface{}); ok {
					actualMap, ok := actualVal.(map[string]interface{})
					if !ok {
						t.Errorf("Expected map for key %s", key)
						continue
					}
					for nestedKey, nestedExpected := range expectedMap {
						if actualMap[nestedKey] != nestedExpected {
							t.Errorf("For %s.%s: expected '%v', got '%v'", key, nestedKey, nestedExpected, actualMap[nestedKey])
						}
					}
				} else {
					if actualVal != expectedVal {
						t.Errorf("For %s: expected '%v', got '%v'", key, expectedVal, actualVal)
					}
				}
			}
		})
	}
}

// TestShouldMaskKey tests key masking logic
func TestShouldMaskKey(t *testing.T) {
	maskKeys := []string{"api_key", "password", "secret"}

	tests := []struct {
		key      string
		expected bool
	}{
		{"api_key", true},
		{"API_KEY", true},
		{"my_api_key", true},
		{"password", true},
		{"user_password", true},
		{"secret_token", true},
		{"username", false},
		{"name", false},
		{"timeout", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := shouldMaskKey(tt.key, maskKeys)
			if result != tt.expected {
				t.Errorf("For key '%s': expected %v, got %v", tt.key, tt.expected, result)
			}
		})
	}
}

// TestEventsMiddleware_Concurrent tests concurrent tool execution
func TestEventsMiddleware_Concurrent(t *testing.T) {
	obs := &mockObserver{}
	opts := DefaultEventsOptions()

	middleware := EventsMiddleware(obs, opts)

	// Execute multiple tools concurrently
	numTools := 10
	var wg sync.WaitGroup

	for i := 0; i < numTools; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			tu := types.ToolUse{
				ID:   fmt.Sprintf("tool-%d", id),
				Name: fmt.Sprintf("tool_%d", id),
				Input: map[string]interface{}{
					"id": id,
				},
			}

			next := func(_ context.Context, _ types.ToolUse) types.ToolResult {
				return types.ToolResult{
					ToolUseID: tu.ID,
					Content:   fmt.Sprintf("output-%d", id),
					IsError:   false,
				}
			}

			ctx := context.Background()
			middleware(ctx, tu, next)
		}(i)
	}

	wg.Wait()

	// Verify all events were captured
	events := obs.getEvents()
	expectedEventCount := numTools * 2 // started + succeeded for each
	if len(events) != expectedEventCount {
		t.Errorf("Expected %d events, got %d", expectedEventCount, len(events))
	}

	// Verify we have both started and succeeded for each tool
	startedCount := 0
	succeededCount := 0
	for _, e := range events {
		if e.Type == types.ToolEventStarted {
			startedCount++
		} else if e.Type == types.ToolEventSucceeded {
			succeededCount++
		}
	}

	if startedCount != numTools {
		t.Errorf("Expected %d started events, got %d", numTools, startedCount)
	}
	if succeededCount != numTools {
		t.Errorf("Expected %d succeeded events, got %d", numTools, succeededCount)
	}
}

// TestEventsMiddleware_NilObserver tests behavior with no observer
func TestEventsMiddleware_NilObserver(t *testing.T) {
	middleware := EventsMiddleware(nil, DefaultEventsOptions())

	tu := types.ToolUse{
		ID:   "test",
		Name: "test_tool",
		Input: map[string]interface{}{},
	}

	executed := false
	next := func(_ context.Context, _ types.ToolUse) types.ToolResult {
		executed = true
		return types.ToolResult{
			ToolUseID: tu.ID,
			Content:   "success",
			IsError:   false,
		}
	}

	ctx := context.Background()
	result := middleware(ctx, tu, next)

	if !executed {
		t.Error("Tool should still execute even with nil observer")
	}
	if result.IsError {
		t.Error("Expected successful execution")
	}
}
