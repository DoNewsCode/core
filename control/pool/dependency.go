package pool

import (
	"github.com/DoNewsCode/core/di"
)

// Providers provide a *Manager to the core.
func Providers() di.Deps {
	return di.Deps{NewManager}
}
