package cron

import (
	"context"
	"time"
)

var (
	prevContextKey = struct{}{}
	nextContextKey = struct{}{}
)

// GetCurrentSchedule returns the current schedule for the given context.
func GetCurrentSchedule(ctx context.Context) time.Time {
	if ctx == nil {
		return time.Time{}
	}
	if t, ok := ctx.Value(prevContextKey).(time.Time); ok {
		return t
	}
	return time.Time{}
}

// GetNextSchedule returns the next schedule for the given context.
func GetNextSchedule(ctx context.Context) time.Time {
	if ctx == nil {
		return time.Time{}
	}
	if t, ok := ctx.Value(nextContextKey).(time.Time); ok {
		return t
	}
	return time.Time{}
}

// MockStartTime allows the user to mock the current time at the beginning of the cron job.
// This is useful for testing.
func MockStartTime(t time.Time) func() time.Time {
	diff := t.Sub(time.Now())
	return func() time.Time {
		return time.Now().Add(diff)
	}
}
