package otredis

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/internal/stub"
	"github.com/DoNewsCode/core/otredis/mocks"

	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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
	g := Gauges{
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

	c.Invoke(func(factory *Factory, g *Gauges) {
		factory.Make("default")
		c := newCollector(factory, g, time.Nanosecond)
		c.collectConnectionStats()
	})
}

func TestObserve(t *testing.T) {
	m := &stub.Gauge{}
	g := Gauges{
		Hits:       m,
		Misses:     m,
		Timeouts:   m,
		TotalConns: m,
		IdleConns:  m,
		StaleConns: m,
	}
	g.Observe(&redis.PoolStats{})
	assert.ElementsMatch(t, m.LabelValues, []string{"dbname", "default"})
}
