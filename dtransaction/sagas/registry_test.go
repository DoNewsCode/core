package sagas

import (
	"context"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
)

func TestRegistry_Recover(t *testing.T) {
	store := NewInProcessStore()
	store.transactions["test"] = []Log{{
		ID:            "0",
		correlationID: "2",
		StartedAt:     time.Now(),
		LogType:       Session,
		StepNumber:    0,
	}, {
		ID:            "1",
		correlationID: "2",
		StartedAt:     time.Now(),
		FinishedAt:    time.Time{},
		StepNumber:    1,
		LogType:       Do,
		StepError:     nil,
	}}
	reg := NewRegistry(store, WithLogger(log.NewNopLogger()))
	reg.Recover(context.Background())
}

func TestRegistry_RecoverWithTimeout(t *testing.T) {
	store := NewInProcessStore()
	store.transactions["test"] = []Log{{
		ID:            "0",
		correlationID: "2",
		StartedAt:     time.Now(),
		LogType:       Session,
	}, {
		ID:            "1",
		correlationID: "2",
		StartedAt:     time.Now(),
		FinishedAt:    time.Time{},
		StepNumber:    0,
		LogType:       Do,
		StepError:     nil,
		StepName:      "foo",
	}}
	reg := NewRegistry(store, WithLogger(log.NewNopLogger()))
	reg.AddStep(&Step{
		Name: "foo",
		Do: func(ctx context.Context, request interface{}) (response interface{}, err error) {
			t.Fatal("should not be called")
			return nil, nil
		},
		Undo: func(ctx context.Context, req interface{}) (response interface{}, err error) {
			t.Fatal("should not be called")
			return nil, nil
		},
	})
	reg.Recover(context.Background())
}
