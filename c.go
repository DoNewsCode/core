/*
Package core is a service container that elegantly bootstrap and coordinate
twelve-factor apps in Go.

Checkout the README.md for an overview of this package.
*/
package core

import (
	"context"
	"fmt"
	"reflect"

	"github.com/DoNewsCode/core/codec/yaml"
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/config/watcher"
	"github.com/DoNewsCode/core/container"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/logging"

	"github.com/go-kit/log"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
)

// C stands for the core of the application. It contains service definitions and
// dependencies. C means to be used in the boostrap phase of the application.
// Do not pass C into services and use it as a service locator.
type C struct {
	appName    contract.AppName
	env        contract.Env
	conf       contract.ConfigAccessor
	logger     logging.LevelLogger
	container  *container.Container
	di         *dig.Container
	baseLogger log.Logger
}

// ConfParser models a parser for configuration. For example, yaml.Parser.
type ConfParser interface {
	Unmarshal([]byte) (map[string]any, error)
	Marshal(map[string]any) ([]byte, error)
}

// ConfProvider models a configuration provider. For example, file.Provider.
type ConfProvider interface {
	ReadBytes() ([]byte, error)
	Read() (map[string]any, error)
}

// ServeFn is an identifier in DI container that represents the function to be called by core.Serve.
type ServeFn func(ctx context.Context) error

// ConfigProvider provides contract.ConfigAccessor to the core.
type ConfigProvider func(configStack []config.ProviderSet, configWatcher contract.ConfigWatcher) contract.ConfigUnmarshaler

// DiProvider provides the *dig.Container to the core.
type DiProvider func(conf contract.ConfigUnmarshaler) *dig.Container

// AppNameProvider provides the contract.AppName to the core.
type AppNameProvider func(conf contract.ConfigUnmarshaler) contract.AppName

// EnvProvider provides the contract.Env to the core.
type EnvProvider func(conf contract.ConfigUnmarshaler) contract.Env

// LoggerProvider provides the log.Logger to the core.
type LoggerProvider func(conf contract.ConfigUnmarshaler, appName contract.AppName, env contract.Env) log.Logger

type coreValues struct {
	// Base Values
	configStack   []config.ProviderSet
	configWatcher contract.ConfigWatcher
	// ConfProvider functions
	configProvider  ConfigProvider
	diProvider      DiProvider
	appNameProvider AppNameProvider
	envProvider     EnvProvider
	loggerProvider  LoggerProvider
}

// CoreOption is the option to modify core attribute.
type CoreOption func(*coreValues)

// WithYamlFile is a two-in-one coreOption. It uses the configuration file as the
// source of configuration, and watches the change of that file for hot reloading.
func WithYamlFile(path string) (CoreOption, CoreOption) {
	return WithConfigStack(file.Provider(path), config.CodecParser{Codec: yaml.Codec{}}),
		WithConfigWatcher(watcher.File{Path: path})
}

// WithInline is a CoreOption that creates a inline config in the configuration stack.
func WithInline(key string, entry any) CoreOption {
	return WithConfigStack(confmap.Provider(map[string]any{
		key: entry,
	}, "."), nil)
}

// WithConfigStack is a CoreOption that defines a configuration layer. See package config for details.
func WithConfigStack(provider ConfProvider, parser ConfParser) CoreOption {
	return func(values *coreValues) {
		values.configStack = append(values.configStack, config.ProviderSet{Parser: parser, Provider: provider})
	}
}

// WithConfigWatcher is a CoreOption that adds a config watcher to the core (for hot reloading configs).
func WithConfigWatcher(watcher contract.ConfigWatcher) CoreOption {
	return func(values *coreValues) {
		values.configWatcher = watcher
	}
}

// SetConfigProvider is a CoreOption to replaces the default ConfigProvider.
func SetConfigProvider(provider ConfigProvider) CoreOption {
	return func(values *coreValues) {
		values.configProvider = provider
	}
}

// SetAppNameProvider is a CoreOption to replaces the default AppNameProvider.
func SetAppNameProvider(provider AppNameProvider) CoreOption {
	return func(values *coreValues) {
		values.appNameProvider = provider
	}
}

// SetEnvProvider is a CoreOption to replaces the default EnvProvider.
func SetEnvProvider(provider EnvProvider) CoreOption {
	return func(values *coreValues) {
		values.envProvider = provider
	}
}

// SetLoggerProvider is a CoreOption to replaces the default LoggerProvider.
func SetLoggerProvider(provider LoggerProvider) CoreOption {
	return func(values *coreValues) {
		values.loggerProvider = provider
	}
}

// SetDiProvider is a CoreOption to replaces the default DiContainer.
func SetDiProvider(provider DiProvider) CoreOption {
	return func(values *coreValues) {
		values.diProvider = provider
	}
}

