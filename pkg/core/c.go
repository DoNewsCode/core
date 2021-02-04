package core

import (
	"os"

	"github.com/DoNewsCode/std/pkg/config"
	"github.com/DoNewsCode/std/pkg/container"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/logging"
	"github.com/go-kit/kit/log"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
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
	return &c
}

func (c *C) AddModule(modules ...interface{}) {
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
			c.Container.AddModule(modules[i])
		}
	}
}

func (c *C) Shutdown() {
	for _, f := range c.CloserProviders {
		f()
	}
}
