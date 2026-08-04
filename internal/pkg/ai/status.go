package ai

import (
	"math"
	"sync"
	"time"
)

type StatusTracker struct {
	mu               sync.Mutex
	enabled          bool
	unavailableUntil time.Time
	cooldown         time.Duration
}

func NewStatusTracker(enabled bool, cooldown time.Duration) *StatusTracker {
	return &StatusTracker{enabled: enabled, cooldown: cooldown}
}

func (s *StatusTracker) MarkUnavailable() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.enabled {
		s.unavailableUntil = time.Now().Add(s.cooldown)
	}
}

func (s *StatusTracker) Status() (available bool, cooldownRemainingSeconds int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.enabled {
		return false, -1
	}
	if s.unavailableUntil.IsZero() {
		return true, 0
	}
	remaining := time.Until(s.unavailableUntil)
	if remaining <= 0 {
		s.unavailableUntil = time.Time{}
		return true, 0
	}
	return false, int64(math.Ceil(remaining.Seconds()))
}

func (s *StatusTracker) Available() bool {
	available, _ := s.Status()
	return available
}
