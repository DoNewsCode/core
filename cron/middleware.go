package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/DoNewsCode/core/cronopts"
	"github.com/DoNewsCode/core/logging"
	"github.com/go-kit/log"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/robfig/cron/v3"
)

type JobMiddleware func(descriptors JobDescriptor) JobDescriptor

func SetName(name string) JobMiddleware {
	return func(descriptors JobDescriptor) JobDescriptor {
		descriptors.Name = name
		return descriptors
	}
}

func ReplaceSchedule(schedule cron.Schedule) JobMiddleware {
	return func(descriptors JobDescriptor) JobDescriptor {
		descriptors.Schedule = schedule
		return descriptors
	}
}

// WrapMetrics returns a new JobDescriptor that will report metrics.
func WrapMetrics(metrics *cronopts.CronJobMetrics) JobMiddleware {
	return func(descriptor JobDescriptor) JobDescriptor {
		descriptor.Run = func(ctx context.Context) error {
			start := time.Now()
			metrics = metrics.Job(descriptor.Name)
			defer metrics.Observe(time.Since(start))
			err := descriptor.Run(ctx)
			if err != nil {
				metrics.Fail()
				return err
			}
			return nil
		}
		return descriptor
	}
}

// WrapLogs returns a new Universal job that will log.
func WrapLogs(logger log.Logger) JobMiddleware {
	return func(descriptor JobDescriptor) JobDescriptor {
		descriptor.Run = func(ctx context.Context) error {
			logger = logging.WithContext(logger, ctx)
			logger = log.With(logger, "job", descriptor.Name, "schedule", descriptor.RawSpec)
			logger.Log("msg", logging.Sprintf("job %s started", descriptor.Name))
			err := descriptor.Run(ctx)
			if err != nil {
				logger.Log("msg", logging.Sprintf("job %s finished with error: %s", descriptor.Name, err))
				return err
			}
			logger.Log("msg", logging.Sprintf("job %s completed", descriptor.Name))
			return nil
		}
		return descriptor
	}
}

// WrapTracing returns a new Universal job that will trace.
func WrapTracing(tracer opentracing.Tracer) JobMiddleware {
	return func(descriptor JobDescriptor) JobDescriptor {
		descriptor.Run = func(ctx context.Context) error {
			span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, fmt.Sprintf("Job: %s", descriptor.Name))
			defer span.Finish()
			span.SetTag("schedule", descriptor.RawSpec)
			err := descriptor.Run(ctx)
			if err != nil {
				ext.Error.Set(span, true)
				return err
			}
			return nil
		}
		return descriptor
	}
}
