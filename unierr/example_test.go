package unierr_test

import (
	"errors"
	"fmt"

	"github.com/DoNewsCode/core/unierr"

	"google.golang.org/grpc/codes"
)

func Example() {
	fmt.Printf("Default status conversion:\n")
	for i := 1; i < 17; i++ {
		err := unierr.Wrap(errors.New(""), codes.Code(i))
		fmt.Printf("GRPC %d <=> HTTP: %d\n", err.GRPCStatus().Code(), err.StatusCode())
	}
	// Output:
	// Default status conversion:
	// GRPC 1 <=> HTTP: 499
	// GRPC 2 <=> HTTP: 500
	// GRPC 3 <=> HTTP: 400
	// GRPC 4 <=> HTTP: 504
	// GRPC 5 <=> HTTP: 404
	// GRPC 6 <=> HTTP: 409
	// GRPC 7 <=> HTTP: 403
	// GRPC 8 <=> HTTP: 429
	// GRPC 9 <=> HTTP: 400
	// GRPC 10 <=> HTTP: 409
	// GRPC 11 <=> HTTP: 400
	// GRPC 12 <=> HTTP: 501
	// GRPC 13 <=> HTTP: 500
	// GRPC 14 <=> HTTP: 500
	// GRPC 15 <=> HTTP: 500
	// GRPC 16 <=> HTTP: 401
}

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
