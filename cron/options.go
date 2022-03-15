package cron

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"math"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/DoNewsCode/core/logging"
	"github.com/go-redis/redis/v8"

	"github.com/go-kit/log"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/robfig/cron/v3"
)

// JobOption is a middleware for cron jobs.
type JobOption func(descriptors *JobDescriptor)

// WithName sets the name of the job.
func WithName(name string) JobOption {
	return func(descriptor *JobDescriptor) {
		descriptor.Name = name
	}
}

// WithSchedule sets the cron schedule of the job.
func WithSchedule(schedule cron.Schedule) JobOption {
	return func(descriptor *JobDescriptor) {
		descriptor.Schedule = schedule
	}
}

// WithMetrics returns a new JobDescriptor that will report metrics.
func WithMetrics(metrics *CronJobMetrics) JobOption {
	return func(descriptor *JobDescriptor) {
		innerRun := descriptor.Run
		descriptor.Run = func(ctx context.Context) error {
			start := time.Now()
			m := metrics.Job(descriptor.Name).Schedule(descriptor.RawSpec)
			defer m.Observe(time.Since(start))
			err := innerRun(ctx)
			if err != nil {
				m.Fail()
				return err
			}
			return nil
		}
	}
}

// WithLogging returns a new Universal job that will log.
func WithLogging(logger log.Logger) JobOption {
	return func(descriptor *JobDescriptor) {
		innerRun := descriptor.Run
		descriptor.Run = func(ctx context.Context) error {
			due := GetCurrentSchedule(ctx)
			delayed := time.Since(due)
			l := logging.WithContext(logger, ctx)
			if delayed > time.Second {
				l = log.With(l, "delayed", delayed)
			}
			l = log.With(l, "job", descriptor.Name, "schedule", descriptor.RawSpec)
			l.Log("msg", logging.Sprintf("job %s started", descriptor.Name))
			err := innerRun(ctx)
			if err != nil {
				l.Log("msg", logging.Sprintf("job %s finished with error: %s", descriptor.Name, err))
				return err
			}
			l.Log("msg", logging.Sprintf("job %s completed", descriptor.Name))
			return nil
		}
	}
}

// WithTracing returns a new Universal job that will trace.
func WithTracing(tracer opentracing.Tracer) JobOption {
	return func(descriptor *JobDescriptor) {
		innerRun := descriptor.Run
		descriptor.Run = func(ctx context.Context) error {
			span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, fmt.Sprintf("Job: %s", descriptor.Name))
			defer span.Finish()
			span.SetTag("schedule", descriptor.RawSpec)
			err := innerRun(ctx)
			if err != nil {
				ext.LogError(span, err)
				return err
			}
			return nil
		}
	}
}

//go:embed start.lua
var startLua string

//go:embed commit.lua
var commitLua string

//go:embed cancel.lua
var cancelLua string

// PersistenceConfig is the configuration for WithPersistence.
type PersistenceConfig struct {
	// How long will the redis lock be held. By default, the lock will be held for a minute.
	LockTTL time.Duration
	// Only missed schedules before this duration can be compensated. By default, it is calculated from the gap between each run.
	MaxRecoverableDuration time.Duration
	// The prefix of keys in redis. Make sure each project use different keys to avoid collision.
	KeyPrefix string
}

