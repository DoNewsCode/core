package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestModule(t *testing.T) {
	type mock struct {
		Module
	}
	assert.Implements(t, (*Module)(nil), mock{})
	assert.Implements(t, (*Module)(nil), &mock{})
}
