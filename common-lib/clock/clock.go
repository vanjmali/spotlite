package clock

import "time"

// Clock provides current time.
// This allows for easier testing by injecting a mock clock.
type Clock interface {
	Now() time.Time
}

// RealClock uses the system time.
type RealClock struct{}

// Now returns the current system time.
func (RealClock) Now() time.Time { return time.Now() }

// FixedClock always returns a fixed time.
type FixedClock struct {
	fixed time.Time
}

// NewFixedClock creates a FixedClock set to the specified initial time.
func NewFixedClock(initialTime time.Time) *FixedClock {
	return &FixedClock{fixed: initialTime}
}

// Now returns the fixed time.
func (fc *FixedClock) Now() time.Time {
	return fc.fixed
}
