package todo

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestNewManager tests the creation of a new todo manager.
func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.maxItems != MaxTodoItems {
		t.Errorf("expected maxItems to be %d, got %d", MaxTodoItems, manager.maxItems)
	}

	if len(manager.items) != 0 {
		t.Errorf("expected items to be empty, got %d items", len(manager.items))
	}
}

// TestManager_Update tests the Update method with various scenarios.
func TestManager_Update(t *testing.T) {
	tests := []struct {
		name    string
		items   []TodoItem
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty list",
			items:   []TodoItem{},
			wantErr: false,
		},
		{
			name: "valid single item",
			items: []TodoItem{
				{
					ID:         "1",
					Content:    "Test task",
					ActiveForm: "Testing task",
					Status:     StatusPending,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				},
			},
			wantErr: false,
		},
		{
			name: "multiple valid items",
			items: []TodoItem{
				{ID: "1", Content: "Task 1", ActiveForm: "Doing task 1", Status: StatusPending},
				{ID: "2", Content: "Task 2", ActiveForm: "Doing task 2", Status: StatusInProgress},
				{ID: "3", Content: "Task 3", ActiveForm: "Doing task 3", Status: StatusCompleted},
			},
			wantErr: false,
		},
		{
			name: "duplicate IDs",
			items: []TodoItem{
				{ID: "1", Content: "Task 1", ActiveForm: "Doing task 1", Status: StatusPending},
				{ID: "1", Content: "Task 2", ActiveForm: "Doing task 2", Status: StatusPending},
			},
			wantErr: true,
			errMsg:  "duplicate todo ID: 1",
		},
		{
			name: "empty ID",
			items: []TodoItem{
				{ID: "", Content: "Task", ActiveForm: "Doing task", Status: StatusPending},
			},
			wantErr: true,
			errMsg:  "todo at index 0 has empty ID",
		},
		{
			name: "empty content",
			items: []TodoItem{
				{ID: "1", Content: "", ActiveForm: "Doing task", Status: StatusPending},
			},
			wantErr: true,
			errMsg:  "todo 1 has empty content",
		},
		{
			name: "empty activeForm",
			items: []TodoItem{
				{ID: "1", Content: "Task", ActiveForm: "", Status: StatusPending},
			},
			wantErr: true,
			errMsg:  "todo 1 has empty activeForm",
		},
		{
			name: "invalid status",
			items: []TodoItem{
				{ID: "1", Content: "Task", ActiveForm: "Doing task", Status: Status("invalid")},
			},
			wantErr: true,
			errMsg:  "todo 1 has invalid status: invalid",
		},
		{
			name: "multiple in-progress tasks",
			items: []TodoItem{
				{ID: "1", Content: "Task 1", ActiveForm: "Doing task 1", Status: StatusInProgress},
				{ID: "2", Content: "Task 2", ActiveForm: "Doing task 2", Status: StatusInProgress},
			},
			wantErr: true,
			errMsg:  "only one task can be in_progress at a time, found 2",
		},
		{
			name: "exceeds max items",
			items: func() []TodoItem {
				items := make([]TodoItem, MaxTodoItems+1)
				for i := 0; i < MaxTodoItems+1; i++ {
					items[i] = TodoItem{
						ID:         fmt.Sprintf("id-%d", i),
						Content:    fmt.Sprintf("Task %d", i),
						ActiveForm: fmt.Sprintf("Doing task %d", i),
						Status:     StatusPending,
					}
				}
				return items
			}(),
			wantErr: true,
			errMsg:  fmt.Sprintf("todo list is limited to %d items, got %d", MaxTodoItems, MaxTodoItems+1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager()
			err := manager.Update(tt.items)

			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Update() error = %v, want %v", err.Error(), tt.errMsg)
			}

			if err == nil && len(manager.items) != len(tt.items) {
				t.Errorf("Update() items count = %d, want %d", len(manager.items), len(tt.items))
			}
		})
	}
}

