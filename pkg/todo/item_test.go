package todo

import (
	"testing"
	"time"
)

// TestStatus_IsValid tests the status validation.
func TestStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{
			name:   "pending is valid",
			status: StatusPending,
			want:   true,
		},
		{
			name:   "in_progress is valid",
			status: StatusInProgress,
			want:   true,
		},
		{
			name:   "completed is valid",
			status: StatusCompleted,
			want:   true,
		},
		{
			name:   "empty string is invalid",
			status: Status(""),
			want:   false,
		},
		{
			name:   "invalid status",
			status: Status("invalid"),
			want:   false,
		},
		{
			name:   "mixed case is invalid",
			status: Status("Pending"),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNewTodoItem tests the creation of new todo items.
func TestNewTodoItem(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		content    string
		activeForm string
		status     Status
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid todo item",
			id:         "task-1",
			content:    "Write tests",
			activeForm: "Writing tests",
			status:     StatusPending,
			wantErr:    false,
		},
		{
			name:       "empty ID",
			id:         "",
			content:    "Write tests",
			activeForm: "Writing tests",
			status:     StatusPending,
			wantErr:    true,
			errMsg:     "todo ID cannot be empty",
		},
		{
			name:       "empty content",
			id:         "task-1",
			content:    "",
			activeForm: "Writing tests",
			status:     StatusPending,
			wantErr:    true,
			errMsg:     "todo content cannot be empty",
		},
		{
			name:       "empty activeForm",
			id:         "task-1",
			content:    "Write tests",
			activeForm: "",
			status:     StatusPending,
			wantErr:    true,
			errMsg:     "todo activeForm cannot be empty",
		},
		{
			name:       "invalid status",
			id:         "task-1",
			content:    "Write tests",
			activeForm: "Writing tests",
			status:     Status("invalid"),
			wantErr:    true,
			errMsg:     "invalid status: invalid",
		},
		{
			name:       "all statuses work",
			id:         "task-1",
			content:    "Write tests",
			activeForm: "Writing tests",
			status:     StatusInProgress,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := NewTodoItem(tt.id, tt.content, tt.activeForm, tt.status)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewTodoItem() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("NewTodoItem() error = %v, want %v", err.Error(), tt.errMsg)
			}

			if err == nil {
				if item.ID != tt.id {
					t.Errorf("ID = %v, want %v", item.ID, tt.id)
				}
				if item.Content != tt.content {
					t.Errorf("Content = %v, want %v", item.Content, tt.content)
				}
				if item.ActiveForm != tt.activeForm {
					t.Errorf("ActiveForm = %v, want %v", item.ActiveForm, tt.activeForm)
				}
				if item.Status != tt.status {
					t.Errorf("Status = %v, want %v", item.Status, tt.status)
				}
				if item.CreatedAt.IsZero() {
					t.Error("CreatedAt should not be zero")
				}
				if item.UpdatedAt.IsZero() {
					t.Error("UpdatedAt should not be zero")
				}
			}
		})
	}
}

// TestTodoItem_UpdateStatus tests status updates on todo items.
func TestTodoItem_UpdateStatus(t *testing.T) {
	item, _ := NewTodoItem("task-1", "Write tests", "Writing tests", StatusPending)
	originalUpdatedAt := item.UpdatedAt

	// Sleep briefly to ensure time difference
	time.Sleep(10 * time.Millisecond)

	tests := []struct {
		name      string
		newStatus Status
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "update to in_progress",
			newStatus: StatusInProgress,
			wantErr:   false,
		},
		{
			name:      "update to completed",
			newStatus: StatusCompleted,
			wantErr:   false,
		},
		{
			name:      "update to invalid status",
			newStatus: Status("invalid"),
			wantErr:   true,
			errMsg:    "invalid status: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new item for each test to avoid state pollution
			testItem, _ := NewTodoItem("task-1", "Write tests", "Writing tests", StatusPending)
			err := testItem.UpdateStatus(tt.newStatus)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateStatus() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("UpdateStatus() error = %v, want %v", err.Error(), tt.errMsg)
			}

			if err == nil {
				if testItem.Status != tt.newStatus {
					t.Errorf("Status = %v, want %v", testItem.Status, tt.newStatus)
				}
				// UpdatedAt should be updated
				if !testItem.UpdatedAt.After(originalUpdatedAt) {
					t.Error("UpdatedAt should be updated after status change")
				}
			}
		})
	}
}

