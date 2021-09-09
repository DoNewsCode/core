package otmongo

import (
	"testing"

	"github.com/DoNewsCode/core"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestIntegration(t *testing.T) {
	t.Parallel()

	var called bool
	c := core.Default()
	c.Provide(Providers(WithConfigInterceptor(func(name string, clientOptions *options.ClientOptions) {
		called = true
	})))
	c.Invoke(func(maker Maker) {
		assert.False(t, called)
		maker.Make("default")
		assert.True(t, called)
	})
}