// TestManager_Add tests adding individual todo items.
func TestManager_Add(t *testing.T) {
	tests := []struct {
		name     string
		existing []TodoItem
		newItem  TodoItem
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "add to empty list",
			existing: []TodoItem{},
			newItem:  TodoItem{ID: "1", Content: "Task", ActiveForm: "Doing task", Status: StatusPending},
			wantErr:  false,
		},
		{
			name:     "add to existing list",
			existing: []TodoItem{{ID: "1", Content: "Task 1", ActiveForm: "Doing task 1", Status: StatusPending}},
			newItem:  TodoItem{ID: "2", Content: "Task 2", ActiveForm: "Doing task 2", Status: StatusPending},
			wantErr:  false,
		},
		{
			name:     "duplicate ID",
			existing: []TodoItem{{ID: "1", Content: "Task", ActiveForm: "Doing task", Status: StatusPending}},
			newItem:  TodoItem{ID: "1", Content: "Another task", ActiveForm: "Doing another task", Status: StatusPending},
			wantErr:  true,
			errMsg:   "todo with ID 1 already exists",
		},
		{
			name:     "add in-progress when none exists",
			existing: []TodoItem{{ID: "1", Content: "Task", ActiveForm: "Doing task", Status: StatusPending}},
			newItem:  TodoItem{ID: "2", Content: "Task 2", ActiveForm: "Doing task 2", Status: StatusInProgress},
			wantErr:  false,
		},
		{
			name:     "add in-progress when one exists",
			existing: []TodoItem{{ID: "1", Content: "Task", ActiveForm: "Doing task", Status: StatusInProgress}},
			newItem:  TodoItem{ID: "2", Content: "Task 2", ActiveForm: "Doing task 2", Status: StatusInProgress},
			wantErr:  true,
			errMsg:   "cannot add in-progress task: 1 is already in progress",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager()
			_ = manager.Update(tt.existing)

			err := manager.Add(tt.newItem)

			if (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Add() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

// TestManager_UpdateStatus tests status updates for todo items.
func TestManager_UpdateStatus(t *testing.T) {
	manager := NewManager()
	items := []TodoItem{
		{ID: "1", Content: "Task 1", ActiveForm: "Doing task 1", Status: StatusPending},
		{ID: "2", Content: "Task 2", ActiveForm: "Doing task 2", Status: StatusPending},
		{ID: "3", Content: "Task 3", ActiveForm: "Doing task 3", Status: StatusInProgress},
	}
	_ = manager.Update(items)

	tests := []struct {
		name      string
		id        string
		newStatus Status
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "update existing item to completed",
			id:        "1",
			newStatus: StatusCompleted,
			wantErr:   false,
		},
		{
			name:      "update to in-progress when one exists",
			id:        "2",
			newStatus: StatusInProgress,
			wantErr:   true,
			errMsg:    "cannot set to in-progress: 3 is already in progress",
		},
		{
			name:      "update non-existent item",
			id:        "999",
			newStatus: StatusCompleted,
			wantErr:   true,
			errMsg:    "todo with ID 999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.UpdateStatus(tt.id, tt.newStatus)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateStatus() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("UpdateStatus() error = %v, want %v", err.Error(), tt.errMsg)
			}

			if err == nil {
				item, _ := manager.Get(tt.id)
				if item.Status != tt.newStatus {
					t.Errorf("UpdateStatus() status = %v, want %v", item.Status, tt.newStatus)
				}
			}
		})
	}
}

// TestManager_Remove tests removing todo items.
func TestManager_Remove(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "remove existing item",
			id:      "2",
			wantErr: false,
		},
		{
			name:    "remove non-existent item",
			id:      "999",
			wantErr: true,
			errMsg:  "todo with ID 999 not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager()
			items := []TodoItem{
				{ID: "1", Content: "Task 1", ActiveForm: "Doing task 1", Status: StatusPending},
				{ID: "2", Content: "Task 2", ActiveForm: "Doing task 2", Status: StatusPending},
				{ID: "3", Content: "Task 3", ActiveForm: "Doing task 3", Status: StatusPending},
			}
			_ = manager.Update(items)

			initialCount := manager.Count()
			err := manager.Remove(tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("Remove() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Remove() error = %v, want %v", err.Error(), tt.errMsg)
			}

			if err == nil {
				if manager.Count() != initialCount-1 {
					t.Errorf("Remove() count = %d, want %d", manager.Count(), initialCount-1)
				}
				if _, err := manager.Get(tt.id); err == nil {
					t.Errorf("Remove() item still exists")
				}
			}
		})
	}
}

// TestManager_Stats tests the statistics calculation.
func TestManager_Stats(t *testing.T) {
	manager := NewManager()
	items := []TodoItem{
		{ID: "1", Content: "Task 1", ActiveForm: "Doing task 1", Status: StatusPending},
		{ID: "2", Content: "Task 2", ActiveForm: "Doing task 2", Status: StatusPending},
		{ID: "3", Content: "Task 3", ActiveForm: "Doing task 3", Status: StatusInProgress},
		{ID: "4", Content: "Task 4", ActiveForm: "Doing task 4", Status: StatusCompleted},
		{ID: "5", Content: "Task 5", ActiveForm: "Doing task 5", Status: StatusCompleted},
	}
	_ = manager.Update(items)

	stats := manager.Stats()

	if stats.Total != 5 {
		t.Errorf("Stats() total = %d, want 5", stats.Total)
	}
	if stats.Pending != 2 {
		t.Errorf("Stats() pending = %d, want 2", stats.Pending)
	}
	if stats.InProgress != 1 {
		t.Errorf("Stats() in_progress = %d, want 1", stats.InProgress)
	}
	if stats.Completed != 2 {
		t.Errorf("Stats() completed = %d, want 2", stats.Completed)
	}
}

