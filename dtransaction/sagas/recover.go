package sagas

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type Registry struct {
	logger log.Logger
	Store  Store
	sagas  map[string]*Saga
}

func NewRegistry(store Store) *Registry {
	return &Registry{
		Store: store,
		sagas: make(map[string]*Saga),
	}
}

func (r *Registry) Register(saga *Saga) {
	r.sagas[saga.Name] = saga
}

func (r *Registry) Recover(ctx context.Context) {
	logs, err := r.Store.UncommittedSagas(ctx)
	if err != nil {
		panic(err)
	}
	for _, log := range logs {
		if _, ok := r.sagas[log.SagaName]; !ok {
			level.Warn(r.logger).Log(
				"msg",
				fmt.Sprintf("saga %s found in store but not registered in codebase", log.SagaName))
			continue
		}
		c := Coordinator{
			CorrelationId: log.CorrelationID,
			Saga:          r.sagas[log.SagaName],
			Store:         r.Store,
		}
		c.compensateStep(ctx, log)
	}
}
