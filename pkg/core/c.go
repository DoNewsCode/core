package core

import (
	"fmt"
	stdlog "log"
	"os"
	"reflect"

	"github.com/DoNewsCode/std/pkg/config"
	"github.com/DoNewsCode/std/pkg/container"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/event"
	"github.com/DoNewsCode/std/pkg/logging"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/knadh/koanf/providers/rawbytes"
	"go.uber.org/dig"
)

type C struct {
	AppName contract.AppName
	Env     contract.Env
	contract.ConfigAccessor
	contract.LevelLogger
	container.Container
	contract.Dispatcher
	di *dig.Container
}

type coreValues struct {
	eventDispatcherProvider func(conf contract.ConfigAccessor) contract.Dispatcher
	diProvider              func(conf contract.ConfigAccessor) *dig.Container
	configProvider          func(cfgFile string) contract.ConfigAccessor
	appNameProvider         func(conf contract.ConfigAccessor) contract.AppName
	envProvider             func(conf contract.ConfigAccessor) contract.Env
	loggerProvider          func(conf contract.ConfigAccessor, appName contract.AppName, env contract.Env) log.Logger
}

type CoreOption func(*coreValues)

func SetConfigProvider(provider func(cfgFile string) contract.ConfigAccessor) CoreOption {
	return func(values *coreValues) {
		values.configProvider = provider
	}
}

func SetAppNameProvider(provider func(conf contract.ConfigAccessor) contract.AppName) CoreOption {
	return func(values *coreValues) {
		values.appNameProvider = provider
	}
}

func SetEnvProvider(provider func(conf contract.ConfigAccessor) contract.Env) CoreOption {
	return func(values *coreValues) {
		values.envProvider = provider
	}
}

func SetLoggerProvider(provider func(
	conf contract.ConfigAccessor,
	appName contract.AppName,
	env contract.Env,
) log.Logger) CoreOption {
	return func(values *coreValues) {
		values.loggerProvider = provider
	}
}

func SetDiProvider(provider func(conf contract.ConfigAccessor) *dig.Container) CoreOption {
	return func(values *coreValues) {
		values.diProvider = provider
	}
}

func SetEventDispatcherProvider(provider func(conf contract.ConfigAccessor) contract.Dispatcher) CoreOption {
	return func(values *coreValues) {
		values.eventDispatcherProvider = provider
	}
}

func New(cfgFilePath string, opts ...CoreOption) *C {
	values := coreValues{
		configProvider:  ProvideConfig,
		appNameProvider: ProvideAppName,
		envProvider:     ProvideEnv,
		loggerProvider:  ProvideLogger,
		diProvider: func(_ contract.ConfigAccessor) *dig.Container {
			return dig.New()
		},
		eventDispatcherProvider: func(_ contract.ConfigAccessor) contract.Dispatcher {
			return &event.SyncDispatcher{}
		},
	}
	for _, f := range opts {
		f(&values)
	}
	conf := values.configProvider(cfgFilePath)
	env := values.envProvider(conf)
	appName := values.appNameProvider(conf)
	logger := values.loggerProvider(conf, appName, env)
	di := values.diProvider(conf)
	dispatcher := values.eventDispatcherProvider(conf)

	var c = C{
		AppName:        appName,
		Env:            env,
		ConfigAccessor: conf,
		LevelLogger:    logging.WithLevel(logger),
		Container:      container.Container{},
		Dispatcher:     dispatcher,
		di:             di,
	}
	populateContainer(&c)
	return &c
}

func populateContainer(c *C) {
	c.Provide(func() contract.Env {
		return c.Env
	})
	c.Provide(func() contract.AppName {
		return c.AppName
	})
	c.Provide(func() contract.ConfigAccessor {
		return c
	})
	c.Provide(func() contract.ConfigRouter {
		if cc, ok := c.ConfigAccessor.(contract.ConfigRouter); ok {
			return cc
		}
		return nil
	})
	c.Provide(func() contract.ConfigWatcher {
		if cc, ok := c.ConfigAccessor.(contract.ConfigWatcher); ok {
			return cc
		}
		return nil
	})
	c.Provide(func() log.Logger {
		return c.LevelLogger
	})
	c.Provide(func() contract.Dispatcher {
		return c.Dispatcher
	})
}

func (c *C) Register(modules ...interface{}) {
	for i := range modules {
		switch modules[i].(type) {
		case error:
			if modules[i].(error) != nil {
				c.Err(modules[i].(error))
				os.Exit(1)
			}
		case func():
			c.CloserProviders = append(c.CloserProviders, modules[i].(func()))
		default:
			c.Container.Register(modules[i])
		}
	}
}

