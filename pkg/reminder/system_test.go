package reminder

import (
	"strings"
	"testing"
	"time"

	"github.com/Zerofisher/goai/pkg/types"
)

// TestNewSystem tests the creation of a new reminder system.
func TestNewSystem(t *testing.T) {
	system := NewSystem(3, 10)

	if system == nil {
		t.Fatal("NewSystem returned nil")
	}

	if system.roundsWithoutTodo != 0 {
		t.Errorf("Expected roundsWithoutTodo to be 0, got %d", system.roundsWithoutTodo)
	}

	if system.reminderInterval != 10 {
		t.Errorf("Expected reminderInterval to be 10, got %d", system.reminderInterval)
	}

	if !system.enabled {
		t.Error("System should be enabled by default")
	}

	if len(system.pendingReminders) != 0 {
		t.Error("Should have no pending reminders initially")
	}
}

// TestNewSystemWithConfig tests creation with custom configuration.
func TestNewSystemWithConfig(t *testing.T) {
	config := Config{
		Enabled:              false,
		InitialReminderAfter: 5,
		ReminderInterval:     15,
		MaxReminders:         5,
	}

	system := NewSystemWithConfig(config)

	if system.enabled {
		t.Error("System should be disabled based on config")
	}

	if system.reminderInterval != 15 {
		t.Errorf("Expected reminderInterval to be 15, got %d", system.reminderInterval)
	}
}

// TestDefaultConfig tests the default configuration.
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if !config.Enabled {
		t.Error("Default config should be enabled")
	}

	if config.InitialReminderAfter != 3 {
		t.Errorf("Expected InitialReminderAfter to be 3, got %d", config.InitialReminderAfter)
	}

	if config.ReminderInterval != 10 {
		t.Errorf("Expected ReminderInterval to be 10, got %d", config.ReminderInterval)
	}

	if config.MaxReminders != 3 {
		t.Errorf("Expected MaxReminders to be 3, got %d", config.MaxReminders)
	}
}

// TestSystem_IncrementRound tests incrementing the round counter.
func TestSystem_IncrementRound(t *testing.T) {
	system := NewSystem(3, 10)

	// Initially should be 0
	if system.GetRoundsWithoutTodo() != 0 {
		t.Errorf("Initial rounds should be 0, got %d", system.GetRoundsWithoutTodo())
	}

	// Increment once
	system.IncrementRound()
	if system.GetRoundsWithoutTodo() != 1 {
		t.Errorf("After one increment should be 1, got %d", system.GetRoundsWithoutTodo())
	}

	// Test IncrementRounds alias
	system.IncrementRounds()
	if system.GetRoundsWithoutTodo() != 2 {
		t.Errorf("After two increments should be 2, got %d", system.GetRoundsWithoutTodo())
	}
}

// TestSystem_Reset tests resetting the reminder system.
func TestSystem_Reset(t *testing.T) {
	system := NewSystem(3, 10)

	// Add some rounds and reminders
	for i := 0; i < 5; i++ {
		system.IncrementRound()
	}
	system.AddReminder("test reminder")

	// Reset
	system.Reset()

	if system.GetRoundsWithoutTodo() != 0 {
		t.Errorf("After reset rounds should be 0, got %d", system.GetRoundsWithoutTodo())
	}

	if len(system.GetPendingReminders()) != 0 {
		t.Error("After reset should have no pending reminders")
	}
}

// TestSystem_ShouldRemind tests the reminder logic.
func TestSystem_ShouldRemind(t *testing.T) {
	tests := []struct {
		name           string
		roundsToAdd    int
		expectedRemind bool
		description    string
	}{
		{
			name:           "no rounds",
			roundsToAdd:    0,
			expectedRemind: false,
			description:    "Should not remind with 0 rounds",
		},
		{
			name:           "before initial",
			roundsToAdd:    2,
			expectedRemind: false,
			description:    "Should not remind before 3 rounds",
		},
		{
			name:           "initial reminder",
			roundsToAdd:    3,
			expectedRemind: true,
			description:    "Should remind at exactly 3 rounds",
		},
		{
			name:           "after initial before interval",
			roundsToAdd:    5,
			expectedRemind: false,
			description:    "Should not remind at 5 rounds",
		},
		{
			name:           "first interval",
			roundsToAdd:    13, // 3 + 10
			expectedRemind: true,
			description:    "Should remind at 3 + 10 rounds",
		},
		{
			name:           "long idle",
			roundsToAdd:    20,
			expectedRemind: true,
			description:    "Should remind after 20 rounds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system := NewSystem(3, 10)

			for i := 0; i < tt.roundsToAdd; i++ {
				system.IncrementRound()
			}

			got := system.ShouldRemind()
			if got != tt.expectedRemind {
				t.Errorf("%s: ShouldRemind() = %v, want %v", tt.description, got, tt.expectedRemind)
			}
		})
	}
}

