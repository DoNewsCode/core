package otmongo

import (
	"github.com/DoNewsCode/core/di"
	"go.mongodb.org/mongo-driver/mongo"
)

// Maker models Factory
type Maker interface {
	Make(name string) (*mongo.Client, error)
}

// Factory is a *di.Factory that creates *mongo.Client using a specific
// configuration entry.
type Factory struct {
	*di.Factory
}

// Make creates *mongo.Client using a specific configuration entry.
func (r Factory) Make(name string) (*mongo.Client, error) {
	client, err := r.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(*mongo.Client), nil
}
