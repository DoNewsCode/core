package otmongo_test

import (
	"context"
	"fmt"

	"github.com/DoNewsCode/std/pkg/core"
	"github.com/DoNewsCode/std/pkg/otmongo"
	"github.com/knadh/koanf/providers/confmap"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func Example() {
	c := core.New(core.WithConfigStack(confmap.Provider(map[string]interface{}{
		"mongo.default.uri": "mongodb://127.0.0.1:27017",
	}, "."), nil))
	c.AddCoreDependencies()
	c.AddDependencyFunc(otmongo.ProvideMongo)
	c.Invoke(func(mongo *mongo.Client) {
		err := mongo.Ping(context.Background(), readpref.Nearest())
		fmt.Println(err)
	})
	// Output:
	// <nil>
}
