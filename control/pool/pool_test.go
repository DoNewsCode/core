package pool

import (
	"context"
	"testing"
	"time"
)

func TestPool_Go(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	m := NewManager()
	p := NewPool(m, 1)
	go m.Run(ctx)
	time.Sleep(time.Millisecond)
	p.Go(context.Background(), func(asyncContext context.Context) {
		cancel()
	})

}

func TestPool_Parallel(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	m := NewManager()
	p := NewPool(m, 2)
	go m.Run(ctx)
	time.Sleep(time.Millisecond)

	var executed = make(chan struct{})
	ts := time.Now()

	// saturate the pool
	p.Go(ctx, func(asyncContext context.Context) {
		time.Sleep(time.Second)
	})
	// fallback to sync mode
	p.Go(ctx, func(asyncContext context.Context) {
		time.Sleep(time.Second)
		close(executed)
	})

	<-executed
	if time.Since(ts) >= 2*time.Second {
		t.Fatal("workers should run in parallel")
	}
}

func TestPool_Overflow(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	m := NewManager()
	p := NewPool(m, 1)
	go m.Run(ctx)
	time.Sleep(time.Millisecond)
	var executed = make(chan struct{})
	ts := time.Now()

	// saturate the pool
	p.Go(ctx, func(asyncContext context.Context) {
		time.Sleep(time.Second)
	})
	// fallback to sync mode
	p.Go(ctx, func(asyncContext context.Context) {
		time.Sleep(time.Second)
		close(executed)
	})
	<-executed
	if time.Since(ts) <= 2*time.Second {
		t.Fatal("pool should have been saturated")
	}
}

func TestPool_ContextExpired(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	m := NewManager()
	p := NewPool(m, 1)
	go m.Run(ctx)
	time.Sleep(time.Millisecond)
	ts := time.Now()

	// saturate the pool
	p.Go(ctx, func(asyncContext context.Context) {
		time.Sleep(time.Second)
	})
	// fallback to sync mode
	p.Go(ctx, func(asyncContext context.Context) {
		time.Sleep(time.Second)
	})
	if time.Since(ts) <= 2*time.Second {
		t.Fatal("pool should have been saturated")
	}
}

func TestPool_ManagerOutOfCapacity(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	m := NewManager()
	m.workers = make(chan *Worker, 1)
	go m.Run(ctx)
	time.Sleep(time.Millisecond)
	var executed = make(chan struct{})

	// saturate the pool
	m.Go(ctx, func(asyncContext context.Context) {
		time.Sleep(time.Millisecond)
	})
	// fallback to sync mode
	m.Go(ctx, func(asyncContext context.Context) {
		time.Sleep(time.Second)
		close(executed)
	})
	<-executed
}

func TestPool_Wait(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	m := NewManager()
	m.workers = make(chan *Worker, 1)
	p := NewPool(m, 2)
	go m.Run(ctx)
	time.Sleep(time.Millisecond)

	ts := time.Now()

	// saturate the pool
	p.Go(ctx, func(asyncContext context.Context) {
		time.Sleep(time.Millisecond)
	})
	// fallback to sync mode
	p.Go(ctx, func(asyncContext context.Context) {
		time.Sleep(time.Second)

	})
	p.Wait()
	if time.Since(ts) <= 1*time.Second {
		t.Fatal("should wait for the pool to finish")
	}

}

func TestPool_contextValue(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := NewManager()
	p := NewPool(m, 1)
	go m.Run(ctx)
	time.Sleep(time.Millisecond)

	key := struct{}{}
	requestContext := context.WithValue(context.Background(), key, "foo")

	p.Go(requestContext, func(asyncContext context.Context) {
		if _, ok := asyncContext.Deadline(); ok {
			t.Fatalf("asyncContext shouldn't have deadline set")
		}
		value := asyncContext.Value(key)
		if value != "foo" {
			t.Fatalf("want foo, got %s", value)
		}
		cancel()
	})
}
