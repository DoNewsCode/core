package cron

import (
	"time"

	"github.com/robfig/cron/v3"
)

// Config is the configuration for the cron package.
type Config struct {
	// Parser is the parser to parse cron expressions.
	Parser cron.ScheduleParser
	// Location is the timezone to use in parsing cron expressions.
	Location *time.Location
	// GlobalOptions are the job options that are applied to all jobs.
	GlobalOptions []JobOption
	// EnableSeconds is whether to enable seconds in the cron expression.
	EnableSeconds bool
}
