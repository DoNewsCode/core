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
	idle  metrics.Gauge
	inUse metrics.Gauge
	open  metrics.Gauge

	dbName bool
	driver bool
}

// NewGauges returns a new Gauges.
func NewGauges(idle, inUse, open metrics.Gauge) *Gauges {
	return &Gauges{
		idle:  idle,
		inUse: inUse,
		open:  open,
	}
}

// DBName sets the dbname label of metrics.
func (g *Gauges) DBName(dbName string) *Gauges {
	withValues := []string{"dbname", dbName}
	return &Gauges{
		idle:   g.idle.With(withValues...),
		inUse:  g.inUse.With(withValues...),
		open:   g.open.With(withValues...),
		dbName: true,
		driver: g.driver,
	}
}

// Driver sets the driver label of metrics.
func (g *Gauges) Driver(driver string) *Gauges {
	withValues := []string{"driver", driver}
	return &Gauges{
		idle:   g.idle.With(withValues...),
		inUse:  g.inUse.With(withValues...),
		open:   g.open.With(withValues...),
		dbName: g.dbName,
		driver: true,
	}
}

// Observe records the DBStats collected. It should be called periodically.
func (g *Gauges) Observe(stats sql.DBStats) {
	if !g.dbName {
		g.idle = g.idle.With("dbname", "")
		g.inUse = g.inUse.With("dbname", "")
		g.open = g.open.With("dbname", "")
	}
	if !g.driver {
		g.idle = g.idle.With("driver", "")
		g.inUse = g.inUse.With("driver", "")
		g.open = g.open.With("driver", "")
	}
	g.idle.Set(float64(stats.Idle))
	g.inUse.Set(float64(stats.InUse))
	g.open.Set(float64(stats.OpenConnections))
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
