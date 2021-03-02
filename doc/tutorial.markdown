## Advanced Tutorial

At this point, We have introduced the basic concept of package core. Let's learn to use package core by more examples.
For readers eager to see the end product, please check out the [skeleton](https://github.com/DoNewsCode/skeleton) demo.

### Phase one: setup a core

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
We can replace the default implementation with the custom ones by provide the following args to the `New` method.

Here is an example swapping every implementation. Normally this is not necessary.

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
By default it reads nothing.

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

In the example above, the configuration is loaded when the system boot up. Sometimes we want to reload configuration while the process is running.
This is can be done by adding a watcher. Currently two kinds of watcher are supported. File watcher and signal watcher.
File watcher reloads the configuration stack when the target file changes, and the signal watcher reloads the stack when USRSIG1 signal is received.

> Due to OS limitations, signal watcher is not available on Windows.

```go
c.Default(
	core.WithConfigStack(file.Provider("./mock/mock.yaml"), yaml.Parser()),
	core.WithConfigWatcher(watcher.File{Path: "./mock/mock.yaml"}),
)
```

> The configuration watch is triggered after the serve command is called.

If the rich options for `core.New()` seems overwhelming, feel free to use the bundled one-liner `WithYamlFile` option.

The above example can be rewritten as:

```go
c.Default(
	core.WithYamlFile("./mock/mock.yaml"),
)
```

## Phase two: Add dependencies

It is often Inevitable for our application to depend on external resources.

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
The return values is added back into the container.

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

`func()` is treated as clean up functions. All returned clean up functions are called in parallel when `c.Shutdown()` is called.

The constructor returns error to indicate initialization failure. Core intentionally panics in this case to raise awareness.

The last but not the least thing to notice is that the dependency graph is build lazily. 
The constructor call is deferred until the return value is directly or indirectly demanded.
That means dependencies for a type can be added to the graph both, before and after the type was added.

The default DiContainer implementation is the [uber/dig](https://pkg.go.dev/go.uber.org/dig). For advanced usage, check out their guide.






