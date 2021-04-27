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

		withValues := []string{"dbname", k}
		d.gauges.Hits.
			With(withValues...).
			Set(float64(stats.Hits))

		d.gauges.Misses.
			With(withValues...).
			Set(float64(stats.Misses))

		d.gauges.Timeouts.
			With(withValues...).
			Set(float64(stats.Timeouts))

		d.gauges.TotalConns.
			With(withValues...).
			Set(float64(stats.TotalConns))
		d.gauges.IdleConns.
			With(withValues...).
			Set(float64(stats.IdleConns))
		d.gauges.StaleConns.
			With(withValues...).
			Set(float64(stats.StaleConns))
	}
}
