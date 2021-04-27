package observability

import (
	"sync"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/otgorm"
	"github.com/DoNewsCode/core/otredis"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

type histogram struct {
	once sync.Once
	*prometheus.Histogram
}

var his histogram

// ProvideHistogramMetrics returns a metrics.Histogram that is designed to measure incoming requests
// to the system. Note it has three labels: "module", "service", "method". If any label is missing,
// the system will panic.
func ProvideHistogramMetrics(appName contract.AppName, env contract.Env) metrics.Histogram {
	his.once.Do(func() {
		his.Histogram = prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
			Namespace: appName.String(),
			Subsystem: env.String(),
			Name:      "request_duration_seconds",
			Help:      "Total time spent serving requests.",
		}, []string{"module", "service", "method"})
	})
	return &his
}

// ProvideGORMMetrics returns a *otgorm.Gauges that measures the connection info in databases.
// It is meant to be consumed by the otgorm.Providers.
func ProvideGORMMetrics(appName contract.AppName, env contract.Env) *otgorm.Gauges {
	return &otgorm.Gauges{
		Idle: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: appName.String(),
			Subsystem: env.String(),
			Name:      "gorm_idle_connections",
			Help:      "number of idle connections",
		}, []string{"dbname", "driver"}),
		Open: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: appName.String(),
			Subsystem: env.String(),
			Name:      "gorm_open_connections",
			Help:      "number of open connections",
		}, []string{"dbname", "driver"}),
		InUse: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: appName.String(),
			Subsystem: env.String(),
			Name:      "gorm_in_use_connections",
			Help:      "number of in use connections",
		}, []string{"dbname", "driver"}),
	}
}

// ProvideRedisMetrics returns a *otredis.Gauges that measures the connection info in redis.
// It is meant to be consumed by the otredis.Providers.
func ProvideRedisMetrics(appName contract.AppName, env contract.Env) *otredis.Gauges {
	return &otredis.Gauges{
		Hits: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: appName.String(),
			Subsystem: env.String(),
			Name:      "redis_hit_connections",
			Help:      "number of times free connection was found in the pool",
		}, []string{"dbname"}),
		Misses: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: appName.String(),
			Subsystem: env.String(),
			Name:      "redis_miss_connections",
			Help:      "number of times free connection was NOT found in the pool",
		}, []string{"dbname"}),
		Timeouts: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: appName.String(),
			Subsystem: env.String(),
			Name:      "redis_timeout_connections",
			Help:      "number of times a wait timeout occurred",
		}, []string{"dbname"}),
		TotalConns: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: appName.String(),
			Subsystem: env.String(),
			Name:      "redis_total_connections",
			Help:      "number of total connections in the pool",
		}, []string{"dbname"}),
		IdleConns: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: appName.String(),
			Subsystem: env.String(),
			Name:      "redis_idle_connections",
			Help:      "number of idle connections in the pool",
		}, []string{"dbname"}),
		StaleConns: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: appName.String(),
			Subsystem: env.String(),
			Name:      "redis_stale_connections",
			Help:      "number of stale connections removed from the pool",
		}, []string{"dbname"}),
	}
}
