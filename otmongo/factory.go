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
type Factory = di.Factory[*mongo.Client]
