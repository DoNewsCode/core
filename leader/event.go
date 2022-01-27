package leader

import "context"

type StatusChanged interface {
	Subscribe(func(ctx context.Context, status *Status) error) int
}
