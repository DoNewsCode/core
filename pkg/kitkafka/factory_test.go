package kitkafka

import (
	"testing"

	"github.com/DoNewsCode/std/pkg/config"
	"github.com/DoNewsCode/std/pkg/di"
	"github.com/stretchr/testify/assert"
)

func TestProvideKafkaReaderFactory(t *testing.T) {
	factory, cleanup := ProvideKafkaReaderFactory(KafkaIn{
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

func TestProvideKafkaWriterFactory(t *testing.T) {
	factory, cleanup := ProvideKafkaWriterFactory(KafkaIn{
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
