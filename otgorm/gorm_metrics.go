package otgorm

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"gorm.io/gorm"
)

type collector struct {
	factory  Factory
	gauges   *Gauges
	interval time.Duration
}

// Gauges is a collection of metrics for database connection info.
type Gauges struct {
	Idle  metrics.Gauge
	InUse metrics.Gauge
	Open  metrics.Gauge
}

// newCollector creates a new database wrapper containing the name of the database,
// it's driver and the (sql) database itself.
func newCollector(factory Factory, gauges *Gauges, interval time.Duration) *collector {
	return &collector{
		factory:  factory,
		gauges:   gauges,
		interval: interval,
	}
}

// collectConnectionStats collects database connections for Prometheus to scrape.
func (d *collector) collectConnectionStats() {
	for k, v := range d.factory.List() {
		conn := v.Conn.(*gorm.DB)
		db, _ := conn.DB()
		stats := db.Stats()
		withValues := []string{"dbname", k, "driver", conn.Name()}
		d.gauges.Idle.
			With(withValues...).
			Set(float64(stats.Idle))

		d.gauges.InUse.
			With(withValues...).
			Set(float64(stats.InUse))

		d.gauges.Open.
			With(withValues...).
			Set(float64(stats.OpenConnections))
	}
}
