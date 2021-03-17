package core

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	conf := provideDefaultConfig()
	_, err := json.Marshal(conf)
	assert.NoError(t, err)
}
