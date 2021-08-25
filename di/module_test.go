package di

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModule(t *testing.T) {
	type mock struct {
		Modular
	}
	assert.Implements(t, (*Modular)(nil), mock{})
	assert.Implements(t, (*Modular)(nil), &mock{})
}
