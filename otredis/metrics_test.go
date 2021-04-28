//go:generate mockery --name=Gauge
package otredis

import (
	"testing"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/otredis/mocks"
	"github.com/golang/mock/gomock"
)

func TestCollector(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock_metrics.NewMockGauge(ctrl)
	var g = Gauges{
		Hits:       m,
		Misses:     m,
		Timeouts:   m,
		TotalConns: m,
		IdleConns:  m,
		StaleConns: m,
	}
	m.EXPECT().With(gomock.Any()).MinTimes(1).Return(m)
	m.EXPECT().Set(gomock.Any()).MinTimes(1)

	c := core.New()
	c.ProvideEssentials()
	c.Provide(Providers())
	c.Provide(di.Deps{func() *Gauges { return &g }})

	c.Invoke(func(factory Factory, g *Gauges) {
		factory.Make("default")
		c := newCollector(factory, g, time.Nanosecond)
		c.collectConnectionStats()
	})
}
