package leader

import "context"

// StatusChanged is a channel for receiving StatusChanged events.
type StatusChanged interface {
	On(func(ctx context.Context, status *Status) error) (unsubscribe func())
}
