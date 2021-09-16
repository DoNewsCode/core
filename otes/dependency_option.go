package otes

import "github.com/olivere/elastic/v7"

type providersOption struct {
	interceptor       EsConfigInterceptor
	clientConstructor func(args ClientArgs) (*elastic.Client, error)
	reloadable        bool
}

// ProvidersOptionFunc is the type of functional providersOption for Providers. Use this type to change how Providers work.
type ProvidersOptionFunc func(options *providersOption)

// WithConfigInterceptor instructs the Providers to accept the
// EsConfigInterceptor so that users can change config during runtime. This can
// be useful when some dynamic computations on configs are required.
func WithConfigInterceptor(interceptor EsConfigInterceptor) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.interceptor = interceptor
	}
}

// WithClientConstructor instructs the Providers to accept an alternative constructor for elasticsearch client.
// Refer to the package elastic for how to construct a custom elastic.Client.
func WithClientConstructor(f func(ClientArgs) (*elastic.Client, error)) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.clientConstructor = f
	}
}

// WithReload toggles whether to reload the factory after receiving the OnReload
// event. By default, the factory won't be reloaded.
func WithReload(shouldReload bool) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.reloadable = shouldReload
	}
}
