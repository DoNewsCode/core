package cronopts

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/robfig/cron/v3"
)

// CronJobMetrics collects metrics for cron jobs.
type CronJobMetrics struct {
	CronJobDurationSeconds metrics.Histogram
	CronJobFailCount       metrics.Counter

	// labels has been set
	module bool
	job    bool
}

func NewCronJobMetrics(histogram metrics.Histogram, counter metrics.Counter) *CronJobMetrics {
	return &CronJobMetrics{
		CronJobDurationSeconds: histogram,
		CronJobFailCount:       counter,
	}
}

// Module specifies the module label for CronJobMetrics.
func (c *CronJobMetrics) Module(module string) *CronJobMetrics {
	return &CronJobMetrics{
		CronJobDurationSeconds: c.CronJobDurationSeconds.With("module", module),
		CronJobFailCount:       c.CronJobFailCount.With("module", module),
		module:                 true,
		job:                    c.job,
	}
}

// Job specifies the job label for CronJobMetrics.
func (c *CronJobMetrics) Job(job string) *CronJobMetrics {
	return &CronJobMetrics{
		CronJobDurationSeconds: c.CronJobDurationSeconds.With("job", job),
		CronJobFailCount:       c.CronJobFailCount.With("job", job),
		module:                 c.module,
		job:                    true,
	}
}

// Measure executes the given function and records the duration and success.
func Measure(c *CronJobMetrics) cron.JobWrapper {
	if !c.module {
		c.CronJobDurationSeconds = c.CronJobDurationSeconds.With("module", "unknown")
	}
	if !c.job {
		c.CronJobDurationSeconds = c.CronJobDurationSeconds.With("job", "unknown")
	}
	return func(job cron.Job) cron.Job {
		return cron.FuncJob(func() {
			start := time.Now()
			defer c.CronJobDurationSeconds.Observe(time.Since(start).Seconds())
			job.Run()
		})
	}
}
