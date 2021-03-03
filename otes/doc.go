/*
Package otes provides es client with opentracing.
For documentation about es usage, see https://github.com/olivere/elastic

Integration

package otes exports the configuration in the following format:

	es:
	  default:
        "url": "http://localhost:9200"
		"index":       "",
		"username":    "",
		"password":    "",
		"shards":      0,
		"replicas":    0,
		"sniff":       false,
		"healthCheck": false,
		"infoLog":     "",
		"errorLog":    "",
		"traceLog":    "",

Add the es dependency to core:

	var c *core.C = core.New()
	c.Provide(otes.Providers())

Then you can invoke etcd from the application.

	c.Invoke(func(client *elastic.Client) {
		// Do something with es
	})

Sometimes there are valid reasons to connect to more than one es server. Inject
otes.Maker to factory a *elastic.Client with a specific configuration entry.

	c.Invoke(function(maker otes.Maker) {
		client, err := maker.Make("default")
		// do something with client
	})
*/
package otes
