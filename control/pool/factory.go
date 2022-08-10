package pool

import (
	"github.com/DoNewsCode/core/di"
)

// Maker models Factory
type Maker interface {
	Make(name string) (*Pool, error)
}

// Factory is the *di.Factory that creates *Pool.
type Factory = di.Factory[*Pool]
