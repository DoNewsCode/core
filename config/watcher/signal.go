package watcher

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

type Signal struct{}

func (s Signal) Watch(ctx context.Context, reload func() error) error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-sigs:
			// Trigger event.
			if err := reload(); err != nil {
				return err
			}
		}
	}
}
