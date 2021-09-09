package di

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntoPopulator(t *testing.T) {
	var target int
	g := NewGraph()
	g.Provide(func() int { return 1 })

	p := IntoPopulator(g)
	err := p.Populate(&target)
	assert.NoError(t, err)
	assert.Equal(t, 1, target)

	err = p.Populate(nil)
	assert.Error(t, err)

	var s string
	err = p.Populate(&s)
	assert.Error(t, err)
}
