package otkafka

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_fromReaderConfig(t *testing.T) {
	reader := fromReaderConfig(ReaderConfig{})
	assert.Equal(t, []string{"127.0.0.1:9092"}, reader.Brokers)
}
