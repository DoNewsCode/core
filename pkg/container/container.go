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

func (c *Container) GetHttpProviders() []func(router *mux.Router) {
	return c.HttpProviders
}

func (c *Container) GetGrpcProviders() []func(server *grpc.Server) {
	return c.GrpcProviders
}

func (c *Container) GetCloserProviders() []func() {
	return c.CloserProviders
}

func (c *Container) GetRunProviders() []func(g *run.Group) {
	return c.RunProviders
}

func (c *Container) GetModules() ifilter.Collection {
	return c.Modules
}

func (c *Container) GetCronProviders() []func(crontab *cron.Cron) {
	return c.CronProviders
}

func (c *Container) GetCommandProviders() []func(command *cobra.Command) {
	return c.CommandProviders
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

func (c *Container) AddModule(module interface{}) {
	if p, ok := module.(func()); ok {
		c.CloserProviders = append(c.CloserProviders, p)
	}
	if p, ok := module.(HttpProvider); ok {
		c.HttpProviders = append(c.HttpProviders, p.ProvideHttp)
	}
	if p, ok := module.(GrpcProvider); ok {
		c.GrpcProviders = append(c.GrpcProviders, p.ProvideGrpc)
	}
	if p, ok := module.(CronProvider); ok {
		c.CronProviders = append(c.CronProviders, p.ProvideCron)
	}
	if p, ok := module.(RunProvider); ok {
		c.RunProviders = append(c.RunProviders, p.ProvideRunGroup)
	}
	if p, ok := module.(CommandProvider); ok {
		c.CommandProviders = append(c.CommandProviders, p.ProvideCommand)
	}
	if p, ok := module.(CloserProvider); ok {
		c.CloserProviders = append(c.CloserProviders, p.ProvideCloser)
	}
	c.Modules = append(c.Modules, module)
}
