package otmongo

import (
	"context"

	"github.com/DoNewsCode/std/pkg/contract"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Mongo(conf contract.ConfigAccessor) (*mongo.Client, error) {
	uri := conf.String("uri")
	opts := options.Client()
	opts.ApplyURI(uri)
	opts.Monitor = NewMonitor()
	return mongo.Connect(context.Background(), opts)
}
