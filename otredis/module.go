package otredis

import (
	"context"
	"time"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/kit/log"
	"github.com/oklog/run"
)

const defaultInterval = 15 * time.Second

// Module is the registration unit for package core.
type Module struct {
	maker     Maker
	env       contract.Env
	logger    log.Logger
	container contract.Container
	collector *collector
	interval  time.Duration
}

// ModuleIn contains the input parameters needed for creating the new module.
type ModuleIn struct {
	di.In

	Maker     Maker
	Env       contract.Env
	Logger    log.Logger
	Container contract.Container
	Collector *collector
	Conf      contract.ConfigAccessor
}

// New creates a Module.
func New(in ModuleIn) Module {
	var duration time.Duration = defaultInterval
	in.Conf.Unmarshal("redisMetrics.interval", &duration)
	return Module{
		maker:     in.Maker,
		env:       in.Env,
		logger:    in.Logger,
		container: in.Container,
		collector: in.Collector,
		interval:  duration,
	}
}

// ProvideRunGroup add a goroutine to periodically scan redis connections and
// report them to metrics collector such as prometheus.
func (m Module) ProvideRunGroup(group *run.Group) {
	if m.collector == nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(m.interval)
	group.Add(func() error {
		for {
			select {
			case <-ticker.C:
				m.collector.collectConnectionStats()
			case <-ctx.Done():
				ticker.Stop()
				return nil
			}
		}
	}, func(err error) {
		cancel()
	})
}
