package otmongo_test

import (
	"context"
	"fmt"

	"github.com/DoNewsCode/std/pkg/core"
	"github.com/DoNewsCode/std/pkg/otmongo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func Example() {
	c := core.New()
	c.AddCoreDependencies()
	c.AddDependencyFunc(otmongo.ProvideMongo)
	c.Invoke(func(mongo *mongo.Client) {
		err := mongo.Ping(context.Background(), readpref.Nearest())
		fmt.Println(err)
	})
	// Output:
	// <nil>
}
