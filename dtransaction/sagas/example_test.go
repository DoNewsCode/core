package sagas_test

import (
	"context"
	"fmt"
	"time"

	"github.com/DoNewsCode/core/dtransaction/sagas"
	"github.com/go-kit/kit/auth/jwt"
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
	correlationId := ctx.Value(sagas.CorrelationId).(string)
	orderTable[correlationId] = request
	return OrderResponse{
		OrderID: "1",
		Sku:     "1",
		Cost:    10.0,
	}, nil
}

func orderCancelEndpoint(ctx context.Context) (err error) {
	correlationId := ctx.Value(sagas.CorrelationId).(string)
	delete(orderTable, correlationId)
	return nil
}

func paymentEndpoint(ctx context.Context, request interface{}) (response interface{}, err error) {
	correlationId := ctx.Value(sagas.CorrelationId).(string)
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

func paymentCancelEndpoint(ctx context.Context) (err error) {
	correlationId := ctx.Value(sagas.CorrelationId).(string)
	delete(paymentTable, correlationId)
	return nil
}

func Example() {
	saga := &sagas.Saga{
		Name:    "example",
		Timeout: 10 * time.Second,
		Steps: []*sagas.Step{
			{
				Name: "Add Order",
				Do: func(ctx context.Context, request interface{}) (response interface{}, err error) {
					resp, err := orderEndpoint(ctx, request.(OrderRequest))
					if err != nil {
						return nil, err
					}
					jwt.ContextToGRPC()
					// Convert the response to next request
					return PaymentRequest{
						OrderID: resp.(OrderResponse).OrderID,
						Sku:     resp.(OrderResponse).Sku,
						Cost:    resp.(OrderResponse).Cost,
					}, nil
				},
				Undo: func(ctx context.Context) error {
					return orderCancelEndpoint(ctx)
				},
			},
			{
				Name: "Make Payment",
				Do: func(ctx context.Context, request interface{}) (response interface{}, err error) {
					resp, err := paymentEndpoint(ctx, request.(PaymentRequest))
					if err != nil {
						return nil, err
					}
					return resp, nil
				},
				Undo: func(ctx context.Context) error {
					return paymentCancelEndpoint(ctx)
				},
			},
		},
	}

	c := sagas.Coordinator{
		Saga:  saga,
		Store: sagas.NewInProcessStore(),
	}

	resp, _ := c.Execute(context.Background(), OrderRequest{Sku: "1"})
	fmt.Println(resp.(PaymentResponse).Success)
	// Output:
	// true

}
