package logging

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"os"
	"strings"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/log/term"
)

var _ contract.LevelLogger = (*levelLogger)(nil)

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
		return log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.Caller(7))
	}
}

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

func (l levelLogger) CheckErr(err error) {
	if err == nil {
		return
	}
	l.Err(err)
	os.Exit(1)
}

func WithLevel(logger log.Logger) levelLogger {
	return levelLogger{logger}
}
