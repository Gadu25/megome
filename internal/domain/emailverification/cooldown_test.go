package emailverification

import (
	"testing"
	"time"
)

func TestRemainingCooldown(t *testing.T) {
	cooldown := time.Minute
	now := time.Now()

	if got := RemainingCooldown(time.Time{}, now, cooldown); got != 0 {
		t.Errorf("zero lastSentAt: expected 0, got %d", got)
	}
	if got := RemainingCooldown(now.Add(-time.Minute), now, cooldown); got != 0 {
		t.Errorf("cooldown elapsed exactly: expected 0, got %d", got)
	}
	if got := RemainingCooldown(now.Add(-61*time.Second), now, cooldown); got != 0 {
		t.Errorf("cooldown exceeded: expected 0, got %d", got)
	}
	if got := RemainingCooldown(now.Add(-30*time.Second), now, cooldown); got != 30 {
		t.Errorf("30s elapsed: expected 30, got %d", got)
	}
	if got := RemainingCooldown(now.Add(-10*time.Millisecond), now, cooldown); got != 60 {
		t.Errorf("10ms elapsed: expected 60 (ceil), got %d", got)
	}
}
