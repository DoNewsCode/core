package stub

import (
	"github.com/go-kit/kit/metrics"
	"sync"
)

// LabelValues contains the set of labels and their corresponding values.
type LabelValues []string

// Label returns the label of given name.
func (l LabelValues) Label(name string) string {
	for i := 0; i < len(l); i += 2 {
		if l[i] == name {
			return l[i+1]
		}
	}
	return ""
}

// Histogram is a stub implementation of the go-kit metrics.Histogram interface.
type Histogram struct {
	sync.Mutex
	LabelValues   LabelValues
	ObservedValue float64
}

// With returns a new Histogram with the given label values.
func (h *Histogram) With(labelValues ...string) metrics.Histogram {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()
	h.LabelValues = labelValues
	return h
}

// Observe records the given value.
func (h *Histogram) Observe(value float64) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()
	h.ObservedValue = value
}

// Gauge is a stub implementation of the go-kit metrics.Gauge interface.
type Gauge struct {
	sync.Mutex
	LabelValues []string
	GaugeValue  float64
}

// With returns a new Gauge with the given label values.
func (g *Gauge) With(labelValues ...string) metrics.Gauge {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()
	g.LabelValues = labelValues
	return g
}

// Set sets the gauge value.
func (g *Gauge) Set(value float64) {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()
	g.GaugeValue = value
}

// Add adds the given value to the gauge.
func (g *Gauge) Add(delta float64) {
	g.Mutex.Lock()
	defer g.Mutex.Unlock()
	g.GaugeValue = g.GaugeValue + delta
}

// Counter is a stub implementation of the go-kit metrics.Counter interface.
type Counter struct {
	sync.Mutex
	LabelValues  []string
	CounterValue float64
}

// With returns a new Counter with the given label values.
func (c *Counter) With(labelValues ...string) metrics.Counter {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.LabelValues = labelValues
	return c
}

// Add adds the given value to the counter.
func (c *Counter) Add(delta float64) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.CounterValue = c.CounterValue + delta
}
