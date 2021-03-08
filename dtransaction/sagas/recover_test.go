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
		CorrelationID: "2",
		SagaName:      "test",
		StartedAt:     time.Now(),
		LogType:       Session,
		StepNumber:    0,
	}, {
		ID:            "1",
		CorrelationID: "2",
		SagaName:      "test",
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
	saga := Saga{
		Name:    "test",
		Timeout: time.Hour,
		Steps: []*Step{
			{
				Name: "foo",
				Do: func(ctx context.Context, request interface{}) (response interface{}, err error) {
					t.Fatal("should not be called")
					return nil, nil
				},
				Undo: func(ctx context.Context) error {
					t.Fatal("should not be called")
					return nil
				},
			},
		},
	}
	store := NewInProcessStore()
	store.transactions["test"] = []Log{{
		ID:            "0",
		CorrelationID: "2",
		SagaName:      "test",
		StartedAt:     time.Now(),
		LogType:       Session,
	}, {
		ID:            "1",
		CorrelationID: "2",
		SagaName:      "test",
		StartedAt:     time.Now(),
		FinishedAt:    time.Time{},
		StepNumber:    0,
		LogType:       Do,
		StepError:     nil,
		StepName:      "foo",
	}}
	reg := NewRegistry(store, WithLogger(log.NewNopLogger()))
	reg.Register(&saga)
	reg.Recover(context.Background())
}
