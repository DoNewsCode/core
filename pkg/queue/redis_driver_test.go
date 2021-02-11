package queue

import (
	"context"
	"errors"
	"flag"
	"math/rand"
	"time"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/event"
	"github.com/DoNewsCode/std/pkg/logging"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	"testing"
)

type MockListener func(ctx context.Context, event contract.Event) error

func (m MockListener) Listen() []contract.Event {
	return event.Of(MockEvent{})
}

func (m MockListener) Process(ctx context.Context, event contract.Event) error {
	return m(ctx, event)
}

type MockEvent struct {
	Value  string
	Called *bool
}

var useRedis = flag.Bool("redis", false, "use real redis for testing")

func setUp() *dispatcher {
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
	dispatcher := WithQueue(&event.SyncDispatcher{}, &driver, UseLogger(logging.NewLogger("logfmt")))
	return dispatcher
}

func TestConsumer_work(t *testing.T) {
	if !*useRedis {
		t.Skip("this test needs redis")
	}
	rand.Seed(int64(time.Now().Unix()))
	cases := []struct {
		name  string
		value contract.Event
		ln    MockListener
	}{
		{
			"simple message",
			event.NewEvent(MockEvent{Value: "hello"}),
			func(ctx context.Context, event contract.Event) error {
				assert.IsType(t, MockEvent{}, event.Data())
				assert.Equal(t, "hello", event.Data().(MockEvent).Value)
				return nil
			},
		},
	}
	for _, cc := range cases {
		c := cc
		t.Run(c.name, func(t *testing.T) {
			dispatcher := setUp()
			dispatcher.Subscribe(c.ln)
			msg, err := dispatcher.packer.Compress(c.value.Data())
			assert.NoError(t, err)
			dispatcher.work(context.Background(), &PersistedEvent{
				Key:         c.value.Type(),
				Value:       msg,
				MaxAttempts: 1,
			})
		})
	}
}

func TestConsumer_Consume(t *testing.T) {
	if !*useRedis {
		t.Skip("this test needs redis")
	}
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
			event.NewEvent(MockEvent{Value: "hello"}),
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
			Persist(event.NewEvent(MockEvent{Value: "hello"})),
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
			Persist(event.NewEvent(MockEvent{Value: "hello", Called: new(bool)}), Defer(2*time.Second)),
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
			Persist(event.NewEvent(MockEvent{Value: "hello", Called: new(bool)}), Defer(time.Second)),
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
			Persist(event.NewEvent(MockEvent{Value: "hello"})),
			func(ctx context.Context, event contract.Event) error {
				return errors.New("some err")
			},
			func() {
				time.Sleep(5 * time.Millisecond)
				info, _ := consumer.driver.Info(context.Background())
				assert.Equal(t, int64(1), info.Failed)
				err := consumer.driver.Flush(context.Background(), "failed")
				assert.NoError(t, err)
			},
		},
		{
			"retry message",
			Persist(event.NewEvent(MockEvent{Value: "hello"}), MaxAttempts(2)),
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
			Persist(event.NewEvent(MockEvent{Value: "hello"}), Timeout(time.Second)),
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
