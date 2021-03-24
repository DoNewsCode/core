/*
Package otredis provides redis client with opentracing.
For documentation about redis usage, see https://github.com/go-redis/redis

package otredis works with redis cluster, redis sentinel and single redis instance.

Integration

package otredis exports the configuration in the following format:

	redis:
	    default:
	        addrs:
	          - 127.0.0.1:6379
	        db: 0
	        username: ""
	        password: ""
	        sentinelPassword: ""
	        maxRetries: 0
	        minRetryBackoff:
	            duration: 0s
	        maxRetryBackoff:
	            duration: 0s
	        dialTimeout:
	            duration: 0s
	        readTimeout:
	            duration: 0s
	        writeTimeout:
	            duration: 0s
	        poolSize: 0
	        minIdleConns: 0
	        maxConnAge:
	            duration: 0s
	        poolTimeout:
	            duration: 0s
	        idleTimeout:
	            duration: 0s
	        idleCheckFrequency:
	            duration: 0s
	        maxRedirects: 0
	        readOnly: false
	        routeByLatency: false
	        routeRandomly: false
	        masterName: ""

To see all available configurations, use the config init command.

Add the redis dependency to core:

	var c *core.C = core.New()
	c.Provide(otredis.Providers())

Then you can invoke redis from the application.

	c.Invoke(func(redisClient redis.UniversalClient) {
		redisClient.Ping(context.Background())
	})

Sometimes there are valid reasons to connect to more than one redis server. Inject
otredis.Maker to factory a redis.UniversalClient with a specific configuration entry.

	c.Invoke(function(maker otredis.Maker) {
		client, err := maker.Make("default")
		// do something with client
	})

*/
package otredis
