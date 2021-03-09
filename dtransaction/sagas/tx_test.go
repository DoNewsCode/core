package sagas

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaga_success(t *testing.T) {
	var value int
	store := NewInProcessStore()
	r := NewRegistry(store)

	ep1 := r.AddStep(&Step{
		"one",
		func(ctx context.Context, req interface{}) (interface{}, error) {
			value++
			return nil, nil
		},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			value--
			return nil, nil
		},
	})
	ep2 := r.AddStep(&Step{
		"two",
		func(ctx context.Context, req interface{}) (interface{}, error) {
			value++
			return nil, nil
		},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			value--
			return nil, nil
		},
	})

	var c, ctx = r.StartTx(context.Background())
	ep1(ctx, nil)
	ep2(ctx, nil)
	c.Commit(ctx)
	assert.Equal(t, 2, value)
}

func TestSaga_failure(t *testing.T) {
	var value int
	store := NewInProcessStore()
	r := NewRegistry(store)

	ep1 := r.AddStep(&Step{
		"one",
		func(ctx context.Context, req interface{}) (interface{}, error) {
			value++
			return nil, nil
		},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			value--
			return nil, nil
		},
	})

	ep2 := r.AddStep(&Step{
		"two",
		func(ctx context.Context, req interface{}) (interface{}, error) {
			value++
			return nil, errors.New("")
		},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			value--
			return nil, nil
		},
	})

	var c, ctx = r.StartTx(context.Background())
	ep1(ctx, nil)
	ep2(ctx, nil)
	c.Rollback(ctx)
	assert.Equal(t, 0, value)
}

func TestSaga_recovery(t *testing.T) {
	var attempt int
	var value int
	var store = &InProcessStore{}
	var r = NewRegistry(store, WithTimeout(0))

	ep1 := r.AddStep(&Step{
		"one",
		func(ctx context.Context, req interface{}) (interface{}, error) {
			value++
			return nil, nil
		},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			value--
			return nil, nil
		},
	})

	ep2 := r.AddStep(&Step{

		"two",
		func(ctx context.Context, req interface{}) (interface{}, error) {
			value++
			return nil, errors.New("")
		},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			if attempt == 0 {
				attempt++
				return nil, errors.New("")
			}
			value--
			return nil, nil

		},
	})

	var c, ctx = r.StartTx(context.Background())
	ep1(ctx, nil)
	ep2(ctx, nil)
	err := c.Rollback(ctx)
	assert.NotNil(t, err)
	assert.NotEmpty(t, err.(*Result).UndoErr)
	assert.Contains(t, err.Error(), "errors encountered while rolling back")
	assert.Equal(t, 1, value)

	r.Recover(ctx)
	assert.Equal(t, 0, value)
}

func TestSaga_panic(t *testing.T) {
	var attempt int
	var value int
	var store = &InProcessStore{}
	var r = NewRegistry(store, WithTimeout(0))

	ep1 := r.AddStep(&Step{
		"one",
		func(ctx context.Context, req interface{}) (interface{}, error) {
			value++
			return nil, nil
		},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			value--
			return nil, nil
		},
	})
	ep2 := r.AddStep(&Step{
		"two",
		func(ctx context.Context, req interface{}) (interface{}, error) {
			value++
			return nil, errors.New("")
		},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			if attempt == 0 {
				attempt++
				panic("err")
			}
			value--
			return nil, nil
		},
	})

	defer func(r *Registry) {
		if rec := recover(); rec != nil {
			r.Recover(context.Background())
			assert.Equal(t, 0, value)
		}
	}(r)

	var _, ctx = r.StartTx(context.Background())
	ep1(ctx, nil)
	ep2(ctx, nil)
}

func TestSaga_shortCircuit(t *testing.T) {
	var value int
	var store = &InProcessStore{}
	var r = NewRegistry(store, WithTimeout(0))

	ep1 := r.AddStep(&Step{
		"one",
		func(ctx context.Context, req interface{}) (interface{}, error) {
			value++
			return nil, nil
		},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			value--
			return nil, nil
		},
	})

	r.AddStep(&Step{
		"two",
		func(ctx context.Context, req interface{}) (interface{}, error) {
			panic("should not reach")
		},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			panic("should not reach")
		},
	})

	var c, ctx = r.StartTx(context.Background())
	ep1(ctx, nil)
	c.Commit(ctx)
	assert.Equal(t, 1, value)
}

func TestSaga_emptyRecover(t *testing.T) {
	var value int
	var attempt int
	var store = &InProcessStore{}
	var r = NewRegistry(store, WithTimeout(0))

	ep := r.AddStep(&Step{
		"two",
		func(ctx context.Context, req interface{}) (interface{}, error) {
			value++
			return nil, errors.New("foo")
		},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			if attempt == 0 {
				attempt++
				value--
				return nil, nil
			}
			panic("err")
		},
	})
	tx, ctx := r.StartTx(context.Background())
	tx.Commit(ctx)
	r.Recover(context.Background())
	assert.Equal(t, 0, value)
	defer func() {
		assert.NotNil(t, recover())
	}()

	ep(ctx, nil)

}
