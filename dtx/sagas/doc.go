/*
Package sagas implements the orchestration based saga pattern.
See https://microservices.io/patterns/data/saga.html

Introduction

A saga is a sequence of local transactions. Each local transaction updates the
database and publishes a message or event to trigger the next local
transaction in the saga. If a local transaction fails because it violates a
business rule then the saga executes a series of compensating transactions
that undo the changes that were made by the preceding local transactions.

Usage

The saga is managed by sagas.Registry. Each saga step has an forward operation
and a rollback counterpart. They must be registered beforehand by calling
Registry.AddStep. A new endpoint will be returned to the caller. Use the
returned endpoint to perform transactional operation.

	store := sagas.NewInProcessStore()
	registry := sagas.NewRegistry(store)
	addOrder := registry.AddStep(&sagas.Step{
		Name: "Add Order",
		Do: func(ctx context.Context, request interface{}) (response interface{}, err error) {
			resp, err := orderEndpoint(ctx, request.(OrderRequest))
			if err != nil {
				return nil, err
			}
			return resp, nil
		},
		Undo: func(ctx context.Context, req interface{}) (response interface{}, err error) {
			return orderCancelEndpoint(ctx, req)
		},
	})
	makePayment := registry.AddStep(&sagas.Step{
		Name: "Make Payment",
		Do: func(ctx context.Context, request interface{}) (response interface{}, err error) {
			resp, err := paymentEndpoint(ctx, request.(PaymentRequest))
			if err != nil {
				return nil, err
			}
			return resp, nil
		},
		Undo: func(ctx context.Context, req interface{}) (response interface{}, err error) {
			return paymentCancelEndpoint(ctx)
		},
	})

Initiate the transaction by calling registry.StartTX. Pass the context returned
to the transaction branches. You can rollback or commit at your will. If the
TX.Rollback is called, the previously registered rollback operations will be
applied automatically, on condition that the forward operation is indeed
executed within the transaction.

	tx, ctx := registry.StartTX(context.Background())
	resp, err := addOrder(ctx, OrderRequest{Sku: "1"})
	if err != nil {
		tx.Rollback(ctx)
	}
	resp, err = makePayment(ctx, PaymentRequest{})
	if err != nil {
		tx.Rollback(ctx)
	}
	tx.Commit(ctx)

Integration

The package leader exports configuration in this format:

	saga:
		sagaTimeoutSecond: 600
		recoverIntervalSecond: 60

To use package sagas with package core:

	var c *core.C = core.Default()
	c.Provide(sagas.Providers)
	c.Invoke(func(registry *sagas.Registry) {
		tx, ctx := registry.StartTX(context.Background())
		resp, err := addOrder(ctx, OrderRequest{Sku: "1"})
		if err != nil {
			tx.Rollback(ctx)
		}
		resp, err = makePayment(ctx, PaymentRequest{})
		if err != nil {
			tx.Rollback(ctx)
		}
		tx.Commit(ctx)
	})
*/
package sagas
