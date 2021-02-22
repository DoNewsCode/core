/*
Package otmongo provides mongo client with opentracing.
For documentation about redis usage, see https://pkg.go.dev/go.mongodb.org/mongo-driver

Integration

package otmongo exports the configuration in the following format:

	redis:
	  default:
	    uri:

Add the mongo dependency to core:

	var c *core.C = core.New()
	c.Provide(otmongo.Provide)

Then you can invoke redis from the application.

	c.Invoke(func(client *mongo.Client) {
		client.Connect(context.Background())
	})

Sometimes there are valid reasons to connect to more than one mongo server. Inject
otmongo.Maker to factory a *mongo.Client with a specific configuration entry.

	c.Invoke(function(maker otredis.Maker) {
		client, err := maker.Make("default")
		// do something with client
	})
*/
package otmongo
