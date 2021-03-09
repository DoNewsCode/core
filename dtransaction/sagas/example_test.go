package sagas_test

import (
	"context"
	"fmt"

	"github.com/DoNewsCode/core/dtransaction"
	"github.com/DoNewsCode/core/dtransaction/sagas"
)

var orderTable = make(map[string]interface{})
var paymentTable = make(map[string]interface{})

type OrderRequest struct {
	Sku string
}

type OrderResponse struct {
	OrderID string
	Sku     string
	Cost    float64
}

type PaymentRequest struct {
	OrderID string
	Sku     string
	Cost    float64
}

type PaymentResponse struct {
	Success bool
}

func orderEndpoint(ctx context.Context, request interface{}) (response interface{}, err error) {
	correlationId := ctx.Value(dtransaction.CorrelationID).(string)
	orderTable[correlationId] = request
	return OrderResponse{
		OrderID: "1",
		Sku:     "1",
		Cost:    10.0,
	}, nil
}

func orderCancelEndpoint(ctx context.Context, request interface{}) (response interface{}, err error) {
	correlationId := ctx.Value(dtransaction.CorrelationID).(string)
	delete(orderTable, correlationId)
	return nil, nil
}

func paymentEndpoint(ctx context.Context, request interface{}) (response interface{}, err error) {
	correlationId := ctx.Value(dtransaction.CorrelationID).(string)
	paymentTable[correlationId] = request
	if request.(PaymentRequest).Cost < 20 {
		return PaymentResponse{
			Success: true,
		}, nil
	}
	return PaymentResponse{
		Success: false,
	}, nil
}

func paymentCancelEndpoint(ctx context.Context) (response interface{}, err error) {
	correlationId := ctx.Value(dtransaction.CorrelationID).(string)
	delete(paymentTable, correlationId)
	return nil, nil
}

func Example() {
	store := sagas.NewInProcessStore()
	registry := sagas.NewRegistry(store)
	addOrder := registry.AddStep(&sagas.Step{
		Name: "Add Order",
		Do: func(ctx context.Context, request interface{}) (response interface{}, err error) {
			resp, err := orderEndpoint(ctx, request.(OrderRequest))
			if err != nil {
				return nil, err
			}
			// Convert the response to next request
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

	tx, ctx := registry.StartTx(context.Background())
	resp, err := addOrder(ctx, OrderRequest{Sku: "1"})
	if err != nil {
		tx.Rollback(ctx)
	}
	resp, err = makePayment(ctx, PaymentRequest{
		OrderID: resp.(OrderResponse).OrderID,
		Sku:     resp.(OrderResponse).Sku,
		Cost:    resp.(OrderResponse).Cost,
	})
	if err != nil {
		tx.Rollback(ctx)
	}
	tx.Commit(ctx)
	fmt.Println(resp.(PaymentResponse).Success)

	// Output:
	// true

}
