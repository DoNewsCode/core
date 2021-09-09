package ots3

type providersOption struct {
	ctor ManagerConstructor
}

// ProvidersOptionFunc is the type of functional providersOption for Providers. Use this type to change how Providers work.
type ProvidersOptionFunc func(options *providersOption)

// ManagerConstructor constructs a manager
type ManagerConstructor func(args ManagerConstructorArgs) (*Manager, error)

// WithManagerConstructor is a provider option to override how s3 manager are constructed.
func WithManagerConstructor(ctor ManagerConstructor) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.ctor = ctor
	}
}
