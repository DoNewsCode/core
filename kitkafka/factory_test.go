package kitkafka

import (
	"testing"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/otkafka"
	"github.com/go-kit/kit/log"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func TestNewSubscriber(t *testing.T) {
	c := core.Default(
		core.SetConfigProvider(func(configStack []config.ProviderSet, configWatcher contract.ConfigWatcher) contract.ConfigAccessor {
			return config.MapAdapter{"kafka.reader": map[string]otkafka.ReaderConfig{
				"default": {
					Brokers:     []string{"127.0.0.1:9092"},
					GroupID:     "kitkafka",
					Topic:       "uppercase",
					StartOffset: kafka.FirstOffset,
				},
			}}
		}),
		core.SetLoggerProvider(func(conf contract.ConfigAccessor, appName contract.AppName, env contract.Env) log.Logger {
			return log.NewNopLogger()
		}),
	)
	defer c.Shutdown()

	c.Provide(otkafka.Providers())

	c.Invoke(func(factory otkafka.ReaderMaker) {
		reader, err := factory.Make("default")
		assert.NoError(t, err)
		server, err := MakeSubscriberServer(
			reader,
			nil,
			WithParallelism(1),
			WithSyncCommit(),
		)
		assert.NoError(t, err)
		assert.NotNil(t, server)
	})

}
