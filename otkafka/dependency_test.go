package otkafka

import (
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
)

func TestProvideConfigs(t *testing.T) {
	c := provideConfig()
	assert.NotEmpty(t, c.Config)
}

func TestProvideReaderFactory(t *testing.T) {
	factory, cleanup := provideReaderFactory(in{
		In: di.In{},
		Conf: config.MapAdapter{"kafka.reader": map[string]ReaderConfig{
			"default": {
				Brokers: []string{"127.0.0.1:9092"},
				Topic:   "Test",
			},
			"alternative": {
				Brokers: []string{"127.0.0.1:9093"},
				Topic:   "Test",
			},
		}},
	})
	def, err := factory.Make("default")
	assert.NoError(t, err)
	assert.NotNil(t, def)
	alt, err := factory.Make("alternative")
	assert.NoError(t, err)
	assert.NotNil(t, alt)
	assert.NotNil(t, cleanup)
	cleanup()
}

func TestProvideWriterFactory(t *testing.T) {
	factory, cleanup := provideWriterFactory(in{
		In: di.In{},
		Conf: config.MapAdapter{"kafka.writer": map[string]WriterConfig{
			"default": {
				Brokers: []string{"127.0.0.1:9092"},
				Topic:   "Test",
			},
			"alternative": {
				Brokers: []string{"127.0.0.1:9092"},
				Topic:   "Test",
			},
		}},
	})
	def, err := factory.Make("default")
	assert.NoError(t, err)
	assert.NotNil(t, def)
	alt, err := factory.Make("alternative")
	assert.NoError(t, err)
	assert.NotNil(t, alt)
	assert.NotNil(t, cleanup)
	cleanup()
}

func TestProvideKafka(t *testing.T) {
	Out, cleanupReader, cleanupWriter, err := provideKafkaFactory(in{
		Logger: log.NewNopLogger(),
		Conf: config.MapAdapter{"kafka.writer": map[string]WriterConfig{
			"default": {
				Brokers: []string{"127.0.0.1:9092"},
				Topic:   "Test",
			},
			"alternative": {
				Brokers: []string{"127.0.0.1:9092"},
				Topic:   "Test",
			},
		}},
	})
	assert.NoError(t, err)
	def, err := Out.WriterMaker.Make("default")
	assert.NoError(t, err)
	assert.NotNil(t, def)
	alt, err := Out.WriterMaker.Make("alternative")
	assert.NoError(t, err)
	assert.NotNil(t, alt)
	cleanupReader()
	cleanupWriter()
}
