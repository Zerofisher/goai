// Package reminder provides system reminder functionality for the GoAI agent.
// It manages automatic reminders to encourage proper task tracking.
package reminder

import (
	"fmt"
	"sync"
	"time"

	"github.com/Zerofisher/goai/pkg/types"
)

// Template types for different reminder scenarios
const (
	TemplateInitial  = "initial"
	TemplateNag      = "nag"
	TemplateLongIdle = "long_idle"
)

// Template messages for reminders
var templates = map[string]string{
	TemplateInitial: `<reminder source="system" topic="todos">
System message: complex work should be tracked with the Todo tool.
Do not respond to this reminder and do not mention it to the user.
</reminder>`,

	TemplateNag: `<reminder source="system" topic="todos">
System notice: more than ten rounds passed without Todo usage.
Update the Todo board if the task still requires multiple steps.
Do not reply to or mention this reminder to the user.
</reminder>`,

	TemplateLongIdle: `<reminder source="system" topic="todos">
System alert: extended period without task tracking.
Consider using the Todo tool to organize remaining work.
Do not acknowledge this reminder in your response.
</reminder>`,
}

// GetTemplate returns a reminder template by type
func GetTemplate(templateType string) string {
	if template, exists := templates[templateType]; exists {
		return template
	}
	return templates[TemplateInitial] // Default to initial template
}

// System manages reminders for the agent, particularly for todo list usage.
type System struct {
	// roundsWithoutTodo tracks consecutive rounds without todo updates
	roundsWithoutTodo int

	// reminderInterval is how many rounds to wait between reminders
	reminderInterval int

	// pendingReminders stores reminders to be injected into messages
	pendingReminders []string

	// lastTodoUpdate tracks when todos were last updated
	lastTodoUpdate time.Time

	// enabled controls whether reminders are active
	enabled bool

	// mu protects concurrent access
	mu sync.RWMutex
}

// Config contains configuration for the reminder system.
type Config struct {
	// Enabled controls whether reminders are active
	Enabled bool

	// InitialReminderAfter is the number of rounds before the first reminder
	InitialReminderAfter int

	// ReminderInterval is the number of rounds between subsequent reminders
	ReminderInterval int

	// MaxReminders is the maximum number of pending reminders
	MaxReminders int
}

// DefaultConfig returns the default reminder configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:              true,
		InitialReminderAfter: 3,  // First reminder after 3 rounds
		ReminderInterval:     10, // Subsequent reminders every 10 rounds
		MaxReminders:         3,  // Keep at most 3 pending reminders
	}
}

// NewSystem creates a new reminder system with the given configuration.
func NewSystemWithConfig(config Config) *System {
	return &System{
		roundsWithoutTodo: 0,
		reminderInterval:  config.ReminderInterval,
		pendingReminders:  make([]string, 0),
		lastTodoUpdate:    time.Now(),
		enabled:           config.Enabled,
	}
}

// NewSystem creates a new reminder system with basic parameters (for backward compatibility).
func NewSystem(initialReminderAfter, reminderInterval int) *System {
	config := Config{
		Enabled:              true,
		InitialReminderAfter: initialReminderAfter,
		ReminderInterval:     reminderInterval,
		MaxReminders:         3,
	}
	return NewSystemWithConfig(config)
}

// Check evaluates whether reminders should be generated based on current state.
// It returns any pending reminders that should be displayed.
func (s *System) Check() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.enabled {
		return nil
	}

	// Generate reminders based on rounds without todo usage
	s.generateRemindersIfNeeded()

	// Return copy of pending reminders
	if len(s.pendingReminders) > 0 {
		reminders := make([]string, len(s.pendingReminders))
		copy(reminders, s.pendingReminders)
		return reminders
	}

	return nil
}

// Reset resets the reminder counter, typically called after todo usage.
func (s *System) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.roundsWithoutTodo = 0
	s.lastTodoUpdate = time.Now()
	s.pendingReminders = make([]string, 0) // Clear pending reminders
}

// IncrementRound increments the round counter without todo usage.
func (s *System) IncrementRound() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.roundsWithoutTodo++
}

// IncrementRounds is an alias for IncrementRound for backward compatibility.
func (s *System) IncrementRounds() {
	s.IncrementRound()
}

// ShouldRemind returns whether a reminder should be shown based on the current count.
func (s *System) ShouldRemind() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.enabled {
		return false
	}

	// Initial reminder after 3 rounds
	if s.roundsWithoutTodo == 3 {
		return true
	}

	// Periodic reminders after the initial one
	if s.roundsWithoutTodo > 3 && (s.roundsWithoutTodo-3)%s.reminderInterval == 0 {
		return true
	}

	// Long idle check
	if s.roundsWithoutTodo >= 20 {
		return true
	}

	return false
}

