package types

import (
	"time"
)

// ToolEventType represents the type of tool event
type ToolEventType string

const (
	// ToolEventStarted indicates the tool execution has started
	ToolEventStarted ToolEventType = "started"
	// ToolEventSucceeded indicates the tool execution succeeded
	ToolEventSucceeded ToolEventType = "succeeded"
	// ToolEventFailed indicates the tool execution failed
	ToolEventFailed ToolEventType = "failed"
	// ToolEventProgress indicates progress during tool execution (reserved for future use)
	ToolEventProgress ToolEventType = "progress"
)

// ToolEvent represents an event during tool execution
type ToolEvent struct {
	// ID is the unique identifier for this tool use (corresponds to ToolUse.ID)
	ID string `json:"id"`

	// Name is the tool name (e.g., "bash", "write_file")
	Name string `json:"name"`

	// Input contains the sanitized/truncated input parameters
	Input map[string]interface{} `json:"input,omitempty"`

	// Output contains the result output (for succeeded events, may be truncated)
	Output string `json:"output,omitempty"`

	// Error contains the error description (for failed events)
	Error string `json:"error,omitempty"`

	// Attempt is the current attempt number (useful with retry middleware)
	Attempt int `json:"attempt,omitempty"`

	// StartedAt is the time when tool execution started
	StartedAt time.Time `json:"started_at"`

	// EndedAt is the time when tool execution ended (nil for started events)
	EndedAt *time.Time `json:"ended_at,omitempty"`

	// Duration is the execution duration (nil for started events)
	Duration *time.Duration `json:"duration,omitempty"`

	// Type is the event type
	Type ToolEventType `json:"type"`

	// Metadata contains optional additional information
	// Examples: exit_code, truncated, output_lines, etc.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
