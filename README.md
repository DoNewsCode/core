Package core is a service container that elegantly bootstrap and coordinate
twelve-factor apps in Go.

![GitHub Workflow Status](https://github.com/DoNewsCode/core/workflows/go/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/DoNewsCode/core.svg)](https://pkg.go.dev/github.com/DoNewsCode/core)
[![codecov](https://codecov.io/gh/DoNewsCode/core/branch/master/graph/badge.svg)](https://codecov.io/gh/go-kit/kit)
[![Go Report Card](https://goreportcard.com/badge/DoNewsCode/core)](https://goreportcard.com/report/go-kit/kit)
[![Sourcegraph](https://sourcegraph.com/github.com/DoNewsCode/core/-/badge.svg)](https://sourcegraph.com/github.com/go-kit/kit?badge)

## Background

The twelve-factor methodology has proven its worth over the years. Since its
invention many fields in technology have changed, many among them are shining
and exciting. In the age of Kubernetes, service mesh and serverless
architectures, the twelve-factor methodology has not faded away, but rather has
happened to be a good fit for nearly all of those powerful platforms.

Scaffolding a twelve-factor go app may not be a difficult task for experienced
engineers, but certainly presents some challenges to juniors. For those who are
capable of setting things up, there are still many decisions left to make, and choices
to be agreed upon within the team.

Package core was created to bootstrap and coordinate such services.

## Overview

Whatever the app is, the bootstrapping phase are roughly composed by:

- Read configuration from out of the binary. Namely, flags, environmental
  variables, and/or configuration files.

- Initialize dependencies. Databases, message queues, service discoveries, etc.

- Define how to run the app. HTTP, RPC, command-lines, cron jobs, or more often mixed.

Package core abstracts those repeated steps, keeping them concise, portable yet explicit. 
Let's see the following snippet:

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
  // Phase One: create a core from a configuration file
  c := core.New(core.WithYamlFile("config.yaml"))

  // Phase two: bind dependencies
  c.Provide(otgorm.Provide)

  // Phase three: define service
  c.AddModule(core.HttpFunc(func(router *mux.Router) {
    router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
      writer.Write([]byte("hello world"))
    })
  }))

  // Phase Four: run!
  c.Serve(context.Background())
}

```

In a few lines, an http service is bootstrapped in the style outlined above.
It is simple, explicit and to some extent, declarative.

The service demonstrated above uses an inline handler function to highlight the point.
Normally, for real projects, we will use modules instead. 
The "module" in package core's glossary is not necessarily a go module (though it can be). It is simply a group of services.

You may note that the http service doesn't really consume the dependency.
That's true.

Let's rewrite the http service to consume the above dependencies.

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

func (m Module) ProvideHttp(router *mux.Router) {
  router.Handle("/", m.H)
}

func main() {
  // Phase One: create a core from a configuration file
  c := core.New(core.WithYamlFile("config.yaml"))

  // Phase two: bind dependencies
  c.Provide(otgorm.Provide)

  // Phase three: define service
  c.AddModuleFunc(New)

  // Phase four: run!
  c.Serve(context.Background())
}
```

Phase three has been replaced by the `c.AddModuleFunc(New)`. `AddModuleFunc` populates the arguments to `New` from dependency containers
and add the returned module instance to the internal module registry.

Now we have a fully workable project, with layers of handler, repository and entity. 
Had this been a DDD workshop, we would be expanding the example even further. 

That being said, let's redirect our attention to other goodies package core has offered:

- Package core is natively support multiplexing modules. 
  You could start you project as a monolith with multiple modules, and gradually migrate them into microservices.

- Package core doesn't lock in transport or framework.
  For instance, you can use go kit to construct your service, and leveraging grpc, ampq, thrift, etc. Non network services like CLI and Cron are also supported.

- Sub packages provide support around service coordination, including but not limited to opentracing integration, metrics exporting, error handling, event-dispatching and leader election.

## Documentation

GoDoc is the primary source of documentation.

## Design Principles

- No package global state.
- Promote dependency injection.
- Testable code.
- Minimalist interface design. Easy to decorate and replace.
- Tries to work with the go ecosystem rather than reinventing the wheel.
- End to end Context passing.

## Non-Goals

- Tries to be a Laravel or Ruby on Rails.
- Tries to care about service details.

## Suggested service framework
- Gin (if http only)
- Go Kit (if multiple transport)
- Go Zero


