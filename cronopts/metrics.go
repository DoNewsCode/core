package cronopts

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/robfig/cron/v3"
)

// CronJobMetrics collects metrics for cron jobs.
type CronJobMetrics struct {
	cronJobDurationSeconds metrics.Histogram
	cronJobFailCount       metrics.Counter

	// labels that has been set
	module string
	job    string
}

// NewCronJobMetrics constructs a new *CronJobMetrics, setting default labels to "unknown".
func NewCronJobMetrics(histogram metrics.Histogram, counter metrics.Counter) *CronJobMetrics {
	return &CronJobMetrics{
		cronJobDurationSeconds: histogram,
		cronJobFailCount:       counter,
		module:                 "unknown",
		job:                    "unknown",
	}
}

// Module specifies the module label for CronJobMetrics.
func (c *CronJobMetrics) Module(module string) *CronJobMetrics {
	return &CronJobMetrics{
		cronJobDurationSeconds: c.cronJobDurationSeconds,
		cronJobFailCount:       c.cronJobFailCount,
		module:                 module,
		job:                    c.job,
	}
}

// Job specifies the job label for CronJobMetrics.
func (c *CronJobMetrics) Job(job string) *CronJobMetrics {
	return &CronJobMetrics{
		cronJobDurationSeconds: c.cronJobDurationSeconds,
		cronJobFailCount:       c.cronJobFailCount,
		module:                 c.module,
		job:                    job,
	}
}

// Fail marks the job as failed.
func (c *CronJobMetrics) Fail() {
	c.cronJobFailCount.With("module", c.module, "job", c.job).Add(1)
}

// Observe records the duration of the job.
func (c *CronJobMetrics) Observe(value float64) {
	c.cronJobDurationSeconds.With("module", c.module, "job", c.job).Observe(value)
}

// Measure wraps the given job and records the duration and success.
func (c *CronJobMetrics) Measure(job cron.Job) cron.Job {
	return cron.FuncJob(func() {
		start := time.Now()
		defer c.cronJobDurationSeconds.With("module", c.module, "job", c.job).Observe(time.Since(start).Seconds())
		job.Run()
	})
}

// Measure returns a job wrapper that wraps the given job and records the duration and success.
func Measure(c *CronJobMetrics) cron.JobWrapper {
	return c.Measure
}
