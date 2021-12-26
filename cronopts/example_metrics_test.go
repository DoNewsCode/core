package cronopts_test

import (
	"context"
	"fmt"
	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/cronopts"
	"github.com/DoNewsCode/core/observability"
	"github.com/robfig/cron/v3"
	"math/rand"
	"time"
)

type CronModule struct {
	metrics *cronopts.CronJobMetrics
}

func NewCronModule(metrics *cronopts.CronJobMetrics) CronModule {
	return CronModule{metrics: metrics.Module("test_module")}
}

func (c CronModule) ProvideCron(crontab *cron.Cron) {
	// Create a new cron job, and measure its execution durations.
	crontab.AddJob("* * * * *", c.metrics.Job("test_job").Measure(cron.FuncJob(func() {
		fmt.Println("running")
		// For 50% chance, the job may fail. Report it to metrics collector.
		if rand.Float64() > 0.5 {
			c.metrics.Fail()
		}
	})))
}

func Example_cronJobMetrics() {
	c := core.Default()
	c.Provide(observability.Providers())
	c.AddModuleFunc(NewCronModule)

	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	c.Serve(ctx)
}
