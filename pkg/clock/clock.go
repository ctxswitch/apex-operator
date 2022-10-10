package clock

import "time"

type Clock interface {
	Now() time.Time
	Since(t time.Time) time.Duration
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}

func (RealClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}

type MockClock struct {
	wall string
}

func NewMockClock(ts string) *MockClock {
	return &MockClock{
		wall: ts,
	}
}

func (m *MockClock) Now() time.Time {
	now, _ := time.Parse(time.RFC3339Nano, m.wall)
	return now
}

func (m *MockClock) Since(t time.Time) time.Duration {
	now, _ := time.Parse(time.RFC3339Nano, m.wall)
	return now.Sub(t)
}

func (m *MockClock) Set(ts string) {
	m.wall = ts
}
