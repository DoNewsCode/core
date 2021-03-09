package sagas

import (
	"context"
	"fmt"
	"time"

	"github.com/DoNewsCode/core/dtransaction"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// Registry holds all transaction sagas in this process. It should be populated during the initialization of the application.
type Registry struct {
	logger log.Logger
	Store  Store
	sagas  map[string]*Saga
	//steps  map[string]*Step
}

// Option is the functional option for NewRegistry.
type Option func(registry *Registry)

// WithLogger is an option that adds a logger to the registry.
func WithLogger(logger log.Logger) Option {
	return func(registry *Registry) {
		registry.logger = logger
	}
}

// NewRegistry creates a new Registry.
func NewRegistry(store Store, opts ...Option) *Registry {
	r := &Registry{
		logger: log.NewNopLogger(),
		Store:  store,
		sagas:  make(map[string]*Saga),
		//steps:  make(map[string]*Step),
	}
	for _, f := range opts {
		f(r)
	}
	return r
}

// Register registers the saga in the registry. The registration should be done
// during the bootstrapping of application.
func (r *Registry) Register(saga *Saga) {
	r.sagas[saga.Name] = saga
}

//// Register registers the saga in the registry. The registration should be done
//// during the bootstrapping of application.
//func (r *Registry) AddStep(saga *Step) endpoint.Endpoint {
//	r.sagas.
//}

// Recover rollbacks all uncommitted sagas by retrieving them in the store.
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
		if log.StartedAt.Add(r.sagas[log.SagaName].Timeout).After(time.Now()) {
			continue
		}
		c := Coordinator{
			correlationId: log.CorrelationID,
			Saga:          r.sagas[log.SagaName],
			Store:         r.Store,
		}
		ctx = context.WithValue(ctx, dtransaction.CorrelationID, c.correlationId)
		c.compensateStep(ctx, log)
	}
}
