//go:generate mockgen -destination=./mocks/metrics_test.go github.com/go-kit/kit/metrics Gauge

package otredis

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-redis/redis/v8"
)

type collector struct {
	factory  *Factory
	gauges   *Gauges
	interval time.Duration
}

// Gauges is a collection of metrics for redis connection info.
type Gauges struct {
	Hits       metrics.Gauge
	Misses     metrics.Gauge
	Timeouts   metrics.Gauge
	TotalConns metrics.Gauge
	IdleConns  metrics.Gauge
	StaleConns metrics.Gauge

	dbName *string
}

// DBName sets the dbname label of redis metrics.
func (g *Gauges) DBName(dbName string) *Gauges {
	return &Gauges{
		Hits:       g.Hits,
		Misses:     g.Misses,
		Timeouts:   g.Timeouts,
		TotalConns: g.TotalConns,
		IdleConns:  g.IdleConns,
		StaleConns: g.StaleConns,
		dbName:     &dbName,
	}
}

// Observe records the redis pool stats. It should be called periodically.
func (g *Gauges) Observe(stats *redis.PoolStats) {
	withValues := []string{"dbname", "default"}
	if g.dbName != nil {
		withValues = []string{"dbname", *g.dbName}
	}

	g.Hits.With(withValues...).Set(float64(stats.Hits))
	g.Misses.With(withValues...).Set(float64(stats.Misses))
	g.Timeouts.With(withValues...).Set(float64(stats.Timeouts))
	g.TotalConns.With(withValues...).Set(float64(stats.TotalConns))
	g.IdleConns.With(withValues...).Set(float64(stats.IdleConns))
	g.StaleConns.With(withValues...).Set(float64(stats.StaleConns))
}

// newCollector creates a new redis wrapper containing the name of the redis.
func newCollector(factory *Factory, gauges *Gauges, interval time.Duration) *collector {
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
