package kitmw_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DoNewsCode/core/key"
	"github.com/DoNewsCode/core/kitmw"
	"github.com/DoNewsCode/core/logging"
	"github.com/go-kit/kit/endpoint"
)

func ExampleMakeErrorConversionMiddleware() {
	var (
		err      error
		original endpoint.Endpoint
		wrapped  endpoint.Endpoint
	)
	original = func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return nil, errors.New("error")
	}
	_, err = original(context.Background(), nil)
	fmt.Printf("%T\n", err)

	wrapped = kitmw.MakeErrorConversionMiddleware(kitmw.ErrorOption{})(original)

	_, err = wrapped(context.Background(), nil)
	fmt.Printf("%T\n", err)

	// Output:
	// *errors.errorString
	// *unierr.Error
}

func ExampleMakeLoggingMiddleware() {
	var (
		original endpoint.Endpoint
		wrapped  endpoint.Endpoint
	)
	original = func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return "respData", nil
	}

	wrapped = kitmw.MakeLoggingMiddleware(
		logging.NewLogger("json"),
		key.New(),
		false,
	)(original)

	wrapped(context.Background(), "reqData")

	// Output:
	// {"request":"reqData","response":"respData"}
}

func ExampleMakeRetryMiddleware() {
	var (
		original endpoint.Endpoint
		wrapped  endpoint.Endpoint
	)
	original = func(ctx context.Context, request interface{}) (response interface{}, err error) {
		fmt.Println("attempt")
		return nil, errors.New("")
	}

	wrapped = kitmw.MakeRetryMiddleware(5, time.Hour)(original)

	wrapped(context.Background(), nil)

	// Output:
	// attempt
	// attempt
	// attempt
	// attempt
	// attempt
}

func ExampleMakeTimeoutMiddleware() {
	var (
		original endpoint.Endpoint
		wrapped  endpoint.Endpoint
	)
	original = func(ctx context.Context, request interface{}) (response interface{}, err error) {
		select {
		case <-ctx.Done():
			return nil, errors.New("timeout")
		case <-time.After(100000 * time.Microsecond):
			return nil, nil
		}
	}

	wrapped = kitmw.MakeTimeoutMiddleware(time.Microsecond)(original)
	_, err := wrapped(context.Background(), nil)
	fmt.Println(err)

	// Output:
	// timeout
}
