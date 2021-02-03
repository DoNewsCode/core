package core

import (
	stdlog "log"
	"os"

	"github.com/DoNewsCode/std/pkg/config"
	"github.com/DoNewsCode/std/pkg/container"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/event"
	"github.com/DoNewsCode/std/pkg/logging"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
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

type Parser interface {
	Unmarshal([]byte) (map[string]interface{}, error)
	Marshal(map[string]interface{}) ([]byte, error)
}

type Provider interface {
	ReadBytes() ([]byte, error)
	Read() (map[string]interface{}, error)
}

type ConfigProvider func(configStack []config.ProviderSet, configWatcher contract.ConfigWatcher) contract.ConfigAccessor
type EventDispatcherProvider func(conf contract.ConfigAccessor) contract.Dispatcher
type DiProvider func(conf contract.ConfigAccessor) *dig.Container
type AppNameProvider func(conf contract.ConfigAccessor) contract.AppName
type EnvProvider func(conf contract.ConfigAccessor) contract.Env
type LoggerProvider func(conf contract.ConfigAccessor, appName contract.AppName, env contract.Env) log.Logger

type coreValues struct {
	// Base Values
	configStack   []config.ProviderSet
	configWatcher contract.ConfigWatcher
	// Provider functions
	configProvider          ConfigProvider
	eventDispatcherProvider EventDispatcherProvider
	diProvider              DiProvider
	appNameProvider         AppNameProvider
	envProvider             EnvProvider
	loggerProvider          LoggerProvider
}

type CoreOption func(*coreValues)

func WithYamlFile(path string) CoreOption {
	return WithConfigStack(file.Provider(path), yaml.Parser())
}

func WithConfigStack(provider Provider, parser Parser) CoreOption {
	return func(values *coreValues) {
		values.configStack = append(values.configStack, config.ProviderSet{Parser: parser, Provider: provider})
	}
}

func WithConfigWatcher(watcher contract.ConfigWatcher) CoreOption {
	return func(values *coreValues) {
		values.configWatcher = watcher
	}
}

func SetConfigProvider(provider ConfigProvider) CoreOption {
	return func(values *coreValues) {
		values.configProvider = provider
	}
}

func SetAppNameProvider(provider AppNameProvider) CoreOption {
	return func(values *coreValues) {
		values.appNameProvider = provider
	}
}

func SetEnvProvider(provider EnvProvider) CoreOption {
	return func(values *coreValues) {
		values.envProvider = provider
	}
}

func SetLoggerProvider(provider LoggerProvider) CoreOption {
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

func New(opts ...CoreOption) *C {
	values := coreValues{
		configStack:             []config.ProviderSet{},
		configWatcher:           nil,
		configProvider:          ProvideConfig,
		appNameProvider:         ProvideAppName,
		envProvider:             ProvideEnv,
		loggerProvider:          ProvideLogger,
		diProvider:              ProvideDi,
		eventDispatcherProvider: ProvideEventDispatcher,
	}
	for _, f := range opts {
		f(&values)
	}
	conf := values.configProvider(values.configStack, values.configWatcher)
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

func (c *C) Shutdown() {
	for _, f := range c.CloserProviders {
		f()
	}
}

func ProvideConfig(configStack []config.ProviderSet, configWatcher contract.ConfigWatcher) contract.ConfigAccessor {
	var (
		stack []config.Option
		err   error
		cfg   contract.ConfigAccessor
	)

	for _, layer := range configStack {
		stack = append(stack, config.WithProviderLayer(layer.Provider, layer.Parser))
	}
	stack = append(stack, config.WithProviderLayer(rawbytes.Provider([]byte(defaultConfig)), yaml.Parser()))
	if configWatcher != nil {
		stack = append(stack, config.WithWatcher(configWatcher))
	}

	cfg, err = config.NewConfig(stack...)
	if err != nil {
		stdlog.Fatal(err)
	}
	return cfg
}

func ProvideEnv(conf contract.ConfigAccessor) contract.Env {
	var env config.Env
	err := conf.Unmarshal("Env", &env)
	if err != nil {
		return config.NewEnv("local")
	}
	return env
}

func ProvideAppName(conf contract.ConfigAccessor) contract.AppName {
	var appName config.AppName
	err := conf.Unmarshal("name", &appName)
	if err != nil {
		return config.AppName("default")
	}
	return appName
}

func ProvideLogger(conf contract.ConfigAccessor, appName contract.AppName, env contract.Env) log.Logger {
	var (
		lvl    string
		format string
		err    error
	)
	err = conf.Unmarshal("level", &lvl)
	if err != nil {
		lvl = "debug"
	}
	err = conf.Unmarshal("format", &format)
	if err != nil {
		format = "logfmt"
	}
	logger := logging.NewLogger(format)
	logger = log.With(logger, "AppName", appName, "Env", env)
	logger = level.NewInjector(logger, level.DebugValue())
	return level.NewFilter(logger, logging.LevelFilter(lvl))
}

func ProvideDi(conf contract.ConfigAccessor) *dig.Container {
	return dig.New()
}

func ProvideEventDispatcher(conf contract.ConfigAccessor) contract.Dispatcher {
	return &event.SyncDispatcher{}
}
