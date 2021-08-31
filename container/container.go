/*
Package container includes the Container type, witch contains a collection of modules.
*/
package container

import (
	"github.com/DoNewsCode/core/contract"
	"github.com/Reasno/ifilter"
	"github.com/gorilla/mux"
	"github.com/oklog/run"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var _ contract.Container = (*Container)(nil)

// CronProvider provides cron jobs.
type CronProvider interface {
	ProvideCron(crontab *cron.Cron)
}

// CommandProvider provides cobra.Command.
type CommandProvider interface {
	ProvideCommand(command *cobra.Command)
}

// HTTPProvider provides http services.
type HTTPProvider interface {
	ProvideHTTP(router *mux.Router)
}

// GRPCProvider provides gRPC services.
type GRPCProvider interface {
	ProvideGRPC(server *grpc.Server)
}

// CloserProvider provides a shutdown function that will be called when service exits.
type CloserProvider interface {
	ProvideCloser()
}

// RunProvider provides a runnable actor. Use it to register any server-like
// actions. For example, kafka consumer can be started here.
type RunProvider interface {
	ProvideRunGroup(group *run.Group)
}

// Container holds all modules registered.
type Container struct {
	httpProviders    []func(router *mux.Router)
	grpcProviders    []func(server *grpc.Server)
	closerProviders  []func()
	runProviders     []func(g *run.Group)
	modules          ifilter.Collection
	cronProviders    []func(crontab *cron.Cron)
	commandProviders []func(command *cobra.Command)
}

// ApplyRouter iterates through every HTTPProvider registered in the container,
// and introduce the router to everyone.
func (c *Container) ApplyRouter(router *mux.Router) {
	for _, p := range c.httpProviders {
		p(router)
	}
}

// ApplyGRPCServer iterates through every GRPCProvider registered in the container,
// and introduce a *grpc.Server to everyone.
func (c *Container) ApplyGRPCServer(server *grpc.Server) {
	for _, p := range c.grpcProviders {
		p(server)
	}
}

// Shutdown iterates through every CloserProvider registered in the container,
// and calls them in the reversed order of registration.
func (c *Container) Shutdown() {
	for i := len(c.closerProviders) - 1; i >= 0; i-- {
		c.closerProviders[i]()
	}
}

// ApplyRunGroup iterates through every RunProvider registered in the container,
// and introduce the *run.Group to everyone.
func (c *Container) ApplyRunGroup(g *run.Group) {
	for _, p := range c.runProviders {
		p(g)
	}
}

// Modules returns all modules in the container. This method is used to scan for
// custom interfaces. For example, The database module use Modules to scan for
// database migrations.
/*
	m.container.Modules().Filter(func(p MigrationProvider) {
		for _, migration := range p.ProvideMigration() {
			if migration.Connection == "" {
				migration.Connection = "default"
			}
			if migration.Connection == connection {
				migrations.Collection = append(migrations.Collection, migration)
			}
		}
	})
*/
func (c *Container) Modules() ifilter.Collection {
	return c.modules
}

// ApplyCron iterates through every CronProvider registered in the container,
// and introduce the *cron.Cron to everyone.
func (c *Container) ApplyCron(crontab *cron.Cron) {
	for _, p := range c.cronProviders {
		p(crontab)
	}
}

// ApplyRootCommand iterates through every CommandProvider registered in the container,
// and introduce the root *cobra.Command to everyone.
func (c *Container) ApplyRootCommand(command *cobra.Command) {
	for _, p := range c.commandProviders {
		p(command)
	}
}

func (c *Container) AddModule(module interface{}) {
	if p, ok := module.(func()); ok {
		c.closerProviders = append(c.closerProviders, p)
		return
	}
	if p, ok := module.(HTTPProvider); ok {
		c.httpProviders = append(c.httpProviders, p.ProvideHTTP)
	}
	if p, ok := module.(GRPCProvider); ok {
		c.grpcProviders = append(c.grpcProviders, p.ProvideGRPC)
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
	c.modules = append(c.modules, module)
}
