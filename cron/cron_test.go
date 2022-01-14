package cron

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCron_sort(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	c := New(WithSeconds())
	var i int32
	c.Add("3 * * * * *", func(ctx context.Context) error {
		assert.Equal(t, 2, atomic.SwapInt32(&i, 3))
		return nil
	})
	c.Add("1 * * * * *", func(ctx context.Context) error {
		assert.Equal(t, 0, atomic.SwapInt32(&i, 1))
		return nil
	})
	c.Add("2 * * * * *", func(ctx context.Context) error {
		assert.Equal(t, 0, atomic.SwapInt32(&i, 2))
		return nil
	})
	c.Run(ctx)
}

func TestCron_no_job(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c := New()
	c.Run(ctx)
}

func TestCron_late_job(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	c := New(WithSeconds())
	go c.Run(ctx)
	time.Sleep(time.Millisecond)
	ch := make(chan struct{})
	c.Add("* * * * * *", func(ctx context.Context) error {
		ch <- struct{}{}
		return nil
	})
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}
}

func TestCron_remove_job(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	c := New(WithSeconds())
	go c.Run(ctx)

	ch := make(chan struct{})
	id, _ := c.Add("* * * * * *", func(ctx context.Context) error {
		ch <- struct{}{}
		return nil
	})
	c.Remove(id)
	select {
	case <-ch:
		t.Fatal("should not be called")
	case <-time.After(2 * time.Second):
	}
}
