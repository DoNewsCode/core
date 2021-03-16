package core

import (
	stdlog "log"

	"github.com/DoNewsCode/core/di"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/events"
	"github.com/DoNewsCode/core/logging"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
)

const defaultConfig = `
name: app
version: 0.1.0
env: local
http:
  addr: :8080
grpc:
  addr: :9090
log:
  level: debug
  format: logfmt
redis:
  default:
    addrs:
      - 127.0.0.1:6379
    db: 0
gorm:
  default:
    database: sqlite
    dsn: ":memory:"
`

// ProvideConfig is the default ConfigProvider for package Core.
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

// ProvideEnv is the default EnvProvider for package Core.
func ProvideEnv(conf contract.ConfigAccessor) contract.Env {
	var env config.Env
	err := conf.Unmarshal("env", &env)
	if err != nil {
		return config.EnvLocal
	}
	return env
}

// ProvideAppName is the default AppNameProvider for package Core.
func ProvideAppName(conf contract.ConfigAccessor) contract.AppName {
	var appName config.AppName
	err := conf.Unmarshal("name", &appName)
	if err != nil {
		return config.AppName("default")
	}
	return appName
}

// ProvideLogger is the default LoggerProvider for package Core.
func ProvideLogger(conf contract.ConfigAccessor, appName contract.AppName, env contract.Env) log.Logger {
	var (
		lvl    string
		format string
		err    error
	)
	err = conf.Unmarshal("log.level", &lvl)
	if err != nil {
		lvl = "debug"
	}
	err = conf.Unmarshal("log.format", &format)
	if err != nil {
		format = "logfmt"
	}
	logger := logging.NewLogger(format)
	logger = level.NewInjector(logger, level.DebugValue())
	return level.NewFilter(logger, logging.LevelFilter(lvl))
}

// ProvideDi is the default DiProvider for package Core.
func ProvideDi(conf contract.ConfigAccessor) DiContainer {
	return di.NewGraph()
}

// ProvideEventDispatcher is the default EventDispatcherProvider for package Core.
func ProvideEventDispatcher(conf contract.ConfigAccessor) contract.Dispatcher {
	return &events.SyncDispatcher{}
}

// provideDefaultConfig exports config for "name", "version", "env", "http", "grpc".
func provideDefaultConfig() []config.ExportedConfig {
	return []config.ExportedConfig{
		{
			Owner: "core",
			Data: map[string]interface{}{
				"name": "app",
			},
			Comment: "The name of the application",
		},
		{
			Owner: "core",
			Data: map[string]interface{}{
				"version": "0.1.0",
			},
			Comment: "The version of the application",
		},
		{
			Owner: "core",
			Data: map[string]interface{}{
				"env": "local",
			},
			Comment: "The environment of the application, one of production, development, staging, testing or local",
		},
		{
			Owner: "core",
			Data: map[string]interface{}{
				"http": map[string]interface{}{
					"addr": ":8080",
				},
			},
			Comment: "The http address",
		},
		{
			Owner: "core",
			Data: map[string]interface{}{
				"grpc": map[string]interface{}{
					"addr": ":9090",
				},
			},
			Comment: "The gRPC address",
		},
		{
			Owner: "core",
			Data: map[string]interface{}{
				"log": map[string]interface{}{"level": "debug", "format": "logfmt"},
			},
			Comment: "The global logging level and format",
		},
	}
}
