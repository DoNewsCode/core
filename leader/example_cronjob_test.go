package leader_test

import (
	"context"
	"fmt"
	"os"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/leader"
	"github.com/DoNewsCode/core/otetcd"

	"github.com/robfig/cron/v3"
)

type CronModule struct {
	Sts *leader.Status
}

func (s CronModule) ProvideCron(crontab *cron.Cron) {
	crontab.AddFunc("* * * * * *", func() {
		if s.Sts.IsLeader() {
			fmt.Println("do work as leader")
		}
	})
}

func Example_cronjob() {
	if os.Getenv("ETCD_ADDR") == "" {
		fmt.Println("set ETCD_ADDR to run this example")
		return
	}
	c := core.Default(core.WithInline("log.level", "none"))
	c.Provide(di.Deps{func() *cron.Cron {
		return cron.New(cron.WithSeconds())
	}})
	c.Provide(otetcd.Providers())
	c.Provide(leader.Providers())
	c.Invoke(func(sts *leader.Status) {
		c.AddModule(CronModule{Sts: sts})
	})
	c.Serve(context.Background())
}
