package sagas

import (
	"context"
	"fmt"
	"time"

	"github.com/DoNewsCode/core/dtx"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/rs/xid"
)

// Step is a step in the Saga.
type Step struct {
	Name string
	Do   endpoint.Endpoint
	Undo endpoint.Endpoint
}

// Registry holds all transaction sagas in this process. It should be populated during the initialization of the application.
type Registry struct {
	logger  log.Logger
	Store   Store
	steps   map[string]*Step
	timeout time.Duration
}

// Option is the functional option for NewRegistry.
type Option func(registry *Registry)

// WithLogger is an option that adds a logger to the registry.
func WithLogger(logger log.Logger) Option {
	return func(registry *Registry) {
		registry.logger = logger
	}
}

// WithTimeout is an option that configures when the unacknowledged steps
// should be marked as stale and become candidates for rollback.
func WithTimeout(duration time.Duration) Option {
	return func(registry *Registry) {
		registry.timeout = duration
	}
}

// NewRegistry creates a new Registry.
func NewRegistry(store Store, opts ...Option) *Registry {
	r := &Registry{
		logger:  log.NewNopLogger(),
		Store:   store,
		timeout: 10 * time.Minute,
		steps:   make(map[string]*Step),
	}
	for _, f := range opts {
		f(r)
	}
	return r
}

// StartTX starts a transaction using saga pattern.
func (r *Registry) StartTX(ctx context.Context) (*TX, context.Context) {
	cid := xid.New().String()
	tx := &TX{
		session: Log{
			ID:            xid.New().String(),
			correlationID: cid,
			StartedAt:     time.Now(),
			LogType:       Session,
		},
		store:         r.Store,
		correlationID: cid,
		rollbacks:     make(map[string]endpoint.Endpoint),
	}
	ctx = context.WithValue(ctx, dtx.CorrelationID, cid)
	ctx = context.WithValue(ctx, TxContextKey, tx)
	must(tx.store.Log(ctx, tx.session))
	return tx, ctx
}

// AddStep registers the saga steps in the registry. The registration should be done
// during the bootstrapping of application.
func (r *Registry) AddStep(step *Step) endpoint.Endpoint {
	r.steps[step.Name] = step
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		logID := xid.New().String()
		tx := TxFromContext(ctx)
		if tx.completed {
			panic("re-executing a completed transaction")
		}
		stepLog := Log{
			ID:            logID,
			correlationID: tx.correlationID,
			StartedAt:     time.Now(),
			LogType:       Do,
			StepName:      step.Name,
			StepParam:     request,
		}
		must(tx.store.Log(ctx, stepLog))
		tx.rollbacks[step.Name] = func(ctx context.Context, _ interface{}) (response interface{}, err error) {
			logID := xid.New().String()
			compensateLog := Log{
				ID:            logID,
				correlationID: tx.correlationID,
				StartedAt:     time.Now(),
				LogType:       Undo,
				StepName:      step.Name,
				StepParam:     request,
			}
			must(tx.store.Log(ctx, compensateLog))
			resp, err := step.Undo(ctx, request)
			must(tx.store.Ack(ctx, logID, err))

			return resp, err
		}
		response, err = step.Do(ctx, request)
		must(tx.store.Ack(ctx, logID, err))
		return response, err
	}
}

// Recover rollbacks all uncommitted sagas by retrieving them in the store.
func (r *Registry) Recover(ctx context.Context) {
	logs, err := r.Store.UncommittedSagas(ctx)
	if err != nil {
		panic(err)
	}
	for _, log := range logs {
		if log.StartedAt.Add(r.timeout).After(time.Now()) {
			continue
		}
		if _, ok := r.steps[log.StepName]; !ok {
			level.Warn(r.logger).Log(
				"msg",
				fmt.Sprintf("saga step %s not registered", log.StepName),
			)
		}
		tx := TX{
			correlationID: log.correlationID,
			store:         r.Store,
		}
		ctx = context.WithValue(ctx, dtx.CorrelationID, tx.correlationID)
		logID := xid.New().String()
		compensateLog := Log{
			ID:            logID,
			correlationID: tx.correlationID,
			StartedAt:     time.Now(),
			LogType:       Undo,
			StepName:      log.StepName,
			StepParam:     log.StepParam,
		}

		must(tx.store.Log(ctx, compensateLog))
		_, err := r.steps[log.StepName].Undo(ctx, log.StepParam)
		must(tx.store.Ack(ctx, logID, err))
	}
}
