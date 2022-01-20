package eventsv2

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvent(t *testing.T) {
	type TestResult struct {
		Success bool
	}

	var myEvent Event[TestResult]

	myEvent.Subscribe(func(ctx context.Context, event TestResult) error {
		assert.True(t, event.Success)
		return nil
	})

	myEvent.Dispatch(context.Background(), TestResult{Success: true})
}
