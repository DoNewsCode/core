package stub

import "github.com/go-kit/kit/metrics"

type Histogram struct {
	LabelValues   []string
	ObservedValue float64
}

func (h *Histogram) With(labelValues ...string) metrics.Histogram {
	h.LabelValues = labelValues
	return h
}

func (h *Histogram) Observe(value float64) {
	h.ObservedValue = value
}

type Gauge struct {
	LabelValues []string
	GaugeValue  float64
}

func (g *Gauge) With(labelValues ...string) metrics.Gauge {
	g.LabelValues = labelValues
	return g
}

func (g *Gauge) Set(value float64) {
	g.GaugeValue = value
}

func (g *Gauge) Add(delta float64) {
	g.GaugeValue = g.GaugeValue + delta
}

type Counter struct {
	LabelValues  []string
	CounterValue float64
}

func (c *Counter) With(labelValues ...string) metrics.Counter {
	c.LabelValues = labelValues
	return c
}

func (c *Counter) Add(delta float64) {
	c.CounterValue = c.CounterValue + delta
}
