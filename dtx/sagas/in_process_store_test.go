package sagas

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DoNewsCode/core/dtx"
	"github.com/stretchr/testify/assert"
)

func TestInProcessStore_Ack(t *testing.T) {
	cases := []struct {
		name    string
		log     Log
		err     error
		asserts func(t *testing.T, log Log, s *InProcessStore)
	}{
		{
			"session without error",
			Log{
				ID:            "1",
				CorrelationID: "2",
				LogType:       Session,
				StartedAt:     time.Now(),
			},
			nil,
			func(t *testing.T, log Log, s *InProcessStore) {
				assert.Len(t, s.transactions, 0)
			},
		},
		{
			"session with error",
			Log{
				ID:            "1",
				CorrelationID: "2",
				LogType:       Session,
				StartedAt:     time.Now(),
			},
			errors.New("foo"),
			func(t *testing.T, log Log, s *InProcessStore) {
				assert.Len(t, s.transactions, 1)
				assert.Error(t, s.transactions[log.CorrelationID][0].StepError)
			},
		},
		{
			"do without error",
			Log{
				ID:            "1",
				CorrelationID: "2",
				LogType:       Do,
				StartedAt:     time.Now(),
			},
			nil,
			func(t *testing.T, log Log, s *InProcessStore) {
				assert.Len(t, s.transactions, 1)
				assert.False(t, s.transactions[log.CorrelationID][0].FinishedAt.IsZero())
				assert.NoError(t, s.transactions[log.CorrelationID][0].StepError)
			},
		},
		{
			"do with error",
			Log{
				ID:            "1",
				CorrelationID: "2",
				LogType:       Do,
				StartedAt:     time.Now(),
			},
			errors.New("foo"),
			func(t *testing.T, log Log, s *InProcessStore) {
				assert.Len(t, s.transactions, 1)
				assert.False(t, s.transactions[log.CorrelationID][0].FinishedAt.IsZero())
				assert.Error(t, s.transactions[log.CorrelationID][0].StepError)
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			store := NewInProcessStore()
			ctx := context.WithValue(context.Background(), dtx.CorrelationID, c.log.CorrelationID)
			store.Log(ctx, c.log)
			store.Ack(ctx, c.log.ID, c.err)
			c.asserts(t, c.log, store)
		})
	}
}

func TestInProcessStore_UncommittedSteps(t *testing.T) {
	store := NewInProcessStore()
	ctx := context.WithValue(context.Background(), dtx.CorrelationID, "2")
	store.Log(ctx, Log{
		ID:            "1",
		CorrelationID: "2",
		StartedAt:     time.Now(),
		LogType:       Session,
	})
	store.Log(ctx, Log{
		ID:            "2",
		CorrelationID: "2",
		StartedAt:     time.Now(),
		LogType:       Do,
	})
	logs, err := store.UnacknowledgedSteps(context.Background(), "2")
	assert.NoError(t, err)
	assert.Len(t, logs, 1)

	store.Log(ctx, Log{
		ID:            "2",
		CorrelationID: "2",
		StartedAt:     time.Now(),
		LogType:       Undo,
	})

	logs, err = store.UnacknowledgedSteps(context.Background(), "2")
	assert.NoError(t, err)
	assert.Len(t, logs, 1)

	store.Ack(ctx, "2", nil)
	logs, err = store.UnacknowledgedSteps(context.Background(), "2")
	assert.NoError(t, err)
	assert.Len(t, logs, 0)
}

func TestInProcessStore_UncommittedSagas(t *testing.T) {
	store := NewInProcessStore()
	store.transactions["test"] = []Log{{
		ID:            "1",
		CorrelationID: "test",
		FinishedAt:    time.Now(),
		LogType:       Session,
		StepError:     nil,
	}}
	logs, err := store.UncommittedSagas(context.Background())
	assert.NoError(t, err)
	assert.Len(t, logs, 0)
}
