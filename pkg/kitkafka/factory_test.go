package kitkafka

import (
	"testing"

	"github.com/DoNewsCode/std/pkg/config"
	"github.com/DoNewsCode/std/pkg/di"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
)

func TestProvideReaderFactory(t *testing.T) {
	factory, cleanup := ProvideReaderFactory(KafkaIn{
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
	factory, cleanup := ProvideWriterFactory(KafkaIn{
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
	Out, cleanupReader, cleanupWriter, err := ProvideKafka(KafkaIn{
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
