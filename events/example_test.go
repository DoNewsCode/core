package events_test

import (
	"context"
	"fmt"

	"github.com/DoNewsCode/core/events"
)

func Example() {
	dispatcher := &events.SyncDispatcher{}

	// Subscribe to a string topic named foo.
	dispatcher.Subscribe(events.Listen("foo", func(ctx context.Context, event interface{}) error {
		fmt.Println(event)
		return nil
	}))

	// Subscribe to a struct topic.
	type Topic struct{}
	dispatcher.Subscribe(events.Listen(Topic{}, func(ctx context.Context, event interface{}) error {
		fmt.Println(event)
		return nil
	}))

	dispatcher.Dispatch(context.Background(), "foo", 100)
	dispatcher.Dispatch(context.Background(), Topic{}, "event")
	// Output:
	// 100
	// event

}
