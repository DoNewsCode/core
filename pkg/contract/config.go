package contract

import "context"

type ConfigRouter interface {
	Route(path string) ConfigAccessor
}

type ConfigAccessor interface {
	String(string) string
	Int(string) int
	Strings(string) []string
	Bool(string) bool
	Get(string) interface{}
	Float64(string) float64
	Unmarshal(path string, o interface{}) error
}

type ConfigWatcher interface {
	Watch(ctx context.Context, reload func() error) error
}



