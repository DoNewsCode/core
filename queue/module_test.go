package queue

import (
	"context"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/events"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestModule_ProvideCommand(t *testing.T) {
	rootCmd, driver := setUpModule()
	cases := []struct {
		name   string
		args   []string
		result QueueInfo
	}{
		{
			"reload",
			[]string{"queue", "reload"},
			QueueInfo{
				Waiting: 1,
			},
		},
		{
			"flush",
			[]string{"queue", "flush"},
			QueueInfo{},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			message := &PersistedEvent{HandleTimeout: time.Hour}
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

func setUpModule() (*cobra.Command, Driver) {
	driver := NewInProcessDriverWithPopInterval(time.Millisecond)
	factory := di.NewFactory(func(name string) (di.Pair, error) {
		queuedDispatcher := WithQueue(
			&events.SyncDispatcher{},
			driver,
		)
		return di.Pair{
			Closer: nil,
			Conn:   queuedDispatcher,
		}, nil
	})
	mod := Module{Factory: &DispatcherFactory{Factory: factory}}
	rootCmd := &cobra.Command{}
	mod.ProvideCommand(rootCmd)
	return rootCmd, driver
}
