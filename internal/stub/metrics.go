package stub

import "github.com/go-kit/kit/metrics"

// Histogram is a stub implementation of the go-kit metrics.Histogram interface.
type Histogram struct {
	LabelValues   []string
	ObservedValue float64
}

// With returns a new Histogram with the given label values.
func (h *Histogram) With(labelValues ...string) metrics.Histogram {
	h.LabelValues = labelValues
	return h
}

//	Observe records the given value.
func (h *Histogram) Observe(value float64) {
	h.ObservedValue = value
}

// Gauge is a stub implementation of the go-kit metrics.Gauge interface.
type Gauge struct {
	LabelValues []string
	GaugeValue  float64
}

// With returns a new Gauge with the given label values.
func (g *Gauge) With(labelValues ...string) metrics.Gauge {
	g.LabelValues = labelValues
	return g
}

// Set sets the gauge value.
func (g *Gauge) Set(value float64) {
	g.GaugeValue = value
}

// Add adds the given value to the gauge.
func (g *Gauge) Add(delta float64) {
	g.GaugeValue = g.GaugeValue + delta
}

// Counter is a stub implementation of the go-kit metrics.Counter interface.
type Counter struct {
	LabelValues  []string
	CounterValue float64
}

// With returns a new Counter with the given label values.
func (c *Counter) With(labelValues ...string) metrics.Counter {
	c.LabelValues = labelValues
	return c
}

// Add adds the given value to the counter.
func (c *Counter) Add(delta float64) {
	c.CounterValue = c.CounterValue + delta
}
