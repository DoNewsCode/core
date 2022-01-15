package cron_test

import (
	"context"
	"fmt"
	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/cron"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/observability"
	"time"
)

type CronModule struct {
	metrics *cron.CronJobMetrics
}

func NewCronModule(metrics *cron.CronJobMetrics) *CronModule {
	return &CronModule{metrics: metrics}
}

func (module *CronModule) ProvideCron(crontab *cron.Cron) {
	crontab.Add("* * * * * *", func(ctx context.Context) error {
		fmt.Println("I am a cron")
		return nil
	}, cron.WithMetrics(module.metrics), cron.WithName("foo"))
}

func Example() {
	c := core.Default(core.WithInline("log.level", "none"))
	c.Provide(observability.Providers())
	c.Provide(
		di.Deps{func() *cron.Cron {
			return cron.New(cron.Config{EnableSeconds: true})
		}},
	)

	c.AddModuleFunc(NewCronModule)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	c.Serve(ctx)
	// Output:
	// I am a cron
}
