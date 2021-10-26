package clihttp_test

import (
	"testing"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/clihttp"
	"github.com/DoNewsCode/core/observability"
)

func TestProviders(t *testing.T) {
	t.Run("no panic", func(t *testing.T) {
		c := core.Default()
		c.Provide(observability.Providers())
		c.Provide(clihttp.Providers())
		c.Invoke(func(client *clihttp.Client) {})
	})
	t.Run("panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				return
			}
			t.Fatal("test should panic")
		}()
		c := core.Default()
		c.Provide(clihttp.Providers())
		c.Invoke(func(client *clihttp.Client) {})
	})
}
