package queue

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/events"
	"github.com/DoNewsCode/std/pkg/logging"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	"testing"
)

type MockListener func(ctx context.Context, event contract.Event) error

func (m MockListener) Listen() []contract.Event {
	return events.From(MockEvent{})
}

func (m MockListener) Process(ctx context.Context, event contract.Event) error {
	return m(ctx, event)
}

type RetryingListener func(ctx context.Context, event contract.Event) error

func (m RetryingListener) Listen() []contract.Event {
	return events.From(RetryingEvent{})
}

func (m RetryingListener) Process(ctx context.Context, event contract.Event) error {
	return m(ctx, event)
}

type AbortedListener func(ctx context.Context, event contract.Event) error

func (m AbortedListener) Listen() []contract.Event {
	return events.From(AbortedEvent{})
}

func (m AbortedListener) Process(ctx context.Context, event contract.Event) error {
	return m(ctx, event)
}

type MockEvent struct {
	Value  string
	Called *bool
}

func setUp() *QueueableDispatcher {
	s := redis.NewUniversalClient(&redis.UniversalOptions{})
	s.FlushAll(context.Background())
	driver := RedisDriver{
		Logger:      logging.NewLogger("logfmt"),
		RedisClient: s,
		ChannelConfig: ChannelConfig{
			Delayed:  "delayed",
			Failed:   "failed",
			Reserved: "reserved",
			Waiting:  "waiting",
			Timeout:  "timeout",
		},
		PopTimeout: time.Second,
		Packer:     packer{},
	}
	dispatcher := WithQueue(&events.SyncDispatcher{}, &driver, UseLogger(logging.NewLogger("logfmt")))
	return dispatcher
}

func TestDispatcher_work(t *testing.T) {
	rand.Seed(int64(time.Now().Unix()))
	cases := []struct {
		name        string
		value       contract.Event
		ln          MockListener
		maxAttempts int
		check       func(int, int)
	}{
		{
			"simple message",
			events.Of(MockEvent{Value: "hello"}),
			func(ctx context.Context, event contract.Event) error {
				assert.IsType(t, MockEvent{}, event.Data())
				assert.Equal(t, "hello", event.Data().(MockEvent).Value)
				return nil
			},
			1,
			func(retries, failed int) {
				assert.Equal(t, 0, retries)
				assert.Equal(t, 0, failed)
			},
		},
		{
			"retry message",
			events.Of(MockEvent{Value: "hello"}),
			func(ctx context.Context, event contract.Event) error {
				assert.IsType(t, MockEvent{}, event.Data())
				assert.Equal(t, "hello", event.Data().(MockEvent).Value)
				return errors.New("foo")
			},
			2,
			func(retries, failed int) {
				assert.Equal(t, 1, retries)
				assert.Equal(t, 0, failed)
			},
		},
		{
			"fail message",
			events.Of(MockEvent{Value: "hello"}),
			func(ctx context.Context, event contract.Event) error {
				assert.IsType(t, MockEvent{}, event.Data())
				assert.Equal(t, "hello", event.Data().(MockEvent).Value)
				return errors.New("foo")
			},
			1,
			func(retries, failed int) {
				assert.Equal(t, 0, retries)
				assert.Equal(t, 1, failed)
			},
		},
	}
	for _, cc := range cases {
		c := cc
		t.Run(c.name, func(t *testing.T) {
			retries := 0
			failed := 0
			dispatcher := setUp()
			dispatcher.Subscribe(c.ln)
			dispatcher.Subscribe(RetryingListener(func(ctx context.Context, event contract.Event) error {
				retries++
				return nil
			}))
			dispatcher.Subscribe(AbortedListener(func(ctx context.Context, event contract.Event) error {
				failed++
				return nil
			}))
			msg, err := dispatcher.packer.Compress(c.value.Data())
			assert.NoError(t, err)
			dispatcher.work(context.Background(), &PersistedEvent{
				Key:         c.value.Type(),
				Value:       msg,
				MaxAttempts: c.maxAttempts,
				Attempts:    1,
			})
			c.check(retries, failed)
		})
	}
}

