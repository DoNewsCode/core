package container

import (
	"github.com/Reasno/ifilter"
	"github.com/gorilla/mux"
	"github.com/oklog/run"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

type Container struct {
	HttpProviders    []func(router *mux.Router)
	GrpcProviders    []func(server *grpc.Server)
	CloserProviders  []func()
	RunProviders     []func(g *run.Group)
	Modules          ifilter.Collection
	CronProviders    []func(crontab *cron.Cron)
	CommandProviders []func(command *cobra.Command)
}

type CronProvider interface {
	ProvideCron(crontab *cron.Cron)
}

type CommandProvider interface {
	ProvideCommand(command *cobra.Command)
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

func (s *Container) AddModule(module interface{}) {
	if p, ok := module.(HttpProvider); ok {
		s.HttpProviders = append(s.HttpProviders, p.ProvideHttp)
	}
	if p, ok := module.(GrpcProvider); ok {
		s.GrpcProviders = append(s.GrpcProviders, p.ProvideGrpc)
	}
	if p, ok := module.(CronProvider); ok {
		s.CronProviders = append(s.CronProviders, p.ProvideCron)
	}
	if p, ok := module.(RunProvider); ok {
		s.RunProviders = append(s.RunProviders, p.ProvideRunGroup)
	}
	if p, ok := module.(CommandProvider); ok {
		s.CommandProviders = append(s.CommandProviders, p.ProvideCommand)
	}
	if p, ok := module.(CloserProvider); ok {
		s.CloserProviders = append(s.CloserProviders, p.ProvideCloser)
	}
	s.Modules = append(s.Modules, module)
}
