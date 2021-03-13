package core

import (
	"context"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/events"
	"io/ioutil"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/otgorm"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestC_Serve(t *testing.T) {
	var called int32
	c := New(
		WithInline("http.addr", ":19998"),
		WithInline("grpc.addr", ":19999"),
	)
	c.ProvideEssentials()
	c.Invoke(func(dispatcher contract.Dispatcher) {
		dispatcher.Subscribe(events.Listen(events.From(OnHTTPServerStart{}), func(ctx context.Context, start contract.Event) error {
			atomic.AddInt32(&called, 1)
			assert.Equal(t, "[::]:19998", start.Data().(OnHTTPServerStart).Listener.Addr().String())
			return nil
		}))
	})
	c.Invoke(func(dispatcher contract.Dispatcher) {
		dispatcher.Subscribe(events.Listen(events.From(OnHTTPServerShutdown{}), func(ctx context.Context, shutdown contract.Event) error {
			atomic.AddInt32(&called, 1)
			assert.Equal(t, "[::]:19998", shutdown.Data().(OnHTTPServerShutdown).Listener.Addr().String())
			return nil
		}))
	})
	c.Invoke(func(dispatcher contract.Dispatcher) {
		dispatcher.Subscribe(events.Listen(events.From(OnGRPCServerStart{}), func(ctx context.Context, start contract.Event) error {
			atomic.AddInt32(&called, 1)
			assert.Equal(t, "[::]:19999", start.Data().(OnGRPCServerStart).Listener.Addr().String())
			return nil
		}))
	})
	c.Invoke(func(dispatcher contract.Dispatcher) {
		dispatcher.Subscribe(events.Listen(events.From(OnGRPCServerShutdown{}), func(ctx context.Context, shutdown contract.Event) error {
			atomic.AddInt32(&called, 1)
			assert.Equal(t, "[::]:19999", shutdown.Data().(OnGRPCServerShutdown).Listener.Addr().String())
			return nil
		}))
	})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	e := c.Serve(ctx)
	assert.NoError(t, e)
	assert.Equal(t, int32(4), atomic.LoadInt32(&called))
}

func TestC_Default(t *testing.T) {
	c := New()
	c.ProvideEssentials()
	c.Provide(otgorm.Providers())
	c.AddModuleFunc(config.New)

	f, _ := ioutil.TempFile("./", "*")

	rootCommand := &cobra.Command{}
	c.ApplyRootCommand(rootCommand)
	rootCommand.SetArgs([]string{"config", "init", "-o", f.Name()})
	rootCommand.Execute()

	output, _ := ioutil.ReadFile(f.Name())
	assert.Contains(t, string(output), "gorm:")
	os.Remove(f.Name())
}
