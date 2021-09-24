# 教程

至此，我们已经介绍了`core`包的基本概念。让我们通过更多的例子来学习如何使用`core`包。
对于渴望看到最终产品的读者，请查看[skeleton](https://github.com/DoNewsCode/skeleton) demo 演示。

## 步骤一：Setup

### 构建核心依赖关系

封装核心的核心元素是核心`c`。它是由几件作品组成的

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

- `AppName`是应用程序名。
- `Env`是一个应用环境，如生产、开发、测试等。
- `ConfigAccessor`是中心配置单例。
- `LevelLogger`是默认的 logger。
- `Container`用于模块注册。
- `Dispatcher`用于在服务之间传输事件。
- `DiContainer`用于依赖注入。

`core.C`中的每个成员都是一个接口。默认实现可以在它们各自的包中找到。 当`core.New()`被调用时，新创建的`C`实例将用默认值填充
。我们可以通过为新方法提供以下参数，将默认实现替换为自定义实现。 通过为`New`方法提供以下参数，我们用自定义实现来替换这些默认实现。

下面是一个替换每个实现的示例。通常这是没有必要的。

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

有另一种方法来引导`core.C`。我们可以使用`core.Default()`代替`core.New()`。 这个`core.Default()`会做`core.New()`
所做的所有事情，此外，还可以自动将核心依赖项(ConfigAccessor, Logger…)添加到`DiContainer`中。 除非你想对 `DI container`
进行细粒度控制，你可以很开心的使用`core.Default()`。

### 导致配置

`New()`和`Default()`还负责从外部读取配置。默认情况下，它不读取任何内容。

这个 configuration 实现由 config package 提供。它将配置视为一个堆栈。`core.C`从 package config 继承了许多选项。

让我们构建一个典型的配置堆栈，顶部是 flag，中间是环境变量，底部是配置文件。

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
    flagSet := flag.NewFlagSet("example", flag.ContinueOnError)
    flagSet.String("log.level", "error", "the log level")

	c := core.Default(
		// 第一: 读取flag
		core.WithConfigStack(basicflag.Provider(flagSet, "."), nil),
		// 第二: 读取环境变量
		core.WithConfigStack(env.Provider("APP_", ".", func(s string) string {
			return strings.ToLower(strings.Replace(s, "APP_", "", 1))
		}), nil),
		// 第三: 读取配置文件
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

在上面的示例中，在系统启动时加载配置。有时我们想在进程运行时重新加载配置。这可以通过添加一个观察者来完成。 目前，支持两种类型的观察者。

- 文件监视程序: 当目标配置文件发生更改时，File watcher 重新加载配置堆栈
- 信号监视程序: 当接收到 USRSIG1 信号时，signal watcher 重新加载堆栈。

> 由于操作系统的限制，信号监视程序在 Windows 上不可用。

```go
c.Default(
	core.WithConfigStack(file.Provider("./mock/mock.yaml"), yaml.Parser()),
	core.WithConfigWatcher(watcher.File{Path: "./mock/mock.yaml"}),
)
```

> 调用 serve 命令后，将触发 configuration watch。

如果`core.New()`的你觉得选项太多了，你也可以仅仅使用`yamlfile`的这一项配置。

上面的例子可以重写为

```go
c.Default(
	core.WithYamlFile("./mock/mock.yaml"),
)
```

## 步骤二: 增加依赖

我们的应用程序常常不可避免地要依赖外部资源。

### provide

Type C 通过`Provide`方法接受依赖项绑定。

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
        // 在 core.Default()中提供了contract.ConfigAccessor(配置读取器)
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

`c.Provide`的参数，在`di.Deps`的类型下，是一系列构造函数。函数的参数应该由`di容器`注入。返回值被添加回容器中。

多个构造函数可以依赖于同一类型。容器为每个保留类型创建一个单例，在直接请求时或者作为另一种类型的依赖项,它最多实例化一次。

构造函数可以返回多个结果以将多个类型添加到容器中。它还可能返回 `errors` 和`func（）`。

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

`func()`作为清除函数。当`c.Shutdown()`被调用时，所有返回的清理函数都被并行调用。

构造函数返回一个错误以指示初始化失败。`core`故意发出 panic，以提高意识。

最后但并非最不重要的一点是，依赖关系图是惰性构建的。构造函数调用被延迟，直到直接或间接地要求返回值。 这意味着在添加类型之前和之后，类型的依赖关系都可以添加到图中。

### invoke

说到需求，`c.Invoke`可以用来实例化依赖关系。

```go
func (c *C) Invoke(function interface{})
```

在实例化依赖项后，`Invoke`运行给定的函数。

函数的任何参数都被视为函数的依赖项。依赖项和它们可能有的依赖项以未指定的顺序被实例化。

默认的双容器实现是[uber/dig](https://pkg.go.dev/go.uber.org/dig) .关于高级用法，请查看他们的指南。

## 步骤三：add functionality.

一个 module 应该是一组功能。他一定包含某些 api，如 http、grpc、Cron 或者其他命令行。

让一组功能依赖于另一组功能是不健康的;边界上下文清晰通常是我们构建微服务的第一课。

在 package core 中，我们有意将依赖关系和 module 的概念分开，这样两个 module 就不会相互依赖。 虽然允许 module 具有共享的依赖关系，但 module
不应该知道共享的所有权。

使用这种方法，我们保留了在微服务周围移动 module 的能力，只要我们能够满足依赖需求。

### 增加 Module

在核心中注册模块有两种方法。您可以使用 DI 容器手动或自动构建模块：

- `c.AddModule`允许你添加一个手动构造的模块。
  ```go
     c.AddModule(srvhttp.DocsModule{})
  ```
- `c.AddModuleFunc`接受模块的构造函数作为参数，并通过注入 DI 容器中的构造函数参数来实例化模块。
  ```go
    func injectModule(conf contract.ConfigAccessor, logger log.Logger) Module, func(), error {
    // build the module, return the module, the cleanup function or possible errors.
    }
    c.AddModuleFunc(injectModule)
  ```

### Module interfaces

当调用`c.Serve(ctx)`或`root command`(见阶段 4)时，所有注册的 module 将被扫描以下接口

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

如果模块实现了任何一个 provider 接口，`core`将会用“registry”调用这个 provider 程序函数，比如，`mux.Router`。然后模块可以注册这些路由。

让我们看一个包含 HTTP、cronjobs 和 closer 的模块：

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

### Custom Module interfaces

当然有可能所有内置的 provider 接口都无法满足业务需求。但是，它可以非常简单的扫描自定义接口：

```go
multiProcessor := thrift.NewTMultiplexedProcessor()
for _, m := range c.Modules() {
	if thriftModule, ok := m.(interface{ProvideThrift(p thrift.TMultiplexedProcessor)}); ok {
    	thriftModule.ProvideThrift(multiProcessor)
    }
}
```

## 步骤四：Serve！

在我们展示的大多数示例中，我们使用`c.Serve`来运行应用程序。如果应用程序是专门的长时间运行的流程，这是合适的。

12-factor-app 的清单表明

> 作为一次性流程运行管理任务

`core`包原生支持一次性进程。这些进程是由 [cobra.Command](https://github.com/spf13/cobra) 提供。 在这个模型中，`serve`
命令只是注册在`root command`下许多子命令中的一个。

下面是一个在 root 下 `groups serve`和`version subcommand`的示例。

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

启动服务:

```go
go run main.go serve
```

打印版本：

```go
go run main.go version
```

你可以也用你自己的服务模块，更换`core.NewServeModule`。

现在您已经完成了本教程，请一定要尝试这个项目，并阅读托管在[pkg.go.dev](https://pkg.go.dev/github.com/DoNewsCode/core) 上的函数级文档。
