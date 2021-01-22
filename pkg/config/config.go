package config

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/DoNewsCode/std/pkg/config/watcher"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/mitchellh/mapstructure"
)

type KoanfAdapter struct {
	contract.ConfigWatcher
	K *koanf.Koanf
}

type configOption struct {
	filePath string
	parser koanf.Parser
	provider koanf.Provider
	watcher Watcher
	delim string
}

type Option func(option *configOption)

func WithDelimiter(delim string) Option {
	return func(option *configOption) {
		option.delim = delim
	}
}

func WithParser(parser koanf.Parser) Option {
	return func(option *configOption) {
		option.parser = parser
	}
}

func WithProvider(provider koanf.Provider) Option {
	return func(option *configOption) {
		option.provider = provider
	}
}

func WithFilePath(path string) Option {
	return func(option *configOption) {
		option.filePath = path
	}
}

func WithWatcher(watcher Watcher) Option {
	return func(option *configOption) {
		option.watcher = watcher
	}
}

func NewConfig(options ...Option) (*KoanfAdapter, error) {
	defaults := configOption{
		filePath: "./config/config.yaml",
		parser: yaml.Parser(),
		delim: ".",
	}

	for _, f := range options {
		f(&defaults)
	}

	if defaults.provider == nil {
		defaults.provider = file.Provider(defaults.filePath)
	}
	if defaults.watcher == nil {
		switch defaults.provider.(type) {
		case *file.File:
			defaults.watcher = watcher.File{
				Path: defaults.filePath,
			}
		default:
		}
	}

	k := koanf.New(defaults.delim)
	err := k.Load(defaults.provider, defaults.parser)
	if err != nil {
		return nil, fmt.Errorf("unable to load config %w", err)
	}
	
	return &KoanfAdapter{
		ConfigWatcher: defaults.watcher,
		K:             k,
	}, nil
}

func (k KoanfAdapter) Unmarshal(path string, o interface{}) error {
	return k.K.UnmarshalWithConf(path, o, koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			Result: o,
			ErrorUnused:      true,
			WeaklyTypedInput: true,
		},
	})
}

func (k KoanfAdapter) Route(s string) contract.ConfigAccessor {
	return KoanfAdapter{
		K: k.K.Cut(s),
	}
}

func (k KoanfAdapter) String(s string) string {
	return k.K.String(s)
}

func (k KoanfAdapter) Int(s string) int {
	return k.K.Int(s)
}

func (k KoanfAdapter) Strings(s string) []string {
	return k.K.Strings(s)
}

func (k KoanfAdapter) Bool(s string) bool {
	return k.K.Bool(s)
}

func (k KoanfAdapter) Get(s string) interface{} {
	return k.K.Get(s)
}

func (k KoanfAdapter) Float64(s string) float64 {
	return k.K.Float64(s)
}

// MapAdapter implements ConfigAccessor and ConfigRouter.
// It is primarily used for testing
type MapAdapter map[string]interface{}

func (m MapAdapter) String(s string) string {
	return m[s].(string)
}

func (m MapAdapter) Int(s string) int {
	return m[s].(int)
}

func (m MapAdapter) Strings(s string) []string {
	return m[s].([]string)
}

func (m MapAdapter) Bool(s string) bool {
	return m[s].(bool)
}

func (m MapAdapter) Get(s string) interface{} {
	return m[s]
}

func (m MapAdapter) Float64(s string) float64 {
	return m[s].(float64)
}

func (m MapAdapter) Unmarshal(path string, o interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()
	var out interface{}
	out = m
	if path != "" {
		out = m[path]
	}
	val := reflect.ValueOf(o)
	if ! val.Elem().CanSet() {
		return errors.New("target cannot be set")
	}
	val.Elem().Set(reflect.ValueOf(out))
	return
}

func (m MapAdapter) Route(s string) contract.ConfigAccessor {
	var v interface{}
	v = m
	if s != "" {
		v = m[s]
	}

	switch v.(type) {
	case map[string]interface{}:
		return MapAdapter(v.(map[string]interface{}))
	case MapAdapter:
		return v.(MapAdapter)
	default:
		panic(fmt.Sprintf("value at path %s is not a valid Router", s))
	}
}

type Watcher interface {
	Watch(ctx context.Context, reload func() error) error
}
