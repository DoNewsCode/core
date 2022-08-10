package pool

import (
	"context"
	"testing"
	"time"

	"github.com/DoNewsCode/core/config"
)

func TestPool_Go(t *testing.T) {
	t.Parallel()

	f, cancel, _ := providePoolFactory()(factoryIn{
		Conf: config.MapAdapter{},
	})
	time.Sleep(time.Millisecond)
	p, _ := f.Factory.Make("default")
	p.Go(context.Background(), func(asyncContext context.Context) {
	})

	cancel()
}

func TestPool_contextValue(t *testing.T) {
	t.Parallel()
	f, cancel, _ := providePoolFactory()(factoryIn{Conf: config.MapAdapter{}})
	time.Sleep(time.Millisecond)

	p, _ := f.Factory.Make("default")

	key := struct{}{}
	requestContext := context.WithValue(context.Background(), key, "foo")
	execute := make(chan struct{})
	p.Go(requestContext, func(asyncContext context.Context) {
		if _, ok := asyncContext.Deadline(); ok {
			t.Fatalf("asyncContext shouldn't have deadline set")
		}
		value := asyncContext.Value(key)
		if value != "foo" {
			t.Fatalf("want foo, got %s", value)
		}
		execute <- struct{}{}
	})
	<-execute
	cancel()
}

func TestPool_Nil_Valid(t *testing.T) {
	var p Pool
	p.Go(context.Background(), func(asyncContext context.Context) {

	})
}
