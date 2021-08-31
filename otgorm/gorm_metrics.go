package otgorm

import (
	"database/sql"
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

	dbName string
	driver string
}

// DBName sets the dbname label of metrics.
func (g *Gauges) DBName(dbName string) *Gauges {
	withValues := []string{"dbname", g.dbName}
	return &Gauges{
		Idle:   g.Idle.With(withValues...),
		InUse:  g.InUse.With(withValues...),
		Open:   g.Open.With(withValues...),
		dbName: dbName,
		driver: g.driver,
	}
}

// Driver sets the driver label of metrics.
func (g *Gauges) Driver(driver string) *Gauges {
	withValues := []string{"driver", driver}
	return &Gauges{
		Idle:   g.Idle.With(withValues...),
		InUse:  g.InUse.With(withValues...),
		Open:   g.Open.With(withValues...),
		dbName: g.dbName,
		driver: driver,
	}
}

// Observe records the DBStats collected. It should be called periodically.
func (g *Gauges) Observe(stats sql.DBStats) {
	g.Idle.Set(float64(stats.Idle))
	g.InUse.Set(float64(stats.InUse))
	g.Open.Set(float64(stats.OpenConnections))
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
		d.gauges.DBName(k).Driver(conn.Name()).Observe(stats)
	}
}
