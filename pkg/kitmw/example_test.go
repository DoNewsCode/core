package kitmw_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/DoNewsCode/std/pkg/kitmw"
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
