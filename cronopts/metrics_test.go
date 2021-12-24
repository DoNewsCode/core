package cronopts

import (
	"testing"
	"time"

	"github.com/go-kit/kit/metrics/generic"
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
)

func TestMeasure(t *testing.T) {
	metrics := NewCronJobMetrics(generic.NewHistogram("foo", 5), generic.NewCounter("bar"))
	metrics = metrics.Module("x").Job("y")
	Measure(metrics)(cron.FuncJob(func() {
		time.Sleep(time.Millisecond)
	})).Run()
	assert.True(t, metrics.module)
	assert.True(t, metrics.job)
	assert.True(t, metrics.CronJobDurationSeconds.(*generic.Histogram).Quantile(0.5) > 0)
}
