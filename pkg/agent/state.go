package agent

import (
	"sync"
	"time"
)

// State manages the agent's runtime state
type State struct {
	// Session information
	sessionID    string
	startTime    time.Time
	lastActivity time.Time

	// Counters
	roundCount    int
	toolCallCount int
	errorCount    int
	toolCalls     map[string]int

	// Status flags
	isProcessing bool
	hasErrors    bool

	// Error tracking
	lastError  error
	errorLog   []ErrorEntry

	// Recovery points
	recoveryPoints []RecoveryPoint

	// Mutex for thread safety
	mu sync.RWMutex
}

// ErrorEntry represents an error that occurred
type ErrorEntry struct {
	Time    time.Time
	Error   error
	Context string
}

// RecoveryPoint represents a point where the agent can recover from
type RecoveryPoint struct {
	ID        string
	Time      time.Time
	Round     int
	StateData map[string]interface{}
}

// NewState creates a new state manager
func NewState() *State {
	return &State{
		sessionID:      generateSessionID(),
		startTime:      time.Now(),
		lastActivity:   time.Now(),
		toolCalls:      make(map[string]int),
		errorLog:       []ErrorEntry{},
		recoveryPoints: []RecoveryPoint{},
	}
}

// GetSessionID returns the current session ID
func (s *State) GetSessionID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessionID
}

// GetUptime returns how long the agent has been running
func (s *State) GetUptime() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Since(s.startTime)
}

// GetLastActivity returns the time of last activity
func (s *State) GetLastActivity() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastActivity
}

// GetIdleTime returns how long since last activity
func (s *State) GetIdleTime() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Since(s.lastActivity)
}

// IncrementRound increments the round counter
func (s *State) IncrementRound() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.roundCount++
	s.lastActivity = time.Now()
}

// GetRoundCount returns the current round count
func (s *State) GetRoundCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.roundCount
}

// RecordToolCall records a tool call
func (s *State) RecordToolCall(toolName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.toolCallCount++
	s.toolCalls[toolName]++
	s.lastActivity = time.Now()
}

// GetToolCallCount returns the total tool call count
func (s *State) GetToolCallCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.toolCallCount
}

// GetToolCallStats returns statistics for each tool
func (s *State) GetToolCallStats() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Create a copy to avoid race conditions
	stats := make(map[string]int)
	for tool, count := range s.toolCalls {
		stats[tool] = count
	}
	return stats
}

// RecordError records an error
func (s *State) RecordError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.errorCount++
	s.lastError = err
	s.hasErrors = true

	entry := ErrorEntry{
		Time:    time.Now(),
		Error:   err,
		Context: "Round " + string(rune(s.roundCount)),
	}

	s.errorLog = append(s.errorLog, entry)

	// Keep only last 100 errors
	if len(s.errorLog) > 100 {
		s.errorLog = s.errorLog[len(s.errorLog)-100:]
	}
}

// GetErrorCount returns the error count
func (s *State) GetErrorCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.errorCount
}

// GetLastError returns the last error
func (s *State) GetLastError() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastError
}

// GetErrorLog returns recent errors
func (s *State) GetErrorLog() []ErrorEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy
	log := make([]ErrorEntry, len(s.errorLog))
	copy(log, s.errorLog)
	return log
}

// HasErrors returns whether any errors have occurred
func (s *State) HasErrors() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hasErrors
}

// SetProcessing sets the processing state
func (s *State) SetProcessing(processing bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isProcessing = processing
	if processing {
		s.lastActivity = time.Now()
	}
}

// IsProcessing returns whether the agent is processing
func (s *State) IsProcessing() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isProcessing
}

// CreateRecoveryPoint creates a recovery point
func (s *State) CreateRecoveryPoint(data map[string]interface{}) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	point := RecoveryPoint{
		ID:        generateRecoveryID(),
		Time:      time.Now(),
		Round:     s.roundCount,
		StateData: data,
	}

	s.recoveryPoints = append(s.recoveryPoints, point)

	// Keep only last 10 recovery points
	if len(s.recoveryPoints) > 10 {
		s.recoveryPoints = s.recoveryPoints[len(s.recoveryPoints)-10:]
	}

	return point.ID
}

// GetRecoveryPoint gets a recovery point by ID
func (s *State) GetRecoveryPoint(id string) *RecoveryPoint {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, point := range s.recoveryPoints {
		if point.ID == id {
			return &point
		}
	}
	return nil
}

// GetLatestRecoveryPoint gets the most recent recovery point
func (s *State) GetLatestRecoveryPoint() *RecoveryPoint {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.recoveryPoints) == 0 {
		return nil
	}

	point := s.recoveryPoints[len(s.recoveryPoints)-1]
	return &point
}

// Reset resets the state
func (s *State) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.roundCount = 0
	s.toolCallCount = 0
	s.errorCount = 0
	s.toolCalls = make(map[string]int)
	s.isProcessing = false
	s.hasErrors = false
	s.lastError = nil
	s.errorLog = []ErrorEntry{}
	s.lastActivity = time.Now()
	// Keep recovery points and session ID
}

// GetStateSummary returns a summary of the current state
func (s *State) GetStateSummary() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"session_id":      s.sessionID,
		"uptime":          s.GetUptime().String(),
		"idle_time":       time.Since(s.lastActivity).String(),
		"round_count":     s.roundCount,
		"tool_call_count": s.toolCallCount,
		"error_count":     s.errorCount,
		"has_errors":      s.hasErrors,
		"is_processing":   s.isProcessing,
		"recovery_points": len(s.recoveryPoints),
	}
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return time.Now().Format("20060102-150405")
}

// generateRecoveryID generates a unique recovery point ID
func generateRecoveryID() string {
	return time.Now().Format("150405.000")
}