func (c *C) Provide(constructor interface{}, opts ...dig.ProvideOption) {
	ftype := reflect.TypeOf(constructor)
	inTypes := make([]reflect.Type, 0)
	outTypes := make([]reflect.Type, 0)
	for i := 0; i < ftype.NumOut(); i++ {
		outT := ftype.Out(i)
		if isCleaup(outT) {
			continue
		}
		outTypes = append(outTypes, outT)
	}
	for i := 0; i < ftype.NumIn(); i++ {
		inT := ftype.In(i)
		inTypes = append(inTypes, inT)
	}
	fnType := reflect.FuncOf(inTypes, outTypes, ftype.IsVariadic() /* variadic */)
	fn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		filteredOuts := make([]reflect.Value, 0)
		outVs := reflect.ValueOf(constructor).Call(args)
		for _, v := range outVs {
			if isCleaup(v.Type()) {
				continue
			}
			filteredOuts = append(filteredOuts, v)
		}
		return filteredOuts
	})
	err := c.di.Provide(fn.Interface())
	if err != nil {
		c.Err(err)
		os.Exit(1)
	}
}

func isCleaup(v reflect.Type) bool {
	if v.Kind() == reflect.Func && v.NumIn() == 0 && v.NumOut() == 0 {
		return true
	}
	return false
}

func isErr(v reflect.Type) bool {
	if v.Implements(_errType) {
		return true
	}
	return false
}

var _errType = reflect.TypeOf((*error)(nil)).Elem()

func (c *C) RegisterFn(function interface{}) {
	c.Provide(function)

	ftype := reflect.TypeOf(function)
	targetTypes := make([]reflect.Type, 0)
	for i := 0; i < ftype.NumOut(); i++ {
		if isErr(ftype.Out(i)) {
			continue
		}
		outT := ftype.Out(i)
		targetTypes = append(targetTypes, outT)
	}
	fnType := reflect.FuncOf(targetTypes, nil, false /* variadic */)
	fn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		for _, arg := range args {
			c.Register(arg.Interface())
		}
		return nil
	})

	c.Invoke(fn.Interface())
}

func (c *C) Invoke(function interface{}, opts ...dig.InvokeOption) {
	err := c.di.Invoke(function, opts...)
	if err != nil {
		c.Err(err)
		os.Exit(1)
	}
}

func (c *C) Populate(targets ...interface{}) error {
	// Validate all targets are non-nil pointers.
	targetTypes := make([]reflect.Type, len(targets))
	for i, t := range targets {
		if t == nil {
			return fmt.Errorf("failed to Populate: target %v is nil", i+1)
		}
		rt := reflect.TypeOf(t)
		if rt.Kind() != reflect.Ptr {
			return fmt.Errorf("failed to Populate: target %v is not a pointer type, got %T", i+1, t)
		}

		targetTypes[i] = reflect.TypeOf(t).Elem()
	}

	// Build a function that looks like:
	//
	// func(t1 T1, t2 T2, ...) {
	//   *targets[0] = t1
	//   *targets[1] = t2
	//   [...]
	// }
	//
	fnType := reflect.FuncOf(targetTypes, nil, false /* variadic */)
	fn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		for i, arg := range args {
			reflect.ValueOf(targets[i]).Elem().Set(arg)
		}
		return nil
	})
	return c.di.Invoke(fn.Interface())
}

func ProvideConfig(cfgFile string) contract.ConfigAccessor {
	var (
		err error
		cfg contract.ConfigAccessor
	)
	if cfgFile == "" {
		cfg, _ = config.NewConfig(config.WithProvider(rawbytes.Provider([]byte(defaultConfig))))
		return cfg
	}
	cfg, err = config.NewConfig(config.WithFilePath(cfgFile))
	if err != nil {
		stdlog.Fatal(err)
	}
	return cfg
}

func ProvideEnv(conf contract.ConfigAccessor) contract.Env {
	envStr := os.Getenv("APP_ENV")
	if envStr != "" {
		return config.NewEnv(envStr)
	}
	err := conf.Unmarshal("env", &envStr)
	if err != nil {
		return config.NewEnv("local")
	}
	return config.NewEnv(envStr)
}

func ProvideAppName(conf contract.ConfigAccessor) contract.AppName {
	appName := os.Getenv("APP_NAME")
	if appName != "" {
		return config.AppName(appName)
	}
	err := conf.Unmarshal("name", &appName)
	if err != nil {
		return config.AppName("default")
	}
	return config.AppName(appName)
}

func ProvideLogger(conf contract.ConfigAccessor, appName contract.AppName, env contract.Env) log.Logger {
	var (
		lvl string
		err error
	)
	err = conf.Unmarshal("level", &lvl)
	if err != nil {
		lvl = "debug"
	}
	logger := logging.NewLogger(env)
	logger = log.With(logger, "appName", appName)
	logger = level.NewInjector(logger, level.DebugValue())
	return level.NewFilter(logger, logging.LevelFilter(lvl))
}
