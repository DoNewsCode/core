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
type Factory struct {
	*di.Factory
}

// Make creates redis.UniversalClient using a specific configuration entry.
func (r Factory) Make(name string) (redis.UniversalClient, error) {
	client, err := r.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(redis.UniversalClient), nil
}