// TestSystem_Check tests the Check method that generates reminders.
func TestSystem_Check(t *testing.T) {
	system := NewSystem(3, 10)

	// Initially no reminders
	reminders := system.Check()
	if len(reminders) != 0 {
		t.Error("Should have no reminders initially")
	}

	// After 3 rounds should generate initial reminder
	for i := 0; i < 3; i++ {
		system.IncrementRound()
	}

	reminders = system.Check()
	if len(reminders) == 0 {
		t.Error("Should have reminders after 3 rounds")
	}

	// Verify it's the initial template
	if len(reminders) > 0 {
		if !strings.Contains(reminders[0], "complex work should be tracked") {
			t.Error("Should contain initial reminder text")
		}
	}
}

// TestSystem_AddReminder tests adding custom reminders.
func TestSystem_AddReminder(t *testing.T) {
	system := NewSystem(3, 10)

	// Add a custom reminder
	system.AddReminder("Custom reminder 1")
	reminders := system.GetPendingReminders()

	if len(reminders) != 1 {
		t.Errorf("Should have 1 reminder, got %d", len(reminders))
	}

	if reminders[0] != "Custom reminder 1" {
		t.Errorf("Expected 'Custom reminder 1', got %s", reminders[0])
	}

	// Add multiple reminders up to limit
	system.AddReminder("Custom reminder 2")
	system.AddReminder("Custom reminder 3")
	system.AddReminder("Custom reminder 4") // Should remove the oldest

	reminders = system.GetPendingReminders()
	if len(reminders) != 3 {
		t.Errorf("Should have at most 3 reminders, got %d", len(reminders))
	}

	// First reminder should have been removed
	if reminders[0] == "Custom reminder 1" {
		t.Error("Oldest reminder should have been removed")
	}
}

// TestSystem_ConsumePendingReminders tests consuming reminders.
func TestSystem_ConsumePendingReminders(t *testing.T) {
	system := NewSystem(3, 10)

	// Add reminders
	system.AddReminder("Reminder 1")
	system.AddReminder("Reminder 2")

	// Consume them
	consumed := system.ConsumePendingReminders()
	if len(consumed) != 2 {
		t.Errorf("Should consume 2 reminders, got %d", len(consumed))
	}

	// Should be cleared after consumption
	remaining := system.GetPendingReminders()
	if len(remaining) != 0 {
		t.Error("Should have no reminders after consumption")
	}
}

// TestSystem_EnableDisable tests enabling and disabling the system.
func TestSystem_EnableDisable(t *testing.T) {
	system := NewSystem(3, 10)

	// Should be enabled by default
	if !system.IsEnabled() {
		t.Error("System should be enabled by default")
	}

	// Disable
	system.Disable()
	if system.IsEnabled() {
		t.Error("System should be disabled after Disable()")
	}

	// Should not remind when disabled
	for i := 0; i < 3; i++ {
		system.IncrementRound()
	}
	if system.ShouldRemind() {
		t.Error("Should not remind when disabled")
	}

	// Re-enable
	system.Enable()
	if !system.IsEnabled() {
		t.Error("System should be enabled after Enable()")
	}

	// Should remind after re-enabling
	if !system.ShouldRemind() {
		t.Error("Should remind after re-enabling with 3 rounds")
	}
}

// TestSystem_Inject tests injecting reminders into messages.
func TestSystem_Inject(t *testing.T) {
	system := NewSystem(3, 10)

	// Add a pending reminder
	system.AddReminder("Test reminder")

	// Create test messages
	messages := []types.Message{
		{
			Role: "system",
			Content: []types.Content{
				{Type: "text", Text: "System message"},
			},
		},
		{
			Role: "user",
			Content: []types.Content{
				{Type: "text", Text: "User message"},
			},
		},
	}

	// Inject reminders
	result := system.Inject(messages)

	// Should inject into the last user message
	if len(result[1].Content) != 2 {
		t.Errorf("User message should have 2 content blocks after injection, got %d", len(result[1].Content))
	}

	// First content should be the reminder
	if result[1].Content[0].Text != "Test reminder" {
		t.Errorf("First content should be reminder, got %s", result[1].Content[0].Text)
	}

	// Original content should be preserved
	if result[1].Content[1].Text != "User message" {
		t.Errorf("Original content should be preserved, got %s", result[1].Content[1].Text)
	}

	// Reminders should be cleared after injection
	remaining := system.GetPendingReminders()
	if len(remaining) != 0 {
		t.Error("Reminders should be cleared after injection")
	}
}

