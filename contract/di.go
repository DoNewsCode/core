package contract

// DIPopulator is an interface that models a container to which users can fetch
// dependencies. It is a syntax sugar to dig.Container. See dig.Container for the
// implementation requirement.
type DIPopulator interface {
	// Populate is just another way of fetching dependencies from container. It
	// accepts a ptr to target, and populates the target from the container.
	Populate(target any) error
}
