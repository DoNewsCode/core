package otkafka

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/segmentio/kafka-go"
)

type readerCollector struct {
	factory  ReaderFactory
	stats    *ReaderStats
	interval time.Duration
}

// AggStats is a gauge group struct.
type AggStats struct {
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

	DialTime   AggStats
	ReadTime   AggStats
	WaitTime   AggStats
	FetchSize  AggStats
	FetchBytes AggStats

	reader string
}

// Reader sets the writer label in WriterStats.
func (r *ReaderStats) Reader(reader string) *ReaderStats {
	withValues := []string{"reader", reader}
	return &ReaderStats{
		Dials:         r.Dials.With(withValues...),
		Fetches:       r.Fetches.With(withValues...),
		Messages:      r.Messages.With(withValues...),
		Bytes:         r.Bytes.With(withValues...),
		Rebalances:    r.Rebalances.With(withValues...),
		Timeouts:      r.Timeouts.With(withValues...),
		Errors:        r.Errors.With(withValues...),
		Offset:        r.Offset.With(withValues...),
		Lag:           r.Lag.With(withValues...),
		MinBytes:      r.MinBytes.With(withValues...),
		MaxBytes:      r.MaxBytes.With(withValues...),
		MaxWait:       r.MaxWait.With(withValues...),
		QueueLength:   r.QueueLength.With(withValues...),
		QueueCapacity: r.QueueCapacity.With(withValues...),
		DialTime: AggStats{
			Min: r.DialTime.Min.With(withValues...),
			Max: r.DialTime.Max.With(withValues...),
			Avg: r.DialTime.Avg.With(withValues...),
		},
		ReadTime: AggStats{
			Min: r.ReadTime.Min.With(withValues...),
			Max: r.ReadTime.Max.With(withValues...),
			Avg: r.ReadTime.Avg.With(withValues...),
		},
		WaitTime: AggStats{
			Min: r.WaitTime.Min.With(withValues...),
			Max: r.WaitTime.Max.With(withValues...),
			Avg: r.WaitTime.Avg.With(withValues...),
		},
		FetchSize: AggStats{
			Min: r.FetchSize.Min.With(withValues...),
			Max: r.FetchSize.Max.With(withValues...),
			Avg: r.FetchSize.Avg.With(withValues...),
		},
		FetchBytes: AggStats{
			Min: r.FetchBytes.Min.With(withValues...),
			Max: r.FetchBytes.Max.With(withValues...),
			Avg: r.FetchBytes.Avg.With(withValues...),
		},
		reader: reader,
	}
}

// Observe records the reader stats. It should be called periodically.
func (r *ReaderStats) Observe(stats kafka.ReaderStats) {
	withValues := []string{"client_id", stats.ClientID, "topic", stats.Topic, "partition", stats.Partition}
	r.Dials.With(withValues...).Add(float64(stats.Dials))
	r.Fetches.With(withValues...).Add(float64(stats.Fetches))
	r.Messages.With(withValues...).Add(float64(stats.Messages))
	r.Bytes.With(withValues...).Add(float64(stats.Bytes))
	r.Rebalances.With(withValues...).Add(float64(stats.Rebalances))
	r.Timeouts.With(withValues...).Add(float64(stats.Timeouts))
	r.Errors.With(withValues...).Add(float64(stats.Errors))

	r.Offset.With(withValues...).Set(float64(stats.Offset))
	r.Lag.With(withValues...).Set(float64(stats.Lag))
	r.MinBytes.With(withValues...).Set(float64(stats.MinBytes))
	r.MaxBytes.With(withValues...).Set(float64(stats.MaxBytes))
	r.MaxWait.With(withValues...).Set(stats.MaxWait.Seconds())
	r.QueueLength.With(withValues...).Set(float64(stats.QueueLength))
	r.QueueCapacity.With(withValues...).Set(float64(stats.QueueCapacity))

	r.DialTime.Min.With(withValues...).Set(stats.DialTime.Min.Seconds())
	r.DialTime.Max.With(withValues...).Set(stats.DialTime.Max.Seconds())
	r.DialTime.Avg.With(withValues...).Set(stats.DialTime.Avg.Seconds())

	r.ReadTime.Min.With(withValues...).Set(stats.ReadTime.Min.Seconds())
	r.ReadTime.Max.With(withValues...).Set(stats.ReadTime.Max.Seconds())
	r.ReadTime.Avg.With(withValues...).Set(stats.ReadTime.Avg.Seconds())

	r.WaitTime.Min.With(withValues...).Set(stats.WaitTime.Min.Seconds())
	r.WaitTime.Max.With(withValues...).Set(stats.WaitTime.Max.Seconds())
	r.WaitTime.Avg.With(withValues...).Set(stats.WaitTime.Avg.Seconds())

	r.FetchSize.Min.With(withValues...).Set(float64(stats.FetchSize.Min))
	r.FetchSize.Max.With(withValues...).Set(float64(stats.FetchSize.Max))
	r.FetchSize.Avg.With(withValues...).Set(float64(stats.FetchSize.Avg))

	r.FetchBytes.Min.With(withValues...).Set(float64(stats.FetchBytes.Min))
	r.FetchBytes.Max.With(withValues...).Set(float64(stats.FetchBytes.Max))
	r.FetchBytes.Avg.With(withValues...).Set(float64(stats.FetchBytes.Avg))
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
		d.stats.Reader(k).Observe(stats)
	}
}
