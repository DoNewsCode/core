package leader_test

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/leader"
)

type AlwaysLeaderDriver struct {
}

func (a AlwaysLeaderDriver) Campaign(ctx context.Context, toLeader func(bool)) error {
	defer toLeader(false)
	toLeader(true)
	<-ctx.Done()
	return nil
}

func (a AlwaysLeaderDriver) Resign(ctx context.Context) error {
	return nil
}

func Example_providers() {
	if os.Getenv("ETCD_ADDR") == "" {
		fmt.Println("set ETCD_ADDR to run this example")
		return
	}
	c := core.Default(core.WithInline("log.level", "none"))
	c.Provide(leader.Providers(leader.WithDriver(AlwaysLeaderDriver{})))

	c.Invoke(func(statusChanged leader.StatusChanged) {
		statusChanged.On(func(ctx context.Context, status *leader.Status) error {
			// Becomes true when campaign succeeds and becomes false when resign
			fmt.Println(status.IsLeader())
			return nil
		})
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c.Serve(ctx)

	// Output:
	// true
	// false
}
