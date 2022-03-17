package cron

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DoNewsCode/core/internal/stub"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	"github.com/go-kit/log"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/robfig/cron/v3"
)

type mockScheduler func(now time.Time) time.Time // mockScheduler is a function that returns the next time to run

func (m mockScheduler) Next(t time.Time) time.Time {
	return m(t)
}

type mockParser struct{}

func (m mockParser) Parse(spec string) (cron.Schedule, error) {
	return mockScheduler(func(now time.Time) time.Time {
		return now.Add(time.Millisecond)
	}), nil
}

func TestJobPersistence(t *testing.T) {
	if os.Getenv("REDIS_ADDR") == "" {
		t.Skip("set REDIS_ADDR to run TestModule_ProvideRunGroup")
		return
	}
	addrs := strings.Split(os.Getenv("REDIS_ADDR"), ",")

	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs: addrs,
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	client.Set(ctx, "test:foo:next", time.Now().Round(time.Second).Format(time.RFC3339), 0)

	c := New(Config{EnableSeconds: true})

	var i int
	c.Add("* * * * * *", func(ctx context.Context) error {
		i++
		return nil
	}, WithPersistence(client, PersistenceConfig{KeyPrefix: "test"}), WithName("foo"))
	c.Run(ctx)

	assert.GreaterOrEqual(t, i, 2)
	t.Log(i)
}

func TestJobOption(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := log.NewSyncLogger(log.NewLogfmtLogger(&buf))
	hist := stub.Histogram{}
	count := stub.Counter{}
	metric := NewCronJobMetrics(&hist, &count)
	tracer := mocktracer.MockTracer{}
	entryCount := 0
	concurrentCount := 0
	var concurrentAccess bool

	for _, ca := range []struct {
		name    string
		stacks  []JobOption
		job     func(context.Context) error
		asserts func(t *testing.T)
	}{
		{
			"name and logging",
			[]JobOption{
				WithName("test"),
				WithLogging(logger),
			},
			func(ctx context.Context) error {
				return nil
			},
			func(t *testing.T) {
				t.Log(buf.String())
				if buf.String() == "" {
					t.Error("Expected logging output")
				}
				if strings.Contains(buf.String(), "test") == false {
					t.Error("Expected test to be in the log output")
				}
				buf = bytes.Buffer{}
			},
		},
		{
			"error and logging",
			[]JobOption{
				WithLogging(logger),
			},
			func(ctx context.Context) error {
				return errors.New("test")
			},
			func(t *testing.T) {
				t.Log(buf.String())
				if buf.String() == "" {
					t.Error("Expected logging output")
				}
				if strings.Contains(buf.String(), "error") == false {
					t.Error("Expected error to be in the log output")
				}
				buf = bytes.Buffer{}
			},
		},
		{
			"metrics",
			[]JobOption{
				WithMetrics(metric),
			},
			func(ctx context.Context) error {
				return nil
			},
			func(t *testing.T) {
				if hist.ObservedValue == 0 {
					t.Error("Expected histogram to be observed")
				}
				if count.CounterValue > 0 {
					t.Error("Expected fail counter to be zero")
				}
				hist = stub.Histogram{}
				count = stub.Counter{}
			},
		},
		{
			"error and metrics",
			[]JobOption{
				WithMetrics(metric),
			},
			func(ctx context.Context) error {
				return errors.New("test")
			},
			func(t *testing.T) {
				if hist.ObservedValue == 0 {
					t.Error("Expected histogram to be observed")
				}
				if count.CounterValue < 1 {
					t.Error("Expected fail counter to be one")
				}
				hist = stub.Histogram{}
				count = stub.Counter{}
			},
		},
		{
			"tracing",
			[]JobOption{
				WithTracing(&tracer),
			},
			func(ctx context.Context) error {
				return nil
			},
			func(t *testing.T) {
				if len(tracer.FinishedSpans()) < 1 {
					t.Error("Expected one span to be finished")
				}
				tracer = mocktracer.MockTracer{}
			},
		},
		{
			"error tracing",
			[]JobOption{
				WithTracing(&tracer),
			},
			func(ctx context.Context) error {
				return errors.New("test")
			},
			func(t *testing.T) {
				if len(tracer.FinishedSpans()) < 1 {
					t.Error("Expected one span to be finished")
				}
				if tracer.FinishedSpans()[0].Tags()["error"] != true {
					t.Error("Expected error tag to be true")
					t.Log(tracer.FinishedSpans()[0].Tags())
				}
				tracer = mocktracer.MockTracer{}
			},
		},
		{
			"panic",
			[]JobOption{
				Recover(logger),
			},
			func(ctx context.Context) error {
				panic("to be recovered")
			},
			func(t *testing.T) {
				if strings.Contains(buf.String(), "to be recovered") == false {
					t.Error("Expected panic to be in the log output")
				}
				buf = bytes.Buffer{}
			},
		},
		{
			"skip if overlap",
			[]JobOption{
				SkipIfOverlap(),
			},
			func(ctx context.Context) error {
				entryCount++
				time.Sleep(6 * time.Millisecond)
				return nil
			},
			func(t *testing.T) {
				if entryCount > 1 {
					t.Errorf("expect entry once, got %d", entryCount)
				}
				entryCount = 0
			},
		},
		{
			"delay if overlap",
			[]JobOption{
				DelayIfOverlap(),
			},
			func(ctx context.Context) error {
				entryCount++
				concurrentCount++
				if concurrentCount > 1 {
					concurrentAccess = true
				}
				time.Sleep(3 * time.Millisecond)
				concurrentCount--
				return nil
			},
			func(t *testing.T) {
				if entryCount < 3 {
					t.Errorf("expect entry at least 3 times, got %d", entryCount)
				}
				if concurrentAccess {
					t.Errorf("conncurrent access not allowed")
				}
				entryCount = 0
				concurrentCount = 0
			},
		},
		{
			"timeout if overlap",
			[]JobOption{
				TimeoutIfOverlap(),
			},
			func(ctx context.Context) error {
				<-ctx.Done()
				return nil
			},
			func(t *testing.T) {
			},
		},
	} {
		ca := ca
		t.Run(ca.name, func(t *testing.T) {
			c := New(Config{Parser: mockParser{}})
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
			defer cancel()
			c.Add("@every 1ms", ca.job, ca.stacks...)
			c.Run(ctx)
			ca.asserts(t)
		})
	}
}
