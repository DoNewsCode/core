package leader_test

import (
	"context"
	"fmt"
	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/events"
	"github.com/DoNewsCode/core/leader"
	"github.com/DoNewsCode/core/otetcd"
	"github.com/oklog/run"
	"time"
)

func Example() {
	c := core.Default()
	c.Provide(otetcd.Providers())
	c.Provide(leader.Providers())
	c.Invoke(func(dispatcher contract.Dispatcher) {
		// This listener will be called twice. Once on becoming the leader and once on resigning the leader.
		dispatcher.Subscribe(events.Listen(events.From(&leader.Status{}), func(ctx context.Context, event contract.Event) error {
			fmt.Println(event.Data().(*leader.Status).IsLeader())
			return nil
		}))
	})
	c.Invoke(func(s *leader.Status) {
		var g run.Group
		timeout(&g, time.Second)
		c.ApplyRunGroup(&g)
		g.Run()
	})

	// Output:
	// true
	// false
}

func timeout(g *run.Group, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	g.Add(func() error {
		<-ctx.Done()
		return ctx.Err()
	}, func(err error) {
		cancel()
	})
}
