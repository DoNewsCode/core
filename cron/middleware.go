package cron

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/DoNewsCode/core/logging"
	"github.com/go-kit/log"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/robfig/cron/v3"
)

// JobOptions is a middleware for cron jobs.
type JobOptions func(descriptors *JobDescriptor)

// WithName sets the name of the job.
func WithName(name string) JobOptions {
	return func(descriptor *JobDescriptor) {
		descriptor.Name = name
	}
}

// WithSchedule sets the cron schedule of the job.
func WithSchedule(schedule cron.Schedule) JobOptions {
	return func(descriptor *JobDescriptor) {
		descriptor.Schedule = schedule
	}
}

// WithMetrics returns a new JobDescriptor that will report metrics.
func WithMetrics(metrics *CronJobMetrics) JobOptions {
	return func(descriptor *JobDescriptor) {
		innerRun := descriptor.Run
		descriptor.Run = func(ctx context.Context) error {
			start := time.Now()
			metrics = metrics.Job(descriptor.Name).Schedule(descriptor.RawSpec)
			defer metrics.Observe(time.Since(start))
			err := innerRun(ctx)
			if err != nil {
				metrics.Fail()
				return err
			}
			return nil
		}
	}
}

// WithLogging returns a new Universal job that will log.
func WithLogging(logger log.Logger) JobOptions {
	return func(descriptor *JobDescriptor) {
		innerRun := descriptor.Run
		descriptor.Run = func(ctx context.Context) error {
			due := GetCurrentSchedule(ctx)
			delayed := due.Sub(time.Now())
			logger = logging.WithContext(logger, ctx)
			if delayed > time.Second {
				log.With(logger, "delayed", delayed)
			}
			logger = log.With(logger, "job", descriptor.Name, "schedule", descriptor.RawSpec)
			logger.Log("msg", logging.Sprintf("job %s started", descriptor.Name))
			err := innerRun(ctx)
			if err != nil {
				logger.Log("msg", logging.Sprintf("job %s finished with error: %s", descriptor.Name, err))
				return err
			}
			logger.Log("msg", logging.Sprintf("job %s completed", descriptor.Name))
			return nil
		}
	}
}

// WithTracing returns a new Universal job that will trace.
func WithTracing(tracer opentracing.Tracer) JobOptions {
	return func(descriptor *JobDescriptor) {
		innerRun := descriptor.Run
		descriptor.Run = func(ctx context.Context) error {
			span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, fmt.Sprintf("Job: %s", descriptor.Name))
			defer span.Finish()
			span.SetTag("schedule", descriptor.RawSpec)
			err := innerRun(ctx)
			if err != nil {
				ext.Error.Set(span, true)
				return err
			}
			return nil
		}
	}
}

// SkipIfOverlap returns a new JobDescriptor that will skip the job if it overlaps with another job.
func SkipIfOverlap() JobOptions {
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
func DelayIfOverlap() JobOptions {
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
func TimeoutIfOverlap() JobOptions {
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
func Recover(logger log.Logger) JobOptions {
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
