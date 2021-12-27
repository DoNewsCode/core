package cronopts

import (
	"testing"
	"time"

	"github.com/DoNewsCode/core/internal/stub"
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
)

func TestMeasure(t *testing.T) {
	histogram := &stub.Histogram{}
	counter := &stub.Counter{}
	metrics := NewCronJobMetrics(histogram, counter)
	metrics = metrics.Module("x").Job("y")
	Measure(metrics)(cron.FuncJob(func() {
		time.Sleep(time.Millisecond)
	})).Run()
	assert.ElementsMatch(t, histogram.LabelValues, []string{"module", "x", "job", "y"})
	assert.True(t, histogram.ObservedValue > 0)
	metrics.Fail()
	assert.ElementsMatch(t, counter.LabelValues, []string{"module", "x", "job", "y"})
	assert.True(t, counter.CounterValue == 1)
}
