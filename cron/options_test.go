package cron

import (
	"bytes"
	"context"
	"errors"
	"github.com/DoNewsCode/core/internal/stub"
	"github.com/go-kit/log"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/robfig/cron/v3"
	"strings"
	"testing"
	"time"
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

func TestJobOption(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := log.NewLogfmtLogger(&buf)
	hist := stub.Histogram{}
	count := stub.Counter{}
	metric := NewCronJobMetrics(&hist, &count)
	tracer := mocktracer.MockTracer{}
	entryCount := 0
	concurrentCount := 0
	var concurrentAccess bool

	for _, ca := range []struct {
		name    string
		stacks  []JobOptions
		job     func(context.Context) error
		asserts func(t *testing.T)
	}{
		{
			"name and logging",
			[]JobOptions{
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
			[]JobOptions{
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
			[]JobOptions{
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
			[]JobOptions{
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
			[]JobOptions{
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
			[]JobOptions{
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
			[]JobOptions{
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
			[]JobOptions{
				SkipIfOverlap(),
			},
			func(ctx context.Context) error {
				entryCount++
				time.Sleep(5 * time.Millisecond)
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
			[]JobOptions{
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
				if entryCount < 4 {
					t.Errorf("expect entry at least 4 times, got %d", entryCount)
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
			[]JobOptions{
				TimeoutIfOverlap(),
			},
			func(ctx context.Context) error {
				entryCount++
				<-ctx.Done()
				return nil
			},
			func(t *testing.T) {
				if entryCount < 3 {
					t.Errorf("expect entry at least 4 times, got %d", entryCount)
				}
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
