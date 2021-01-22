package core

import (
	stdlog "log"

	"errors"

	"github.com/go-kit/kit/log"
	"github.com/DoNewsCode/std/pkg/config"
	"github.com/DoNewsCode/std/pkg/container"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/logging"
)

type C struct {
	contract.ConfigRouter
	contract.ConfigAccessor
	contract.LevelLogger
	container.Container
}

func New(env contract.Env) *C {
	logger := logging.NewLogger(env)
	return &C{
		ConfigRouter:   nil,
		ConfigAccessor: nil,
		LevelLogger: logging.WithLevel(logger),
		Container: container.Container{},
	}
}

func NewWithConf(cfgFile string) *C {
	conf, err := ProvideConfig(cfgFile)
	if err != nil {
		stdlog.Fatal(err)
	}

	logger := ProvideLogger(conf)
	return &C{
		ConfigRouter:   conf,
		ConfigAccessor: conf,
		LevelLogger: logging.WithLevel(logger),
		Container: container.Container{},
	}
}

func ProvideConfig(cfgFile string) (interface{
	contract.ConfigRouter
	contract.ConfigAccessor
}, error) {
	if cfgFile != "" {
		return config.NewConfig(config.WithFilePath(cfgFile))
	}
	return nil, errors.New("configuration path not provided")
}

func ProvideLogger(conf contract.ConfigAccessor) log.Logger {
	var env config.Env
	err := conf.Unmarshal("env", &env)
	if err != nil {
		env = config.NewEnv("local")
	}
	logger := logging.NewLogger(env)
	return logger
}
