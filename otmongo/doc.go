/*
Package otmongo provides mongo client with opentracing.
For documentation about redis usage, see https://pkg.go.dev/go.mongodb.org/mongo-driver

Integration

package otmongo exports the configuration factoryIn the following format:

	mongo:
	  default:
	    uri:

Add the mongo dependency to core:

	var c *core.C = core.New()
	c.Provide(otmongo.Providers())

Then you can invoke redis from the application.

	c.Invoke(func(client *mongo.Client) {
		client.Connect(context.Background())
	})

Sometimes there are valid reasons to connect to more than one mongo server. Inject
otmongo.Maker to factory a *mongo.Client with a specific configuration entry.

	c.Invoke(function(maker otmongo.Maker) {
		client, err := maker.Make("default")
		// do something with client
	})
*/
package otmongo
