package contract

import "context"

// ConfigRouter enables modular configuration by giving every piece of configuration a path.
type ConfigRouter interface {
	Route(path string) ConfigAccessor
}

// ConfigAccessor models a the basic configuration. If the configuration is hot reloaded,
// ConfigAccessor should fetch the latest info.
type ConfigAccessor interface {
	String(string) string
	Int(string) int
	Strings(string) []string
	Bool(string) bool
	Get(string) interface{}
	Float64(string) float64
	Unmarshal(path string, o interface{}) error
}

// ConfigWatcher is an interface for hot-reload provider.
type ConfigWatcher interface {
	Watch(ctx context.Context, reload func() error) error
}