// New creates a new bare-bones C.
func New(opts ...CoreOption) *C {
	values := coreValues{
		configStack:     []config.ProviderSet{},
		configWatcher:   nil,
		configProvider:  ProvideConfig,
		appNameProvider: ProvideAppName,
		envProvider:     ProvideEnv,
		loggerProvider:  ProvideLogger,
		diProvider:      ProvideDi,
	}
	for _, f := range opts {
		f(&values)
	}
	conf := values.configProvider(values.configStack, values.configWatcher)
	env := values.envProvider(conf)
	appName := values.appNameProvider(conf)
	logger := values.loggerProvider(conf, appName, env)
	diContainer := values.diProvider(conf)

	c := C{
		appName:    appName,
		env:        env,
		conf:       config.WithAccessor(conf),
		logger:     logging.WithLevel(logger),
		container:  &container.Container{},
		di:         diContainer,
		baseLogger: logger,
	}
	return &c
}

// Default creates a core.C under its default state. Core dependencies are
// already provided, and the config module and serve module are bundled.
func Default(opts ...CoreOption) *C {
	c := New(opts...)
	c.ProvideEssentials()
	return c
}

// AddModule adds a module to the core.
//
// A Module is a group of functionality. It must provide some runnable stuff:
// http handlers, grpc handlers, cron jobs, one-time command, etc.
//
// Optionally if the module embeds di.In, its fields will be injected by DI
// container. The semantics of injection follows the same rule of dig.Invoke.
// Note that the module added in this way will not retain any original field
// values, i.e. the module will only contain fields populated by DI container.
func (c *C) AddModule(module any) {
	t := reflect.TypeOf(module)
	if t.Kind() == reflect.Ptr && dig.IsIn(t.Elem()) {
		err := di.IntoPopulator(c.di).Populate(module)
		if err != nil {
			panic(err)
		}
		c.container.AddModule(module)
		return
	}
	if dig.IsIn(t) {
		copy := reflect.New(t)
		err := di.IntoPopulator(c.di).Populate(copy.Interface())
		if err != nil {
			panic(err)
		}
		c.container.AddModule(copy.Elem().Interface())
		return
	}
	c.container.AddModule(module)
}

// Provide adds dependencies provider to the core. Note the dependency provider
// must be a function in the form of:
//
//  func(foo Foo) Bar
//
// where foo is the upstream dependency and Bar is the provided type. The order
// for providers doesn't matter. They are only executed lazily when the Invoke is
// called.
//
// This method internally calls uber's dig library. Consult dig's documentation
// for details. (https://pkg.go.dev/go.uber.org/dig)
//
// The difference is, core.Provide has been made to accommodate the convention
// from google/wire (https://github.com/google/wire). All "func()" returned by
// constructor are treated as clean up functions. It also examines if the
// dependency implements the modular interface. If so, this dependency will be
// added the module collection.
func (c *C) Provide(deps di.Deps) {
	for _, dep := range deps {
		c.provide(dep)
	}
}

func (c *C) provide(constructor any) {
	var (
		options        []dig.ProvideOption
		shouldMakeFunc bool
	)

	if op, ok := constructor.(di.OptionalProvider); ok {
		constructor = op.Constructor
		options = op.Options
	}

	ftype := reflect.TypeOf(constructor)
	if ftype == nil {
		panic("can't provide an untyped nil")
	}
	if ftype.Kind() != reflect.Func {
		panic(fmt.Sprintf("must provide constructor function, got %v (type %v)", constructor, ftype))
	}

	inTypes := make([]reflect.Type, 0)
	outTypes := make([]reflect.Type, 0)
	for i := 0; i < ftype.NumOut(); i++ {
		outT := ftype.Out(i)
		if isCleanup(outT) {
			shouldMakeFunc = true
			continue
		}
		if isModule(outT) {
			shouldMakeFunc = true
		}
		outTypes = append(outTypes, outT)
	}

	for i := 0; i < ftype.NumIn(); i++ {
		inT := ftype.In(i)
		if isModule(inT) {
			shouldMakeFunc = true
		}
		inTypes = append(inTypes, inT)
	}

	// no cleanup or module, we can use normal dig.
	if !shouldMakeFunc {
		err := c.di.Provide(constructor, options...)
		if err != nil {
			panic(err)
		}
		return
	}

	// has cleanup or module, use reflect.MakeFunc as interceptor.
	fnType := reflect.FuncOf(inTypes, outTypes, ftype.IsVariadic() /* variadic */)
	fn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		filteredOuts := make([]reflect.Value, 0)
		outVs := reflect.ValueOf(constructor).Call(args)
		for _, v := range outVs {
			vType := v.Type()
			if isCleanup(vType) {
				c.AddModule(cleanup(v.Interface().(func())))
				continue
			}
			if isModule(vType) {
				c.AddModule(v.Interface().(di.Modular).Module())
			}
			filteredOuts = append(filteredOuts, v)
		}
		return filteredOuts
	})
	options = append([]dig.ProvideOption{dig.LocationForPC(reflect.ValueOf(constructor).Pointer())}, options...)
	err := c.di.Provide(fn.Interface(), options...)
	if err != nil {
		panic(err)
	}
}

