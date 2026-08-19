package emailverification

import (
	"math"
	"time"
)

// RemainingCooldown returns the number of whole seconds left before the resend
// cooldown expires, rounded up, or 0 if the cooldown has already elapsed.
func RemainingCooldown(lastSentAt, now time.Time, cooldown time.Duration) int64 {
	if lastSentAt.IsZero() {
		return 0
	}
	remaining := cooldown - now.Sub(lastSentAt)
	if remaining <= 0 {
		return 0
	}
	return int64(math.Ceil(remaining.Seconds()))
}
