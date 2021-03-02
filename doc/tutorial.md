## Advanced Tutorial

At this point, We have introduced the basic concept of package core. Let's learn to use package core with more examples.
For readers eager to see the end product, please check out the [skeleton](https://github.com/DoNewsCode/skeleton) demo.

### Phase one: Setup

#### Construct the core dependencies

The central element in package core is `core.C`. It is composed by several pieces:

```go
type C struct {
	AppName contract.AppName
	Env     contract.Env
	contract.ConfigAccessor
	logging.LevelLogger
	contract.Container
	contract.Dispatcher
	di DiContainer
}
```

* AppName is the application name.
* Env is the application environment, such as production, development, testing.
* ConfigAccessor is the central configuration singleton.
* LevelLogger is the default logger.
* Container is used for module registration.
* Dispatcher is used for transmitting events between services.
* DiContainer is used for dependencies injection.

Every member in the `core.C` is an interface.
The default implementation can be found in their respective packages.
When `core.New()` is called, a newly created C instance will be filled with the default values.
We can replace the default implementation with the custom ones by providing the following args to the `New` method.

Here is an example of swapping every implementation. Normally this is not necessary.

```go
core.New(
  SetConfigProvider(configImplementation)
  SetAppNameProvider(appNameImplementation)
  SetEnvProvider(envImplementation)
  SetLoggerProvider(loggerImplementation)
  SetDiProvider(diImplementation)
  SetEventDispatcherProvider(eventDispatcherImplementation)
)
```

There is another way to bootstrap the core.C. We can use `core.Default()` instead of `core.New()`.
The `core.Default()` will do everything that `core.New()` does,
plus adding core dependencies (ConfigAccessor, Logger...) to the `DiContainer` automatically.
Unless you want fine-grained control over the DI container, you can happily use `core.Default()`.

#### Load configurations

`New()` and `Default()` is also responsible for reading configurations from outside.
By default, it reads nothing.

The configuration implementation is provided by package config.
It views the configuration as a stack.
`core.C` inherits many of the options from package config.

Let's build a typical configuration stack with flags on top,
environment variables in the middle and configuration file at the bottom.

```go
package main

import (
	"context"
	"flag"
	"net/http"
	"strings"

	"github.com/DoNewsCode/core"
	"github.com/gorilla/mux"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/basicflag"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
)

func main() {

	fs := flag.NewFlagSet("example", flag.ContinueOnError)
	fs.String("log.level", "error", "the log level")

	c := core.Default(
		core.WithConfigStack(basicflag.Provider(fs, "."), nil),
		core.WithConfigStack(env.Provider("APP_", ".", func(s string) string {
			return strings.ToLower(strings.Replace(s, "APP_", "", 1))
		}), nil),
		core.WithConfigStack(file.Provider("./mock/mock.json"), json.Parser()),
	)

	c.AddModule(core.HttpFunc(func(router *mux.Router) {
		router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			writer.Write([]byte("hello world"))
		})
	}))

	c.Serve(context.Background())
}
```

In the example above, the configuration is loaded when the system boots up. Sometimes we want to reload the configuration while the process is running.
This is can be done by adding a watcher. Currently, two kinds of watchers are supported. File watcher and signal watcher.
File watcher reloads the configuration stack when the target file changes, and the signal watcher reloads the stack when USRSIG1 signal is received.

> Due to OS limitations, the signal watcher is not available on Windows.

```go
c.Default(
	core.WithConfigStack(file.Provider("./mock/mock.yaml"), yaml.Parser()),
	core.WithConfigWatcher(watcher.File{Path: "./mock/mock.yaml"}),
)
```

> The configuration watch is triggered after the serve command is called.

If the rich options for `core.New()` seem overwhelming, feel free to use the bundled one-liner `WithYamlFile` option.

The above example can be rewritten as:

```go
c.Default(
	core.WithYamlFile("./mock/mock.yaml"),
)
```

### Phase two: Add dependencies

It is often Inevitable for our application to depend on external resources.


#### Provide

Type C accepts dependencies bindings via the `Provide` methods.

```go
package main

import (
	"context"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
)

type RemoteService struct {
	Conf contract.ConfigAccessor
}

type RemoteModule struct {
	Service RemoteService
}

func main() {

	c := core.Default()

	c.Provide(di.Deps{
		// contract.ConfigAccessor is provided in core.Default()
		func(conf contract.ConfigAccessor) RemoteService {
			return RemoteService{
				Conf: conf,
			}
		},
		func(svc RemoteService) RemoteModule {
			return RemoteModule{
				Service: svc,
			}
		},
	})

	c.Invoke(func(module RemoteModule) {
		c.AddModule(module)
	})

	c.Serve(context.Background())
}
```

The argument of `c.Provide`, under the type of `di.Deps`, is a list of constructor functions.
The arguments of the functions are supposed to be injected by the di container.
The return values are added back into the container.

Multiple constructors can rely on the same type.
The container creates a singleton for each retained type,
instantiating it at most once when requested directly or as a dependency of another type.

Constructors can return multiple results to add multiple types to the container. 
It may also return errors and `func()`.

```go
	c.Provide(di.Deps{
		func(conf contract.ConfigAccessor) (RemoteService, func(), error) {
			s := RemoteService{
				Conf: conf,
			}
			return s, s.Close, nil
		},
	})
```

`func()` is treated as clean up functions. All returned clean-up functions are called in parallel when `c.Shutdown()` is called.

The constructor returns an error to indicate initialization failure. Core intentionally panics in this case to raise awareness.

The last but not the least thing to notice is that the dependency graph is built lazily. 
The constructor call is deferred until the return value is directly or indirectly demanded.
That means dependencies for a type can be added to the graph both, before and after the type was added.

#### Invoke

Speaking of the demand, `c.Invoke` can be used to instantiate dependencies.

```go
func (c *C) Invoke(function interface{})
```

Invoke runs the given function after instantiating its dependencies.

Any arguments that the function are treated as its dependencies. 
The dependencies are instantiated in an unspecified order along with any
dependencies that they might have.

The default `DiContainer` implementation is the [uber/dig](https://pkg.go.dev/go.uber.org/dig). 
For advanced usage, check out their guide.

### Phase three: Add functionality.

A module is a group of functionality. It must have certain APIs, such as HTTP, gRPC, Cron, or command line.

It is not healthy to have one group of functionality depends on another; Bounded context is usually the first
lesson we learn building microservices.

In package core, we deliberately separated the concept of dependency 
and module so that no two modules can depend on one other. Though
modules are allowed to have shared dependencies, the module should have 
no idea about the shared ownership.

Using this methodology, we retain the ability to move modules around microservices, 
as long as we are able to meet the dependency requirement.

#### Add Module

There are two ways to register a module in the core. You can build the module manually or 
autopilot with DI container:

* `c.AddModule` allows you to add a manually constructed module.

```go
c.AddModule(srvhttp.DocsModule{})
```

* `c.AddModuleFunc` accepts the module's constructor as argument, 
and instantiate the module by injecting the constructor parameters from DI container.

```go
func injectModule(conf contract.ConfigAccessor, logger log.Logger) Module, func(), error {
	// build the module, return the module, the cleanup function or possible errors.
}
c.AddModuleFunc(injectModule)
```

#### Module interfaces

When `c.Serve(ctx)` or the root command (see phase four) is called, 
all registered modules will be scanned for the following interfaces:

```go
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

// CloserProvider provides a shutdown function that will be called when the service exits.
type CloserProvider interface {
	ProvideCloser()
}

// RunProvider provides a runnable actor. Use it to register any server-like
// actions. For example, kafka consumer can be started here.
type RunProvider interface {
	ProvideRunGroup(group *run.Group)
}
```

If the module implements any of the provider interfaces, 
the core will call this provider function with a "registry", say, mux.Router.
The module can then register its routes.

Let's see a module with both HTTP, cronjobs, and a closer:

```go
package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DoNewsCode/core"
	"github.com/gorilla/mux"
	"github.com/robfig/cron/v3"
)

type RemoteModule struct {}

func (r RemoteModule) ProvideCloser() {
	fmt.Println("closing")
}

func (r RemoteModule) ProvideCron(crontab *cron.Cron) {
	crontab.AddFunc("* * * * *", func() {
		fmt.Println("cron triggered")
	})
}

func (r RemoteModule) ProvideHTTP(router *mux.Router) {
	router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("hello world"))
	})
}

func main() {
	c := core.Default()
	defer c.Shutdown()

	c.AddModule(RemoteModule{})
	c.Serve(context.Background())
}

```

#### Custom module interfaces

It is completely possible that none of the built-in provider interfaces satisfy the business requirement.
However, it is deadly simple to scan for your custom interfaces:

```go
multiProcessor := thrift.NewTMultiplexedProcessor()
for _, m := range c.Modules() {
	if thriftModule, ok := m.(interface{ProvideThrift(p thrift.TMultiplexedProcessor)}); ok {
    	thriftModule.ProvideThrift(multiProcessor)
    }
}
```

### Phase four: Serve!

In most of the examples we have shown, we use `c.Serve` the run the application. 
This is appropriate if the application is exclusively long-running process.

The manifest of twelve-factor apps states: 

> Run admin/management tasks as one-off processes

Package core natively supports one-off processes. 
Those processes were conducted by [cobra.Command](https://github.com/spf13/cobra).
In this model, the `serve` command is only one of many subcommands registered under the root command.

Below is an example that groups serve and version subcommand under the root.

```go
package main

import (
	"fmt"

	"github.com/DoNewsCode/core"
	"github.com/spf13/cobra"
)

type RemoteModule struct {}

func (r RemoteModule) ProvideCommand(command *cobra.Command) {
	cmd := &cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("0.1.0")
		},
	}
	command.AddCommand(cmd)
}

func main() {
	c := core.Default()
	defer c.Shutdown()

	c.AddModule(RemoteModule{})
	c.AddModuleFunc(core.NewServeModule)
	rootCmd := &cobra.Command{
		Use: "root",
		Short: "A demo command",
	}
	c.ApplyRootCommand(rootCmd)
	rootCmd.Execute()
}
```

To start the server:

```bash
go run main.go serve
```

To print the version:

```bash
go run main.go version
```

You can replace the `core.NewServeModule` with your own serve module too.

Now that you have finished the tutorial, be sure to try out the project,
and read function level documentation hosted at [pkg.go.dev](https://pkg.go.dev/github.com/DoNewsCode/core).










