package cron

import (
	"context"
	"time"
)

var (
	prevContextKey = struct{}{}
	nextContextKey = struct{}{}
)

func GetCurrentSchedule(ctx context.Context) time.Time {
	if ctx == nil {
		return time.Time{}
	}
	if t, ok := ctx.Value(prevContextKey).(time.Time); ok {
		return t
	}
	return time.Time{}
}

func GetNextSchedule(ctx context.Context) time.Time {
	if ctx == nil {
		return time.Time{}
	}
	if t, ok := ctx.Value(nextContextKey).(time.Time); ok {
		return t
	}
	return time.Time{}
}
