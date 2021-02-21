package di

// Module is a special embed interface. If a type embeds the Module, that
// embedding type is a core module and it will be added to the core
// automatically. This property is examined and executed by the core.Provide in
// package core. Other than that, the interface bears no further meaning.
//
// If a module is not created by core.Provide, there is no need to embed this interface.
// When in doubt, don't embed this.
type Module interface {
	moduleSentinel()
}
