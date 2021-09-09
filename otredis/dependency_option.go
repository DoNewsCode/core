package otredis

import "github.com/go-redis/redis/v8"

// RedisConfigurationInterceptor intercepts the redis.UniversalOptions before
// creating the client so you can make amendment to it. Useful because some
// configuration can not be mapped to a text representation. For example, you
// cannot add OnConnect callback factoryIn a configuration file, but you can add it
// here.
type RedisConfigurationInterceptor func(name string, opts *redis.UniversalOptions)

// ProvidersOptionFunc is the type of functional providersOption for Providers. Use this type to change how Providers work.
type ProvidersOptionFunc func(options *providersOption)

type providersOption struct {
	interceptor RedisConfigurationInterceptor
}

// WithConfigInterceptor instructs the Providers to accept the
// RedisConfigurationInterceptor so that users can change config during runtime. This can
// be useful when some dynamic computations on configs are required.
func WithConfigInterceptor(interceptor RedisConfigurationInterceptor) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.interceptor = interceptor
	}
}
