package stub

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetrics(t *testing.T) {
	h := &Histogram{}
	g := &Gauge{}
	c := &Counter{}

	h.With("foo", "bar").Observe(1.0)
	assert.ElementsMatch(t, h.LabelValues, []string{"foo", "bar"})
	assert.Equal(t, h.ObservedValue, 1.0)

	g.With("foo", "bar").Set(1.0)
	g.Add(1)
	assert.Equal(t, g.GaugeValue, 2.0)
	assert.ElementsMatch(t, g.LabelValues, []string{"foo", "bar"})

	c.With("foo", "bar").Add(1)
	c.Add(1)
	assert.ElementsMatch(t, c.LabelValues, []string{"foo", "bar"})
	assert.Equal(t, c.CounterValue, 2.0)
}
