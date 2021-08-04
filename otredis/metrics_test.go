package otredis

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/otredis/mocks"
	"github.com/golang/mock/gomock"
)

func TestCollector(t *testing.T) {
	if os.Getenv("REDIS_ADDR") == "" {
		t.Skip("set REDIS_ADDR to run TestCollector")
		return
	}
	addrs := strings.Split(os.Getenv("REDIS_ADDR"), ",")

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

	c := core.New(
		core.WithInline("redis.default.addrs", addrs),
		core.WithInline("redisMetrics.interval", "1ms"),
		core.WithInline("log.level", "none"),
	)
	c.ProvideEssentials()
	c.Provide(Providers())
	c.Provide(di.Deps{func() *Gauges { return &g }})

	c.Invoke(func(factory Factory, g *Gauges) {
		factory.Make("default")
		c := newCollector(factory, g, time.Nanosecond)
		c.collectConnectionStats()
	})
}
