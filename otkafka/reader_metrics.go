package otkafka

import (
	"github.com/go-kit/kit/metrics"
	"github.com/segmentio/kafka-go"
	"time"
)

type readerCollector struct {
	factory  ReaderFactory
	stats    *ReaderStats
	interval time.Duration
}

// ThreeStats is a gauge group struct.
type ThreeStats struct {
	Min metrics.Gauge
	Max metrics.Gauge
	Avg metrics.Gauge
}

// ReaderStats is a collection of metrics for kafka reader info.
type ReaderStats struct {
	Dials      metrics.Counter
	Fetches    metrics.Counter
	Messages   metrics.Counter
	Bytes      metrics.Counter
	Rebalances metrics.Counter
	Timeouts   metrics.Counter
	Errors     metrics.Counter

	Offset        metrics.Gauge
	Lag           metrics.Gauge
	MinBytes      metrics.Gauge
	MaxBytes      metrics.Gauge
	MaxWait       metrics.Gauge
	QueueLength   metrics.Gauge
	QueueCapacity metrics.Gauge

	DialTime   ThreeStats
	ReadTime   ThreeStats
	WaitTime   ThreeStats
	FetchSize  ThreeStats
	FetchBytes ThreeStats
}

// newCollector creates a new kafka reader wrapper containing the name of the reader.
func newReaderCollector(factory ReaderFactory, stats *ReaderStats, interval time.Duration) *readerCollector {
	return &readerCollector{
		factory:  factory,
		stats:    stats,
		interval: interval,
	}
}

// collectConnectionStats collects kafka reader info for Prometheus to scrape.
func (d *readerCollector) collectConnectionStats() {
	for k, v := range d.factory.List() {
		reader := v.Conn.(*kafka.Reader)
		stats := reader.Stats()
		withValues := []string{"reader", k, "client_id", stats.ClientID, "topic", stats.Topic, "partition", stats.Partition}

		d.stats.Dials.With(withValues...).Add(float64(stats.Dials))
		d.stats.Fetches.With(withValues...).Add(float64(stats.Fetches))
		d.stats.Messages.With(withValues...).Add(float64(stats.Messages))
		d.stats.Bytes.With(withValues...).Add(float64(stats.Bytes))
		d.stats.Rebalances.With(withValues...).Add(float64(stats.Rebalances))
		d.stats.Timeouts.With(withValues...).Add(float64(stats.Timeouts))
		d.stats.Errors.With(withValues...).Add(float64(stats.Errors))

		d.stats.Offset.With(withValues...).Set(float64(stats.Offset))
		d.stats.Lag.With(withValues...).Set(float64(stats.Lag))
		d.stats.MinBytes.With(withValues...).Set(float64(stats.MinBytes))
		d.stats.MaxBytes.With(withValues...).Set(float64(stats.MaxBytes))
		d.stats.MaxWait.With(withValues...).Set(stats.MaxWait.Seconds())
		d.stats.QueueLength.With(withValues...).Set(float64(stats.QueueLength))
		d.stats.QueueCapacity.With(withValues...).Set(float64(stats.QueueCapacity))

		d.stats.DialTime.Min.With(withValues...).Set(stats.DialTime.Min.Seconds())
		d.stats.DialTime.Max.With(withValues...).Set(stats.DialTime.Max.Seconds())
		d.stats.DialTime.Avg.With(withValues...).Set(stats.DialTime.Avg.Seconds())

		d.stats.ReadTime.Min.With(withValues...).Set(stats.ReadTime.Min.Seconds())
		d.stats.ReadTime.Max.With(withValues...).Set(stats.ReadTime.Max.Seconds())
		d.stats.ReadTime.Avg.With(withValues...).Set(stats.ReadTime.Avg.Seconds())

		d.stats.WaitTime.Min.With(withValues...).Set(stats.WaitTime.Min.Seconds())
		d.stats.WaitTime.Max.With(withValues...).Set(stats.WaitTime.Max.Seconds())
		d.stats.WaitTime.Avg.With(withValues...).Set(stats.WaitTime.Avg.Seconds())

		d.stats.FetchSize.Min.With(withValues...).Set(float64(stats.FetchSize.Min))
		d.stats.FetchSize.Max.With(withValues...).Set(float64(stats.FetchSize.Max))
		d.stats.FetchSize.Avg.With(withValues...).Set(float64(stats.FetchSize.Avg))

		d.stats.FetchBytes.Min.With(withValues...).Set(float64(stats.FetchBytes.Min))
		d.stats.FetchBytes.Max.With(withValues...).Set(float64(stats.FetchBytes.Max))
		d.stats.FetchBytes.Avg.With(withValues...).Set(float64(stats.FetchBytes.Avg))
	}
}
