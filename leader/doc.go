/*
Package leader provides a simple leader election implementation.

Introduction

Leader election is particularly useful if the state cannot be rule out of your
application. For example, you want to run some cron jobs to scan the database,
but you have more than one instances up and running. Running cron jobs on all
instances many not be desired. With the help of package leader, you can opt to
only run such jobs on the leader node. When the leader goes down, a new leader will
be elected. The cron job runner is therefore highly available.

Usage

The package leader exports configuration in this format:

	leader:
	  etcdName: default

To use package leader with package core:

	var c *core.C = core.Default()
	c.Provide(otetcd.Providers) // to provide the underlying driver
	c.Provide(leader.Providers)
	c.Invoke(func(status *leader.Status) {
		if ! status.IsLeader {
			return
		}
		// DO SOMETHING ON LEADER
	})
*/
package leader
