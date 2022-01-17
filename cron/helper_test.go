package cron

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetCurrentSchedule(t *testing.T) {
	assert.True(t, GetCurrentSchedule(context.Background()).IsZero())
}

func TestGetNextSchedule(t *testing.T) {
	assert.True(t, GetNextSchedule(context.Background()).IsZero())
}
