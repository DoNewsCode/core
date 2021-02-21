/*
Package container includes the Container type, witch contains a collection of modules.
*/
package container

import (
	"sync"

	"github.com/Reasno/ifilter"
	"github.com/gorilla/mux"
	"github.com/oklog/run"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

type Container struct {
	httpProviders    []func(router *mux.Router)
	grpcProviders    []func(server *grpc.Server)
	closerProviders  []func()
	runProviders     []func(g *run.Group)
	modules          ifilter.Collection
	cronProviders    []func(crontab *cron.Cron)
	commandProviders []func(command *cobra.Command)
}

func (c *Container) ApplyRouter(router *mux.Router) {
	for _, p := range c.httpProviders {
		p(router)
	}
}

func (c *Container) ApplyGRPCServer(server *grpc.Server) {
	for _, p := range c.grpcProviders {
		p(server)
	}
}

func (c *Container) Shutdown() {
	var wg sync.WaitGroup
	for _, p := range c.closerProviders {
		wg.Add(1)
		p := p
		go func() {
			p()
			wg.Done()
		}()
	}
	wg.Wait()
}

func (c *Container) ApplyRunGroup(g *run.Group) {
	for _, p := range c.runProviders {
		p(g)
	}
}

func (c *Container) Modules() ifilter.Collection {
	return c.modules
}

func (c *Container) ApplyCron(crontab *cron.Cron) {
	for _, p := range c.cronProviders {
		p(crontab)
	}
}

func (c *Container) ApplyRootCommand(command *cobra.Command) {
	for _, p := range c.commandProviders {
		p(command)
	}
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

func (c *Container) AddModule(module interface{}) {
	if p, ok := module.(func()); ok {
		c.closerProviders = append(c.closerProviders, p)
	}
	if p, ok := module.(HttpProvider); ok {
		c.httpProviders = append(c.httpProviders, p.ProvideHttp)
	}
	if p, ok := module.(GrpcProvider); ok {
		c.grpcProviders = append(c.grpcProviders, p.ProvideGrpc)
	}
	if p, ok := module.(CronProvider); ok {
		c.cronProviders = append(c.cronProviders, p.ProvideCron)
	}
	if p, ok := module.(RunProvider); ok {
		c.runProviders = append(c.runProviders, p.ProvideRunGroup)
	}
	if p, ok := module.(CommandProvider); ok {
		c.commandProviders = append(c.commandProviders, p.ProvideCommand)
	}
	if p, ok := module.(CloserProvider); ok {
		c.closerProviders = append(c.closerProviders, p.ProvideCloser)
	}
	if _, ok := module.(func()); !ok {
		c.modules = append(c.modules, module)
	}

}
