package clihttp_test

import (
	"net/http"
	"testing"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/clihttp"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/observability"
	"github.com/stretchr/testify/assert"
)

type mockDoer bool

func (m *mockDoer) Do(req *http.Request) (*http.Response, error) {
	*m = true
	resp := http.Response{}
	return &resp, nil
}

func TestProviders(t *testing.T) {
	t.Run("no panic", func(t *testing.T) {
		c := core.Default()
		c.Provide(observability.Providers())
		c.Provide(clihttp.Providers())
		c.Invoke(func(client *clihttp.Client) {})
		c.Invoke(func(client contract.HttpDoer) {})
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
	t.Run("replace doer", func(t *testing.T) {
		var mock mockDoer
		c := core.Default()
		c.Provide(observability.Providers())
		c.Provide(clihttp.Providers(clihttp.WithClientConstructor(func(args clihttp.ClientArgs) (contract.HttpDoer, error) {
			return &mock, nil
		})))
		c.Invoke(func(client *clihttp.Client) {
			req, _ := http.NewRequest(http.MethodGet, "", nil)
			client.Do(req)
			assert.True(t, bool(mock))
		})
	})
	t.Run("additional options", func(t *testing.T) {
		c := core.Default()
		c.Provide(observability.Providers())
		c.Provide(clihttp.Providers(clihttp.WithClientOption(clihttp.WithRequestLogThreshold(10))))
		c.Invoke(func(client *clihttp.Client) {})
	})
}
