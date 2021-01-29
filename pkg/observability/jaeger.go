package observability

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/uber/jaeger-client-go"
)

type JaegerLogAdapter struct {
	Logging log.Logger
}

func (l JaegerLogAdapter) Infof(msg string, args ...interface{}) {
	level.Info(l.Logging).Log("msg", fmt.Sprintf(msg, args...))
}

func (l JaegerLogAdapter) Error(msg string) {
	level.Error(l.Logging).Log("msg", msg)
}

func ProvideJaegerLogAdapter(l log.Logger) jaeger.Logger {
	return &JaegerLogAdapter{Logging: l}
}
