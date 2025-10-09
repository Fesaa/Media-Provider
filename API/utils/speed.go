package utils

import (
	"sync"
	"time"
)

// SpeedTracker tracks progress and speed for work items
type SpeedTracker struct {
	lock sync.RWMutex

	lastCheck time.Time
	startTime time.Time
	max       int
	cur       int

	// For tracking intermediate progress of current work item
	intermediate     *SpeedTracker
	intermediateLock sync.RWMutex
}

// NewSpeedTracker creates a new speed tracker with a maximum number of items
func NewSpeedTracker(max int) *SpeedTracker {
	now := time.Now()
	return &SpeedTracker{
		max:       max,
		lastCheck: now,
		startTime: now,
	}
}

// Increment records one instance of work is finished
func (s *SpeedTracker) Increment() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.cur++
	s.lastCheck = time.Now()
}

// IncrementIntermediate increments the intermediate tracker if it exists
func (s *SpeedTracker) IncrementIntermediate() {
	s.intermediateLock.Lock()
	defer s.intermediateLock.Unlock()

	if s.intermediate != nil {
		s.intermediate.Increment()
	}
}

// Progress returns the completion percentage (0-100)
// If an intermediate tracker exists, includes its fractional progress
func (s *SpeedTracker) Progress() float64 {
	s.lock.RLock()
	cur := s.cur
	s.lock.RUnlock()

	if s.max == 0 {
		return 0
	}

	progress := float64(cur)
	intermediateProgress := 0.0

	// Add fractional progress from intermediate tracker
	s.intermediateLock.RLock()
	if s.intermediate != nil {
		intermediateProgress = s.intermediate.Progress() / float64(s.max)
	}
	s.intermediateLock.RUnlock()

	return (progress/float64(s.max))*100 + intermediateProgress
}

// Speed returns items per second
func (s *SpeedTracker) Speed() float64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	elapsed := time.Since(s.startTime).Seconds()
	if elapsed == 0 {
		return 0
	}
	return float64(s.cur) / elapsed
}

// IntermediateSpeed returns the speed of the intermediate tracker if it exists
func (s *SpeedTracker) IntermediateSpeed() float64 {
	s.intermediateLock.RLock()
	defer s.intermediateLock.RUnlock()
	if s.intermediate != nil {
		return s.intermediate.Speed()
	}
	return 0
}

// SetIntermediate sets the intermediate progress tracker for the current work item
func (s *SpeedTracker) SetIntermediate(max int) {
	s.intermediateLock.Lock()
	defer s.intermediateLock.Unlock()

	s.intermediate = NewSpeedTracker(max)
}

// ClearIntermediate removes the intermediate tracker (call when work item completes)
func (s *SpeedTracker) ClearIntermediate() {
	s.intermediateLock.Lock()
	defer s.intermediateLock.Unlock()

	s.intermediate = nil
}

func (s *SpeedTracker) EstimatedTimeRemaining() float64 {
	if s.cur == 0 {
		return 0
	}

	return float64(s.max-s.cur) * s.Speed()
}
