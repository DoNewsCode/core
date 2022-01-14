package cron

import (
	"time"

	"github.com/robfig/cron/v3"
)

// Option represents a modification to the default behavior of a Cron.
type Option func(*Cron)

// WithLocation overrides the timezone of the cron instance.
func WithLocation(loc *time.Location) Option {
	return func(c *Cron) {
		c.location = loc
	}
}

// WithSeconds overrides the parser used for interpreting job schedules to
// include a seconds field as the first one.
func WithSeconds() Option {
	return WithParser(cron.NewParser(
		cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	))
}

// WithParser overrides the parser used for interpreting job schedules.
func WithParser(p cron.ScheduleParser) Option {
	return func(c *Cron) {
		c.parser = p
	}
}

// WithGlobalMiddleware specifies Job wrappers to apply to all jobs added to this cron.
func WithGlobalMiddleware(middleware ...JobMiddleware) Option {
	return func(c *Cron) {
		c.globalMiddleware = middleware
	}
}
