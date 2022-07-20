package cron

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeOnceScheduler struct {
	returned int32
	runAfter time.Duration
}

func (s *fakeOnceScheduler) Next(t time.Time) time.Time {
	if old := atomic.SwapInt32(&s.returned, 1); old == 0 {
		return t.Add(s.runAfter)
	}
	return time.Time{}
}

func TestCron_heapsort(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Millisecond)
	defer cancel()

	c := New(Config{EnableSeconds: true})
	var i int32

	c.Add("3 * * * * *", func(ctx context.Context) error {
		assert.Equal(t, int32(2), atomic.SwapInt32(&i, 3))
		return nil
	}, WithSchedule(&fakeOnceScheduler{runAfter: 3 * time.Millisecond}))
	c.Add("1 * * * * *", func(ctx context.Context) error {
		assert.Equal(t, int32(0), atomic.SwapInt32(&i, 1))
		return nil
	}, WithSchedule(&fakeOnceScheduler{runAfter: 1 * time.Millisecond}))
	c.Add("2 * * * * *", func(ctx context.Context) error {
		assert.Equal(t, int32(0), atomic.SwapInt32(&i, 2))
		return nil
	}, WithSchedule(&fakeOnceScheduler{runAfter: 2 * time.Millisecond}))
	c.Run(ctx)
}

func TestCron_no_job(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c := New(Config{})
	c.Run(ctx)
}

func TestCron_late_job(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	c := New(Config{EnableSeconds: true})

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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Millisecond)
	defer cancel()

	c := New(Config{})
	go c.Run(ctx)

	ch := make(chan struct{})
	id, _ := c.Add("* * * * * *", func(ctx context.Context) error {
		ch <- struct{}{}
		return nil
	}, WithSchedule(&fakeOnceScheduler{runAfter: time.Millisecond}))
	c.Remove(id)
	select {
	case <-ch:
		t.Fatal("should not be called")
	case <-time.After(2 * time.Millisecond):
	}
}

func TestCron_nowFunc(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	fakeNow, _ := time.ParseInLocation("2006-01-02 15:04:05", "2029-01-01 00:00:00", time.Local)
	c := New(Config{NowFunc: MockStartTime(fakeNow.Add(-time.Millisecond))})
	go c.Run(ctx)

	ch := make(chan struct{})
	c.Add("0 0 * * *", func(ctx context.Context) error {
		ch <- struct{}{}
		close(ch)
		return nil
	})
	select {
	case <-ch:
	case <-time.After(2 * time.Millisecond):
		t.Fatal("timeout")
	}
}
