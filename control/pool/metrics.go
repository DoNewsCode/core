package pool

import (
	"github.com/go-kit/kit/metrics"
)

func NewCounter(syncCounter, asyncCounter metrics.Counter) *Counter {
	return &Counter{
		syncCount:  syncCounter,
		asyncCount: asyncCounter,
		poolName:   "default",
	}
}

// Counter is a collection of metrics in pool.Pool.
type Counter struct {
	syncCount  metrics.Counter
	asyncCount metrics.Counter

	poolName string
}

func (c *Counter) PoolName(name string) *Counter {
	if c == nil {
		return nil
	}
	return &Counter{
		syncCount:  c.syncCount,
		asyncCount: c.asyncCount,
		poolName:   name,
	}
}

// IncSyncJob records the sync jobs count.
func (c *Counter) IncSyncJob() {
	if c == nil {
		return
	}
	c.syncCount.With("pool_name", c.poolName).Add(1)
}

// IncAsyncJob records the async jobs count.
func (c *Counter) IncAsyncJob() {
	if c == nil {
		return
	}
	c.asyncCount.With("pool_name", c.poolName).Add(1)
}
