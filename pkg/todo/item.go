// Package todo provides Todo list management for the GoAI Agent.
// It manages task tracking with status transitions and rendering.
package todo

import (
	"fmt"
	"time"
)

// Status represents the state of a todo item.
type Status string

const (
	// StatusPending indicates a todo that has not been started.
	StatusPending Status = "pending"
	// StatusInProgress indicates a todo that is currently being worked on.
	StatusInProgress Status = "in_progress"
	// StatusCompleted indicates a todo that has been finished.
	StatusCompleted Status = "completed"
)

// ValidStatuses contains all valid todo statuses.
var ValidStatuses = []Status{StatusPending, StatusInProgress, StatusCompleted}

// IsValid checks if a status is valid.
func (s Status) IsValid() bool {
	for _, valid := range ValidStatuses {
		if s == valid {
			return true
		}
	}
	return false
}

// TodoItem represents a single task in the todo list.
type TodoItem struct {
	// ID is the unique identifier for the todo item.
	ID string `json:"id"`

	// Content is the task description.
	Content string `json:"content"`

	// ActiveForm is the present continuous form of the task (e.g., "Running tests").
	ActiveForm string `json:"activeForm"`

	// Status is the current state of the todo item.
	Status Status `json:"status"`

	// CreatedAt is when the todo was created.
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt is when the todo was last modified.
	UpdatedAt time.Time `json:"updatedAt"`
}

// NewTodoItem creates a new todo item with the given parameters.
func NewTodoItem(id, content, activeForm string, status Status) (*TodoItem, error) {
	if id == "" {
		return nil, fmt.Errorf("todo ID cannot be empty")
	}

	if content == "" {
		return nil, fmt.Errorf("todo content cannot be empty")
	}

	if activeForm == "" {
		return nil, fmt.Errorf("todo activeForm cannot be empty")
	}

	if !status.IsValid() {
		return nil, fmt.Errorf("invalid status: %s", status)
	}

	now := time.Now()
	return &TodoItem{
		ID:         id,
		Content:    content,
		ActiveForm: activeForm,
		Status:     status,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// UpdateStatus changes the status of the todo item.
func (t *TodoItem) UpdateStatus(newStatus Status) error {
	if !newStatus.IsValid() {
		return fmt.Errorf("invalid status: %s", newStatus)
	}

	// Validate status transitions
	if !t.isValidTransition(newStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", t.Status, newStatus)
	}

	t.Status = newStatus
	t.UpdatedAt = time.Now()
	return nil
}

// isValidTransition checks if a status transition is allowed.
func (t *TodoItem) isValidTransition(newStatus Status) bool {
	// Allow any transition for now, but this can be restricted later
	// For example: completed -> pending might not be allowed
	return true
}

// IsCompleted returns true if the todo is completed.
func (t *TodoItem) IsCompleted() bool {
	return t.Status == StatusCompleted
}

// IsInProgress returns true if the todo is in progress.
func (t *TodoItem) IsInProgress() bool {
	return t.Status == StatusInProgress
}

// IsPending returns true if the todo is pending.
func (t *TodoItem) IsPending() bool {
	return t.Status == StatusPending
}