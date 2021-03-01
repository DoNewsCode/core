/*
Package otetcd provides etcd client with opentracing.
For documentation about etcd usage, see https://github.com/etcd-io/etcd/tree/master/client/v3

Integration

package otetcd exports the configuration in the following format:

	etcd:
	  default:
        autoSyncIntervalSecond: 0
        dialKeepAliveTimeSecond: 0
        dialKeepAliveTimeoutSecond: 0
        dialTimeoutSecond: 0
        endpoints:
        - 127.0.0.1:2379
        maxCallRecvMsgSize: 0
        maxCallSendMsgSize: 0
        password: ""
        permitWithoutStream: false
        rejectOldCluster: false
        username: ""

Add the etcd dependency to core:

	var c *core.C = core.New()
	c.Provide(otetcd.Providers())

Then you can invoke etcd from the application.

	c.Invoke(func(client *clientv3.Client) {
		// Do something with etcd v3
	})

Sometimes there are valid reasons to connect to more than one etcd server. Inject
otetcd.Maker to factory a *clientv3.Client with a specific configuration entry.

	c.Invoke(function(maker otetcd.Maker) {
		client, err := maker.Make("default")
		// do something with client
	})
*/
package otetcd
