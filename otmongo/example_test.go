package otmongo_test

import (
	"context"
	"fmt"
	"os"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/otmongo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func Example() {
	if os.Getenv("MONGO_ADDR") == "" {
		fmt.Println("set MONGO_ADDR to run this example")
		return
	}
	c := core.New()
	c.ProvideEssentials()
	c.Provide(otmongo.Providers())
	c.Invoke(func(mongo *mongo.Client) {
		err := mongo.Ping(context.Background(), readpref.Nearest())
		fmt.Println(err)
	})
	// Output:
	// <nil>
}
