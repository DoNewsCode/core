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
	assert.True(t, nowFunc().Equal(fakeNow))
	time.Sleep(time.Second)
	assert.True(t, nowFunc().Second() == fakeNow.Add(time.Second).Second())
}
