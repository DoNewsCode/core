/*
Package logging provides a kitlog compatible logger.

This package is mostly a thin wrapper around kitlog
(http://github.com/go-kit/log). kitlog provides a minimalist, contextual,
fully composable logger. However, it is too unopinionated, hence requiring some
efforts and coordination to set up a good practise.

Integration

Package logging is bundled in core. Enable logging as dependency by calling:

	var c *core.C = core.New()
	c.ProvideEssentials()

See example for usage.
*/
package logging

import (
	"context"
	"os"
	"strings"

	"github.com/DoNewsCode/core/ctxmeta"
	"github.com/opentracing/opentracing-go"

	"github.com/DoNewsCode/core/contract"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/go-kit/log/term"
)

// LevelLogger is an alias of contract.LevelLogger
type LevelLogger = contract.LevelLogger

var _ LevelLogger = (*levelLogger)(nil)

// NewLogger constructs a log.Logger based on the given format. The support
// formats are "json" and "logfmt".
func NewLogger(format string) (logger log.Logger) {
	switch strings.ToLower(format) {
	case "json":
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
		return logger
	default:
		// Color by level value
		colorFn := func(keyvals ...interface{}) term.FgBgColor {
			for i := 0; i < len(keyvals)-1; i += 2 {
				if keyvals[i] != "level" {
					continue
				}
				if value, ok := keyvals[i+1].(level.Value); ok {
					switch value.String() {
					case "debug":
						return term.FgBgColor{Fg: term.DarkGray}
					case "info":
						return term.FgBgColor{Fg: term.Gray}
					case "warn":
						return term.FgBgColor{Fg: term.Yellow}
					case "error":
						return term.FgBgColor{Fg: term.Red}
					case "crit":
						return term.FgBgColor{Fg: term.Gray, Bg: term.DarkRed}
					default:
						return term.FgBgColor{}
					}
				}
			}
			return term.FgBgColor{}
		}
		logger = term.NewLogger(os.Stdout, log.NewLogfmtLogger, colorFn)
		logger = log.With(log.NewSyncLogger(logger), "ts", log.DefaultTimestampUTC)

		return logger
	}
}

// LevelFilter filters the log output based on its level.
// Allowed levels are "debug", "info", "warn", "error", or "none"
func LevelFilter(levelCfg string) level.Option {
	switch levelCfg {
	case "debug":
		return level.AllowDebug()
	case "info":
		return level.AllowInfo()
	case "warn":
		return level.AllowWarn()
	case "error":
		return level.AllowError()
	case "none":
		return level.AllowNone()
	default:
		return level.AllowAll()
	}
}

type span interface {
	LogKV(alternatingKeyValues ...interface{})
}

type spanLogger struct {
	span span
	base log.Logger
	kvs  []interface{}
}

func (s spanLogger) Log(keyvals ...interface{}) error {
	for k := range s.kvs {
		if f, ok := s.kvs[k].(log.Valuer); ok {
			s.kvs[k] = f()
		}
	}
	s.kvs = append(s.kvs, keyvals...)
	s.span.LogKV(s.kvs...)
	return s.base.Log(keyvals...)
}

// WithContext decorates the log.Logger with information form context. If there is an opentracing span
// in the context, the span will receive the logger output as well.
func WithContext(logger log.Logger, ctx context.Context) log.Logger {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return WithBaggage(logger, ctx)
	}
	return WithBaggage(spanLogger{span: span, base: logger}, ctx)
}

// WithBaggage decorates the log.Logger with information form context.
func WithBaggage(logger log.Logger, ctx context.Context) log.Logger {
	var args []interface{}

	bag := ctxmeta.GetBaggage(ctx)

	for _, kv := range bag.Slice() {
		args = append(args, kv.Key, kv.Val)
	}

	base := log.With(logger, args...)

	return base
}

type levelLogger struct {
	log.Logger
}

func (l levelLogger) Debugf(s string, i ...interface{}) {
	_ = level.Debug(l).Log("msg", Sprintf(s, i...))
}

func (l levelLogger) Infof(s string, i ...interface{}) {
	_ = level.Info(l).Log("msg", Sprintf(s, i...))
}

func (l levelLogger) Warnf(s string, i ...interface{}) {
	_ = level.Warn(l).Log("msg", Sprintf(s, i...))
}

func (l levelLogger) Errf(s string, i ...interface{}) {
	_ = level.Error(l).Log("msg", Sprintf(s, i...))
}

func (l levelLogger) Debugw(s string, fields ...interface{}) {
	m := append(fields, "msg", s)
	_ = level.Debug(l).Log(m...)
}

func (l levelLogger) Infow(s string, fields ...interface{}) {
	m := append(fields, "msg", s)
	_ = level.Info(l).Log(m...)
}

func (l levelLogger) Warnw(s string, fields ...interface{}) {
	m := append(fields, "msg", s)
	_ = level.Warn(l).Log(m...)
}

func (l levelLogger) Errw(s string, fields ...interface{}) {
	m := append(fields, "msg", s)
	_ = level.Error(l).Log(m...)
}

func (l levelLogger) Debug(args ...interface{}) {
	_ = level.Debug(l).Log("msg", Sprint(args...))
}

func (l levelLogger) Info(args ...interface{}) {
	_ = level.Info(l).Log("msg", Sprint(args...))
}

func (l levelLogger) Warn(args ...interface{}) {
	_ = level.Warn(l).Log("msg", Sprint(args...))
}

func (l levelLogger) Err(args ...interface{}) {
	_ = level.Error(l).Log("msg", Sprint(args...))
}

// WithLevel decorates the logger and returns a contract.LevelLogger.
//
// Note: Don't inject contract.LevelLogger to dependency consumers directly as
// this will weaken the powerful abstraction of log.Logger. Only inject
// log.Logger, and converts log.Logger to contract.LevelLogger within the
// boundary of dependency consumer if desired.
func WithLevel(logger log.Logger) LevelLogger {
	if l, ok := logger.(LevelLogger); ok {
		return l
	}
	return levelLogger{log.With(logger, "caller", log.Caller(5))}
}
