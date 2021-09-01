package ctxmeta_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DoNewsCode/core/ctxmeta"
)

// This example demonstrates how to use Unmarshal to retrieve metadata into an
// arbitrary type.
func Example_Unmarshal() {
	type DomainError struct {
		Code   int
		Reason string
	}

	bag, _ := ctxmeta.Inject(context.Background())
	derr := DomainError{Code: http.StatusTeapot, Reason: "Earl Gray exception"}
	bag.Set("err", derr)

	if target := (DomainError{}); bag.Unmarshal("err", &target) == nil {
		fmt.Printf("DomainError Code=%d Reason=%q\n", target.Code, target.Reason)
	}

	// Output:
	// DomainError Code=418 Reason="Earl Gray exception"
}
