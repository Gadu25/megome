package ai

import (
	"testing"
	"time"
)

func TestStatusTrackerAvailableInitially(t *testing.T) {
	status := NewStatusTracker(true, time.Minute)
	available, remaining := status.Status()
	if !available {
		t.Error("expected available=true initially")
	}
	if remaining != 0 {
		t.Errorf("expected remaining=0, got %d", remaining)
	}
}

func TestStatusTrackerUnavailableThenRecovers(t *testing.T) {
	status := NewStatusTracker(true, 50*time.Millisecond)
	status.MarkUnavailable()

	available, remaining := status.Status()
	if available {
		t.Error("expected available=false after MarkUnavailable")
	}
	if remaining <= 0 {
		t.Errorf("expected positive remaining, got %d", remaining)
	}

	time.Sleep(100 * time.Millisecond)
	available, remaining = status.Status()
	if !available {
		t.Error("expected available=true after cooldown elapses")
	}
	if remaining != 0 {
		t.Errorf("expected remaining=0 after recovery, got %d", remaining)
	}
}

func TestStatusTrackerDisabledAlwaysUnavailable(t *testing.T) {
	status := NewStatusTracker(false, time.Minute)
	available, remaining := status.Status()
	if available {
		t.Error("expected available=false when disabled")
	}
	if remaining != -1 {
		t.Errorf("expected remaining=-1 when disabled, got %d", remaining)
	}
}
