// Package cronopts contains the options for cron.
package cronopts

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// CronLogAdapter is an adapter between kitlog and cron logger interface
type CronLogAdapter struct {
	Logging log.Logger
}

// Info implements cron.Logger
func (c CronLogAdapter) Info(msg string, keysAndValues ...interface{}) {
	_ = level.Info(c.Logging).Log(append([]interface{}{"msg", msg}, keysAndValues...)...)
}

// Error implements cron.Logger
func (c CronLogAdapter) Error(err error, msg string, keysAndValues ...interface{}) {
	_ = level.Error(c.Logging).Log(append([]interface{}{"msg", msg, "err", err}, keysAndValues...)...)
}
