package sagas_test

import (
	"context"
	"fmt"

	"github.com/DoNewsCode/core/dtx"
	"github.com/DoNewsCode/core/dtx/sagas"
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
	correlationID := ctx.Value(dtx.CorrelationID).(string)
	orderTable[correlationID] = request
	return OrderResponse{
		OrderID: "1",
		Sku:     "1",
		Cost:    10.0,
	}, nil
}

func orderCancelEndpoint(ctx context.Context, request interface{}) (err error) {
	correlationID := ctx.Value(dtx.CorrelationID).(string)
	delete(orderTable, correlationID)
	return nil
}

func paymentEndpoint(ctx context.Context, request interface{}) (response interface{}, err error) {
	correlationID := ctx.Value(dtx.CorrelationID).(string)
	paymentTable[correlationID] = request
	if request.(PaymentRequest).Cost < 20 {
		return PaymentResponse{
			Success: true,
		}, nil
	}
	return PaymentResponse{
		Success: false,
	}, nil
}

func paymentCancelEndpoint(ctx context.Context) (err error) {
	correlationID := ctx.Value(dtx.CorrelationID).(string)
	delete(paymentTable, correlationID)
	return nil
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
		Undo: func(ctx context.Context, req interface{}) (err error) {
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
		Undo: func(ctx context.Context, req interface{}) (err error) {
			return paymentCancelEndpoint(ctx)
		},
	})

	tx, ctx := registry.StartTX(context.Background())
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
		_ = tx.Rollback(ctx)
	}
	tx.Commit(ctx)
	fmt.Println(resp.(PaymentResponse).Success)

	// Output:
	// true

}
