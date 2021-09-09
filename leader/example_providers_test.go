package leader_test

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/events"
	"github.com/DoNewsCode/core/leader"
)

type AlwaysLeaderDriver struct{}

func (a AlwaysLeaderDriver) Campaign(ctx context.Context) error {
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

	c.Invoke(func(dispatcher contract.Dispatcher, sts *leader.Status) {
		dispatcher.Subscribe(events.Listen(leader.OnStatusChanged, func(ctx context.Context, event interface{}) error {
			// Becomes true when campaign succeeds and becomes false when resign
			fmt.Println(event.(leader.OnStatusChangedPayload).Status.IsLeader())
			return nil
		}))
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c.Serve(ctx)

	// Output:
	// true
	// false
}
