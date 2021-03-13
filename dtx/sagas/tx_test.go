package sagas

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
)

func TestSaga_success(t *testing.T) {
	var value int
	store := NewInProcessStore()
	r := NewRegistry(store)

	ep1 := r.AddStep(&Step{
		Name: "one",
		Do: func(ctx context.Context, req interface{}) (interface{}, error) {
			value++
			return nil, nil
		},
		Undo: func(ctx context.Context, req interface{}) error {
			value--
			return nil
		},
	})
	ep2 := r.AddStep(&Step{
		Name: "two",
		Do: func(ctx context.Context, req interface{}) (interface{}, error) {
			value++
			return nil, nil
		},
		Undo: func(ctx context.Context, req interface{}) error {
			value--
			return nil
		},
	})

	var c, ctx = r.StartTX(context.Background())
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
		Name: "one",
		Do: func(ctx context.Context, req interface{}) (interface{}, error) {
			value++
			return nil, nil
		},
		Undo: func(ctx context.Context, req interface{}) error {
			value--
			return nil
		},
	})

	ep2 := r.AddStep(&Step{
		Name: "two",
		Do: func(ctx context.Context, req interface{}) (interface{}, error) {
			value++
			return nil, errors.New("")
		},
		Undo: func(ctx context.Context, req interface{}) error {
			value--
			return nil
		},
	})

	var c, ctx = r.StartTX(context.Background())
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
	var errTest = errors.New("test")
	ep1 := r.AddStep(&Step{
		Name: "one",
		Do: func(ctx context.Context, req interface{}) (interface{}, error) {
			value++
			return nil, nil
		},
		Undo: func(ctx context.Context, req interface{}) error {
			value--
			return nil
		},
	})

	ep2 := r.AddStep(&Step{

		Name: "two",
		Do: func(ctx context.Context, req interface{}) (interface{}, error) {
			value++
			return nil, errors.New("")
		},
		Undo: func(ctx context.Context, req interface{}) error {
			if attempt == 0 {
				attempt++
				return errTest
			}
			value--
			return nil

		},
	})

	var c, ctx = r.StartTX(context.Background())
	ep1(ctx, nil)
	ep2(ctx, nil)
	err := c.Rollback(ctx)
	assert.NotNil(t, err)
	assert.Len(t, err.(*multierror.Error).Errors, 1)
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
		Name: "one",
		Do: func(ctx context.Context, req interface{}) (interface{}, error) {
			value++
			return nil, nil
		},
		Undo: func(ctx context.Context, req interface{}) error {
			value--
			return nil
		},
	})
	ep2 := r.AddStep(&Step{
		Name: "two",
		Do: func(ctx context.Context, req interface{}) (interface{}, error) {
			value++
			return nil, errors.New("")
		},
		Undo: func(ctx context.Context, req interface{}) error {
			if attempt == 0 {
				attempt++
				panic("err")
			}
			value--
			return nil
		},
	})

	defer func(r *Registry) {
		if rec := recover(); rec != nil {
			r.Recover(context.Background())
			assert.Equal(t, 0, value)
		}
	}(r)

	var _, ctx = r.StartTX(context.Background())
	ep1(ctx, nil)
	ep2(ctx, nil)
}

func TestSaga_shortCircuit(t *testing.T) {
	var value int
	var store = &InProcessStore{}
	var r = NewRegistry(store, WithTimeout(0))

	ep1 := r.AddStep(&Step{
		Name: "one",
		Do: func(ctx context.Context, req interface{}) (interface{}, error) {
			value++
			return nil, nil
		},
		Undo: func(ctx context.Context, req interface{}) error {
			value--
			return nil
		},
	})

	r.AddStep(&Step{
		Name: "two",
		Do: func(ctx context.Context, req interface{}) (interface{}, error) {
			panic("should not reach")
		},
		Undo: func(ctx context.Context, req interface{}) error {
			panic("should not reach")
		},
	})

	var c, ctx = r.StartTX(context.Background())
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
		Name: "two",
		Do: func(ctx context.Context, req interface{}) (interface{}, error) {
			value++
			return nil, errors.New("foo")
		},
		Undo: func(ctx context.Context, req interface{}) error {
			if attempt == 0 {
				attempt++
				value--
				return nil
			}
			panic("err")
		},
	})
	tx, ctx := r.StartTX(context.Background())
	tx.Commit(ctx)
	r.Recover(context.Background())
	assert.Equal(t, 0, value)
	defer func() {
		assert.NotNil(t, recover())
	}()

	ep(ctx, nil)

}

func TestSaga_decode_encode(t *testing.T) {
	var store = &InProcessStore{}
	var r = NewRegistry(store, WithTimeout(0))
	var called int
	ep := r.AddStep(&Step{
		Name: "two",
		Do: func(ctx context.Context, req interface{}) (interface{}, error) {
			assert.Equal(t, "FOO", req)
			return nil, errors.New("foo")
		},
		Undo: func(ctx context.Context, req interface{}) error {
			assert.Equal(t, "FOO", req)
			return nil
		},
		DecodeParam: func(bytes []byte) (interface{}, error) {
			called++
			return strings.ToUpper(string(bytes)), nil
		},
		EncodeParam: func(i interface{}) ([]byte, error) {
			called++
			return []byte(strings.ToLower(i.(string))), nil
		},
	})
	tx, ctx := r.StartTX(context.Background())
	ep(ctx, "FOO")
	tx.Rollback(ctx)
	assert.LessOrEqual(t, 2, called)
	assert.Equal(t, "foo", string(store.transactions[tx.correlationID][1].StepParam.([]byte)))

}
