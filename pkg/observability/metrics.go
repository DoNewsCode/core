package observability

import (
	"sync"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

type histogram struct {
	once sync.Once
	*prometheus.Histogram
}

var his histogram

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
