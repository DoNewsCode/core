package otmongo

import "go.mongodb.org/mongo-driver/mongo/options"

// MongoConfigInterceptor is an injection type hint that allows user to make last
// minute modification to mongo configuration. This is useful when some
// configuration cannot be easily expressed factoryIn a text form. For example, the
// options.ContextDialer.
type MongoConfigInterceptor func(name string, clientOptions *options.ClientOptions)

type providersOption struct {
	interceptor MongoConfigInterceptor
	reloadable  bool
}

// ProvidersOptionFunc is the type of functional providersOption for Providers. Use this type to change how Providers work.
type ProvidersOptionFunc func(options *providersOption)

// WithConfigInterceptor instructs the Providers to accept the
// MongoConfigInterceptor so that users can change config during runtime. This can
// be useful when some dynamic computations on configs are required.
func WithConfigInterceptor(interceptor MongoConfigInterceptor) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.interceptor = interceptor
	}
}

// WithReload toggles whether the factory should reload cached instances upon
// OnReload event.
func WithReload(shouldReload bool) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.reloadable = shouldReload
	}
}
