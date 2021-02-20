package contract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	tenant := MapTenant{}
	assert.Equal(t, map[string]interface{}{}, tenant.KV())
	assert.Equal(t, "map[]", tenant.String())
}
