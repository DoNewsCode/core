package sagas

import (
	"context"

	"github.com/DoNewsCode/core/contract"
	"github.com/hashicorp/go-multierror"
)

type contextKey string

// TxContextKey is the context key for TX.
const TxContextKey contextKey = "coordinator"

// Store is the interface to persist logs of transactions.
type Store interface {
	Log(ctx context.Context, log Log) error
	Ack(ctx context.Context, id string, err error) error
	UnacknowledgedSteps(ctx context.Context, correlationID string) ([]Log, error)
	UncommittedSagas(ctx context.Context) ([]Log, error)
}

// TX is a distributed transaction coordinator. It should be initialized
// by directly assigning its public members.
type TX struct {
	store         Store
	dispatcher    contract.Dispatcher
	correlationID string
	session       Log
	rollbacks     map[string]onRollbackPayload
	undoErr       *multierror.Error
	completed     bool
}

// Commit commits the current transaction.
func (tx *TX) Commit(ctx context.Context) error {
	tx.completed = true
	return tx.store.Ack(ctx, tx.session.ID, nil)
}

// Rollback rollbacks the current transaction.
func (tx *TX) Rollback(ctx context.Context) error {
	for name, event := range tx.rollbacks {
		err := tx.dispatcher.Dispatch(ctx, onRollback(name), event)
		if err != nil {
			tx.undoErr = multierror.Append(tx.undoErr, err)
		}
	}
	tx.completed = true
	return tx.undoErr.ErrorOrNil()
}

// TxFromContext returns the tx instance from context.
func TxFromContext(ctx context.Context) *TX {
	return ctx.Value(TxContextKey).(*TX)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
