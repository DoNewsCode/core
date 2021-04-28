package otkafka

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_fromReaderConfig(t *testing.T) {
	reader := fromReaderConfig(ReaderConfig{})
	assert.Equal(t, os.Getenv("KAFKA_ADDR"), reader.Brokers[0])
}
