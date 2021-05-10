package otkafka

import (
	"github.com/DoNewsCode/core/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_fromReaderConfig(t *testing.T) {
	reader := fromReaderConfig(ReaderConfig{})
	assert.Equal(t, config.EnvDefaultKafkaAddrs, reader.Brokers)
}