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

	err = p.Populate(s)
	assert.Error(t, err)
}

type Stub struct{}

func (s Stub) Foo() {}

type Fooer interface {
	Foo()
}

func TestBind(t *testing.T) {
	ctor := func() Stub {
		return Stub{}
	}
	g := NewGraph()
	g.Provide(ctor)
	g.Provide(Bind(new(Stub), new(Fooer)))
	err := g.Invoke(func(f Fooer) {
		assert.NotNil(t, f)
	})
	assert.NoError(t, err)
}
