package di

// Module is a special interface. A type implements Module to indicate it is a
// valid core module. This property is examined and executed by the core.Provide
// in package core. Other than that, the interface bears no further meaning.
//
// If a module is not created by core.Provide, there is no need to implements this interface.
type Module interface {
	// ModuleSentinel marks the type as a core module
	ModuleSentinel()
}

// Deps is a set of providers grouped together. This is used by core.Provide
// method to identify provider sets.
type Deps []interface{}
