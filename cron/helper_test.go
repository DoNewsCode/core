package cron

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCurrentSchedule(t *testing.T) {
	assert.True(t, GetCurrentSchedule(context.Background()).IsZero())
}

func TestGetNextSchedule(t *testing.T) {
	assert.True(t, GetNextSchedule(context.Background()).IsZero())
}
