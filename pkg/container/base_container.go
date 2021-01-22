package container

import (
	"github.com/gorilla/mux"
	"github.com/oklog/run"
	"google.golang.org/grpc"
)

type BaseContainer struct {
	HttpProviders   []func(router *mux.Router)
	GrpcProviders   []func(server *grpc.Server)
	CloserProviders []func()
	RunProviders    []func(g *run.Group)
}

type HttpProvider interface {
	ProvideHttp(router *mux.Router)
}

type GrpcProvider interface {
	ProvideGrpc(server *grpc.Server)
}

type CloserProvider interface {
	ProvideCloser()
}

type RunProvider interface {
	ProvideRunGroup(group *run.Group)
}

type HttpFunc func(router *mux.Router)

func (h HttpFunc) ProvideHttp(router *mux.Router) {
	h(router)
}

func (s *BaseContainer) Register(app interface{}) {
	if p, ok := app.(HttpProvider); ok {
		s.HttpProviders = append(s.HttpProviders, p.ProvideHttp)
	}
	if p, ok := app.(GrpcProvider); ok {
		s.GrpcProviders = append(s.GrpcProviders, p.ProvideGrpc)
	}
	if p, ok := app.(RunProvider); ok {
		s.RunProviders = append(s.RunProviders, p.ProvideRunGroup)
	}
	if p, ok := app.(CloserProvider); ok {
		s.CloserProviders = append(s.CloserProviders, p.ProvideCloser)
	}
}
