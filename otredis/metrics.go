//go:generate mockgen -destination=./mocks/metrics.go github.com/go-kit/kit/metrics Gauge

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
	Hits       metrics.Gauge
	Misses     metrics.Gauge
	Timeouts   metrics.Gauge
	TotalConns metrics.Gauge
	IdleConns  metrics.Gauge
	StaleConns metrics.Gauge

	dbName string
}

// DBName sets the dbname label of redis metrics.
func (g *Gauges) DBName(dbName string) *Gauges {
	withValues := []string{"dbname", dbName}
	return &Gauges{
		Hits:       g.Hits.With(withValues...),
		Misses:     g.Misses.With(withValues...),
		Timeouts:   g.Timeouts.With(withValues...),
		TotalConns: g.TotalConns.With(withValues...),
		IdleConns:  g.IdleConns.With(withValues...),
		StaleConns: g.StaleConns.With(withValues...),
		dbName:     dbName,
	}
}

// Observe records the redis pool stats. It should be called periodically.
func (g *Gauges) Observe(stats *redis.PoolStats) {

	g.Hits.Set(float64(stats.Hits))

	g.Misses.Set(float64(stats.Misses))

	g.Timeouts.Set(float64(stats.Timeouts))

	g.TotalConns.Set(float64(stats.TotalConns))

	g.IdleConns.Set(float64(stats.IdleConns))

	g.StaleConns.Set(float64(stats.StaleConns))
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
