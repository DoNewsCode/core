Package core 是一个服务容器，它优雅地引导和协调 Go 中的 12 个因素应用程序。

# 背景简介

十二因素法多年来证明了它的价值。自从它发明以来，许多技术领域都发生了变化，其中许多领域都焕发着光辉和激动人心。
在`Kubernetes`、`mesh服务` 和`无服务器架构`的时代，十二要素方法论并没有消失，而是恰好适合于几乎所有这些强大的平台。

对于经验丰富的工程师来说，构建一个包含 12 个要素的 go 应用程序可能不是一项困难的任务，但对年轻人来说无疑是一个挑战。
对于那些有能力安排事情的人来说，还有很多决定要做，还有很多选择要在团队内部达成一致。

包核心是为了引导和协调这些服务而创建的。

# 概述

无论 app 是什么，引导阶段大致由以下部分组成：

- 从二进制文件中读取配置。即标志、环境变量和/或配置文件。
- 初始化依赖项。数据库、消息队列、服务发现等。
- 定义如何运行应用程序。HTTP、RPC、命令行、cron jobs 或更常见的混合。

本包`core`对那些重复的步骤进行了抽象，使它们简洁、可移植但又显式。让我们看看下面的代码片段：

```go
package main

import (
  "context"
  "net/http"

  "github.com/DoNewsCode/core"
  "github.com/DoNewsCode/core/observability"
  "github.com/DoNewsCode/core/otgorm"
  "github.com/gorilla/mux"
)

func main() {
  // 步骤一: 创建core从配置文件
  c := core.New(core.WithYamlFile("config.yaml"))

  // 步骤二: 绑定依赖
  c.Provide(otgorm.Providers())

  // 步骤三: 定义服务
  c.AddModule(core.HttpFunc(func(router *mux.Router) {
    router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
      writer.Write([]byte("hello world"))
    })
  }))

  // 步骤四: 启动
  c.Serve(context.Background())
}
```

在几行代码中，HTTP 服务是以上面概述的样式引导的。它是简单的、明确的，在某种程度上是声明性的。

上面演示的服务使用内联处理程序函数突出显示该点。通常，对于实际项目，我们将使用模块代替。
package Core 术语表中的`module`不一定是`go模块`（尽管可以）。它只是一组服务。

您可能会注意到，HTTP 服务并不真正使用依赖关系。那是真的。

让我们重写 HTTP 服务以使用上述依赖项。

```go
package main

import (
  "context"
  "net/http"

  "github.com/DoNewsCode/core"
  "github.com/DoNewsCode/core/otgorm"
  "github.com/DoNewsCode/core/srvhttp"
  "github.com/gorilla/mux"
  "gorm.io/gorm"
)

type User struct {
  Id   string
  Name string
}

type Repository struct {
  DB *gorm.DB
}

func (r Repository) Find(id string) (*User, error) {
  var user User
  if err := r.DB.First(&user, id).Error; err != nil {
    return nil, err
  }
  return &user, nil
}

type Handler struct {
  R Repository
}

func (h Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
  encoder := srvhttp.NewResponseEncoder(writer)
  encoder.Encode(h.R.Find(request.URL.Query().Get("id")))
}

type Module struct {
  H Handler
}

func New(db *gorm.DB) Module {
  return Module{Handler{Repository{db}}}
}

func (m Module) ProvideHTTP(router *mux.Router) {
  router.HandleFunc("/", m.H)
}

func main() {
  // 步骤一: 从配置文件创建核心
  c := core.New(core.WithYamlFile("config.yaml"))

  // 步骤二: 绑定依赖
  c.Provide(otgorm.Providers())

  // 步骤三: 定义服务
  c.AddModuleFunc(New)

  // 步骤四: 启动
  c.Serve(context.Background())
}
```

第三阶段已经被 c.AddModuleFunc（新）所取代。AddModuleFunc 将参数从依赖项容器填充到 New，并将返回的模块实例添加到内部模块注册表。

当调用 c.Serve（）时，将扫描所有注册的模块以查找实现的接口。示例中的模块实现接口：

```go
type HTTPProvider interface {
	ProvideHTTP(router *mux.Router)
}
```

因此，核心知道该模块想要公开 HTTP 服务，并随后通过路由器调用 ProvideHttp。您可以注册多个模块，每个模块可以实现一个或多个服务。

现在我们有了一个完全可行的项目，包括处理程序、存储库和实体层。如果这是一个 DDD 研讨会，我们将进一步扩展这个例子。

也就是说，让我们把注意力转移到 core 提供的其他商品上：

- `core`本身支持多路模块复用。您可以将项目作为一个包含多个模块的整体开始，然后逐渐将它们迁移到微服务中。
- `core`不锁定在传输或框架中。例如，您可以使用 go kit 构建服务，并利用 gRPC、AMPQ、thrift 等。还支持 CLI 和 Cron 等非网络服务。
- 子包围绕服务协调提供支持，包括但不限于分布式跟踪、度量导出、错误处理、事件调度和领导人选举。

# 文档

- [Tutorial](https://github.com/DoNewsCode/core/blob/master/doc/tutorial.md)
- [GoDoc](https://pkg.go.dev/github.com/DoNewsCode/core)
- [Demo Project](https://github.com/DoNewsCode/skeleton)
- [Contributing](https://github.com/DoNewsCode/core/blob/master/doc/contributing.markdown)

# 设计原则

- 无包全局状态。
- 促进依赖注入。
- 可测试代码。
- 极简界面设计。易于装饰和更换。
- 与围棋生态系统合作，而不是重新发明轮子。
- 端到端上下文传递。

# 非目标

- 非目标
- 尝试成为 Spring、Laravel 或 Ruby-on-Rails。
- 尽量关心服务细节。
- 尝试重新实现现代平台提供的功能。

# 建议的服务框架

- Gin（如果仅限于 HTTP）
- Go 套件（如果有多个运输工具）
- 归零
