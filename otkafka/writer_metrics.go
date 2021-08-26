package otkafka

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/segmentio/kafka-go"
)

type writerCollector struct {
	factory  WriterFactory
	stats    *WriterStats
	interval time.Duration
}

// WriterStats is a collection of metrics for kafka writer info.
type WriterStats struct {
	Writes   metrics.Counter
	Messages metrics.Counter
	Bytes    metrics.Counter
	Errors   metrics.Counter

	MaxAttempts  metrics.Gauge
	MaxBatchSize metrics.Gauge
	BatchTimeout metrics.Gauge
	ReadTimeout  metrics.Gauge
	WriteTimeout metrics.Gauge
	RequiredAcks metrics.Gauge
	Async        metrics.Gauge

	BatchTime  AggStats
	WriteTime  AggStats
	WaitTime   AggStats
	Retries    AggStats
	BatchSize  AggStats
	BatchBytes AggStats
}

// newCollector creates a new kafka writer wrapper containing the name of the reader.
func newWriterCollector(factory WriterFactory, stats *WriterStats, interval time.Duration) *writerCollector {
	return &writerCollector{
		factory:  factory,
		stats:    stats,
		interval: interval,
	}
}

// collectConnectionStats collects kafka writer info for Prometheus to scrape.
func (d *writerCollector) collectConnectionStats() {
	for k, v := range d.factory.List() {
		writer := v.Conn.(*kafka.Writer)
		stats := writer.Stats()
		withValues := []string{"writer", k, "topic", stats.Topic}

		d.stats.Writes.With(withValues...).Add(float64(stats.Writes))
		d.stats.Messages.With(withValues...).Add(float64(stats.Messages))
		d.stats.Bytes.With(withValues...).Add(float64(stats.Bytes))
		d.stats.Errors.With(withValues...).Add(float64(stats.Errors))

		d.stats.BatchTime.Min.With(withValues...).Add(stats.BatchTime.Min.Seconds())
		d.stats.BatchTime.Max.With(withValues...).Add(stats.BatchTime.Max.Seconds())
		d.stats.BatchTime.Avg.With(withValues...).Add(stats.BatchTime.Avg.Seconds())

		d.stats.WriteTime.Min.With(withValues...).Add(stats.WriteTime.Min.Seconds())
		d.stats.WriteTime.Max.With(withValues...).Add(stats.WriteTime.Max.Seconds())
		d.stats.WriteTime.Avg.With(withValues...).Add(stats.WriteTime.Avg.Seconds())

		d.stats.WaitTime.Min.With(withValues...).Add(stats.WaitTime.Min.Seconds())
		d.stats.WaitTime.Max.With(withValues...).Add(stats.WaitTime.Max.Seconds())
		d.stats.WaitTime.Avg.With(withValues...).Add(stats.WaitTime.Avg.Seconds())

		d.stats.Retries.Min.With(withValues...).Add(float64(stats.Retries.Min))
		d.stats.Retries.Max.With(withValues...).Add(float64(stats.Retries.Max))
		d.stats.Retries.Avg.With(withValues...).Add(float64(stats.Retries.Avg))

		d.stats.BatchSize.Min.With(withValues...).Add(float64(stats.BatchSize.Min))
		d.stats.BatchSize.Max.With(withValues...).Add(float64(stats.BatchSize.Max))
		d.stats.BatchSize.Avg.With(withValues...).Add(float64(stats.BatchSize.Avg))

		d.stats.BatchBytes.Min.With(withValues...).Add(float64(stats.BatchBytes.Min))
		d.stats.BatchBytes.Max.With(withValues...).Add(float64(stats.BatchBytes.Max))
		d.stats.BatchBytes.Avg.With(withValues...).Add(float64(stats.BatchBytes.Avg))

		d.stats.MaxAttempts.With(withValues...).Set(float64(stats.MaxAttempts))
		d.stats.MaxBatchSize.With(withValues...).Set(float64(stats.MaxBatchSize))
		d.stats.BatchTimeout.With(withValues...).Set(stats.BatchTimeout.Seconds())
		d.stats.ReadTimeout.With(withValues...).Set(stats.ReadTimeout.Seconds())
		d.stats.WriteTimeout.With(withValues...).Set(stats.WriteTimeout.Seconds())
		d.stats.RequiredAcks.With(withValues...).Set(float64(stats.RequiredAcks))
		var async float64
		if stats.Async {
			async = 1.0
		}
		d.stats.Async.With(withValues...).Set(async)
	}
}
