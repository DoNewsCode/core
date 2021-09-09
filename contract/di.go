package contract

// DIProvider is an interface that models a container to which users can add
// dependencies. See di.Graph for the implementation requirement.
type DIProvider interface {
	Provide(constructor interface{}) error
}

// DIInvoker is an interface that models a container to which users can fetch
// dependencies. See di.Graph for the implementation requirement.
type DIInvoker interface {
	Invoke(function interface{}) error
}

// DIPopulator is an interface that models a container to which users can fetch
// dependencies. It is an syntax sugar to DIInvoker. See di.Graph for the
// implementation requirement.
type DIPopulator interface {
	// Populate is just another way of fetching dependencies from container. It
	// accepts a ptr to target, and populates the target from the container.
	Populate(target interface{}) error
}

// DIContainer is a container roughly modeled after dig.Container.
type DIContainer interface {
	DIProvider
	DIInvoker
}
