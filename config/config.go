package config

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/DoNewsCode/core/contract"
	"github.com/knadh/koanf"
	"github.com/mitchellh/mapstructure"
)

// KoanfAdapter is a implementation of contract.Config based on Koanf (https://github.com/knadh/koanf).
type KoanfAdapter struct {
	layers    []ProviderSet
	watcher   contract.ConfigWatcher
	delimiter string
	K         *koanf.Koanf
}

// ProviderSet is a configuration layer formed by a parser and a provider.
type ProviderSet struct {
	Parser   koanf.Parser
	Provider koanf.Provider
}

// Option is the functional option type for KoanfAdapter
type Option func(option *KoanfAdapter)

// WithProviderLayer is an option for *KoanfAdapter that adds a layer to the bottom of the configuration stack.
// This option can be used multiple times, thus forming the whole stack. The layer on top has higher priority.
func WithProviderLayer(provider koanf.Provider, parser koanf.Parser) Option {
	return func(option *KoanfAdapter) {
		option.layers = append(option.layers, ProviderSet{Provider: provider, Parser: parser})
	}
}

// WithWatcher is an option for *KoanfAdapter that adds a config watcher. The watcher should notify the configurations
// whenever a reload event is triggered.
func WithWatcher(watcher contract.ConfigWatcher) Option {
	return func(option *KoanfAdapter) {
		option.watcher = watcher
	}
}

// WithDelimiter changes the default delimiter of Koanf. See Koanf's doc to learn more about delimiters.
func WithDelimiter(delimiter string) Option {
	return func(option *KoanfAdapter) {
		option.delimiter = delimiter
	}
}

// NewConfig creates a new *KoanfAdapter.
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

// Reload reloads the whole configuration stack. It reloads layer by layer, so if
// an error occurred, Reload will return early and abort the rest of the
// reloading.
func (k KoanfAdapter) Reload() error {
	for i := len(k.layers) - 1; i >= 0; i-- {
		err := k.K.Load(k.layers[i].Provider, k.layers[i].Parser)
		if err != nil {
			return fmt.Errorf("unable to load config %w", err)
		}
	}
	return nil
}

// Watch uses the internal watcher to watch the configuration reload signals.
// This function should be registered in the run group. If the watcher is nil,
// this call will block until context expired.
func (k KoanfAdapter) Watch(ctx context.Context) error {
	if k.watcher == nil {
		<-ctx.Done()
		return ctx.Err()
	}
	return k.watcher.Watch(ctx, k.Reload)
}

// Unmarshal unmarshals a given key path into the given struct using the mapstructure lib.
// If no path is specified, the whole map is unmarshalled. `koanf` is the struct field tag used to match field names.
func (k KoanfAdapter) Unmarshal(path string, o interface{}) error {
	return k.K.UnmarshalWithConf(path, o, koanf.UnmarshalConf{
		Tag: "json",
		DecoderConfig: &mapstructure.DecoderConfig{
			Result:           o,
			ErrorUnused:      true,
			WeaklyTypedInput: true,
		},
	})
}

// Route cuts the config map at a given key path into a sub map and returns a new contract.ConfigAccessor instance
// with the cut config map loaded. For instance, if the loaded config has a path that looks like parent.child.sub.a.b,
// `Route("parent.child")` returns a new contract.ConfigAccessor instance with the config map `sub.a.b` where
// everything above `parent.child` are cut out.
func (k KoanfAdapter) Route(s string) contract.ConfigAccessor {
	return KoanfAdapter{
		K: k.K.Cut(s),
	}
}

// String returns the string value of a given key path or "" if the path does not exist or if the value is not a valid string
func (k KoanfAdapter) String(s string) string {
	return k.K.String(s)
}

// Int returns the int value of a given key path or 0 if the path does not exist or if the value is not a valid int.
func (k KoanfAdapter) Int(s string) int {
	return k.K.Int(s)
}

// Strings returns the []string slice value of a given key path or an empty []string slice if the path does not exist
// or if the value is not a valid string slice.
func (k KoanfAdapter) Strings(s string) []string {
	return k.K.Strings(s)
}

// Bool returns the bool value of a given key path or false if the path does not exist or if the value is not a valid bool representation.
// Accepted string representations of bool are the ones supported by strconv.ParseBool.
func (k KoanfAdapter) Bool(s string) bool {
	return k.K.Bool(s)
}

// Get returns the raw, uncast interface{} value of a given key path in the config map. If the key path does not exist, nil is returned.
func (k KoanfAdapter) Get(s string) interface{} {
	return k.K.Get(s)
}

// Float64 returns the float64 value of a given key path or 0 if the path does not exist or if the value is not a valid float64.
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
