package core

import (
	stdlog "log"

	"errors"

	"github.com/DoNewsCode/std/pkg/config"
	"github.com/DoNewsCode/std/pkg/container"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/event"
	"github.com/DoNewsCode/std/pkg/logging"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type C struct {
	contract.ConfigRouter
	contract.ConfigAccessor
	contract.LevelLogger
	container.Container
	contract.Dispatcher
}

func New(env contract.Env) *C {
	logger := logging.NewLogger(env)
	return &C{
		ConfigRouter:   nil,
		ConfigAccessor: nil,
		LevelLogger:    logging.WithLevel(logger),
		Container:      container.Container{},
		Dispatcher:     &event.Dispatcher{},
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
		LevelLogger:    logging.WithLevel(logger),
		Container:      container.Container{},
		Dispatcher:     &event.Dispatcher{},
	}
}

func ProvideConfig(cfgFile string) (interface {
	contract.ConfigRouter
	contract.ConfigAccessor
}, error) {
	if cfgFile != "" {
		return config.NewConfig(config.WithFilePath(cfgFile))
	}
	return nil, errors.New("configuration path not provided")
}

func ProvideLogger(conf contract.ConfigAccessor) log.Logger {
	var (
		env config.Env
		lvl string
		err error
	)
	err = conf.Unmarshal("env", &env)
	if err != nil {
		env = config.NewEnv("local")
	}
	err = conf.Unmarshal("level", &lvl)
	if err != nil {
		lvl = "debug"
	}
	logger := logging.NewLogger(env)
	logger = level.NewInjector(logger, level.DebugValue())
	return level.NewFilter(logger, logging.LevelFilter(lvl))
}
