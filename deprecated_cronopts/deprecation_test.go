package cronopts_test

import (
	"context"
	"fmt"
	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/observability"
	deprecatedcron "github.com/robfig/cron/v3"
	"testing"
	"time"
)

type CronModule struct{}

func (module *CronModule) ProvideCron(crontab *deprecatedcron.Cron) {
	crontab.AddFunc("* * * * * *", func() {
		fmt.Println("Cron job ran")
	})
}

func Test_deprecation(t *testing.T) {
	c := core.Default(core.WithInline("log.level", "none"))
	c.Provide(observability.Providers())
	c.Provide(
		di.Deps{func() *deprecatedcron.Cron {
			return deprecatedcron.New(deprecatedcron.WithSeconds())
		}},
	)

	c.AddModule(CronModule{})
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	c.Serve(ctx)
	// Output:
	// Cron job ran
}
