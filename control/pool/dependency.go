package pool

import (
	"github.com/DoNewsCode/core/di"
)

// Providers provide a *pool.Pool to the core.
func Providers(options ...ProviderOptionFunc) di.Deps {
	return di.Deps{func() *Pool {
		return NewPool(options...)
	}}
}

// ProviderOptionFunc is the functional option to Providers.
type ProviderOptionFunc func(pool *Pool)

// WithConcurrency sets the maximum concurrency for the pool.
func WithConcurrency(concurrency int) ProviderOptionFunc {
	return func(pool *Pool) {
		pool.concurrency = concurrency
	}
}
