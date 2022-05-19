package contract

import (
	"context"
	"time"
)

// ConfigRouter enables modular configuration by giving every piece of configuration a path.
type ConfigRouter interface {
	Route(path string) ConfigUnmarshaler
}

// ConfigUnmarshaler is a minimum config interface that can be used to retrieve
// configuration from external system. If the configuration is hot reloaded,
// ConfigUnmarshaler should fetch the latest info.
type ConfigUnmarshaler interface {
	Unmarshal(path string, o any) error
}

// ConfigAccessor builds upon the ConfigUnmarshaler and provides a richer set of
// API.
// Note: it is recommended to inject ConfigUnmarshaler as the dependency
// and call config.Upgrade to get the ConfigAccessor. The interface area of
// ConfigUnmarshaler is much smaller and thus much easier to customize.
type ConfigAccessor interface {
	ConfigUnmarshaler
	String(string) string
	Int(string) int
	Strings(string) []string
	Bool(string) bool
	Get(string) any
	Float64(string) float64
	Duration(string) time.Duration
}

// ConfigWatcher is an interface for hot-reload provider.
type ConfigWatcher interface {
	Watch(ctx context.Context, reload func() error) error
}
