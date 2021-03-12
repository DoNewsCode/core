package sagas

import (
	"context"
	"encoding/json"
	"time"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/dtx"
	"github.com/DoNewsCode/core/events"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/rs/xid"
)

// Step is a step in the Saga.
type Step struct {
	Name        string
	Do          func(context.Context, interface{}) (interface{}, error)
	Undo        func(ctx context.Context, req interface{}) error
	EncodeParam func(interface{}) ([]byte, error)
	DecodeParam func([]byte) (interface{}, error)
}

// Codec is an interface for serialization and deserialization.
type Codec interface {
	Unmarshal(data []byte, value interface{}) error
	Marshal(value interface{}) ([]byte, error)
}

// Registry holds all transaction sagas in this process. It should be populated during the initialization of the application.
type Registry struct {
	logger     log.Logger
	Store      Store
	timeout    time.Duration
	codec      Codec
	dispatcher contract.Dispatcher
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
		dispatcher: &events.SyncDispatcher{},
		logger:     log.NewNopLogger(),
		Store:      store,
		timeout:    10 * time.Minute,
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
		dispatcher:    r.dispatcher,
		correlationID: cid,
		rollbacks:     make(map[string]rollbackEvent),
	}
	ctx = context.WithValue(ctx, dtx.CorrelationID, cid)
	ctx = context.WithValue(ctx, TxContextKey, tx)
	must(tx.store.Log(ctx, tx.session))
	return tx, ctx
}

// AddStep registers the saga steps in the registry. The registration should be done
// during the bootstrapping of application.
func (r *Registry) AddStep(step *Step) func(context.Context, interface{}) (interface{}, error) {
	r.dispatcher.Subscribe(events.Listen(
		[]contract.Event{rollbackEvent{name: step.Name, request: []byte{}}},
		func(ctx context.Context, event contract.Event) error {
			request := event.Data()
			logID := xid.New().String()
			tx := TxFromContext(ctx)

			compensateLog := Log{
				ID:            logID,
				correlationID: tx.correlationID,
				StartedAt:     time.Now(),
				LogType:       Undo,
				StepName:      step.Name,
				StepParam:     event.Data(),
			}
			if _, ok := event.Data().([]byte); step.DecodeParam != nil && ok {
				var err error
				request, err = step.DecodeParam(event.Data().([]byte))
				if err != nil {
					return errors.Wrap(err, "unable to encode step parameter")
				}
			}
			must(tx.store.Log(ctx, compensateLog))
			err := step.Undo(ctx, request)
			must(tx.store.Ack(ctx, logID, err))

			return err
		}))
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		data := request
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
		if step.EncodeParam != nil {
			data, err = step.EncodeParam(request)
			if err != nil {
				return nil, errors.Wrap(err, "unable to encode step parameter")
			}
			stepLog.StepParam = data
		}

		must(tx.store.Log(ctx, stepLog))
		tx.rollbacks[step.Name] = rollbackEvent{name: step.Name, request: data}
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
		tx := TX{
			correlationID: log.correlationID,
			store:         r.Store,
		}
		ctx = context.WithValue(ctx, dtx.CorrelationID, tx.correlationID)
		ctx = context.WithValue(ctx, TxContextKey, &tx)
		_ = r.dispatcher.Dispatch(ctx, rollbackEvent{name: log.StepName, request: log.StepParam})
	}
}

func (r *Registry) marshal(param interface{}) ([]byte, error) {
	if r.codec != nil {
		return r.codec.Marshal(param)
	}
	return json.Marshal(param)
}

func (r *Registry) unmarshal(bytes []byte, param interface{}) error {
	if r.codec != nil {
		return r.codec.Unmarshal(bytes, param)
	}
	return json.Unmarshal(bytes, param)
}
