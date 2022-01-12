package jobs_test

import (
	"context"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/cronopts"
	"github.com/DoNewsCode/core/cronopts/jobs"
	"github.com/DoNewsCode/core/observability"
	"github.com/go-kit/log"
	"github.com/opentracing/opentracing-go"
	"github.com/robfig/cron/v3"
)

type CronJobModule struct {
	mJob jobs.Universal
}

func (c CronJobModule) ProvideCron(crontab *cron.Cron) {
	crontab.AddJob("@every 1s", c.mJob)
}

// NewCronJobModule creates a new module that provides a cron job with metrics, logging and tracing.
func NewCronJobModule(tracer opentracing.Tracer, metrics *cronopts.CronJobMetrics, logger log.Logger) CronJobModule {
	return CronJobModule{
		mJob: jobs.New("cronjob", func(ctx context.Context) error {
			return nil
		}, jobs.WithMetrics(metrics), jobs.WithTracing(tracer), jobs.WithLogs(logger)),
	}
}

func Example() {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()

	c := core.Default(core.WithInline("log.level", "none"))
	c.Provide(observability.Providers())
	c.AddModuleFunc(NewCronJobModule)
	c.Serve(ctx)

	// Output:
}
