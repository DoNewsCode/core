package dtransaction

import (
	"context"
	"net/http"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestHTTPToContext(t *testing.T) {
	reqFunc := HTTPToContext()

	// When the header doesn't exist
	ctx := reqFunc(context.Background(), &http.Request{})

	if ctx.Value(CorrelationID) != nil {
		t.Error("Context shouldn't contain the CorrelationID")
	}

	head := http.Header{}
	// Authorization header is correct
	head.Set(header, "foobar")
	ctx = reqFunc(context.Background(), &http.Request{Header: head})

	token := ctx.Value(CorrelationID).(string)
	if token != "foobar" {
		t.Errorf("Context doesn't contain the expected encoded token value; expected: %s, got: %s", "foobar", token)
	}
}

func TestContextToHTTP(t *testing.T) {
	reqFunc := ContextToHTTP()

	// No JWT Token is passed in the context
	ctx := context.Background()
	r := http.Request{}
	reqFunc(ctx, &r)

	token := r.Header.Get(header)
	if token != "" {
		t.Error("header key should not exist in metadata")
	}

	// Correct JWT Token is passed in the context
	ctx = context.WithValue(context.Background(), CorrelationID, "foobar")
	r = http.Request{Header: http.Header{}}
	reqFunc(ctx, &r)

	token = r.Header.Get(header)
	expected := "foobar"

	if token != expected {
		t.Errorf("Authorization header does not contain the expected JWT token; expected %s, got %s", expected, token)
	}
}

func TestGRPCToContext(t *testing.T) {
	md := metadata.MD{}
	reqFunc := GRPCToContext()

	// No Authorization header is passed
	ctx := reqFunc(context.Background(), md)
	token := ctx.Value(CorrelationID)
	if token != nil {
		t.Error("Context should not contain a correlation ID")
	}

	md[headerHTTP2] = []string{"foobar"}
	ctx = reqFunc(context.Background(), md)
	token, ok := ctx.Value(CorrelationID).(string)
	if !ok {
		t.Fatal("Correlation ID not passed to context correctly")
	}

	if token != "foobar" {
		t.Errorf("Correlation ID did not match: expecting %s got %s", "foobar", token)
	}
}

func TestContextToGRPC(t *testing.T) {
	reqFunc := ContextToGRPC()

	// No JWT Token is passed in the context
	ctx := context.Background()
	md := metadata.MD{}
	reqFunc(ctx, &md)

	_, ok := md[headerHTTP2]
	if ok {
		t.Error("authorization key should not exist in metadata")
	}

	// Correct JWT Token is passed in the context
	ctx = context.WithValue(context.Background(), CorrelationID, "foobar")
	md = metadata.MD{}
	reqFunc(ctx, &md)

	token, ok := md[headerHTTP2]
	if !ok {
		t.Fatal("JWT Token not passed to metadata correctly")
	}

	if token[0] != "foobar" {
		t.Errorf("JWT tokens did not match: expecting %s got %s", "foobar", token[0])
	}
}
