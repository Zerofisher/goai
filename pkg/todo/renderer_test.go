package todo

import (
	"fmt"
	"strings"
	"testing"
)

// TestNewRenderer tests the creation of new renderers.
func TestNewRenderer(t *testing.T) {
	tests := []struct {
		name           string
		enableColors   bool
		expectedColors bool
	}{
		{
			name:           "renderer with colors",
			enableColors:   true,
			expectedColors: true,
		},
		{
			name:           "renderer without colors",
			enableColors:   false,
			expectedColors: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewRenderer(tt.enableColors)
			if renderer == nil {
				t.Fatal("NewRenderer returned nil")
			}
			if renderer.enableColors != tt.expectedColors {
				t.Errorf("enableColors = %v, want %v", renderer.enableColors, tt.expectedColors)
			}
		})
	}
}

// TestRenderer_Render tests the rendering of todo items.
func TestRenderer_Render(t *testing.T) {
	tests := []struct {
		name         string
		items        []TodoItem
		enableColors bool
		contains     []string
		notContains  []string
	}{
		{
			name:         "empty list",
			items:        []TodoItem{},
			enableColors: false,
			contains:     []string{"No todos yet"},
		},
		{
			name: "single pending item",
			items: []TodoItem{
				{ID: "1", Content: "Write tests", ActiveForm: "Writing tests", Status: StatusPending},
			},
			enableColors: false,
			contains:     []string{"☐", "Write tests"},
		},
		{
			name: "single in-progress item",
			items: []TodoItem{
				{ID: "1", Content: "Run tests", ActiveForm: "Running tests", Status: StatusInProgress},
			},
			enableColors: false,
			contains:     []string{"☐", "Run tests", "(Running tests)"},
		},
		{
			name: "single completed item",
			items: []TodoItem{
				{ID: "1", Content: "Deploy", ActiveForm: "Deploying", Status: StatusCompleted},
			},
			enableColors: false,
			contains:     []string{"☒", "Deploy"},
		},
		{
			name: "mixed status items",
			items: []TodoItem{
				{ID: "1", Content: "Task 1", ActiveForm: "Doing task 1", Status: StatusPending},
				{ID: "2", Content: "Task 2", ActiveForm: "Doing task 2", Status: StatusInProgress},
				{ID: "3", Content: "Task 3", ActiveForm: "Doing task 3", Status: StatusCompleted},
			},
			enableColors: false,
			contains:     []string{"Task 1", "Task 2", "(Doing task 2)", "Task 3", "☐", "☒"},
		},
		{
			name: "with colors enabled - pending",
			items: []TodoItem{
				{ID: "1", Content: "Pending task", ActiveForm: "Working on task", Status: StatusPending},
			},
			enableColors: true,
			contains:     []string{"☐", "Pending task", "[38;2;176;176;176m"}, // Pending color
		},
		{
			name: "with colors enabled - in progress",
			items: []TodoItem{
				{ID: "1", Content: "Active task", ActiveForm: "Working on task", Status: StatusInProgress},
			},
			enableColors: true,
			contains:     []string{"☐", "Working on task", "[38;2;120;200;255m"}, // Progress color
		},
		{
			name: "with colors enabled - completed",
			items: []TodoItem{
				{ID: "1", Content: "Done task", ActiveForm: "Finishing task", Status: StatusCompleted},
			},
			enableColors: true,
			contains:     []string{"☒", "Done task", "[38;2;34;139;34m", "[9m"}, // Completed color and strikethrough
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewRenderer(tt.enableColors)
			output := renderer.Render(tt.items)

			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Render() output missing expected string: %q\nGot: %s", expected, output)
				}
			}

			for _, unexpected := range tt.notContains {
				if strings.Contains(output, unexpected) {
					t.Errorf("Render() output contains unexpected string: %q\nGot: %s", unexpected, output)
				}
			}
		})
	}
}

// TestRenderer_RenderWithStats tests rendering with statistics.
func TestRenderer_RenderWithStats(t *testing.T) {
	tests := []struct {
		name         string
		items        []TodoItem
		stats        Stats
		enableColors bool
		checkStats   bool
	}{
		{
			name:         "empty list with stats",
			items:        []TodoItem{},
			stats:        Stats{Total: 0, Completed: 0, InProgress: 0, Pending: 0},
			enableColors: false,
			checkStats:   true,
		},
		{
			name: "items with stats",
			items: []TodoItem{
				{ID: "1", Content: "Task 1", ActiveForm: "Doing task 1", Status: StatusPending},
				{ID: "2", Content: "Task 2", ActiveForm: "Doing task 2", Status: StatusInProgress},
				{ID: "3", Content: "Task 3", ActiveForm: "Doing task 3", Status: StatusCompleted},
				{ID: "4", Content: "Task 4", ActiveForm: "Doing task 4", Status: StatusPending},
			},
			stats:        Stats{Total: 4, Completed: 1, InProgress: 1, Pending: 2},
			enableColors: false,
			checkStats:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewRenderer(tt.enableColors)
			output := renderer.RenderWithStats(tt.items, tt.stats)

			// Check for stats line
			if tt.checkStats && len(tt.items) > 0 {
				if !strings.Contains(output, "Total:") {
					t.Error("RenderWithStats() should contain 'Total:' in stats")
				}
				if !strings.Contains(output, "Completed:") {
					t.Error("RenderWithStats() should contain 'Completed:' in stats")
				}
				if !strings.Contains(output, "In Progress:") {
					t.Error("RenderWithStats() should contain 'In Progress:' in stats")
				}
				if !strings.Contains(output, "Pending:") {
					t.Error("RenderWithStats() should contain 'Pending:' in stats")
				}
			}
		})
	}
}