// ProvideEssentials adds the default core dependencies to the core.
func (c *C) ProvideEssentials() {
	type coreDependencies struct {
		di.Out

		Env               contract.Env
		AppName           contract.AppName
		Container         contract.Container
		ConfigUnmarshaler contract.ConfigUnmarshaler
		ConfigAccessor    contract.ConfigAccessor
		ConfigRouter      contract.ConfigRouter
		ConfigWatcher     contract.ConfigWatcher
		DIPopulator       contract.DIPopulator
		Logger            log.Logger
		LevelLogger       logging.LevelLogger
		Lifecycles        lifecycleOut
		DefaultConfigs    []config.ExportedConfig `group:"config,flatten"`
	}

	c.provide(func() coreDependencies {
		coreDependencies := coreDependencies{
			Env:               c.env,
			AppName:           c.appName,
			Container:         c.container,
			ConfigUnmarshaler: c.conf,
			ConfigAccessor:    c.conf,
			Logger:            c.baseLogger,
			LevelLogger:       c.logger,
			DIPopulator:       di.IntoPopulator(c.di),
			Lifecycles:        provideLifecycle(),
			DefaultConfigs:    provideDefaultConfig(),
		}
		if cc, ok := c.conf.(contract.ConfigRouter); ok {
			coreDependencies.ConfigRouter = cc
		}
		if cc, ok := c.conf.(contract.ConfigWatcher); ok {
			coreDependencies.ConfigWatcher = cc
		}
		return coreDependencies
	})
}

// Serve runs the serve command bundled in the core.
// For larger projects, consider use full-featured ServeModule instead of calling serve directly.
func (c *C) Serve(ctx context.Context) error {
	type serveFn struct {
		di.In

		ServeFn ServeFn `optional:"true"`
	}
	return c.di.Invoke(func(fn serveFn, in serveIn) error {
		if fn.ServeFn != nil {
			return fn.ServeFn(ctx)
		}
		cmd := newServeCmd(in)
		return cmd.ExecuteContext(ctx)
	})
}

// Shutdown iterates through every CloserProvider registered in the container,
// and calls them in the reversed order of registration.
func (c *C) Shutdown() {
	modules := c.Modules()
	for i := range modules {
		if closer, ok := modules[len(modules)-i-1].(CloserProvider); ok {
			closer.ProvideCloser()
		}
	}
}

// AddModuleFunc add the module after Invoking its constructor. Clean up
// functions and errors are handled automatically.
func (c *C) AddModuleFunc(constructor any) {
	c.provide(constructor)
	ftype := reflect.TypeOf(constructor)
	targetTypes := make([]reflect.Type, 0)
	for i := 0; i < ftype.NumOut(); i++ {
		if isErr(ftype.Out(i)) {
			continue
		}
		if isCleanup(ftype.Out(i)) {
			continue
		}
		outT := ftype.Out(i)
		targetTypes = append(targetTypes, outT)
	}

	fnType := reflect.FuncOf(targetTypes, nil, false /* variadic */)
	fn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		for _, arg := range args {
			c.AddModule(arg.Interface())
		}
		return nil
	})

	err := c.di.Invoke(fn.Interface())
	if err != nil {
		panic(err)
	}
}

// ApplyRootCommand iterates through every CommandProvider registered in the container,
// and introduce the root *cobra.Command to everyone.
func (c *C) ApplyRootCommand(command *cobra.Command) {
	modules := c.Modules()
	for i := range modules {
		if p, ok := modules[i].(CommandProvider); ok {
			p.ProvideCommand(command)
		}
	}
}

// Invoke runs the given function after instantiating its dependencies. Any
// arguments that the function has are treated as its dependencies. The
// dependencies are instantiated in an unspecified order along with any
// dependencies that they might have. The function may return an error to
// indicate failure. The error will be returned to the caller as-is.
//
// It internally calls uber's dig library. Consult dig's documentation for
// details. (https://pkg.go.dev/go.uber.org/dig)
func (c *C) Invoke(function any) {
	err := c.di.Invoke(function)
	if err != nil {
		panic(err)
	}
}

// Modules returns every Module registered in the core container,
func (c *C) Modules() []any {
	return c.container.Modules()
}

// Decorate provides a decorator for a type that has already been provided in the Container.
// See the documentation for uber's dig library for details.
func (c *C) Decorate(function any) {
	err := c.di.Decorate(function)
	if err != nil {
		panic(err)
	}
}

// Populate populates the given addr with the dependencies of the dereferenced type.
func (c *C) Populate(Addr any) {
	err := di.IntoPopulator(c.di).Populate(Addr)
	if err != nil {
		panic(err)
	}
}

func isCleanup(v reflect.Type) bool {
	if v.Kind() == reflect.Func && v.NumIn() == 0 && v.NumOut() == 0 {
		return true
	}
	return false
}

type cleanup func()

func (c cleanup) ProvideCloser() {
	c()
}

var _errType = reflect.TypeOf((*error)(nil)).Elem()

func isErr(v reflect.Type) bool {
	return v.Implements(_errType)
}

var _moduleType = reflect.TypeOf((*di.Modular)(nil)).Elem()

func isModule(v reflect.Type) bool {
	return v.Implements(_moduleType)
}
