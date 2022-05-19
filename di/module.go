package di

// Modular is a special interface. A type implements Modular to indicate it contains a
// valid core module. This property is examined and executed by the core.Provide
// in package core. Other than that, the interface bears no further meaning.
//
// If the module is not a by-product of dependency construction via core.Provide,
// then there is no need to use this mechanism.
type Modular interface {
	// Module returns a core module
	Module() any
}

// Deps is a set of providers grouped together. This is used by core.Provide
// method to identify provider sets.
type Deps []any
