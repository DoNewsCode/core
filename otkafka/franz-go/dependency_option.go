package franz_go

// Interceptor is an interceptor that makes last minute change to a *kafka.ReaderConfig
// during kafka.Reader's creation
type Interceptor func(name string, config *Config)

type providersOption struct {
	reloadable  bool
	interceptor Interceptor
}

// ProvidersOptionFunc is the type of functional providersOption for Providers. Use this type to change how Providers work.
type ProvidersOptionFunc func(options *providersOption)

// WithInterceptor instructs the Providers to accept the
// ReaderInterceptor so that users can change reader config during runtime. This can
// be useful when some dynamic computations on configs are required.
func WithInterceptor(interceptor Interceptor) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.interceptor = interceptor
	}
}

// WithReload toggles whether the writer factory should reload cached instances upon
// OnReload event.
func WithReload(shouldReload bool) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.reloadable = shouldReload
	}
}
