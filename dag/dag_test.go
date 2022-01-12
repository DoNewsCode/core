package dag

import (
	"context"
	"errors"
	"os"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
)

func TestDag_AddVertex(t *testing.T) {
	dag := New()
	dag.AddVertex(func(ctx context.Context) error { return nil }, WithName("foo"))
	assert.Equal(t, dag.order(), 1)
	dag.AddVertex(func(ctx context.Context) error { return nil })
	assert.Equal(t, dag.order(), 2)
}

func TestDag_addEdge(t *testing.T) {
	t.Run("normal addEdge", func(t *testing.T) {
		dag := New()
		e1 := dag.AddVertex(func(ctx context.Context) error { return nil })
		e2 := dag.AddVertex(func(ctx context.Context) error { return nil })
		assert.NoError(t, dag.AddEdge(e1, e2))
	})

	t.Run("circular deps", func(t *testing.T) {
		dag := New()
		e1 := dag.AddVertex(func(ctx context.Context) error { return nil })
		e2 := dag.AddVertex(func(ctx context.Context) error { return nil })
		e3 := dag.AddVertex(func(ctx context.Context) error { return nil })
		dag.AddEdge(e1, e2)
		dag.AddEdge(e2, e3)
		assert.Error(t, dag.AddEdge(e3, e1))
	})

	t.Run("invalid edge", func(t *testing.T) {
		dag := New()
		dag.AddVertex(func(ctx context.Context) error { return nil })
		assert.Error(t, dag.AddEdge(0, 1))
		assert.Error(t, dag.AddEdge(1, 2))
	})
}

func TestDag_Run(t *testing.T) {
	t.Parallel()
	makeJob := func(expectedLevel int32, level *int32, executed *int32) func(ctx context.Context) error {

		return func(ctx context.Context) error {
			c := atomic.LoadInt32(level)
			if c+1 != expectedLevel {
				t.Errorf("expected %d, got %d", expectedLevel, c+1)
				return errors.New("wrong expected level")
			}
			runtime.Gosched()
			time.Sleep(time.Millisecond * 10)
			atomic.AddInt32(executed, 1)
			atomic.StoreInt32(level, expectedLevel)
			return nil
		}
	}

	cases := []struct {
		name string
		dag  func(executed *int32) *DAG
	}{
		{
			"simple",
			func(executed *int32) *DAG {
				var level int32
				dag := New()
				dag.AddVertex(makeJob(1, &level, executed), WithLogger(log.NewLogfmtLogger(os.Stdout)))
				dag.AddVertex(makeJob(2, &level, executed), WithLogger(log.NewLogfmtLogger(os.Stdout)))
				dag.AddEdge(0, 1)
				return dag
			},
		},
		{
			"no edges",
			func(executed *int32) *DAG {
				var level int32
				dag := New()
				dag.AddVertex(makeJob(1, &level, executed))
				dag.AddVertex(makeJob(1, &level, executed))
				return dag
			},
		},
		{
			"diamond",
			func(executed *int32) *DAG {
				var level int32
				dag := New()
				dag.AddVertex(makeJob(1, &level, executed))
				dag.AddVertex(makeJob(2, &level, executed))
				dag.AddVertex(makeJob(2, &level, executed))
				dag.AddVertex(makeJob(3, &level, executed))
				dag.AddEdge(0, 1)
				dag.AddEdge(0, 2)
				dag.AddEdge(1, 3)
				dag.AddEdge(2, 3)
				return dag
			},
		},
		{
			"imbalanced",
			func(executed *int32) *DAG {
				var level int32
				dag := New()
				dag.AddVertex(makeJob(1, &level, executed))
				dag.AddVertex(makeJob(2, &level, executed))
				dag.AddVertex(makeJob(3, &level, executed))
				dag.AddVertex(makeJob(2, &level, executed))
				dag.AddEdge(0, 1)
				dag.AddEdge(1, 2)
				dag.AddEdge(0, 3)
				return dag
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			executed := int32(0)
			dag := c.dag(&executed)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			assert.NoError(t, dag.Run(ctx))
			assert.Equal(t, dag.order(), int(executed))
		})
	}
}

func TestDag_Run_short_circuits(t *testing.T) {
	makeJob := func(err error) func(ctx context.Context) error {
		return func(ctx context.Context) error {
			return err
		}
	}
	var errExpected = errors.New("expected")
	dag := New()
	dag.AddVertex(makeJob(nil))
	dag.AddVertex(makeJob(nil))
	dag.AddVertex(makeJob(errExpected))
	dag.AddVertex(makeJob(errors.New("should not reach here")))
	dag.AddEdge(0, 1)
	dag.AddEdge(0, 2)
	dag.AddEdge(1, 3)
	dag.AddEdge(2, 3)
	err := dag.Run(context.Background())
	assert.ErrorIs(t, err, errExpected)
}
