package observability

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/uber/jaeger-client-go"
)

// JaegerLogAdapter is an adapter that bridges kitlog and Jaeger.
type JaegerLogAdapter struct {
	Logging log.Logger
}

// Infof implements jaeger's logger
func (l JaegerLogAdapter) Infof(msg string, args ...any) {
	level.Info(l.Logging).Log("msg", fmt.Sprintf(msg, args...))
}

// Error implements jaeger's logger
func (l JaegerLogAdapter) Error(msg string) {
	level.Error(l.Logging).Log("msg", msg)
}

// ProvideJaegerLogAdapter returns a valid jaeger.Logger.
func ProvideJaegerLogAdapter(l log.Logger) jaeger.Logger {
	return &JaegerLogAdapter{Logging: log.With(l, "tag", "observability")}
}
