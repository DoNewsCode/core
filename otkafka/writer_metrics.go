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

	writer bool
}

// Writer sets the writer label in WriterStats.
func (w *WriterStats) Writer(writer string) *WriterStats {
	withValues := []string{"writer", writer}
	return &WriterStats{
		Writes:       w.Writes.With(withValues...),
		Messages:     w.Messages.With(withValues...),
		Bytes:        w.Bytes.With(withValues...),
		Errors:       w.Errors.With(withValues...),
		MaxAttempts:  w.MaxAttempts.With(withValues...),
		MaxBatchSize: w.MaxBatchSize.With(withValues...),
		BatchTimeout: w.BatchTimeout.With(withValues...),
		ReadTimeout:  w.ReadTimeout.With(withValues...),
		WriteTimeout: w.WriteTimeout.With(withValues...),
		RequiredAcks: w.RequiredAcks.With(withValues...),
		BatchTime: AggStats{
			Min: w.BatchTime.Min.With(withValues...),
			Max: w.BatchTime.Max.With(withValues...),
			Avg: w.BatchTime.Avg.With(withValues...),
		},
		WriteTime: AggStats{
			Min: w.WriteTime.Min.With(withValues...),
			Max: w.WriteTime.Max.With(withValues...),
			Avg: w.WriteTime.Avg.With(withValues...),
		},
		Retries: AggStats{
			Min: w.Retries.Min.With(withValues...),
			Max: w.Retries.Max.With(withValues...),
			Avg: w.Retries.Avg.With(withValues...),
		},
		WaitTime: AggStats{
			Min: w.WaitTime.Min.With(withValues...),
			Max: w.WaitTime.Max.With(withValues...),
			Avg: w.WaitTime.Avg.With(withValues...),
		},
		BatchSize: AggStats{
			Min: w.BatchSize.Min.With(withValues...),
			Max: w.BatchSize.Max.With(withValues...),
			Avg: w.BatchSize.Avg.With(withValues...),
		},
		BatchBytes: AggStats{
			Min: w.BatchBytes.Min.With(withValues...),
			Max: w.BatchBytes.Max.With(withValues...),
			Avg: w.BatchBytes.Avg.With(withValues...),
		},
		Async:  w.Async.With(withValues...),
		writer: true,
	}
}

// Observe records the writer stats. It should called periodically.
func (w *WriterStats) Observe(stats kafka.WriterStats) *WriterStats {
	withValues := []string{"topic", stats.Topic}
	if !w.writer {
		withValues = append(withValues, "writer", "")
	}

	w.Writes.With(withValues...).Add(float64(stats.Writes))
	w.Messages.With(withValues...).Add(float64(stats.Messages))
	w.Bytes.With(withValues...).Add(float64(stats.Bytes))
	w.Errors.With(withValues...).Add(float64(stats.Errors))

	w.BatchTime.Min.With(withValues...).Add(stats.BatchTime.Min.Seconds())
	w.BatchTime.Max.With(withValues...).Add(stats.BatchTime.Max.Seconds())
	w.BatchTime.Avg.With(withValues...).Add(stats.BatchTime.Avg.Seconds())

	w.WriteTime.Min.With(withValues...).Add(stats.WriteTime.Min.Seconds())
	w.WriteTime.Max.With(withValues...).Add(stats.WriteTime.Max.Seconds())
	w.WriteTime.Avg.With(withValues...).Add(stats.WriteTime.Avg.Seconds())

	w.WaitTime.Min.With(withValues...).Add(stats.WaitTime.Min.Seconds())
	w.WaitTime.Max.With(withValues...).Add(stats.WaitTime.Max.Seconds())
	w.WaitTime.Avg.With(withValues...).Add(stats.WaitTime.Avg.Seconds())

	w.Retries.Min.With(withValues...).Add(float64(stats.Retries.Min))
	w.Retries.Max.With(withValues...).Add(float64(stats.Retries.Max))
	w.Retries.Avg.With(withValues...).Add(float64(stats.Retries.Avg))

	w.BatchSize.Min.With(withValues...).Add(float64(stats.BatchSize.Min))
	w.BatchSize.Max.With(withValues...).Add(float64(stats.BatchSize.Max))
	w.BatchSize.Avg.With(withValues...).Add(float64(stats.BatchSize.Avg))

	w.BatchBytes.Min.With(withValues...).Add(float64(stats.BatchBytes.Min))
	w.BatchBytes.Max.With(withValues...).Add(float64(stats.BatchBytes.Max))
	w.BatchBytes.Avg.With(withValues...).Add(float64(stats.BatchBytes.Avg))

	w.MaxAttempts.With(withValues...).Set(float64(stats.MaxAttempts))
	w.MaxBatchSize.With(withValues...).Set(float64(stats.MaxBatchSize))
	w.BatchTimeout.With(withValues...).Set(stats.BatchTimeout.Seconds())
	w.ReadTimeout.With(withValues...).Set(stats.ReadTimeout.Seconds())
	w.WriteTimeout.With(withValues...).Set(stats.WriteTimeout.Seconds())
	w.RequiredAcks.With(withValues...).Set(float64(stats.RequiredAcks))
	var async float64
	if stats.Async {
		async = 1.0
	}
	w.Async.With(withValues...).Set(async)
	return w
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
		d.stats.Writer(k).Observe(stats)
	}
}