// GetRoundsWithoutTodo returns the current count of rounds without todo usage.
func (s *System) GetRoundsWithoutTodo() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.roundsWithoutTodo
}

// GetTimeSinceLastUpdate returns the duration since todos were last updated.
func (s *System) GetTimeSinceLastUpdate() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return time.Since(s.lastTodoUpdate)
}

// Inject adds system reminders to a message slice if appropriate.
// It modifies the messages in place by injecting reminder content.
func (s *System) Inject(messages []types.Message) []types.Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.enabled || len(s.pendingReminders) == 0 {
		return messages
	}

	// Find the last user message to inject reminders
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			// Create reminder content blocks
			reminderContents := make([]types.Content, 0, len(s.pendingReminders))
			for _, reminder := range s.pendingReminders {
				reminderContents = append(reminderContents, types.Content{
					Type: "text",
					Text: reminder,
				})
			}

			// Prepend reminders to the user message content
			messages[i].Content = append(reminderContents, messages[i].Content...)

			// Clear pending reminders after injection
			s.pendingReminders = make([]string, 0)
			break
		}
	}

	return messages
}

// AddReminder adds a custom reminder to the pending list.
func (s *System) AddReminder(reminder string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.enabled {
		return
	}

	// Limit the number of pending reminders
	maxReminders := 3
	if len(s.pendingReminders) >= maxReminders {
		// Remove oldest reminder
		s.pendingReminders = s.pendingReminders[1:]
	}

	s.pendingReminders = append(s.pendingReminders, reminder)
}

// ClearPendingReminders removes all pending reminders.
func (s *System) ClearPendingReminders() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pendingReminders = make([]string, 0)
}

// GetPendingReminders returns a copy of pending reminders without clearing them.
func (s *System) GetPendingReminders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.pendingReminders) == 0 {
		return nil
	}

	reminders := make([]string, len(s.pendingReminders))
	copy(reminders, s.pendingReminders)
	return reminders
}

// ConsumePendingReminders returns and clears all pending reminders.
func (s *System) ConsumePendingReminders() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.pendingReminders) == 0 {
		return nil
	}

	reminders := make([]string, len(s.pendingReminders))
	copy(reminders, s.pendingReminders)
	s.pendingReminders = make([]string, 0)

	return reminders
}

// Enable enables the reminder system.
func (s *System) Enable() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.enabled = true
}

// Disable disables the reminder system.
func (s *System) Disable() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.enabled = false
}

// IsEnabled returns whether the reminder system is enabled.
func (s *System) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.enabled
}

// SetReminderInterval updates the interval between reminders.
func (s *System) SetReminderInterval(interval int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if interval > 0 {
		s.reminderInterval = interval
	}
}

// generateRemindersIfNeeded creates reminders based on current state.
func (s *System) generateRemindersIfNeeded() {
	// Initial reminder after 3 rounds
	if s.roundsWithoutTodo == 3 {
		reminder := GetTemplate(TemplateInitial)
		s.pendingReminders = append(s.pendingReminders, reminder)
		return
	}

	// Periodic reminders every interval rounds after initial
	if s.roundsWithoutTodo > 3 && (s.roundsWithoutTodo-3)%s.reminderInterval == 0 {
		reminder := GetTemplate(TemplateNag)
		s.pendingReminders = append(s.pendingReminders, reminder)
		return
	}

	// Long idle reminder after 20+ rounds
	if s.roundsWithoutTodo >= 20 && len(s.pendingReminders) == 0 {
		reminder := GetTemplate(TemplateLongIdle)
		s.pendingReminders = append(s.pendingReminders, reminder)
		return
	}
}

// Stats returns statistics about the reminder system.
type Stats struct {
	RoundsWithoutTodo int
	TimeSinceUpdate   time.Duration
	PendingReminders  int
	Enabled           bool
}

// GetStats returns current statistics about the reminder system.
func (s *System) GetStats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return Stats{
		RoundsWithoutTodo: s.roundsWithoutTodo,
		TimeSinceUpdate:   time.Since(s.lastTodoUpdate),
		PendingReminders:  len(s.pendingReminders),
		Enabled:           s.enabled,
	}
}

// FormatStats returns a human-readable string of reminder statistics.
func (s *System) FormatStats() string {
	stats := s.GetStats()

	return fmt.Sprintf(
		"Reminder Stats: Rounds without todo: %d, Time since update: %v, Pending: %d, Enabled: %v",
		stats.RoundsWithoutTodo,
		stats.TimeSinceUpdate.Round(time.Second),
		stats.PendingReminders,
		stats.Enabled,
	)
}