package di_test

import (
	"reflect"
	"testing"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/di"

	"github.com/stretchr/testify/assert"
	"go.uber.org/dig"
)

type Stub struct{}

func (s Stub) Foo() {}

type Fooer interface {
	Foo()
}

func ctor() Stub {
	return Stub{}
}

func badConstructor(f Fooer) Stub {
	return Stub{}
}

func TestBind(t *testing.T) {
	ctor := func() Stub {
		return Stub{}
	}
	g := core.New()
	g.Provide(di.Deps{ctor, di.Bind(new(Stub), new(Fooer))})
	g.Invoke(func(f Fooer) {
		assert.NotNil(t, f)
	})
}

func TestName(t *testing.T) {
	g := core.New()
	g.Provide(di.Deps{di.Name(ctor, "foo")})
	g.Invoke(func(injected struct {
		dig.In
		Stub Stub `name:"foo"`
	}) {
	})
}

func TestAs(t *testing.T) {
	g := core.New()
	g.Provide(di.Deps{di.As(ctor, new(Fooer))})
	g.Invoke(func(injected struct {
		dig.In
		Stub Fooer
	}) {
	})
}

func TestLocationForPC(t *testing.T) {
	g := core.New()
	inlineCtor := func(f Fooer) Stub {
		return Stub{}
	}
	g.Provide(di.Deps{di.LocationForPC(inlineCtor, reflect.ValueOf(badConstructor).Pointer())})

	defer func() {
		if r := recover(); r != nil {
			assert.Contains(t, r.(error).Error(), "badConstructor")
			return
		}
		t.Fatal("test should panic")
	}()

	g.Invoke(func(injected struct {
		dig.In
		Stub Stub
	}) {
	})
}

func TestChainedOptions(t *testing.T) {
	g := core.New()
	g.Provide(di.Deps{di.As(di.Name(di.LocationForPC(ctor, reflect.ValueOf(ctor).Pointer()), "foo"), new(Fooer))})

	g.Invoke(func(injected struct {
		di.In
		Stub Fooer `name:"foo"`
	}) {
	})
}
