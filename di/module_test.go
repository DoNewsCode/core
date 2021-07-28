package di

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModule(t *testing.T) {
	type mock struct {
		Module
	}
	assert.Implements(t, (*Module)(nil), mock{})
	assert.Implements(t, (*Module)(nil), &mock{})
}
