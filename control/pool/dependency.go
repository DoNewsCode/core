package pool

import (
	"github.com/DoNewsCode/core/di"
)

// Providers provide a *pool.Pool to the core.
func Providers() di.Deps {
	return di.Deps{NewManager}
}
