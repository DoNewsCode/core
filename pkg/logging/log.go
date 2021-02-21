/*
Package logging provides a kitlog compatible logger.

This package is mostly a thin wrapper around kitlog
(http://github.com/go-kit/kit/log). kitlog provides a minimalist, contextual,
fully composable logger. However it is too unopinionated, hence requiring some
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
	"fmt"
	"os"
	"strings"

	"github.com/DoNewsCode/std/pkg/config"
	"github.com/DoNewsCode/std/pkg/di"
	"github.com/opentracing/opentracing-go"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/log/term"
)

var _ contract.LevelLogger = (*levelLogger)(nil)

// NewLogger constructs a log.Logger based on the given format. The support
// formats are "json" and "logfmt".
func NewLogger(format string) (logger log.Logger) {
	switch strings.ToLower(format) {
	case "json":
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
		logger = moduleLogger{Logger: logger}
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
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = moduleLogger{Logger: logger}
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

type spanLogger struct {
	span opentracing.Span
	base log.Logger
}

func (s spanLogger) Log(keyvals ...interface{}) error {
	s.span.LogKV(keyvals...)
	return s.base.Log(keyvals...)
}

// WithContext decorates the log.Logger with information form context. If there is a opentracing span
// in the context, the span will receive the logger output as well.
func WithContext(logger log.Logger, ctx context.Context) log.Logger {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return withContext(logger, ctx)
	}
	return spanLogger{span: span, base: withContext(logger, ctx)}
}

func withContext(logger log.Logger, ctx context.Context) log.Logger {
	transport, _ := ctx.Value(contract.TransportKey).(string)
	requestUrl, _ := ctx.Value(contract.RequestUrlKey).(string)
	ip, _ := ctx.Value(contract.IpKey).(string)
	tenant, ok := ctx.Value(contract.TenantKey).(contract.Tenant)
	if !ok {
		tenant = contract.MapTenant{}
	}
	args := []interface{}{"transport", transport, "requestUrl", requestUrl, "clientIp", ip}
	for k, v := range tenant.KV() {
		args = append(args, k, v)
	}

	return log.With(
		logger,
		args...,
	)
}

type levelLogger struct {
	depth int
	log.Logger
}

func (l levelLogger) Debugf(s string, i ...interface{}) {
	s = fmt.Sprintf(s, i...)
	_ = log.With(level.Debug(l), "caller", log.Caller(l.depth)).Log("msg", s)
}

func (l levelLogger) Infof(s string, i ...interface{}) {
	s = fmt.Sprintf(s, i...)
	_ = log.With(level.Info(l), "caller", log.Caller(l.depth)).Log("msg", s)
}

func (l levelLogger) Warnf(s string, i ...interface{}) {
	s = fmt.Sprintf(s, i...)
	_ = log.With(level.Warn(l), "caller", log.Caller(l.depth)).Log("msg", s)
}

func (l levelLogger) Errf(s string, i ...interface{}) {
	s = fmt.Sprintf(s, i...)
	_ = log.With(level.Error(l), "caller", log.Caller(l.depth)).Log("err", s)
}

func (l levelLogger) Debug(s string) {
	_ = log.With(level.Debug(l), "caller", log.Caller(l.depth)).Log("err", s)
}

func (l levelLogger) Info(s string) {
	_ = log.With(level.Info(l), "caller", log.Caller(l.depth)).Log("msg", s)
}

func (l levelLogger) Warn(s string) {
	_ = log.With(level.Warn(l), "caller", log.Caller(4)).Log("msg", s)
}

func (l levelLogger) Err(err string) {
	_ = log.With(level.Error(l), "caller", log.Caller(4)).Log("err", err)
}

// WithLevel decorates the logger and returns a contract.LevelLogger.
//
// Note: Don't inject contract.LevelLogger to dependency consumers directly as
// this will weakens the powerful abstraction of log.Logger. Only inject
// log.Logger, and converts log.Logger to contract.LevelLogger within the
// boundary of dependency consumer if desired.
func WithLevel(logger log.Logger) levelLogger {
	if l, ok := logger.(levelLogger); ok {
		return l
	}
	return levelLogger{4, logger}
}

type moduleLogger struct {
	di.Module
	log.Logger
}

func (m moduleLogger) ProvideConfig() []config.ExportedConfig {
	return []config.ExportedConfig{
		{
			Owner: "log",
			Data: map[string]interface{}{
				"log": map[string]interface{}{"level": "debug", "format": "logfmt"},
			},
			Comment: "The global logging level and format",
		},
	}
}
