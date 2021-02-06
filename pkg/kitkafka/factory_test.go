package kitkafka

import (
	"github.com/DoNewsCode/std/pkg/config"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"go.uber.org/dig"
	"testing"
)

func TestProvideKafkaReaderFactory(t *testing.T) {
	factory, cleanup := ProvideKafkaReaderFactory(KafkaParam{
		In: dig.In{},
		Conf: config.MapAdapter{"kafka.reader": map[string]kafka.ReaderConfig{
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
	factory, cleanup := ProvideKafkaWriterFactory(KafkaParam{
		In: dig.In{},
		Conf: config.MapAdapter{"kafka.writer": map[string]kafka.Writer{
			"default": {
				Addr:  kafka.TCP("localhost:9092"),
				Topic: "Test",
			},
			"alternative": {
				Addr:  kafka.TCP("localhost:9092"),
				Topic: "Test",
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
