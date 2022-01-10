package otfranz

import (
	"github.com/go-kit/log"
	"github.com/twmb/franz-go/pkg/kgo"
)

// FranzLogAdapter return an log adapter bridging kitlog and kgo.Logger.
func FranzLogAdapter(lvl string, logger log.Logger) kgo.Logger {
	return &logAdapter{
		logger: logger,
		level:  newLogLevel(lvl),
	}
}

// newLogLevel trans core level config to kgo.LogLevel
func newLogLevel(level string) kgo.LogLevel {
	switch level {
	case "debug":
		return kgo.LogLevelDebug
	case "info":
		return kgo.LogLevelInfo
	case "warn":
		return kgo.LogLevelWarn
	case "error":
		return kgo.LogLevelError
	default:
		return kgo.LogLevelNone
	}
}

// logAdapter is an log adapter bridging kitlog and kgo.Logger.
type logAdapter struct {
	logger log.Logger
	level  kgo.LogLevel
}

// Level implements kgo.Logger.
func (w *logAdapter) Level() kgo.LogLevel {
	return w.level
}

// Log implements kgo.Logger.
func (w *logAdapter) Log(lvl kgo.LogLevel, msg string, keyvals ...interface{}) {
	if w.Level() < lvl {
		return
	}
	kvs := []interface{}{
		"msg", msg,
	}
	kvs = append(kvs, keyvals...)
	_ = w.logger.Log(kvs...)
}
