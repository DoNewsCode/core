package unierr_test

import (
	"errors"
	"fmt"
	"github.com/DoNewsCode/std/pkg/unierr"
	"google.golang.org/grpc/codes"
)

func ExampleError_StatusCode() {
	err := errors.New("my stuff is missing")
	unifiedError := unierr.Wrap(err, codes.NotFound)

	httpStatus := unifiedError.StatusCode()
	fmt.Println(httpStatus)
	bytes, _ := unifiedError.MarshalJSON()
	fmt.Println(string(bytes))
	// Output:
	// 404
	// {"code":5,"message":"my stuff is missing"}
}

func ExampleError_GRPCStatus() {
	err := errors.New("my stuff is missing")
	unifiedError := unierr.Wrap(err, codes.NotFound)

	grpcStatus := unifiedError.GRPCStatus()
	fmt.Println(grpcStatus.Code())
	fmt.Println(grpcStatus.Message())
	// Output:
	// NotFound
	// my stuff is missing
}
