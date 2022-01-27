package core

import (
	"fmt"
	stdlog "log"
	"net"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/contract/lifecycle"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/events"
	"github.com/DoNewsCode/core/logging"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"go.uber.org/dig"
)

const defaultConfig = `
name: app
env: local
http:
  addr: :8080
  disable: false
grpc:
  addr: :9090
  disable: false
cron:
  disable: false
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
func ProvideConfig(configStack []config.ProviderSet, configWatcher contract.ConfigWatcher) contract.ConfigUnmarshaler {
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
func ProvideEnv(conf contract.ConfigUnmarshaler) contract.Env {
	return config.NewEnvFromConf(conf)
}

// ProvideAppName is the default AppNameProvider for package Core.
func ProvideAppName(conf contract.ConfigUnmarshaler) contract.AppName {
	var appName config.AppName
	err := conf.Unmarshal("name", &appName)
	if err != nil {
		return config.AppName("default")
	}
	return appName
}

// ProvideLogger is the default LoggerProvider for package Core.
func ProvideLogger(conf contract.ConfigUnmarshaler, appName contract.AppName, env contract.Env) log.Logger {
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
	logger = level.NewFilter(logger, logging.LevelFilter(lvl))
	logger = level.NewInjector(logger, level.DebugValue())
	return logger
}

// ProvideDi is the default DiProvider for package Core.
func ProvideDi(conf contract.ConfigUnmarshaler) *dig.Container {
	return dig.New()
}

type lifecycleOut struct {
	di.Out
	lifecycle.ConfigReload
	lifecycle.HTTPServerStart
	lifecycle.HTTPServerShutdown
	lifecycle.GRPCServerStart
	lifecycle.GRPCServerShutdown
}

func provideLifecycle() lifecycleOut {
	return lifecycleOut{
		ConfigReload:       &events.Event[contract.ConfigUnmarshaler]{},
		HTTPServerStart:    &events.Event[lifecycle.HTTPServerStartPayload]{},
		HTTPServerShutdown: &events.Event[lifecycle.HTTPServerShutdownPayload]{},
		GRPCServerStart:    &events.Event[lifecycle.GRPCServerStartPayload]{},
		GRPCServerShutdown: &events.Event[lifecycle.GRPCServerShutdownPayload]{},
	}
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
			Validate: func(data map[string]interface{}) error {
				_, err := getString(data, "name")
				if err != nil {
					return fmt.Errorf("the name field is not valid: %w", err)
				}
				return nil
			},
		},
		{
			Owner: "core",
			Data: map[string]interface{}{
				"env": "local",
			},
			Comment: "The environment of the application, one of production, development, staging, testing or local",
			Validate: func(data map[string]interface{}) error {
				str, err := getString(data, "env")
				if err != nil {
					return fmt.Errorf("the env field is not valid: %w", err)
				}
				if config.NewEnv(str) != config.EnvUnknown {
					return nil
				}
				return fmt.Errorf(
					"the env field must be one of \"production\", \"development\", \"staging\", \"testing\" or \"local\", got %s", str)
			},
		},
		{
			Owner: "core",
			Data: map[string]interface{}{
				"http": map[string]interface{}{
					"addr":    ":8080",
					"disable": false,
				},
			},
			Comment: "The http address",
			Validate: func(data map[string]interface{}) error {
				disable, err := getBool(data, "http", "disable")
				if err != nil {
					return fmt.Errorf("the http.disable field is not valid: %w", err)
				}
				if disable {
					return nil
				}
				str, err := getString(data, "http", "addr")
				if err != nil {
					return fmt.Errorf("the http.addr field is not valid: %w", err)
				}
				if _, err := net.ResolveTCPAddr("tcp", str); err != nil {
					return fmt.Errorf("the http.addr field must be an valid address like :8080, got %s", str)
				}
				return nil
			},
		},
		{
			Owner: "core",
			Data: map[string]interface{}{
				"grpc": map[string]interface{}{
					"addr":    ":9090",
					"disable": false,
				},
			},
			Comment: "The gRPC address",
			Validate: func(data map[string]interface{}) error {
				disable, err := getBool(data, "grpc", "disable")
				if err != nil {
					return fmt.Errorf("the grpc.disable field is not valid: %w", err)
				}
				if disable {
					return nil
				}
				str, err := getString(data, "grpc", "addr")
				if err != nil {
					return fmt.Errorf("the grpc.addr field is not valid: %w", err)
				}
				if _, err := net.ResolveTCPAddr("tcp", str); err != nil {
					return fmt.Errorf("the grpc.addr field must be an valid address like :9090, got %s", str)
				}
				return nil
			},
		},
		{
			Owner: "core",
			Data: map[string]interface{}{
				"cron": map[string]interface{}{
					"disable": false,
				},
			},
			Comment: "The cron job runner",
			Validate: func(data map[string]interface{}) error {
				_, err := getBool(data, "cron", "disable")
				if err != nil {
					return fmt.Errorf("the cron.disable field is not valid: %w", err)
				}
				return nil
			},
		},
		{
			Owner: "core",
			Data: map[string]interface{}{
				"log": map[string]interface{}{"level": "debug", "format": "logfmt"},
			},
			Comment: "The global logging level and format",
			Validate: func(data map[string]interface{}) error {
				lvl, err := getString(data, "log", "level")
				if err != nil {
					return fmt.Errorf("the log.level field is not valid: %w", err)
				}
				if !isValidLevel(lvl) {
					return fmt.Errorf("allowed levels are \"debug\", \"info\", \"warn\", \"error\", or \"none\", got \"%s\"", lvl)
				}
				format, err := getString(data, "log", "format")
				if err != nil {
					return fmt.Errorf("the log.format field is not valid: %w", err)
				}
				if !isValidFormat(format) {
					return fmt.Errorf("the log format is not supported")
				}
				return nil
			},
		},
	}
}
