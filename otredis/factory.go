package otredis

import (
	"github.com/DoNewsCode/core/di"

	"github.com/go-redis/redis/v8"
)

// Maker is models Factory
type Maker interface {
	Make(name string) (redis.UniversalClient, error)
}

// Factory is a *di.Factory that creates redis.UniversalClient using a
// specific configuration entry.
type Factory = di.Factory[redis.UniversalClient]
