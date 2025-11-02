// Package todo provides a tool for managing todo lists in the GoAI agent.
package todo

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Zerofisher/goai/pkg/reminder"
	"github.com/Zerofisher/goai/pkg/todo"
	"github.com/Zerofisher/goai/pkg/tools"
)

// TodoTool implements the Tool interface for managing todo lists.
type TodoTool struct {
	*tools.BaseTool
	manager  *todo.Manager
	reminder *reminder.System
}

// NewTodoTool creates a new todo tool instance.
func NewTodoTool(manager *todo.Manager, reminder *reminder.System) *TodoTool {
	schema := buildTodoSchema()

	return &TodoTool{
		BaseTool: tools.NewBaseTool(
			"todo_write",
			"Update the shared todo list with tasks and their statuses (pending | in_progress | completed)",
			schema,
		),
		manager:  manager,
		reminder: reminder,
	}
}

// buildTodoSchema creates the input schema for the todo tool.
func buildTodoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"todos": map[string]interface{}{
				"type":        "array",
				"description": "The list of todo items to update",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"content": map[string]interface{}{
							"type":        "string",
							"description": "The task description (imperative form, e.g., 'Run tests')",
							"minLength":   1,
						},
						"activeForm": map[string]interface{}{
							"type":        "string",
							"description": "Present continuous form shown during execution (e.g., 'Running tests')",
							"minLength":   1,
						},
						"status": map[string]interface{}{
							"type":        "string",
							"description": "The current status of the todo item",
							"enum":        []string{"pending", "in_progress", "completed"},
						},
					},
					"required":             []string{"content", "status", "activeForm"},
					"additionalProperties": false,
				},
				"maxItems": todo.MaxTodoItems,
			},
		},
		"required":             []string{"todos"},
		"additionalProperties": false,
	}
}

// Execute runs the todo tool with the given input.
func (t *TodoTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
	// First validate the input
	if err := t.Validate(input); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	// Extract the todos array
	todosRaw, exists := input["todos"]
	if !exists {
		return "", fmt.Errorf("todos field is required")
	}

	todosArray, ok := todosRaw.([]interface{})
	if !ok {
		return "", fmt.Errorf("todos must be an array")
	}

	// Convert to TodoItem slice
	items := make([]todo.TodoItem, 0, len(todosArray))
	for i, itemRaw := range todosArray {
		itemMap, ok := itemRaw.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("todo at index %d must be an object", i)
		}

		item, err := t.parseTodoItem(itemMap, i)
		if err != nil {
			return "", fmt.Errorf("failed to parse todo at index %d: %w", i, err)
		}

		items = append(items, *item)
	}

	// Update the manager
	if err := t.manager.Update(items); err != nil {
		return "", fmt.Errorf("failed to update todos: %w", err)
	}

	// Reset reminder system since todos were updated
	if t.reminder != nil {
		t.reminder.Reset()
	}

	// Generate response with current state and statistics
	renderer := todo.NewRenderer(true) // Enable colors for terminal output
	output := renderer.Render(t.manager.GetAll())

	stats := t.manager.Stats()
	summary := t.generateSummary(stats)

	return fmt.Sprintf("%s\n\n%s", output, summary), nil
}

// Validate checks if the input is valid for the todo tool.
func (t *TodoTool) Validate(input map[string]interface{}) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	todosRaw, exists := input["todos"]
	if !exists {
		return fmt.Errorf("todos field is required")
	}

	todosArray, ok := todosRaw.([]interface{})
	if !ok {
		return fmt.Errorf("todos must be an array")
	}

	if len(todosArray) > todo.MaxTodoItems {
		return fmt.Errorf("todo list is limited to %d items, got %d", todo.MaxTodoItems, len(todosArray))
	}

	// Track in-progress count for validation
	inProgressCount := 0
	seenIDs := make(map[string]bool)

	for i, itemRaw := range todosArray {
		itemMap, ok := itemRaw.(map[string]interface{})
		if !ok {
			return fmt.Errorf("todo at index %d must be an object", i)
		}

		// Validate required fields
		content, hasContent := itemMap["content"].(string)
		if !hasContent || content == "" {
			return fmt.Errorf("todo at index %d: content is required and cannot be empty", i)
		}

		activeForm, hasActiveForm := itemMap["activeForm"].(string)
		if !hasActiveForm || activeForm == "" {
			return fmt.Errorf("todo at index %d: activeForm is required and cannot be empty", i)
		}

		statusStr, hasStatus := itemMap["status"].(string)
		if !hasStatus {
			return fmt.Errorf("todo at index %d: status is required", i)
		}

		status := todo.Status(statusStr)
		if !status.IsValid() {
			return fmt.Errorf("todo at index %d: invalid status '%s'", i, statusStr)
		}

		// Count in-progress items
		if status == todo.StatusInProgress {
			inProgressCount++
		}

		// Check for duplicate IDs if provided
		if idRaw, hasID := itemMap["id"]; hasID {
			id := fmt.Sprintf("%v", idRaw)
			if seenIDs[id] {
				return fmt.Errorf("duplicate todo ID: %s", id)
			}
			seenIDs[id] = true
		}
	}

	// Validate constraint: only one task can be in progress
	if inProgressCount > 1 {
		return fmt.Errorf("only one task can be in_progress at a time, found %d", inProgressCount)
	}

	return nil
}

// parseTodoItem converts a map to a TodoItem.
func (t *TodoTool) parseTodoItem(itemMap map[string]interface{}, index int) (*todo.TodoItem, error) {
	// Extract required fields
	content, _ := itemMap["content"].(string)
	activeForm, _ := itemMap["activeForm"].(string)
	statusStr, _ := itemMap["status"].(string)

	// Generate ID if not provided
	id := strconv.Itoa(index + 1)
	if idRaw, hasID := itemMap["id"]; hasID {
		id = fmt.Sprintf("%v", idRaw)
	}

	status := todo.Status(statusStr)

	return todo.NewTodoItem(id, content, activeForm, status)
}

// generateSummary creates a summary message based on the current statistics.
func (t *TodoTool) generateSummary(stats todo.Stats) string {
	if stats.Total == 0 {
		return "Todo list is now empty."
	}

	parts := []string{
		fmt.Sprintf("Updated todo list with %d items", stats.Total),
	}

	if stats.Completed > 0 {
		parts = append(parts, fmt.Sprintf("%d completed", stats.Completed))
	}

	if stats.InProgress > 0 {
		parts = append(parts, fmt.Sprintf("%d in progress", stats.InProgress))
	}

	if stats.Pending > 0 {
		parts = append(parts, fmt.Sprintf("%d pending", stats.Pending))
	}

	// Calculate completion percentage
	if stats.Total > 0 && stats.Completed > 0 {
		percentage := (stats.Completed * 100) / stats.Total
		parts = append(parts, fmt.Sprintf("(%d%% complete)", percentage))
	}

	// Build the final message
	result := fmt.Sprintf("✅ %s", parts[0])
	if len(parts) > 1 {
		result += " - " + strings.Join(parts[1:], ", ")
	}

	return result
}

// GetManager returns the underlying todo manager.
func (t *TodoTool) GetManager() *todo.Manager {
	return t.manager
}

// GetReminder returns the reminder system.
func (t *TodoTool) GetReminder() *reminder.System {
	return t.reminder
}

// MarshalJSON implements json.Marshaler for the tool definition.
func (t *TodoTool) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"name":         t.Name(),
		"description":  t.Description(),
		"input_schema": t.InputSchema(),
	})
}