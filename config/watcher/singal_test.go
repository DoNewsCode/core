//go:build !windows
// +build !windows

package watcher

import (
	"context"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSignal(t *testing.T) {
	t.Parallel()
	var (
		called bool
		ch     = make(chan struct{})
	)
	sig := Signal{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sig.Watch(ctx, func() error {
			cancel()
			called = true
			ch <- struct{}{}
			return nil
		})
	}()
	time.Sleep(time.Second)
	syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
	<-ch
	assert.True(t, called)
}
