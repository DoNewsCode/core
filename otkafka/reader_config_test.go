package otkafka

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_fromReaderConfig(t *testing.T) {
	reader := fromReaderConfig(ReaderConfig{})
	assert.Equal(t, envDefaultKafkaAddrs, reader.Brokers)
}
