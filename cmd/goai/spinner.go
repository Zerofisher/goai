package main

import (
	"fmt"
	"sync"
	"time"
)

// Color constants (matching those in interactive.go)
const (
	spinnerResetColor   = "\033[0m"
	spinnerPrimaryColor = "\033[36m"  // Cyan
)

// Spinner provides a loading animation for long-running operations.
type Spinner struct {
	frames  []string
	message string
	active  bool
	mu      sync.Mutex
	stop    chan bool
}

// NewSpinner creates a new spinner instance.
func NewSpinner() *Spinner {
	return &Spinner{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		stop:   make(chan bool, 1),
	}
}

// Start begins the spinner animation with the given message.
func (s *Spinner) Start(message string) {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.message = message
	s.mu.Unlock()

	go s.animate()
}

// Stop stops the spinner animation.
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return
	}

	s.active = false
	s.stop <- true

	// Clear the spinner line
	fmt.Printf("\r\033[K")
}

// animate runs the spinner animation loop.
func (s *Spinner) animate() {
	frameIndex := 0
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.mu.Lock()
			if !s.active {
				s.mu.Unlock()
				return
			}
			frame := s.frames[frameIndex]
			msg := s.message
			s.mu.Unlock()

			// Print spinner frame
			fmt.Printf("\r%s%s %s...%s", spinnerPrimaryColor, frame, msg, spinnerResetColor)

			frameIndex = (frameIndex + 1) % len(s.frames)
		}
	}
}

// UpdateMessage updates the spinner message while it's running.
func (s *Spinner) UpdateMessage(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.message = message
}

// IsActive returns whether the spinner is currently running.
func (s *Spinner) IsActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

// WithSpinner runs a function with a spinner animation.
func WithSpinner(message string, fn func() error) error {
	spinner := NewSpinner()
	spinner.Start(message)
	defer spinner.Stop()
	return fn()
}