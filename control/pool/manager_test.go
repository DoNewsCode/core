package pool

import (
	"context"
	"testing"
	"time"
)

func TestManager_Go(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	m := NewManager()
	go m.Run(ctx)

	var executed = make(chan struct{})
	m.Go(ctx, func(ctx context.Context) {
		close(executed)
	})
	<-executed
}