// WithPersistence ensures the job will be run at least once by committing
// successful runs into redis. If one or more schedule is missed, the driver will
// compensate the missing run(s) in the next schedule. Users should use
// GetCurrentSchedule method to determine the targeted schedule of the current
// run, instead of relying on time.Now.
func WithPersistence(redis redis.UniversalClient, config PersistenceConfig) JobOption {
	if config.LockTTL == 0 {
		config.LockTTL = time.Minute
	}
	return func(descriptor *JobDescriptor) {
		innerRun := descriptor.Run
		descriptor.Run = func(ctx context.Context) error {
			var (
				expectedNext       time.Time
				expectedNextString interface{}
				err                error
				current            = GetCurrentSchedule(ctx)
				next               = GetNextSchedule(ctx)
			)

			hostname, _ := os.Hostname()
			expire := calculateNextTTL(config, current, next)

			cancel := func() {
				redis.Eval(ctx, cancelLua, []string{descriptor.Name}, []string{hostname})
			}

			for next.Sub(descriptor.next) <= 0 {
				keys := []string{strings.Join([]string{config.KeyPrefix, descriptor.Name}, ":")}
				argv := []string{hostname, fmt.Sprintf("%.0f", config.LockTTL.Round(time.Second).Seconds())}
				expectedNextString, err = redis.Eval(ctx, startLua, keys, argv).Result()
				if err != nil {
					cancel()
					return fmt.Errorf("failed to start job: %w", err)
				}
				if expectedNextString == -2 {
					cancel()
					return errors.New("job is already running")
				}
				if expectedNextString == -1 {
					expectedNext = current
				} else {
					expectedNext, err = time.Parse(time.RFC3339, expectedNextString.(string))
					if err != nil {
						cancel()
						return fmt.Errorf("could not parse expected next time: %s", err)
					}
				}

				current = expectedNext
				next = descriptor.Schedule.Next(current)
				ctx = context.WithValue(ctx, prevContextKey, current)
				ctx = context.WithValue(ctx, nextContextKey, next)
				err = innerRun(ctx)
				if err != nil {
					cancel()
					return err
				}
				if err := redis.Eval(ctx, commitLua, keys, []string{hostname, next.Format(time.RFC3339), expire}).Err(); err != nil {
					cancel()
					return fmt.Errorf("failed to commit job: %w", err)
				}
			}
			return nil
		}
	}
}

// SkipIfOverlap returns a new JobDescriptor that will skip the job if it overlaps with another job.
func SkipIfOverlap() JobOption {
	ch := make(chan struct{}, 1)
	return func(descriptor *JobDescriptor) {
		innerRun := descriptor.Run
		descriptor.Run = func(ctx context.Context) error {
			select {
			case ch <- struct{}{}:
				defer func() {
					<-ch
				}()
				return innerRun(ctx)
			default:
				return errors.New("skipped due to overlap")
			}
		}
	}
}

// DelayIfOverlap returns a new JobDescriptor that will delay the job if it overlaps with another job.
func DelayIfOverlap() JobOption {
	ch := make(chan struct{}, 1)
	return func(descriptor *JobDescriptor) {
		innerRun := descriptor.Run
		descriptor.Run = func(ctx context.Context) error {
			ch <- struct{}{}
			defer func() {
				<-ch
			}()
			return innerRun(ctx)
		}
	}
}

// TimeoutIfOverlap returns a new JobDescriptor that will cancel the job's context if the next schedule is due.
func TimeoutIfOverlap() JobOption {
	return func(descriptor *JobDescriptor) {
		innerRun := descriptor.Run
		descriptor.Run = func(ctx context.Context) error {
			if !GetNextSchedule(ctx).IsZero() {
				ctx, cancel := context.WithDeadline(ctx, GetNextSchedule(ctx))
				defer cancel()
				return innerRun(ctx)
			}
			return innerRun(ctx)
		}
	}
}

// Recover returns a new JobDescriptor that will recover from panics.
func Recover(logger log.Logger) JobOption {
	return func(descriptor *JobDescriptor) {
		innerRun := descriptor.Run
		descriptor.Run = func(ctx context.Context) error {
			defer func() {
				if r := recover(); r != nil {
					logging.WithContext(logger, ctx).Log("msg", "job panicked", "err", r, "stack", debug.Stack())
				}
			}()
			return innerRun(ctx)
		}
	}
}

func calculateNextTTL(config PersistenceConfig, current, next time.Time) string {
	if config.MaxRecoverableDuration != 0 {
		return fmt.Sprintf("%0.f", config.MaxRecoverableDuration.Seconds())
	}
	return fmt.Sprintf("%0.f", math.Max(3600, 2*float64(next.Sub(current).Seconds())+1))
}
