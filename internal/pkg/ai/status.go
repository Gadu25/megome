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

func (s *StatusTracker) MarkUnavailable(customCooldown time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.enabled {
		cooldown := customCooldown
		if cooldown <= 0 {
			cooldown = s.cooldown
		}
		s.unavailableUntil = time.Now().Add(cooldown)
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
