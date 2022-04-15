package otredis

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/di"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestModule_ProvideCommand(t *testing.T) {
	if os.Getenv("REDIS_ADDR") == "" {
		t.Skip("set REDIS_ADDR to run TestModule_ProvideRunGroup")
		return
	}
	addrs := strings.Split(os.Getenv("REDIS_ADDR"), ",")

	c := core.New(core.WithInline("redis.default.addrs", addrs))
	c.ProvideEssentials()
	c.Provide(di.Deps{
		provideRedisFactory(&providersOption{}),
		di.Bind(new(*Factory), new(Maker)),
	})
	c.AddModuleFunc(New)
	rootCmd := cobra.Command{}
	c.ApplyRootCommand(&rootCmd)
	assert.True(t, rootCmd.HasSubCommands())

	cases := []struct {
		name   string
		args   []string
		before func(t *testing.T, maker Maker)
		after  func(t *testing.T, maker Maker)
	}{
		{
			"cleanup",
			[]string{"redis", "cleanup", "1ms"},
			func(t *testing.T, maker Maker) {
				client, _ := maker.Make("default")
				client.Set(context.Background(), "foo", "bar", 0)
				time.Sleep(time.Second)
			},
			func(t *testing.T, maker Maker) {
				client, _ := maker.Make("default")
				results, _ := client.Keys(context.Background(), "*").Result()
				assert.Empty(t, results)
			},
		},
		{
			"cleanup",
			[]string{"redis", "cleanup", "1ms", "-p", "bar"},
			func(t *testing.T, maker Maker) {
				client, _ := maker.Make("default")
				client.Set(context.Background(), "foo", "bar", 0)
				client.Set(context.Background(), "bar", "bar", 0)
				time.Sleep(time.Second)
			},
			func(t *testing.T, maker Maker) {
				client, _ := maker.Make("default")
				results, _ := client.Keys(context.Background(), "*").Result()
				assert.Len(t, results, 1)
			},
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			c.Invoke(func(maker Maker) {
				cc.before(t, maker)
			})
			rootCmd.SetArgs(cc.args)
			rootCmd.Execute()
			c.Invoke(func(maker Maker) {
				cc.after(t, maker)
			})
		})
	}
}
