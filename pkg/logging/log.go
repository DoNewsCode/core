package logging

import (
	"context"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/log/term"
	"github.com/DoNewsCode/std/pkg/contract"
)

func NewLogger(env contract.Env) (logger log.Logger) {
	if !env.IsLocal() {
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
		return log.With(logger)
	}
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
	return log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
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

func WithContext(logger log.Logger, ctx context.Context) log.Logger {

	transport, _ := ctx.Value(contract.TransportKey).(string)
	requestUrl, _ := ctx.Value(contract.RequestUrlKey).(string)
	tenant, ok := ctx.Value(contract.TenantKey).(contract.Tenant)
	if !ok {
		tenant = contract.MapTenant{}
	}
	args := []interface{}{"transport", transport, "requestUrl", requestUrl}
	for k, v := range tenant.KV() {
		args = append(args, k, v)
	}

	return log.With(
		logger,
		args...
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

func WithLevel(logger log.Logger) contract.LevelLogger {
	return levelLogger{logger}
}
