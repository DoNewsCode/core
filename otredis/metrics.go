//go:generate mockgen -destination=./mocks/metrics_test.go github.com/go-kit/kit/metrics Gauge

package otredis

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-redis/redis/v8"
)

type collector struct {
	factory  Factory
	gauges   *Gauges
	interval time.Duration
}

// Gauges is a collection of metrics for redis connection info.
type Gauges struct {
	hits       metrics.Gauge
	misses     metrics.Gauge
	timeouts   metrics.Gauge
	totalConns metrics.Gauge
	idleConns  metrics.Gauge
	staleConns metrics.Gauge

	dbName string
}

// NewGauges constructs a new *Gauges. The default dbName label is set to "default".
func NewGauges(
	hits metrics.Gauge,
	misses metrics.Gauge,
	timeouts metrics.Gauge,
	totalConns metrics.Gauge,
	idleConns metrics.Gauge,
	staleConns metrics.Gauge,
) *Gauges {
	return &Gauges{
		hits:       hits,
		misses:     misses,
		timeouts:   timeouts,
		totalConns: totalConns,
		idleConns:  idleConns,
		staleConns: staleConns,
		dbName:     "default",
	}
}

// DBName sets the dbname label of redis metrics.
func (g *Gauges) DBName(dbName string) *Gauges {
	return &Gauges{
		hits:       g.hits,
		misses:     g.misses,
		timeouts:   g.timeouts,
		totalConns: g.totalConns,
		idleConns:  g.idleConns,
		staleConns: g.staleConns,
		dbName:     dbName,
	}
}

// Observe records the redis pool stats. It should be called periodically.
func (g *Gauges) Observe(stats *redis.PoolStats) {
	withValues := []string{"dbname", g.dbName}

	g.hits.With(withValues...).Set(float64(stats.Hits))

	g.misses.With(withValues...).Set(float64(stats.Misses))

	g.timeouts.With(withValues...).Set(float64(stats.Timeouts))

	g.totalConns.With(withValues...).Set(float64(stats.TotalConns))

	g.idleConns.With(withValues...).Set(float64(stats.IdleConns))

	g.staleConns.With(withValues...).Set(float64(stats.StaleConns))
}

// newCollector creates a new redis wrapper containing the name of the redis.
func newCollector(factory Factory, gauges *Gauges, interval time.Duration) *collector {
	return &collector{
		factory:  factory,
		gauges:   gauges,
		interval: interval,
	}
}

// collectConnectionStats collects redis connections for Prometheus to scrape.
func (d *collector) collectConnectionStats() {
	for k, v := range d.factory.List() {
		conn := v.Conn.(redis.UniversalClient)
		stats := conn.PoolStats()
		d.gauges.DBName(k).Observe(stats)
	}
}
