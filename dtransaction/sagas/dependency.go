package sagas

import (
	"context"
	"time"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
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
		recoverInterval
		SagaEndpoints
*/
func Providers() di.Deps {
	return []interface{}{provide}
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
	di.Module

	Registry      *Registry
	Interval      recoverInterval
	SagaEndpoints SagaEndpoints
}

// provide creates a new saga module.
func provide(in in) out {
	if in.Store == nil {
		in.Store = NewInProcessStore()
	}
	timeoutSec := in.Conf.Float64("sagas.defaultSagaTimeoutSecond")
	if timeoutSec == 0 {
		timeoutSec = 600
	}
	registry := NewRegistry(
		in.Store,
		WithLogger(in.Logger),
		WithTimeout(time.Duration(timeoutSec)*time.Second),
	)
	eps := make(SagaEndpoints)

	for i := range in.Steps {
		eps[in.Steps[i].Name] = registry.AddStep(in.Steps[i])
	}

	recoverSec := in.Conf.Float64("sagas.recoverIntervalSecond")
	if recoverSec == 0 {
		recoverSec = 60
	}
	return out{Registry: registry, Interval: recoverInterval(time.Duration(recoverSec) * time.Second), SagaEndpoints: eps}
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
