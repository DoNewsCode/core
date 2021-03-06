package sagas

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaga_success(t *testing.T) {
	var value int
	var mySaga = &Saga{
		Name: "test",
		steps: []*Step{
			{
				"one",
				func(ctx context.Context, correlationId string) error {
					value++
					return nil
				},
				func(ctx context.Context, correlationId string) error {
					value--
					return nil
				},
			},
			{
				"two",
				func(ctx context.Context, correlationId string) error {
					value++
					return nil
				},
				func(ctx context.Context, correlationId string) error {
					value--
					return nil
				},
			},
		},
	}
	var c = Coordinator{
		CorrelationId: "test",
		Saga:          mySaga,
		Store:         &InProcessStore{},
	}
	c.Execute(context.Background())
	assert.Equal(t, 2, value)
}

func TestSaga_failure(t *testing.T) {
	var value int
	var mySaga = &Saga{
		Name: "test",
		steps: []*Step{
			{
				"one",
				func(ctx context.Context, correlationId string) error {
					value++
					return nil
				},
				func(ctx context.Context, correlationId string) error {
					value--
					return nil
				},
			},
			{
				"two",
				func(ctx context.Context, correlationId string) error {
					value++
					return errors.New("")
				},
				func(ctx context.Context, correlationId string) error {
					value--
					return nil
				},
			},
		},
	}
	var c = Coordinator{
		CorrelationId: "test",
		Saga:          mySaga,
		Store:         &InProcessStore{},
	}
	result := c.Execute(context.Background())
	assert.Error(t, result.(*Result).DoErr)
	assert.Equal(t, 0, value)
}

func TestSaga_recovery(t *testing.T) {
	var attempt int
	var value int
	var store = &InProcessStore{}
	var mySaga = &Saga{
		Name: "test",
		steps: []*Step{
			{
				"one",
				func(ctx context.Context, correlationId string) error {
					value++
					return nil
				},
				func(ctx context.Context, correlationId string) error {
					value--
					return nil
				},
			},
			{
				"two",
				func(ctx context.Context, correlationId string) error {
					value++
					return errors.New("")
				},
				func(ctx context.Context, correlationId string) error {
					if attempt == 0 {
						attempt++
						return errors.New("")
					}
					value--
					return nil
				},
			},
		},
	}
	var c = Coordinator{
		CorrelationId: "test",
		Saga:          mySaga,
		Store:         store,
	}
	result := c.Execute(context.Background())
	assert.NotEmpty(t, result.(*Result).UndoErr)
	assert.Contains(t, result.Error(), "additional errors encountered while rolling back")
	assert.Equal(t, 1, value)

	var r = NewRegistry(store)
	r.Register(mySaga)
	r.Recover(context.Background())
	assert.Equal(t, 0, value)
}

func TestSaga_panic(t *testing.T) {
	var attempt int
	var value int
	var store = &InProcessStore{}
	var mySaga = &Saga{
		Name: "test",
		steps: []*Step{
			{
				"one",
				func(ctx context.Context, correlationId string) error {
					value++
					return nil
				},
				func(ctx context.Context, correlationId string) error {
					value--
					return nil
				},
			},
			{
				"two",
				func(ctx context.Context, correlationId string) error {
					value++
					return errors.New("")
				},
				func(ctx context.Context, correlationId string) error {
					if attempt == 0 {
						attempt++
						panic("err")
					}
					value--
					return nil
				},
			},
		},
	}
	var c = Coordinator{
		CorrelationId: "test",
		Saga:          mySaga,
		Store:         store,
	}
	var r = NewRegistry(store)
	r.Register(mySaga)

	defer func(r *Registry) {
		if rec := recover(); rec != nil {
			r.Recover(context.Background())
			assert.Equal(t, 0, value)
		}
	}(r)
	c.Execute(context.Background())
}

func TestSaga_shortCircuit(t *testing.T) {
	var value int
	var store = &InProcessStore{}
	var mySaga = &Saga{
		Name: "test",
		steps: []*Step{
			{
				"one",
				func(ctx context.Context, correlationId string) error {
					value++
					return errors.New("foo")
				},
				func(ctx context.Context, correlationId string) error {
					value--
					return nil
				},
			},
			{
				"two",
				func(ctx context.Context, correlationId string) error {
					panic("should not reach")
				},
				func(ctx context.Context, correlationId string) error {
					panic("should not reach")
				},
			},
		},
	}
	var c = Coordinator{
		CorrelationId: "test",
		Saga:          mySaga,
		Store:         store,
	}
	result := c.Execute(context.Background())
	assert.Error(t, result.(*Result).DoErr)
	assert.Equal(t, 0, value)
}

func TestSaga_emptyRecover(t *testing.T) {
	var value int
	var attempt int
	var store = &InProcessStore{}
	var mySaga = &Saga{
		Name: "test",
		steps: []*Step{
			{
				"one",
				func(ctx context.Context, correlationId string) error {
					value++
					return errors.New("foo")
				},
				func(ctx context.Context, correlationId string) error {
					if attempt == 0 {
						attempt++
						value--
						return nil
					}
					panic("err")
				},
			},
		},
	}
	var c = Coordinator{
		CorrelationId: "test",
		Saga:          mySaga,
		Store:         store,
	}
	result := c.Execute(context.Background())
	assert.Error(t, result.(*Result).DoErr)
	assert.Equal(t, 0, value)

	var r = NewRegistry(store)
	r.Register(mySaga)
	r.Recover(context.Background())
	assert.Equal(t, 0, value)
}
