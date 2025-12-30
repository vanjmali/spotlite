package clock

import "time"

// Clock provides current time.
type Clock interface {
	Now() time.Time
}

// RealClock uses the system time.
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

type FixedClock struct {
	fixed time.Time
}

func NewFixedClock(initialTime time.Time) *FixedClock {
	return &FixedClock{fixed: initialTime}
}

func (fc *FixedClock) Now() time.Time {
	return fc.fixed
}
