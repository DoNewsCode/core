package otkafka

import (
	"os"
	"strings"
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/kit/log"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func TestProvideConfigs(t *testing.T) {
	c := provideConfig()
	assert.NotEmpty(t, c.Config)
}

func TestProvideReaderFactory(t *testing.T) {
	if os.Getenv("KAFKA_ADDR") == "" {
		t.Skip("set KAFKA_ADDR to run TestProvideReaderFactory")
		return
	}
	addrs := strings.Split(os.Getenv("KAFKA_ADDR"), ",")
	factory, cleanup := provideReaderFactory(factoryIn{
		Conf: config.MapAdapter{"kafka.reader": map[string]ReaderConfig{
			"default": {
				Brokers: addrs,
				Topic:   "Test",
			},
			"alternative": {
				Brokers: addrs,
				Topic:   "Test",
			},
		}},
	}, func(name string, reader *kafka.ReaderConfig) {})
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
	if os.Getenv("KAFKA_ADDR") == "" {
		t.Skip("set KAFKA_ADDR to run TestProvideReaderFactory")
		return
	}
	addrs := strings.Split(os.Getenv("KAFKA_ADDR"), ",")
	factory, cleanup := provideWriterFactory(factoryIn{
		In: di.In{},
		Conf: config.MapAdapter{"kafka.writer": map[string]WriterConfig{
			"default": {
				Brokers: addrs,
				Topic:   "Test",
			},
			"alternative": {
				Brokers: addrs,
				Topic:   "Test",
			},
		}},
	}, func(name string, writer *kafka.Writer) {})
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
	Out, cleanupReader, cleanupWriter, err := provideKafkaFactory(&providersOption{})(factoryIn{
		Logger: log.NewNopLogger(),
		Conf: config.MapAdapter{"kafka.writer": map[string]WriterConfig{
			"default": {
				Brokers: nil,
				Topic:   "Test",
			},
			"alternative": {
				Brokers: nil,
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
