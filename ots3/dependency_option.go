package ots3

type providersOption struct {
	ctor       ManagerConstructor
	reloadable bool
}

// ProvidersOptionFunc is the type of functional providersOption for Providers. Use this type to change how Providers work.
type ProvidersOptionFunc func(options *providersOption)

// ManagerConstructor constructs a manager
type ManagerConstructor func(args ManagerArgs) (*Manager, error)

// WithManagerConstructor is a provider option to override how s3 manager are constructed.
func WithManagerConstructor(ctor ManagerConstructor) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.ctor = ctor
	}
}

// WithReload toggles whether the factory should reload cached instances upon
// OnReload event.
func WithReload(shouldReload bool) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.reloadable = shouldReload
	}
}
