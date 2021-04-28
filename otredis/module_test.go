package otredis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"os"
	"testing"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/di"
	mock_metrics "github.com/DoNewsCode/core/otredis/mocks"
	"github.com/golang/mock/gomock"
)

func TestMain(m *testing.M) {
	if os.Getenv("REDIS_ADDR") == "" {
		fmt.Println("Set env REDIS_ADDR to run otredis tests")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestModule_ProvideRunGroup(t *testing.T) {
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
		core.WithInline("redis.default.addrs", []string{os.Getenv("REDIS_ADDR")}),
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
