package kitmw

import (
	"context"
	"net"
	"net/http"
	"testing"

	"github.com/DoNewsCode/core/contract"
	helloworld_v1 "github.com/DoNewsCode/core/kitmw/testdata"
	"github.com/go-kit/kit/endpoint"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type greeterServer struct {
	helloworld_v1.UnimplementedGreeterServer
}

func (g greeterServer) SayHello(ctx context.Context, request *helloworld_v1.HelloRequest) (*helloworld_v1.HelloReply, error) {
	return &helloworld_v1.HelloReply{}, nil
}

func MakeEndpoint(s greeterServer) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		return s.SayHello(ctx, request.(*helloworld_v1.HelloRequest))
	}
}

type gs struct {
	helloworld_v1.UnimplementedGreeterServer
	sayHello grpctransport.Server
}

func (g gs) mustEmbedUnimplementedGreeterServer() {
	panic("implement me")
}

func (g gs) SayHello(ctx context.Context, request *helloworld_v1.HelloRequest) (*helloworld_v1.HelloReply, error) {
	_, rep, err := g.sayHello.ServeGRPC(ctx, request)
	if err != nil {
		return nil, err
	}
	return rep.(*helloworld_v1.HelloReply), nil
}

func TestTransportToGRPCContext(t *testing.T) {
	server := grpc.NewServer(grpc.UnaryInterceptor(Interceptor))
	greeter := greeterServer{}
	ep := MakeEndpoint(greeter)
	sayHello := grpctransport.NewServer(
		ep,
		func(ctx context.Context, i interface{}) (request interface{}, err error) {
			return i, nil
		},
		func(ctx context.Context, i interface{}) (response interface{}, err error) {
			assert.Equal(t, "127.0.0.1", ctx.Value(contract.IpKey).(string))
			assert.Equal(t, GRPCKind, ctx.Value(contract.TransportKey).(string))
			assert.Equal(t, "/helloworld.v1.Greeter/SayHello", ctx.Value(contract.RequestUrlKey).(string))
			return i, nil
		},
		grpctransport.ServerBefore(TransportToGRPCContext()),
		grpctransport.ServerBefore(IPToGRPCContext()),
	)

	helloworld_v1.RegisterGreeterServer(server, gs{sayHello: *sayHello})
	ln, _ := net.Listen("tcp", ":9090")
	defer ln.Close()
	go server.Serve(ln)

	conn, err := grpc.Dial(":9090", grpc.WithInsecure())
	assert.NoError(t, err)
	client := helloworld_v1.NewGreeterClient(conn)
	_, _ = client.SayHello(context.Background(), &helloworld_v1.HelloRequest{})
}

func TestTransportToHTTPContext(t *testing.T) {
	greeter := greeterServer{}
	ep := MakeEndpoint(greeter)
	sayHello := httptransport.NewServer(
		ep,
		func(ctx context.Context, req *http.Request) (request interface{}, err error) {
			return &helloworld_v1.HelloRequest{}, nil
		},
		func(ctx context.Context, resp http.ResponseWriter, i interface{}) error {
			assert.Equal(t, "127.0.0.1", ctx.Value(contract.IpKey).(string))
			assert.Equal(t, "/hello", ctx.Value(contract.RequestUrlKey).(string))
			assert.Equal(t, HTTPKind, ctx.Value(contract.TransportKey).(string))
			return nil
		},
		httptransport.ServerBefore(TransportToHTTPContext()),
		httptransport.ServerBefore(IPToHTTPContext()),
		httptransport.ServerBefore(RequestURLToHTTPContext()),
	)

	ln, _ := net.Listen("tcp", ":8181")
	defer ln.Close()
	go http.Serve(ln, sayHello)

	http.Get("http://127.0.0.1:8181/hello")
}

func TestRealIP(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-FORWARDED-FOR", "10.0.0.1, 1.1.1.1")
	ip := realIP(req)
	assert.Equal(t, "10.0.0.1", ip)
	req.Header.Set("X-FORWARDED-FOR", "10.0.0.1")
	ip = realIP(req)
	assert.Equal(t, "10.0.0.1", ip)
}