// TestRenderer_RenderCompact tests compact rendering.
func TestRenderer_RenderCompact(t *testing.T) {
	tests := []struct {
		name     string
		items    []TodoItem
		contains []string
	}{
		{
			name:     "empty list",
			items:    []TodoItem{},
			contains: []string{"No todos"},
		},
		{
			name: "some items",
			items: []TodoItem{
				{ID: "1", Content: "Task 1", ActiveForm: "Doing task 1", Status: StatusPending},
				{ID: "2", Content: "Task 2", ActiveForm: "Doing task 2", Status: StatusCompleted},
			},
			contains: []string{"[", "]", "1/2 completed"},
		},
		{
			name: "all completed",
			items: []TodoItem{
				{ID: "1", Content: "Task 1", ActiveForm: "Doing task 1", Status: StatusCompleted},
				{ID: "2", Content: "Task 2", ActiveForm: "Doing task 2", Status: StatusCompleted},
			},
			contains: []string{"2/2 completed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewRenderer(false)
			output := renderer.RenderCompact(tt.items)

			for _, exp := range tt.contains {
				if !strings.Contains(output, exp) {
					t.Errorf("RenderCompact() missing expected string: %q\nGot: %s", exp, output)
				}
			}
		})
	}
}

// TestRenderer_RenderMarkdown tests Markdown rendering.
func TestRenderer_RenderMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		items    []TodoItem
		contains []string
	}{
		{
			name:     "empty list",
			items:    []TodoItem{},
			contains: []string{"- [ ] No todos yet"},
		},
		{
			name: "mixed items",
			items: []TodoItem{
				{ID: "1", Content: "Pending task", Status: StatusPending},
				{ID: "2", Content: "Active task", Status: StatusInProgress},
				{ID: "3", Content: "Done task", Status: StatusCompleted},
			},
			contains: []string{
				"## Todo List",
				"- [ ] Pending task",
				"- [ ] Active task *(in progress)*",
				"- [x] Done task",
				"### Statistics",
				"Total: 3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewRenderer(false)
			output := renderer.RenderMarkdown(tt.items)

			for _, exp := range tt.contains {
				if !strings.Contains(output, exp) {
					t.Errorf("RenderMarkdown() missing expected string: %q\nGot: %s", exp, output)
				}
			}
		})
	}
}

// TestRenderer_ColorConsistency tests that color codes are consistent.
func TestRenderer_ColorConsistency(t *testing.T) {
	renderer := NewRenderer(true)

	// Test that the same status always produces the same color
	items := []TodoItem{
		{ID: "1", Content: "Task 1", ActiveForm: "Doing 1", Status: StatusPending},
		{ID: "2", Content: "Task 2", ActiveForm: "Doing 2", Status: StatusPending},
	}

	output := renderer.Render(items)

	// Both items should be rendered with the same color code for pending status
	pendingColor := "\x1b[38;2;176;176;176m"
	if !strings.Contains(output, pendingColor) {
		t.Error("Pending items should have consistent color codes")
	}
}

// TestRenderer_LargeList tests rendering a large number of items.
func TestRenderer_LargeList(t *testing.T) {
	items := make([]TodoItem, 20) // Maximum allowed
	for i := 0; i < 20; i++ {
		status := StatusPending
		if i%3 == 1 {
			status = StatusCompleted
		}
		items[i] = TodoItem{
			ID:         fmt.Sprintf("id-%d", i),
			Content:    fmt.Sprintf("Task %d", i),
			ActiveForm: fmt.Sprintf("Working on task %d", i),
			Status:     status,
		}
	}

	renderer := NewRenderer(false)
	renderer.SetShowStats(false) // Disable stats for line count test
	output := renderer.Render(items)

	// Should have 20 lines (one per item)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 20 {
		t.Errorf("Expected 20 lines for 20 items, got %d", len(lines))
	}

	// Check stats with RenderWithStats
	stats := Stats{
		Total:      20,
		Completed:  6,
		InProgress: 0,
		Pending:    14,
	}
	outputWithStats := renderer.RenderWithStats(items, stats)
	if !strings.Contains(outputWithStats, "Total: 20") {
		t.Error("Stats should show Total: 20")
	}
}

// TestRenderer_SetShowStats tests the SetShowStats configuration.
func TestRenderer_SetShowStats(t *testing.T) {
	renderer := NewRenderer(false)
	items := []TodoItem{
		{ID: "1", Content: "Task 1", ActiveForm: "Doing task 1", Status: StatusPending},
		{ID: "2", Content: "Task 2", ActiveForm: "Doing task 2", Status: StatusCompleted},
	}

	// Test with stats enabled (default)
	output := renderer.Render(items)
	if !strings.Contains(output, "Total:") {
		t.Error("Render() should show stats by default")
	}

	// Test with stats disabled
	renderer.SetShowStats(false)
	output = renderer.Render(items)
	if strings.Contains(output, "Total:") {
		t.Error("Render() should not show stats when disabled")
	}

	// Re-enable and test
	renderer.SetShowStats(true)
	output = renderer.Render(items)
	if !strings.Contains(output, "Total:") {
		t.Error("Render() should show stats when re-enabled")
	}
}