// TestSystem_GetStats tests statistics retrieval.
func TestSystem_GetStats(t *testing.T) {
	system := NewSystem(3, 10)

	// Add some state
	for i := 0; i < 5; i++ {
		system.IncrementRound()
	}
	system.AddReminder("Test")

	stats := system.GetStats()

	if stats.RoundsWithoutTodo != 5 {
		t.Errorf("Stats should show 5 rounds, got %d", stats.RoundsWithoutTodo)
	}

	if stats.PendingReminders != 1 {
		t.Errorf("Stats should show 1 pending reminder, got %d", stats.PendingReminders)
	}

	if !stats.Enabled {
		t.Error("Stats should show system as enabled")
	}
}

// TestSystem_FormatStats tests formatting statistics.
func TestSystem_FormatStats(t *testing.T) {
	system := NewSystem(3, 10)

	// Add some state
	for i := 0; i < 5; i++ {
		system.IncrementRound()
	}

	formatted := system.FormatStats()

	if !strings.Contains(formatted, "Rounds without todo: 5") {
		t.Error("Formatted stats should mention rounds")
	}

	if !strings.Contains(formatted, "Enabled: true") {
		t.Error("Formatted stats should mention enabled status")
	}
}

// TestSystem_SetReminderInterval tests updating the reminder interval.
func TestSystem_SetReminderInterval(t *testing.T) {
	system := NewSystem(3, 10)

	// Change interval
	system.SetReminderInterval(5)

	// Test with new interval
	// Should remind at 3 (initial) and 3+5=8
	for i := 0; i < 8; i++ {
		system.IncrementRound()
	}

	if !system.ShouldRemind() {
		t.Error("Should remind at 8 rounds with interval of 5")
	}
}

// TestGetTemplate tests template retrieval.
func TestGetTemplate(t *testing.T) {
	tests := []struct {
		name         string
		templateType string
		shouldContain string
	}{
		{
			name:         "initial template",
			templateType: TemplateInitial,
			shouldContain: "complex work should be tracked",
		},
		{
			name:         "nag template",
			templateType: TemplateNag,
			shouldContain: "more than ten rounds passed",
		},
		{
			name:         "long idle template",
			templateType: TemplateLongIdle,
			shouldContain: "extended period without task tracking",
		},
		{
			name:         "unknown template defaults to initial",
			templateType: "unknown",
			shouldContain: "complex work should be tracked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := GetTemplate(tt.templateType)
			if !strings.Contains(template, tt.shouldContain) {
				t.Errorf("Template should contain '%s', got: %s", tt.shouldContain, template)
			}
		})
	}
}

// TestSystem_GetTimeSinceLastUpdate tests time tracking.
func TestSystem_GetTimeSinceLastUpdate(t *testing.T) {
	system := NewSystem(3, 10)

	// Initially should be near zero
	duration := system.GetTimeSinceLastUpdate()
	if duration > time.Second {
		t.Error("Initial time since update should be less than a second")
	}

	// Sleep a bit
	time.Sleep(100 * time.Millisecond)

	duration = system.GetTimeSinceLastUpdate()
	if duration < 100*time.Millisecond {
		t.Error("Time since update should be at least 100ms")
	}

	// Reset should update the time
	system.Reset()
	duration = system.GetTimeSinceLastUpdate()
	if duration > time.Second {
		t.Error("After reset, time since update should be near zero")
	}
}

// TestSystem_ClearPendingReminders tests clearing reminders.
func TestSystem_ClearPendingReminders(t *testing.T) {
	system := NewSystem(3, 10)

	// Add reminders
	system.AddReminder("Reminder 1")
	system.AddReminder("Reminder 2")

	// Verify they exist
	if len(system.GetPendingReminders()) != 2 {
		t.Error("Should have 2 pending reminders")
	}

	// Clear them
	system.ClearPendingReminders()

	// Verify they're gone
	if len(system.GetPendingReminders()) != 0 {
		t.Error("Should have no pending reminders after clearing")
	}
}

// TestSystem_ConcurrentAccess tests thread safety.
func TestSystem_ConcurrentAccess(t *testing.T) {
	system := NewSystem(3, 10)

	// Run concurrent operations
	done := make(chan bool, 4)

	// Increment rounds
	go func() {
		for i := 0; i < 100; i++ {
			system.IncrementRound()
		}
		done <- true
	}()

	// Add reminders
	go func() {
		for i := 0; i < 50; i++ {
			system.AddReminder("reminder")
		}
		done <- true
	}()

	// Check status
	go func() {
		for i := 0; i < 100; i++ {
			_ = system.ShouldRemind()
			_ = system.GetStats()
		}
		done <- true
	}()

	// Reset periodically
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(10 * time.Millisecond)
			system.Reset()
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 4; i++ {
		<-done
	}

	// System should still be in a valid state
	stats := system.GetStats()
	if stats.RoundsWithoutTodo < 0 {
		t.Error("Rounds should not be negative after concurrent access")
	}
}