// TestManager_GetInProgress tests getting the in-progress item.
func TestManager_GetInProgress(t *testing.T) {
	manager := NewManager()

	// Test with no in-progress item
	item, exists := manager.GetInProgress()
	if exists {
		t.Error("GetInProgress() should return false when no item is in progress")
	}
	if item != nil {
		t.Error("GetInProgress() should return nil when no item is in progress")
	}

	// Test with in-progress item
	items := []TodoItem{
		{ID: "1", Content: "Task 1", ActiveForm: "Doing task 1", Status: StatusPending},
		{ID: "2", Content: "Task 2", ActiveForm: "Doing task 2", Status: StatusInProgress},
		{ID: "3", Content: "Task 3", ActiveForm: "Doing task 3", Status: StatusCompleted},
	}
	_ = manager.Update(items)

	item, exists = manager.GetInProgress()
	if !exists {
		t.Error("GetInProgress() should return true when an item is in progress")
	}
	if item == nil || item.ID != "2" {
		t.Error("GetInProgress() should return the in-progress item")
	}
}

// TestManager_ConcurrentAccess tests concurrent access to the manager.
func TestManager_ConcurrentAccess(t *testing.T) {
	manager := NewManager()
	items := []TodoItem{
		{ID: "1", Content: "Task 1", ActiveForm: "Doing task 1", Status: StatusPending},
		{ID: "2", Content: "Task 2", ActiveForm: "Doing task 2", Status: StatusPending},
	}
	_ = manager.Update(items)

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = manager.GetAll()
			_ = manager.Stats()
			_ = manager.Count()
			_, _ = manager.Get("1")
		}()
	}

	// Concurrent writes (status updates)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			if err := manager.UpdateStatus(id, StatusCompleted); err != nil {
				errors <- err
			}
		}(fmt.Sprintf("%d", i%2+1))
	}

	wg.Wait()
	close(errors)

	// Check for race condition errors
	for err := range errors {
		if err != nil && err.Error() != "cannot set to in-progress: 1 is already in progress" &&
			err.Error() != "cannot set to in-progress: 2 is already in progress" {
			t.Errorf("Unexpected error during concurrent access: %v", err)
		}
	}
}

// TestManager_Clear tests clearing all items.
func TestManager_Clear(t *testing.T) {
	manager := NewManager()
	items := []TodoItem{
		{ID: "1", Content: "Task 1", ActiveForm: "Doing task 1", Status: StatusPending},
		{ID: "2", Content: "Task 2", ActiveForm: "Doing task 2", Status: StatusPending},
	}
	_ = manager.Update(items)

	if manager.Count() != 2 {
		t.Errorf("Initial count = %d, want 2", manager.Count())
	}

	manager.Clear()

	if manager.Count() != 0 {
		t.Errorf("After clear count = %d, want 0", manager.Count())
	}
}

// TestManager_MaxItemsLimit tests the maximum items limit.
func TestManager_MaxItemsLimit(t *testing.T) {
	manager := NewManager()

	// Add items up to the limit
	for i := 0; i < MaxTodoItems; i++ {
		err := manager.Add(TodoItem{
			ID:         fmt.Sprintf("id-%d", i),
			Content:    fmt.Sprintf("Task %d", i),
			ActiveForm: fmt.Sprintf("Doing task %d", i),
			Status:     StatusPending,
		})
		if err != nil {
			t.Fatalf("Failed to add item %d: %v", i, err)
		}
	}

	// Try to add one more
	err := manager.Add(TodoItem{
		ID:         "overflow",
		Content:    "Overflow task",
		ActiveForm: "Doing overflow task",
		Status:     StatusPending,
	})

	if err == nil {
		t.Error("Expected error when exceeding max items limit")
	}

	expectedErr := fmt.Sprintf("todo list is at maximum capacity (%d items)", MaxTodoItems)
	if err.Error() != expectedErr {
		t.Errorf("Error = %v, want %v", err.Error(), expectedErr)
	}
}