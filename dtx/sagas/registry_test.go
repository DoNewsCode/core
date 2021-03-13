package sagas

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
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
		Undo: func(ctx context.Context, req interface{}) (err error) {
			t.Fatal("should not be called")
			return nil
		},
	})
	reg.Recover(context.Background())
}

func TestRegistry_RecoverSerialized(t *testing.T) {
	store := NewInProcessStore()
	store.transactions["test"] = []Log{{
		ID:            "0",
		correlationID: "2",
		StartedAt:     time.Now().Add(-time.Hour),
		LogType:       Session,
	}, {
		ID:            "1",
		correlationID: "2",
		StartedAt:     time.Now().Add(-time.Hour),
		FinishedAt:    time.Time{},
		StepNumber:    0,
		LogType:       Do,
		StepError:     nil,
		StepName:      "foo",
		StepParam:     []byte(`foo`),
	}}
	reg := NewRegistry(store, WithLogger(log.NewNopLogger()))
	reg.AddStep(&Step{
		Name: "foo",
		Do: func(ctx context.Context, request interface{}) (response interface{}, err error) {
			t.Fatal("should not be called")
			return nil, nil
		},
		Undo: func(ctx context.Context, req interface{}) (err error) {
			assert.Equal(t, "FOO", req.(string))
			return nil
		},
		DecodeParam: func(bytes []byte) (interface{}, error) {
			return strings.ToUpper(string(bytes)), nil
		},
	})
	reg.Recover(context.Background())
}
