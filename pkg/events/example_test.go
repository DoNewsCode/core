package events_test

import (
	"context"
	"fmt"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/events"
)

func Example() {
	dispatcher := &events.SyncDispatcher{}
	// Subscribe to int event.
	dispatcher.Subscribe(events.Listen(events.From(0), func(ctx context.Context, event contract.Event) error {
		fmt.Println(event.Data())
		return nil
	}))
	// Subscribe to string event.
	dispatcher.Subscribe(events.Listen(events.From(""), func(ctx context.Context, event contract.Event) error {
		fmt.Println(event.Data())
		return nil
	}))
	dispatcher.Dispatch(context.Background(), events.Of(100))
	dispatcher.Dispatch(context.Background(), events.Of("event"))
	// Output:
	// 100
	// event

}