func TestDispatcher_Consume(t *testing.T) {
	consumer := setUp()

	var called string
	cases := []struct {
		name   string
		evt    contract.Event
		ln     MockListener
		called func()
	}{
		{
			"ordinary message",
			events.Of(MockEvent{Value: "hello"}),
			func(ctx context.Context, event contract.Event) error {
				assert.IsType(t, MockEvent{}, event.Data())
				assert.Equal(t, "hello", event.Data().(MockEvent).Value)
				called = "ordinary message"
				return nil
			},
			func() {
				assert.Equal(t, "ordinary message", called)
				called = ""
			},
		},
		{
			"persist message",
			Persist(events.Of(MockEvent{Value: "hello"})),
			func(ctx context.Context, event contract.Event) error {
				assert.IsType(t, MockEvent{}, event.Data())
				assert.Equal(t, "hello", event.Data().(MockEvent).Value)
				called = "persist message"
				return nil
			},
			func() {
				time.Sleep(5 * time.Millisecond)
				assert.Equal(t, "persist message", called)
				called = ""
			},
		},
		{
			"deferred message",
			Persist(events.Of(MockEvent{Value: "hello", Called: new(bool)}), Defer(2*time.Second)),
			func(ctx context.Context, event contract.Event) error {
				called = "deferred message"
				return nil
			},
			func() {
				time.Sleep(1 * time.Second)
				assert.NotEqual(t, "deferred message", called)
				called = ""
				time.Sleep(2 * time.Second)
			},
		},
		{
			"deferred message but called",
			Persist(events.Of(MockEvent{Value: "hello", Called: new(bool)}), Defer(time.Second)),
			func(ctx context.Context, event contract.Event) error {
				called = "deferred message but called"
				return nil
			},
			func() {
				time.Sleep(2 * time.Second)
				assert.Equal(t, "deferred message but called", called)
				called = ""
			},
		},
		{
			"failed message",
			Persist(events.Of(MockEvent{Value: "hello"})),
			func(ctx context.Context, event contract.Event) error {
				return errors.New("some err")
			},
			func() {
				time.Sleep(20 * time.Millisecond)
				info, _ := consumer.driver.Info(context.Background())
				assert.Equal(t, int64(1), info.Failed)
				err := consumer.driver.Flush(context.Background(), "failed")
				assert.NoError(t, err)
			},
		},
		{
			"retry message",
			Persist(events.Of(MockEvent{Value: "hello"}), MaxAttempts(2)),
			func(ctx context.Context, event contract.Event) error {
				if called != "retry message" {
					called = "retry message"
					return errors.New("some err")
				}
				return nil
			},
			func() {
				time.Sleep(5 * time.Millisecond)
				info, _ := consumer.driver.Info(context.Background())
				assert.Equal(t, int64(0), info.Failed)
			},
		},
		{
			"reload message",
			Persist(events.Of(MockEvent{Value: "hello"}), Timeout(time.Second)),
			func(ctx context.Context, event contract.Event) error {
				if called != "reload message" {
					return errors.New("some err")
				}
				return nil
			},
			func() {
				time.Sleep(5 * time.Millisecond)
				called = "reload message"
				num, _ := consumer.driver.Reload(context.Background(), "failed")
				assert.Equal(t, int64(1), num)
				time.Sleep(5 * time.Millisecond)
				info, _ := consumer.driver.Info(context.Background())
				assert.Equal(t, int64(0), info.Failed)
				called = ""
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dispatcher := setUp()
			ctx, cancel := context.WithCancel(context.Background())
			go dispatcher.Consume(ctx)
			defer cancel()
			dispatcher.Subscribe(c.ln)
			err := dispatcher.Dispatch(context.Background(), c.evt)
			assert.NoError(t, err)
			c.called()
		})
	}
}
