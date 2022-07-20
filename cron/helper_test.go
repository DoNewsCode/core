package cron

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetCurrentSchedule(t *testing.T) {
	assert.True(t, GetCurrentSchedule(context.Background()).IsZero())
}

func TestGetNextSchedule(t *testing.T) {
	assert.True(t, GetNextSchedule(context.Background()).IsZero())
}

func TestMockStartTimeFunc(t *testing.T) {
	fakeNow, _ := time.ParseInLocation("2006-01-02 15:04:05", "2029-01-01 00:00:00", time.Local)
	nowFunc := MockStartTimeFunc(fakeNow)
	now1 := nowFunc()
	now2 := time.Now()
	
	assert.True(t, now1.Equal(fakeNow))
	time.Sleep(time.Millisecond)
	assert.True(t, nowFunc().Sub(now1) == time.Since(now2))
}
