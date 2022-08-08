package pool

import (
	"context"
	"testing"
	"time"

	"github.com/DoNewsCode/core/config"
)

func TestPool_Go(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())

	f, _, _ := providePoolFactory()(factoryIn{
		Conf: config.MapAdapter{},
	})
	go f.Factory.run(ctx)

	p, _ := f.Factory.Make("default")
	p.Go(context.Background(), func(asyncContext context.Context) {
		cancel()
	})

}

func TestPool_CapLimit(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f, _, _ := providePoolFactory()(factoryIn{Conf: config.MapAdapter{
		"pool": map[string]any{
			"default": map[string]any{
				"cap":     1,
				"timeout": "1s",
			},
		},
	}})
	go f.Factory.run(ctx)

	p, _ := f.Factory.Make("default")

	ts := time.Now()
	var executed = make(chan struct{})

	// job1
	p.Go(ctx, func(asyncContext context.Context) {
		time.Sleep(time.Second)
	})
	if p.WorkerCount() != 1 {
		t.Fatal("worker count should be 1")
	}
	// job2
	p.Go(ctx, func(asyncContext context.Context) {
		close(executed)
	})
	if p.WorkerCount() != 2 {
		t.Fatal("worker count should be 2")
	}
	<-executed
	// job channel not be blocked, so the interval should be less than 1 second
	if time.Since(ts) >= time.Second {
		t.Fatal("timeout: sync mode should be used")
	}
	time.Sleep(time.Second)
	if p.WorkerCount() != 1 {
		t.Fatal("worker should be recycle")
	}
}

func TestPool_contextValue(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f, _, _ := providePoolFactory()(factoryIn{Conf: config.MapAdapter{}})
	go f.Factory.run(ctx)

	p, _ := f.Factory.Make("default")

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
