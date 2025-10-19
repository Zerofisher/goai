package todo

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Zerofisher/goai/pkg/reminder"
	"github.com/Zerofisher/goai/pkg/todo"
)

// TestNewTodoTool tests the creation of a new todo tool.
func TestNewTodoTool(t *testing.T) {
	manager := todo.NewManager()
	reminderSys := reminder.NewSystem(3, 10)
	tool := NewTodoTool(manager, reminderSys)

	if tool == nil {
		t.Fatal("NewTodoTool returned nil")
	}

	if tool.Name() != "TodoWrite" {
		t.Errorf("Expected name 'TodoWrite', got %s", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Description should not be empty")
	}

	if tool.InputSchema() == nil {
		t.Error("InputSchema should not be nil")
	}

	if tool.GetManager() != manager {
		t.Error("GetManager should return the same manager instance")
	}

	if tool.GetReminder() != reminderSys {
		t.Error("GetReminder should return the same reminder instance")
	}
}

// TestTodoTool_Execute tests the Execute method with various inputs.
func TestTodoTool_Execute(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid single todo",
			input: map[string]interface{}{
				"todos": []interface{}{
					map[string]interface{}{
						"content":    "Write tests",
						"activeForm": "Writing tests",
						"status":     "pending",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple todos",
			input: map[string]interface{}{
				"todos": []interface{}{
					map[string]interface{}{
						"content":    "Task 1",
						"activeForm": "Doing task 1",
						"status":     "pending",
					},
					map[string]interface{}{
						"content":    "Task 2",
						"activeForm": "Doing task 2",
						"status":     "in_progress",
					},
					map[string]interface{}{
						"content":    "Task 3",
						"activeForm": "Doing task 3",
						"status":     "completed",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
			errMsg:  "validation failed",
		},
		{
			name:    "missing todos field",
			input:   map[string]interface{}{},
			wantErr: true,
			errMsg:  "todos field is required",
		},
		{
			name: "empty content",
			input: map[string]interface{}{
				"todos": []interface{}{
					map[string]interface{}{
						"content":    "",
						"activeForm": "Doing task",
						"status":     "pending",
					},
				},
			},
			wantErr: true,
			errMsg:  "content is required and cannot be empty",
		},
		{
			name: "invalid status",
			input: map[string]interface{}{
				"todos": []interface{}{
					map[string]interface{}{
						"content":    "Task",
						"activeForm": "Doing task",
						"status":     "invalid",
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid status",
		},
		{
			name: "multiple in-progress",
			input: map[string]interface{}{
				"todos": []interface{}{
					map[string]interface{}{
						"content":    "Task 1",
						"activeForm": "Doing task 1",
						"status":     "in_progress",
					},
					map[string]interface{}{
						"content":    "Task 2",
						"activeForm": "Doing task 2",
						"status":     "in_progress",
					},
				},
			},
			wantErr: true,
			errMsg:  "only one task can be in_progress at a time",
		},
		{
			name: "exceeds max items",
			input: map[string]interface{}{
				"todos": func() []interface{} {
					items := make([]interface{}, todo.MaxTodoItems+1)
					for i := 0; i < todo.MaxTodoItems+1; i++ {
						items[i] = map[string]interface{}{
							"content":    "Task",
							"activeForm": "Doing task",
							"status":     "pending",
						}
					}
					return items
				}(),
			},
			wantErr: true,
			errMsg:  "todo list is limited to",
		},
		{
			name: "todos not an array",
			input: map[string]interface{}{
				"todos": "not an array",
			},
			wantErr: true,
			errMsg:  "todos must be an array",
		},
		{
			name: "todo item not an object",
			input: map[string]interface{}{
				"todos": []interface{}{
					"not an object",
				},
			},
			wantErr: true,
			errMsg:  "must be an object",
		},
		{
			name: "empty todo list",
			input: map[string]interface{}{
				"todos": []interface{}{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := todo.NewManager()
			reminderSys := reminder.NewSystem(3, 10)
			tool := NewTodoTool(manager, reminderSys)

			output, err := tool.Execute(context.Background(), tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errMsg != "" {
				if !containsString(err.Error(), tt.errMsg) {
					t.Errorf("Execute() error = %v, want error containing %v", err.Error(), tt.errMsg)
				}
			}

			if err == nil && output == "" {
				t.Error("Execute() should return non-empty output on success")
			}
		})
	}
}

// TestTodoTool_Validate tests the validation logic.
func TestTodoTool_Validate(t *testing.T) {
	manager := todo.NewManager()
	tool := NewTodoTool(manager, nil)

	tests := []struct {
		name    string
		input   map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid input",
			input: map[string]interface{}{
				"todos": []interface{}{
					map[string]interface{}{
						"content":    "Task",
						"activeForm": "Doing task",
						"status":     "pending",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
		{
			name:    "missing todos",
			input:   map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "missing activeForm",
			input: map[string]interface{}{
				"todos": []interface{}{
					map[string]interface{}{
						"content": "Task",
						"status":  "pending",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing status",
			input: map[string]interface{}{
				"todos": []interface{}{
					map[string]interface{}{
						"content":    "Task",
						"activeForm": "Doing task",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tool.Validate(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestTodoTool_parseTodoItem tests the parsing of todo items.
func TestTodoTool_parseTodoItem(t *testing.T) {
	manager := todo.NewManager()
	tool := NewTodoTool(manager, nil)

	tests := []struct {
		name    string
		itemMap map[string]interface{}
		index   int
		wantErr bool
	}{
		{
			name: "valid item without ID",
			itemMap: map[string]interface{}{
				"content":    "Task",
				"activeForm": "Doing task",
				"status":     "pending",
			},
			index:   0,
			wantErr: false,
		},
		{
			name: "valid item with ID",
			itemMap: map[string]interface{}{
				"id":         "custom-id",
				"content":    "Task",
				"activeForm": "Doing task",
				"status":     "pending",
			},
			index:   0,
			wantErr: false,
		},
		{
			name: "empty content",
			itemMap: map[string]interface{}{
				"content":    "",
				"activeForm": "Doing task",
				"status":     "pending",
			},
			index:   0,
			wantErr: true,
		},
		{
			name: "invalid status",
			itemMap: map[string]interface{}{
				"content":    "Task",
				"activeForm": "Doing task",
				"status":     "invalid",
			},
			index:   0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := tool.parseTodoItem(tt.itemMap, tt.index)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseTodoItem() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && item == nil {
				t.Error("parseTodoItem() should return non-nil item on success")
			}
		})
	}
}

// TestTodoTool_generateSummary tests the summary generation.
func TestTodoTool_generateSummary(t *testing.T) {
	manager := todo.NewManager()
	tool := NewTodoTool(manager, nil)

	tests := []struct {
		name     string
		stats    todo.Stats
		contains []string
	}{
		{
			name:     "empty stats",
			stats:    todo.Stats{Total: 0},
			contains: []string{"Todo list is now empty"},
		},
		{
			name: "mixed stats",
			stats: todo.Stats{
				Total:      5,
				Completed:  2,
				InProgress: 1,
				Pending:    2,
			},
			contains: []string{"Updated todo list with 5 items", "2 completed", "1 in progress", "2 pending", "40% complete"},
		},
		{
			name: "all completed",
			stats: todo.Stats{
				Total:      3,
				Completed:  3,
				InProgress: 0,
				Pending:    0,
			},
			contains: []string{"Updated todo list with 3 items", "3 completed", "100% complete"},
		},
		{
			name: "none completed",
			stats: todo.Stats{
				Total:      3,
				Completed:  0,
				InProgress: 0,
				Pending:    3,
			},
			contains: []string{"Updated todo list with 3 items", "3 pending"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := tool.generateSummary(tt.stats)

			for _, expected := range tt.contains {
				if !containsString(summary, expected) {
					t.Errorf("generateSummary() missing expected string: %q\nGot: %s", expected, summary)
				}
			}
		})
	}
}

// TestTodoTool_ReminderReset tests that the reminder system is reset on updates.
func TestTodoTool_ReminderReset(t *testing.T) {
	manager := todo.NewManager()
	reminderSys := reminder.NewSystem(3, 10)
	tool := NewTodoTool(manager, reminderSys)

	// Increment rounds to trigger reminder (need exactly 3 rounds)
	for i := 0; i < 3; i++ {
		reminderSys.IncrementRounds()
	}

	// Verify reminder is needed
	if !reminderSys.ShouldRemind() {
		t.Errorf("Reminder should be needed after 3 rounds, got %d rounds", reminderSys.GetRoundsWithoutTodo())
	}

	// Execute todo update
	input := map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "Task",
				"activeForm": "Doing task",
				"status":     "pending",
			},
		},
	}

	_, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify reminder was reset
	if reminderSys.ShouldRemind() {
		t.Error("Reminder should be reset after todo update")
	}
}

// TestTodoTool_MarshalJSON tests JSON marshaling.
func TestTodoTool_MarshalJSON(t *testing.T) {
	manager := todo.NewManager()
	tool := NewTodoTool(manager, nil)

	data, err := tool.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result["name"] != "TodoWrite" {
		t.Errorf("Expected name 'TodoWrite', got %v", result["name"])
	}

	if result["description"] == nil || result["description"] == "" {
		t.Error("Description should not be empty")
	}

	if result["input_schema"] == nil {
		t.Error("Input schema should not be nil")
	}
}

// TestTodoTool_Integration tests the full integration with manager.
func TestTodoTool_Integration(t *testing.T) {
	manager := todo.NewManager()
	tool := NewTodoTool(manager, nil)

	// Add initial todos
	input1 := map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "Task 1",
				"activeForm": "Doing task 1",
				"status":     "pending",
			},
			map[string]interface{}{
				"content":    "Task 2",
				"activeForm": "Doing task 2",
				"status":     "pending",
			},
		},
	}

	output1, err := tool.Execute(context.Background(), input1)
	if err != nil {
		t.Fatalf("First Execute failed: %v", err)
	}

	if !containsString(output1, "2 pending") {
		t.Error("Output should indicate 2 pending tasks")
	}

	// Update statuses
	input2 := map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "Task 1",
				"activeForm": "Doing task 1",
				"status":     "in_progress",
			},
			map[string]interface{}{
				"content":    "Task 2",
				"activeForm": "Doing task 2",
				"status":     "pending",
			},
		},
	}

	output2, err := tool.Execute(context.Background(), input2)
	if err != nil {
		t.Fatalf("Second Execute failed: %v", err)
	}

	if !containsString(output2, "1 in progress") {
		t.Error("Output should indicate 1 in progress task")
	}

	// Complete a task
	input3 := map[string]interface{}{
		"todos": []interface{}{
			map[string]interface{}{
				"content":    "Task 1",
				"activeForm": "Doing task 1",
				"status":     "completed",
			},
			map[string]interface{}{
				"content":    "Task 2",
				"activeForm": "Doing task 2",
				"status":     "in_progress",
			},
		},
	}

	output3, err := tool.Execute(context.Background(), input3)
	if err != nil {
		t.Fatalf("Third Execute failed: %v", err)
	}

	if !containsString(output3, "1 completed") {
		t.Error("Output should indicate 1 completed task")
	}

	if !containsString(output3, "50% complete") {
		t.Error("Output should indicate 50% completion")
	}

	// Verify manager state
	stats := manager.Stats()
	if stats.Total != 2 {
		t.Errorf("Manager should have 2 items, got %d", stats.Total)
	}
	if stats.Completed != 1 {
		t.Errorf("Manager should have 1 completed, got %d", stats.Completed)
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}