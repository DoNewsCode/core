package config

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/knadh/koanf"
	"github.com/mitchellh/mapstructure"
)

type KoanfAdapter struct {
	layers    []ProviderSet
	watcher   contract.ConfigWatcher
	delimiter string
	K         *koanf.Koanf
}

type ProviderSet struct {
	Parser   koanf.Parser
	Provider koanf.Provider
}

type Option func(option *KoanfAdapter)

func WithProviderLayer(provider koanf.Provider, parser koanf.Parser) Option {
	return func(option *KoanfAdapter) {
		option.layers = append(option.layers, ProviderSet{Provider: provider, Parser: parser})
	}
}

func WithWatcher(watcher contract.ConfigWatcher) Option {
	return func(option *KoanfAdapter) {
		option.watcher = watcher
	}
}

func WithDelimiter(delimiter string) Option {
	return func(option *KoanfAdapter) {
		option.delimiter = delimiter
	}
}

func NewConfig(options ...Option) (*KoanfAdapter, error) {
	adapter := KoanfAdapter{delimiter: "."}

	for _, f := range options {
		f(&adapter)
	}

	adapter.K = koanf.New(adapter.delimiter)

	if err := adapter.Reload(); err != nil {
		return nil, err
	}

	return &adapter, nil
}

func (k KoanfAdapter) Reload() error {
	for i := len(k.layers) - 1; i >= 0; i-- {
		err := k.K.Load(k.layers[i].Provider, k.layers[i].Parser)
		if err != nil {
			return fmt.Errorf("unable to load config %w", err)
		}
	}
	return nil
}

func (k KoanfAdapter) Watch(ctx context.Context) error {
	return k.watcher.Watch(ctx, k.Reload)
}

func (k KoanfAdapter) Unmarshal(path string, o interface{}) error {
	return k.K.UnmarshalWithConf(path, o, koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			Result:           o,
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
	if !val.Elem().CanSet() {
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
