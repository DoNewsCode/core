package sagas

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

type contextKey string

// TxContextKey is the context key for TX.
const TxContextKey contextKey = "coordinator"

// Store is the interface to persist logs of transactions.
type Store interface {
	Log(ctx context.Context, log Log) error
	Ack(ctx context.Context, id string, err error) error
	UnacknowledgedSteps(ctx context.Context, correlationId string) ([]Log, error)
	UncommittedSagas(ctx context.Context) ([]Log, error)
}

// TX is a distributed transaction coordinator. It should be initialized
// by directly assigning its public members.
type TX struct {
	store         Store
	correlationId string
	session       Log
	rollbacks     map[string]endpoint.Endpoint
	doErr         []error
	undoErr       []error
	completed     bool
}

// Commit commits the current transaction.
func (tx *TX) Commit(ctx context.Context) {
	must(tx.store.Ack(ctx, tx.session.ID, nil))
	tx.completed = true
}

// Rollback rollbacks the current transaction.
func (tx *TX) Rollback(ctx context.Context) error {
	for _, call := range tx.rollbacks {
		_, err := call(ctx, nil)
		if err != nil {
			tx.undoErr = append(tx.undoErr, err)
		}
	}
	tx.completed = true
	if len(tx.undoErr) >= 1 {
		return &Result{DoErr: tx.doErr, UndoErr: tx.undoErr}
	}
	return nil
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
