package cron

import (
	"github.com/go-kit/kit/metrics"
	"time"
)

// CronJobMetrics collects metrics for cron jobs.
type CronJobMetrics struct {
	cronJobDurationSeconds metrics.Histogram
	cronJobFailCount       metrics.Counter

	// labels that has been set
	module   string
	job      string
	schedule string
}

// NewCronJobMetrics constructs a new *CronJobMetrics, setting default labels to "unknown".
func NewCronJobMetrics(histogram metrics.Histogram, counter metrics.Counter) *CronJobMetrics {
	return &CronJobMetrics{
		cronJobDurationSeconds: histogram,
		cronJobFailCount:       counter,
		module:                 "unknown",
		job:                    "unknown",
		schedule:               "unknown",
	}
}

// Module specifies the module label for CronJobMetrics.
func (c *CronJobMetrics) Module(module string) *CronJobMetrics {
	return &CronJobMetrics{
		cronJobDurationSeconds: c.cronJobDurationSeconds,
		cronJobFailCount:       c.cronJobFailCount,
		module:                 module,
		job:                    c.job,
		schedule:               c.schedule,
	}
}

// Job specifies the job label for CronJobMetrics.
func (c *CronJobMetrics) Job(job string) *CronJobMetrics {
	return &CronJobMetrics{
		cronJobDurationSeconds: c.cronJobDurationSeconds,
		cronJobFailCount:       c.cronJobFailCount,
		module:                 c.module,
		job:                    job,
		schedule:               c.schedule,
	}
}

// Schedule specifies the schedule label for CronJobMetrics.
func (c *CronJobMetrics) Schedule(schedule string) *CronJobMetrics {
	return &CronJobMetrics{
		cronJobDurationSeconds: c.cronJobDurationSeconds,
		cronJobFailCount:       c.cronJobFailCount,
		module:                 c.module,
		job:                    c.job,
		schedule:               schedule,
	}
}

// Fail marks the job as failed.
func (c *CronJobMetrics) Fail() {
	c.cronJobFailCount.With("module", c.module, "job", c.job, "schedule", c.schedule).Add(1)
}

// Observe records the duration of the job.
func (c *CronJobMetrics) Observe(duration time.Duration) {
	c.cronJobDurationSeconds.With("module", c.module, "job", c.job, "schedule", c.schedule).Observe(duration.Seconds())
}
