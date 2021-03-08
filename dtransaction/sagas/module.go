package sagas

import (
	"context"
	"time"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/kit/log"
	"github.com/oklog/run"
)

// Module is a saga registry
type Module struct {
	*Registry
	RecoverInterval time.Duration
}

// In is the injection parameter for saga module.
type In struct {
	di.In

	Conf   contract.ConfigAccessor
	Logger log.Logger
	Store  Store   `optional:"true"`
	Sagas  []*Saga `group:"saga"`
}

// New creates a new saga module
func New(in In) Module {
	if in.Store == nil {
		in.Store = NewInProcessStore()
	}
	registry := NewRegistry(in.Store, WithLogger(in.Logger))
	for i := range in.Sagas {
		timeoutSec := in.Conf.Float64("sagas.defaultSagaTimeoutSecond")
		if timeoutSec == 0 {
			timeoutSec = 360
		}
		timeout := time.Duration(timeoutSec) * time.Second
		if in.Sagas[i].Timeout == 0 {
			in.Sagas[i].Timeout = timeout
		}
		registry.Register(in.Sagas[i])
	}
	recoverSec := in.Conf.Float64("sagas.recoverIntervalSecond")
	if recoverSec == 0 {
		recoverSec = 60
	}
	return Module{Registry: registry, RecoverInterval: time.Duration(recoverSec) * time.Second}
}

// ProvideRunGroup implements the RunProvider
func (m Module) ProvideRunGroup(group *run.Group) {
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(m.RecoverInterval)
	group.Add(func() error {
		m.Registry.Recover(ctx)
		for {
			select {
			case <-ticker.C:
				m.Registry.Recover(ctx)
			case <-ctx.Done():
				return nil
			}
		}
	}, func(err error) {
		cancel()
		ticker.Stop()
	})
}
