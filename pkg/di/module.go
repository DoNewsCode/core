package di

// Module is a special embed interface. If a type embeds the Module, that
// embedding type is a core module. This property is examined by the
// core.AddDependencyFunc in package core.
// Other than that, the interface bears
// no further meaning.
//
// If a module is not created by core.AddDependencyFunc, there is no need to embed this interface.
// When in doubt, don't embed this.
type Module interface {
	moduleSentinel()
}
