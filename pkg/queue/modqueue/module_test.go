package modqueue

import (
	"context"
	"github.com/DoNewsCode/std/pkg/async"
	"github.com/DoNewsCode/std/pkg/events"
	"github.com/DoNewsCode/std/pkg/queue"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestModule_ProvideCommand(t *testing.T) {
	rootCmd, driver := setUp()
	cases := []struct {
		name   string
		args   []string
		result queue.QueueInfo
	}{
		{
			"reload",
			[]string{"queue", "reload"},
			queue.QueueInfo{
				Waiting: 1,
			},
		},
		{
			"flush",
			[]string{"queue", "flush"},
			queue.QueueInfo{},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			message := &queue.PersistedEvent{HandleTimeout: time.Hour}
			driver.Push(context.Background(), message, 0)
			driver.Pop(context.Background())
			driver.Fail(context.Background(), message)
			rootCmd.SetArgs(c.args)
			rootCmd.Execute()
			info, _ := driver.Info(context.Background())
			assert.Equal(t, c.result, info)
			driver.Pop(context.Background())
		})
	}
}

func setUp() (*cobra.Command, queue.Driver) {
	driver := queue.NewInProcessDriverWithPopInterval(time.Millisecond)
	factory := async.NewFactory(func(name string) (async.Pair, error) {
		queuedDispatcher := queue.WithQueue(
			&events.SyncDispatcher{},
			driver,
		)
		return async.Pair{
			Closer: nil,
			Conn:   queuedDispatcher,
		}, nil
	})
	mod := Module{Factory: &DispatcherFactory{Factory: factory}}
	rootCmd := &cobra.Command{}
	mod.ProvideCommand(rootCmd)
	return rootCmd, driver
}
