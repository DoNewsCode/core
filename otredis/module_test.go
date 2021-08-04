package otredis

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/di"
	mock_metrics "github.com/DoNewsCode/core/otredis/mocks"
	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
)

func TestModule_ProvideRunGroup(t *testing.T) {
	if os.Getenv("REDIS_ADDR") == "" {
		t.Skip("set REDIS_ADDR to run TestModule_ProvideRunGroup")
		return
	}
	addrs := strings.Split(os.Getenv("REDIS_ADDR"), ",")
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	withValues := []interface{}{
		gomock.Eq("dbname"),
		gomock.Eq("default"),
	}

	m := mock_metrics.NewMockGauge(ctrl)
	m.EXPECT().With(withValues...).Return(m).MinTimes(1)
	m.EXPECT().Set(gomock.Any()).MinTimes(1)

	c := core.New(
		core.WithInline("redis.default.addrs", addrs),
		core.WithInline("redisMetrics.interval", "1ms"),
		core.WithInline("log.level", "none"),
	)
	c.ProvideEssentials()
	c.Provide(di.Deps{func() *Gauges {
		return &Gauges{
			Hits:       m,
			Misses:     m,
			Timeouts:   m,
			TotalConns: m,
			IdleConns:  m,
			StaleConns: m,
		}
	}})
	c.Provide(Providers())
	c.AddModuleFunc(New)

	ctx, cancel := context.WithCancel(context.Background())

	c.Invoke(func(cli redis.UniversalClient) {
		cli.ClientID(ctx)
	})

	go c.Serve(ctx)
	<-time.After(10 * time.Millisecond)
	cancel()
	<-time.After(1000 * time.Millisecond)
}
