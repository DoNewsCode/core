package otmongo_test

import (
	"context"
	"fmt"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/otmongo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func Example() {
	c := core.New()
	c.ProvideEssentials()
	c.Provide(otmongo.Provide)
	c.Invoke(func(mongo *mongo.Client) {
		err := mongo.Ping(context.Background(), readpref.Nearest())
		fmt.Println(err)
	})
	// Output:
	// <nil>
}
