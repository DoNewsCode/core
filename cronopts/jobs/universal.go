// Package jobs contains a universal job type that implements cron.Job interface.
// It is designed to be used with cron.New() and cron.AddJob() methods. Compared
// to anonymous jobs, this job type supports go idioms like context and error,
// and offers hook point for common observability concerns such as metrics,
// logging and tracing.
package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/DoNewsCode/core/cronopts"
	"github.com/DoNewsCode/core/dag"
	"github.com/DoNewsCode/core/logging"
	"github.com/go-kit/log"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// Universal is a generic job that can be used to run any task. It implements the
// cron.Job interface and supports context parameter and error propagation. The
// Name parameter allows common observability concerns such as logging, tracing
// and metrics to take advantage.
type Universal struct {
	Name string
	Do   func(ctx context.Context) error
}

// New returns a new Universal job.
func New(name string, do func(ctx context.Context) error, wrapper ...func(universal Universal) Universal) Universal {
	base := Universal{
		Name: name,
		Do:   do,
	}
	for _, w := range wrapper {
		base = w(base)
	}
	return base
}

// NewFromDAG returns a new Universal job from a DAG.
func NewFromDAG(name string, dag *dag.DAG, wrapper ...func(universal Universal) Universal) Universal {
	return New(name, dag.Run, wrapper...)
}

// Run implements the cron.Job interface.
func (s Universal) Run() {
	_ = s.Do(context.Background())
}

// WithMetrics returns a new Universal job that will report metrics.
func WithMetrics(metrics *cronopts.CronJobMetrics) func(universal Universal) Universal {
	return func(universal Universal) Universal {
		return Universal{
			Name: universal.Name,
			Do: func(ctx context.Context) error {
				start := time.Now()
				metrics = metrics.Job(universal.Name)
				defer metrics.Observe(time.Since(start))
				err := universal.Do(ctx)
				if err != nil {
					metrics.Fail()
					return err
				}
				return nil
			},
		}
	}
}

// WithLogs returns a new Universal job that will log.
func WithLogs(logger log.Logger) func(universal Universal) Universal {
	return func(universal Universal) Universal {
		return Universal{
			Name: universal.Name,
			Do: func(ctx context.Context) error {
				logger = logging.WithContext(logger, ctx)
				logger = log.With(logger, "job", universal.Name)
				logger.Log("msg", logging.Sprintf("job %s started", universal.Name))
				err := universal.Do(ctx)
				if err != nil {
					logger.Log("msg", logging.Sprintf("job %s finished with error %s", universal.Name, err))
					return err
				}
				logger.Log("msg", logging.Sprintf("job %s completed", universal.Name))
				return nil
			},
		}
	}
}

// WithTracing returns a new Universal job that will trace.
func WithTracing(tracer opentracing.Tracer) func(universal Universal) Universal {
	return func(universal Universal) Universal {
		return Universal{
			Name: universal.Name,
			Do: func(ctx context.Context) error {
				span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, fmt.Sprintf("Job: %s", universal.Name))
				defer span.Finish()
				err := universal.Do(ctx)
				if err != nil {
					ext.Error.Set(span, true)
					return err
				}
				return nil
			},
		}
	}
}
