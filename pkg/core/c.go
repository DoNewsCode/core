package core

import (
	"fmt"
	"reflect"

	"github.com/DoNewsCode/std/pkg/config"
	"github.com/DoNewsCode/std/pkg/config/watcher"
	"github.com/DoNewsCode/std/pkg/container"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/di"
	"github.com/DoNewsCode/std/pkg/logging"
	"github.com/go-kit/kit/log"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
)

type C struct {
	AppName contract.AppName
	Env     contract.Env
	contract.ConfigAccessor
	contract.LevelLogger
	contract.Container
	contract.Dispatcher
	di contract.DiContainer
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
type DiProvider func(conf contract.ConfigAccessor) contract.DiContainer
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

func WithYamlFile(path string) (CoreOption, CoreOption) {
	return WithConfigStack(file.Provider(path), yaml.Parser()),
		WithConfigWatcher(watcher.File{Path: path})
}

// WithInline is an coreoption that creates a inline config in the configuration stack.
func WithInline(key, entry string) CoreOption {
	return WithConfigStack(confmap.Provider(map[string]interface{}{
		key: entry,
	}, "."), nil)
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

func SetDiProvider(provider func(conf contract.ConfigAccessor) contract.DiContainer) CoreOption {
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
	diContainer := values.diProvider(conf)
	dispatcher := values.eventDispatcherProvider(conf)

	var c = C{
		AppName:        appName,
		Env:            env,
		ConfigAccessor: conf,
		LevelLogger:    logging.WithLevel(logger),
		Container:      &container.Container{},
		Dispatcher:     dispatcher,
		di:             diContainer,
	}
	return &c
}

func (c *C) AddModule(modules ...interface{}) {
	for i := range modules {
		switch modules[i].(type) {
		case error:
			panic(modules[i].(error))
		default:
			c.Container.AddModule(modules[i])
		}
	}
}

func (c *C) addDependency(dep interface{}) {
	inTypes := make([]reflect.Type, 0)
	outTypes := make([]reflect.Type, 0)
	depType := reflect.TypeOf(dep)
	if isModule(depType) {
		c.AddModule(dep)
	}
	outTypes = append(outTypes, reflect.TypeOf(dep))
	fnType := reflect.FuncOf(inTypes, outTypes, false /* variadic */)
	fn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		return []reflect.Value{reflect.ValueOf(dep)}
	})
	_ = c.di.Provide(fn.Interface())
}

func (c *C) AddDependencyFunc(constructor interface{}) {
	var shoudMakeFunc bool

	ftype := reflect.TypeOf(constructor)
	if ftype.Kind() != reflect.Func {
		panic("AddDependencyFunc only accepts function as argument")
	}
	inTypes := make([]reflect.Type, 0)
	outTypes := make([]reflect.Type, 0)
	for i := 0; i < ftype.NumOut(); i++ {
		outT := ftype.Out(i)
		if isCleanup(outT) {
			shoudMakeFunc = true
			continue
		}
		if isModule(outT) {
			shoudMakeFunc = true
		}
		outTypes = append(outTypes, outT)
	}

	for i := 0; i < ftype.NumIn(); i++ {
		inT := ftype.In(i)
		if isModule(inT) {
			shoudMakeFunc = true
		}
		inTypes = append(inTypes, inT)
	}

	// no cleanup or module, we can use normal dig.
	if !shoudMakeFunc {
		err := c.di.Provide(constructor)
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
				c.AddModule(v.Interface())
				continue
			}
			if isModule(vType) {
				c.AddModule(v.Interface())
			}
			filteredOuts = append(filteredOuts, v)
		}
		return filteredOuts
	})
	err := c.di.Provide(fn.Interface())
	if err != nil {
		panic(err)
	}
}

type CoreDependencies struct {
	di.Out

	Env            contract.Env
	AppName        contract.AppName
	Container      contract.Container
	ConfigAccessor contract.ConfigAccessor
	ConfigRouter   contract.ConfigRouter
	ConfigWatcher  contract.ConfigWatcher
	Logger         log.Logger
	Dispatcher     contract.Dispatcher
	DefaultConfigs []config.ExportedConfig `group:"config,flatten"`
}

func (c *C) AddCoreDependencies() {
	c.AddDependencyFunc(func() CoreDependencies {
		coreDependencies := CoreDependencies{
			Env:            c.Env,
			AppName:        c.AppName,
			Container:      c.Container,
			ConfigAccessor: c.ConfigAccessor,
			Logger:         c.LevelLogger,
			Dispatcher:     c.Dispatcher,
			DefaultConfigs: provideDefaultConfig(),
		}
		if cc, ok := c.ConfigAccessor.(contract.ConfigRouter); ok {
			coreDependencies.ConfigRouter = cc
		}
		if cc, ok := c.ConfigAccessor.(contract.ConfigWatcher); ok {
			coreDependencies.ConfigWatcher = cc
		}

		return coreDependencies
	})
}

func (c *C) AddModuleFunc(function interface{}) {
	c.AddDependencyFunc(function)
	ftype := reflect.TypeOf(function)
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

func (c *C) Invoke(function interface{}) error {
	return c.di.Invoke(function)
}

func (c *C) populate(targets ...interface{}) error {
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

	fnType := reflect.FuncOf(targetTypes, nil, false)
	fn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		for i, arg := range args {
			reflect.ValueOf(targets[i]).Elem().Set(arg)
		}
		return nil
	})
	return c.Invoke(fn.Interface())
}

func isCleanup(v reflect.Type) bool {
	if v.Kind() == reflect.Func && v.NumIn() == 0 && v.NumOut() == 0 {
		return true
	}
	return false
}

var _errType = reflect.TypeOf((*error)(nil)).Elem()

func isErr(v reflect.Type) bool {
	return v.Implements(_errType)
}

var _moduleType = reflect.TypeOf((*di.Module)(nil)).Elem()

func isModule(v reflect.Type) bool {
	return v.Implements(_moduleType)
}