// TestTodoItem_StatusHelpers tests the status helper methods.
func TestTodoItem_StatusHelpers(t *testing.T) {
	tests := []struct {
		name         string
		status       Status
		isCompleted  bool
		isInProgress bool
		isPending    bool
	}{
		{
			name:         "pending status",
			status:       StatusPending,
			isCompleted:  false,
			isInProgress: false,
			isPending:    true,
		},
		{
			name:         "in_progress status",
			status:       StatusInProgress,
			isCompleted:  false,
			isInProgress: true,
			isPending:    false,
		},
		{
			name:         "completed status",
			status:       StatusCompleted,
			isCompleted:  true,
			isInProgress: false,
			isPending:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, _ := NewTodoItem("task-1", "Test", "Testing", tt.status)

			if got := item.IsCompleted(); got != tt.isCompleted {
				t.Errorf("IsCompleted() = %v, want %v", got, tt.isCompleted)
			}

			if got := item.IsInProgress(); got != tt.isInProgress {
				t.Errorf("IsInProgress() = %v, want %v", got, tt.isInProgress)
			}

			if got := item.IsPending(); got != tt.isPending {
				t.Errorf("IsPending() = %v, want %v", got, tt.isPending)
			}
		})
	}
}

// TestTodoItem_isValidTransition tests status transition validation.
func TestTodoItem_isValidTransition(t *testing.T) {
	tests := []struct {
		name        string
		fromStatus  Status
		toStatus    Status
		wantAllowed bool
	}{
		// Currently all transitions are allowed
		{
			name:        "pending to in_progress",
			fromStatus:  StatusPending,
			toStatus:    StatusInProgress,
			wantAllowed: true,
		},
		{
			name:        "in_progress to completed",
			fromStatus:  StatusInProgress,
			toStatus:    StatusCompleted,
			wantAllowed: true,
		},
		{
			name:        "completed to pending",
			fromStatus:  StatusCompleted,
			toStatus:    StatusPending,
			wantAllowed: true, // Currently allowed, could be restricted later
		},
		{
			name:        "pending to completed",
			fromStatus:  StatusPending,
			toStatus:    StatusCompleted,
			wantAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, _ := NewTodoItem("task-1", "Test", "Testing", tt.fromStatus)
			got := item.isValidTransition(tt.toStatus)

			if got != tt.wantAllowed {
				t.Errorf("isValidTransition(%v -> %v) = %v, want %v",
					tt.fromStatus, tt.toStatus, got, tt.wantAllowed)
			}
		})
	}
}

// TestTodoItem_TimestampBehavior tests the timestamp behavior.
func TestTodoItem_TimestampBehavior(t *testing.T) {
	// Test that CreatedAt and UpdatedAt are set on creation
	item, _ := NewTodoItem("task-1", "Test", "Testing", StatusPending)

	if item.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set on creation")
	}

	if item.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set on creation")
	}

	// Initially, CreatedAt and UpdatedAt should be very close (within a second)
	diff := item.UpdatedAt.Sub(item.CreatedAt)
	if diff > time.Second {
		t.Errorf("CreatedAt and UpdatedAt should be close on creation, diff: %v", diff)
	}

	originalCreatedAt := item.CreatedAt
	originalUpdatedAt := item.UpdatedAt

	// Sleep to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Update status and check timestamps
	_ = item.UpdateStatus(StatusInProgress)

	if !item.CreatedAt.Equal(originalCreatedAt) {
		t.Error("CreatedAt should not change on status update")
	}

	if !item.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated after status change")
	}
}

// TestValidStatuses tests that ValidStatuses contains all expected values.
func TestValidStatuses(t *testing.T) {
	expectedStatuses := []Status{StatusPending, StatusInProgress, StatusCompleted}

	if len(ValidStatuses) != len(expectedStatuses) {
		t.Errorf("ValidStatuses length = %d, want %d", len(ValidStatuses), len(expectedStatuses))
	}

	statusMap := make(map[Status]bool)
	for _, s := range ValidStatuses {
		statusMap[s] = true
	}

	for _, expected := range expectedStatuses {
		if !statusMap[expected] {
			t.Errorf("ValidStatuses missing expected status: %s", expected)
		}
	}
}