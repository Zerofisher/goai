package todo

import (
	"fmt"
	"sync"
	"time"
)

const (
	// MaxTodoItems is the maximum number of todos allowed.
	MaxTodoItems = 20
)

// Stats contains statistics about the todo list.
type Stats struct {
	Total      int `json:"total"`
	Completed  int `json:"completed"`
	InProgress int `json:"in_progress"`
	Pending    int `json:"pending"`
}

// Manager manages a collection of todo items.
type Manager struct {
	items    []TodoItem
	maxItems int
	mu       sync.RWMutex
}

// NewManager creates a new todo manager with default settings.
func NewManager() *Manager {
	return &Manager{
		items:    make([]TodoItem, 0),
		maxItems: MaxTodoItems,
	}
}

// Update replaces the entire todo list with new items.
// It validates all constraints including:
// - Maximum item limit
// - Only one task can be in progress
// - No duplicate IDs
// - All fields are valid
func (m *Manager) Update(items []TodoItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(items) > m.maxItems {
		return fmt.Errorf("todo list is limited to %d items, got %d", m.maxItems, len(items))
	}

	// Validate items
	seenIDs := make(map[string]bool)
	inProgressCount := 0

	for i := range items {
		item := &items[i]

		// Check for duplicate IDs
		if seenIDs[item.ID] {
			return fmt.Errorf("duplicate todo ID: %s", item.ID)
		}
		seenIDs[item.ID] = true

		// Validate required fields
		if item.ID == "" {
			return fmt.Errorf("todo at index %d has empty ID", i)
		}

		if item.Content == "" {
			return fmt.Errorf("todo %s has empty content", item.ID)
		}

		if item.ActiveForm == "" {
			return fmt.Errorf("todo %s has empty activeForm", item.ID)
		}

		// Validate status
		if !item.Status.IsValid() {
			return fmt.Errorf("todo %s has invalid status: %s", item.ID, item.Status)
		}

		// Count in-progress tasks
		if item.Status == StatusInProgress {
			inProgressCount++
		}

		// Set timestamps if not provided
		if item.CreatedAt.IsZero() {
			item.CreatedAt = time.Now()
		}
		if item.UpdatedAt.IsZero() {
			item.UpdatedAt = time.Now()
		}
	}

	// Enforce constraint: only one task can be in progress
	if inProgressCount > 1 {
		return fmt.Errorf("only one task can be in_progress at a time, found %d", inProgressCount)
	}

	m.items = items
	return nil
}

// Add adds a new todo item to the list.
func (m *Manager) Add(item TodoItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.items) >= m.maxItems {
		return fmt.Errorf("todo list is at maximum capacity (%d items)", m.maxItems)
	}

	// Check for duplicate ID
	for _, existing := range m.items {
		if existing.ID == item.ID {
			return fmt.Errorf("todo with ID %s already exists", item.ID)
		}
	}

	// Check in-progress constraint
	if item.Status == StatusInProgress {
		for _, existing := range m.items {
			if existing.Status == StatusInProgress {
				return fmt.Errorf("cannot add in-progress task: %s is already in progress", existing.ID)
			}
		}
	}

	// Set timestamps
	now := time.Now()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = now
	}

	m.items = append(m.items, item)
	return nil
}

// UpdateStatus changes the status of a specific todo item.
func (m *Manager) UpdateStatus(id string, status Status) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var item *TodoItem
	for i := range m.items {
		if m.items[i].ID == id {
			item = &m.items[i]
			break
		}
	}

	if item == nil {
		return fmt.Errorf("todo with ID %s not found", id)
	}

	// Check in-progress constraint
	if status == StatusInProgress {
		for i := range m.items {
			if m.items[i].ID != id && m.items[i].Status == StatusInProgress {
				return fmt.Errorf("cannot set to in-progress: %s is already in progress", m.items[i].ID)
			}
		}
	}

	return item.UpdateStatus(status)
}

// Remove removes a todo item by ID.
func (m *Manager) Remove(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, item := range m.items {
		if item.ID == id {
			// Remove item by slicing
			m.items = append(m.items[:i], m.items[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("todo with ID %s not found", id)
}

// Get retrieves a todo item by ID.
func (m *Manager) Get(id string) (*TodoItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, item := range m.items {
		if item.ID == id {
			// Return a copy to prevent external modification
			itemCopy := item
			return &itemCopy, nil
		}
	}

	return nil, fmt.Errorf("todo with ID %s not found", id)
}

// GetAll returns a copy of all todo items.
func (m *Manager) GetAll() []TodoItem {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]TodoItem, len(m.items))
	copy(result, m.items)
	return result
}

// Stats returns statistics about the todo list.
func (m *Manager) Stats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := Stats{
		Total: len(m.items),
	}

	for _, item := range m.items {
		switch item.Status {
		case StatusCompleted:
			stats.Completed++
		case StatusInProgress:
			stats.InProgress++
		case StatusPending:
			stats.Pending++
		}
	}

	return stats
}

// Clear removes all todo items.
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items = make([]TodoItem, 0)
}

// GetInProgress returns the currently in-progress todo item, if any.
func (m *Manager) GetInProgress() (*TodoItem, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, item := range m.items {
		if item.Status == StatusInProgress {
			itemCopy := item
			return &itemCopy, true
		}
	}

	return nil, false
}

// HasInProgress returns true if there is a todo in progress.
func (m *Manager) HasInProgress() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, item := range m.items {
		if item.Status == StatusInProgress {
			return true
		}
	}

	return false
}

// Count returns the total number of todo items.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.items)
}

// Render returns a string representation of the todo list.
// This method delegates to the Renderer for actual formatting.
func (m *Manager) Render() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	renderer := NewRenderer(false) // Default to no colors for basic rendering
	return renderer.Render(m.items)
}