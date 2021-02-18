package logging

import (
	"context"
	"os"
	"strings"

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
		logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.Caller(7))
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
	log.Logger
}

func (l levelLogger) Debug(s string) {
	_ = level.Debug(l).Log("err", s)
}

func (l levelLogger) Info(s string) {
	_ = level.Info(l).Log("msg", s)
}

func (l levelLogger) Warn(s string) {
	_ = level.Warn(l).Log("msg", s)
}

func (l levelLogger) Err(err error) {
	_ = level.Error(l).Log("err", err)
}

// WithLevel decorates the logger and returns a contract.LevelLogger.
//
// Note: Don't inject contract.LevelLogger to dependency consumers directly as
// this will weakens the powerful abstraction of log.Logger. Only inject
// log.Logger, and converts log.Logger to contract.LevelLogger within the
// boundary of dependency consumer if desired.
func WithLevel(logger log.Logger) levelLogger {
	return levelLogger{logger}
}

type moduleLogger struct {
	di.Module
	log.Logger
}

func (m moduleLogger) ProvideConfig() []contract.ExportedConfig {
	return []contract.ExportedConfig{
		{
			Name: "log",
			Data: map[string]interface{}{
				"log": map[string]interface{}{"level": "debug", "format": "logfmt"},
			},
			Comment: "The global logging level and format",
		},
	}
}
