package pool

import (
	"context"
	"testing"
	"time"

	"github.com/oklog/run"
)

func TestPool_Go(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	p := NewPool(WithConcurrency(1))
	go p.Go(context.Background(), func(asyncContext context.Context) {
		cancel()
	})
	p.Run(ctx)
}

func TestPool_FallbackToSyncMode(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	p := NewPool(WithConcurrency(1))
	p.Run(ctx)

	var executed = make(chan struct{})
	go func() {
		// saturate the pool
		p.Go(ctx, func(asyncContext context.Context) {
			time.Sleep(time.Second)
		})
		// fallback to sync mode
		p.Go(ctx, func(asyncContext context.Context) {
			close(executed)
		})
	}()
	<-executed
}

func TestPool_contextValue(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	p := NewPool(WithConcurrency(1))
	key := struct{}{}
	requestContext := context.WithValue(context.Background(), key, "foo")
	go p.Go(requestContext, func(asyncContext context.Context) {
		if _, ok := asyncContext.Deadline(); ok {
			t.Fatalf("asyncContext shouldn't have deadline set")
		}
		value := asyncContext.Value(key)
		if value != "foo" {
			t.Fatalf("want foo, got %s", value)
		}
	})
	p.Run(ctx)
}

func TestPool_ProvideRunGroup(t *testing.T) {
	t.Parallel()
	p := NewPool(WithConcurrency(1))
	var group run.Group
	group.Add(func() error { return nil }, func(err error) {})
	p.ProvideRunGroup(&group)
	group.Run()
}
