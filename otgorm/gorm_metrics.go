package otgorm

import (
	"database/sql"
	"time"

	"github.com/go-kit/kit/metrics"
)

type collector struct {
	factory  *Factory
	gauges   *Gauges
	interval time.Duration
}

// Gauges is a collection of metrics for database connection info.
type Gauges struct {
	idle  metrics.Gauge
	inUse metrics.Gauge
	open  metrics.Gauge

	dbName string
	driver string
}

// NewGauges returns a new Gauges.
func NewGauges(idle, inUse, open metrics.Gauge) *Gauges {
	return &Gauges{
		idle:   idle,
		inUse:  inUse,
		open:   open,
		dbName: "unknown",
		driver: "default",
	}
}

// DBName sets the dbname label of metrics.
func (g *Gauges) DBName(dbName string) *Gauges {
	return &Gauges{
		idle:   g.idle,
		inUse:  g.inUse,
		open:   g.open,
		dbName: dbName,
		driver: g.driver,
	}
}

// Driver sets the driver label of metrics.
func (g *Gauges) Driver(driver string) *Gauges {
	return &Gauges{
		idle:   g.idle,
		inUse:  g.inUse,
		open:   g.open,
		dbName: g.dbName,
		driver: driver,
	}
}

// Observe records the DBStats collected. It should be called periodically.
func (g *Gauges) Observe(stats sql.DBStats) {
	g.idle.With("dbname", g.dbName, "driver", g.driver).Set(float64(stats.Idle))
	g.inUse.With("dbname", g.dbName, "driver", g.driver).Set(float64(stats.InUse))
	g.open.With("dbname", g.dbName, "driver", g.driver).Set(float64(stats.OpenConnections))
}

// newCollector creates a new database wrapper containing the name of the database,
// it's driver and the (sql) database itself.
func newCollector(factory *Factory, gauges *Gauges, interval time.Duration) *collector {
	return &collector{
		factory:  factory,
		gauges:   gauges,
		interval: interval,
	}
}

// collectConnectionStats collects database connections for Prometheus to scrape.
func (d *collector) collectConnectionStats() {
	for k, v := range d.factory.List() {
		conn := v.Conn
		db, _ := conn.DB()
		stats := db.Stats()
		d.gauges.DBName(k).Driver(conn.Name()).Observe(stats)
	}
}
