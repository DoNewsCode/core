package kitmw_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/DoNewsCode/std/pkg/kitmw"
	"github.com/go-kit/kit/endpoint"
)

func ExampleMakeErrorConversionMiddleware() {
	var ep endpoint.Endpoint = func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return nil, errors.New("f")
	}
	wrapped := kitmw.MakeErrorConversionMiddleware(kitmw.ErrorOption{})(ep)
	_, err := wrapped(context.Background(), nil)
	fmt.Printf("%T", err)
	// Output:
	// *unierr.Error
}
