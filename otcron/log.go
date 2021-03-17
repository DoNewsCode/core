package otcron

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// CronLogAdapter is an adapter between kitlog and cron logger interface
type CronLogAdapter struct {
	Logging log.Logger
}

// Info implements cron.Logger
func (r CronLogAdapter) Info(msg string, keysAndValues ...interface{}) {
	_ = level.Info(r.Logging).Log("msg", msg, keysAndValues)
}

// Error implements cron.Logger
func (r CronLogAdapter) Error(err error, msg string, keysAndValues ...interface{}) {
	_ = level.Error(r.Logging).Log("msg", msg, "err", err, keysAndValues)
}
