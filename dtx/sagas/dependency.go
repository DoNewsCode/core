package sagas

import (
	"context"
	"time"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/oklog/run"
)

/*
Providers returns a set of dependency providers.
	Depends On:
		contract.ConfigAccessor
		log.Logger
		Store   `optional:"true"`
		[]*Step `group:"saga"`
	Provide:
		*Registry
		SagaEndpoints
*/
func Providers() di.Deps {
	return []interface{}{provide, provideConfig}
}

// in is the injection parameter for saga module.
type in struct {
	di.In

	Conf   contract.ConfigAccessor
	Logger log.Logger
	Store  Store   `optional:"true"`
	Steps  []*Step `group:"saga"`
}

type recoverInterval time.Duration

// SagaEndpoints is a collection of all registered endpoint in the saga registry
type SagaEndpoints map[string]endpoint.Endpoint

type out struct {
	di.Out

	Registry      *Registry
	Interval      recoverInterval
	SagaEndpoints SagaEndpoints
}

// provide creates a new saga module.
func provide(in in) out {
	if in.Store == nil {
		in.Store = NewInProcessStore()
	}
	var conf configuration
	err := in.Conf.Unmarshal("sagas", &conf)
	if err != nil {
		level.Warn(in.Logger).Log("err", err)
	}
	timeout := conf.getSagaTimeout().Duration
	recoverVal := conf.getRecoverInterval().Duration

	registry := NewRegistry(
		in.Store,
		WithLogger(in.Logger),
		WithTimeout(timeout),
	)
	eps := make(SagaEndpoints)

	for i := range in.Steps {
		eps[in.Steps[i].Name] = registry.AddStep(in.Steps[i])
	}

	return out{Registry: registry, Interval: recoverInterval(recoverVal), SagaEndpoints: eps}
}

// ProvideRunGroup implements the RunProvider.
func (m out) ProvideRunGroup(group *run.Group) {
	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(time.Duration(m.Interval))
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

func (m out) ModuleSentinel() {}
