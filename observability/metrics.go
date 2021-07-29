package observability

import (
	"sync"

	"github.com/DoNewsCode/core/otkafka"

	"github.com/DoNewsCode/core/otgorm"
	"github.com/DoNewsCode/core/otredis"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

type histogram struct {
	once sync.Once
	*prometheus.Histogram
}

var his histogram

// ProvideHistogramMetrics returns a metrics.Histogram that is designed to measure incoming requests
// to the system. Note it has three labels: "module", "service", "method". If any label is missing,
// the system will panic.
func ProvideHistogramMetrics() metrics.Histogram {
	his.once.Do(func() {
		his.Histogram = prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Total time spent serving requests.",
		}, []string{"module", "service", "method"})
	})
	return &his
}

// ProvideGORMMetrics returns a *otgorm.Gauges that measures the connection info in databases.
// It is meant to be consumed by the otgorm.Providers.
func ProvideGORMMetrics() *otgorm.Gauges {
	return &otgorm.Gauges{
		Idle: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "gorm_idle_connections",
			Help: "number of idle connections",
		}, []string{"dbname", "driver"}),
		Open: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "gorm_open_connections",
			Help: "number of open connections",
		}, []string{"dbname", "driver"}),
		InUse: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "gorm_in_use_connections",
			Help: "number of in use connections",
		}, []string{"dbname", "driver"}),
	}
}

// ProvideRedisMetrics returns a *otredis.Gauges that measures the connection info in redis.
// It is meant to be consumed by the otredis.Providers.
func ProvideRedisMetrics() *otredis.Gauges {
	return &otredis.Gauges{
		Hits: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "redis_hit_connections",
			Help: "number of times free connection was found in the pool",
		}, []string{"dbname"}),
		Misses: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "redis_miss_connections",
			Help: "number of times free connection was NOT found in the pool",
		}, []string{"dbname"}),
		Timeouts: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "redis_timeout_connections",
			Help: "number of times a wait timeout occurred",
		}, []string{"dbname"}),
		TotalConns: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "redis_total_connections",
			Help: "number of total connections in the pool",
		}, []string{"dbname"}),
		IdleConns: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "redis_idle_connections",
			Help: "number of idle connections in the pool",
		}, []string{"dbname"}),
		StaleConns: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "redis_stale_connections",
			Help: "number of stale connections removed from the pool",
		}, []string{"dbname"}),
	}
}

// ProvideKafkaReaderMetrics returns a *otkafka.ReaderStats that measures the reader info in kafka.
// It is meant to be consumed by the otkafka.Providers.
func ProvideKafkaReaderMetrics() *otkafka.ReaderStats {
	labels := []string{"reader", "client_id", "topic", "partition"}

	return &otkafka.ReaderStats{
		Dials: prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_reader_dial_count",
			Help: "",
		}, labels),
		Fetches: prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_reader_fetch_count",
			Help: "",
		}, labels),
		Messages: prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_reader_message_count",
			Help: "",
		}, labels),
		Bytes: prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_reader_message_bytes",
			Help: "",
		}, labels),
		Rebalances: prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_reader_rebalance_count",
			Help: "",
		}, labels),
		Timeouts: prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_reader_timeout_count",
			Help: "",
		}, labels),
		Errors: prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_reader_error_count",
			Help: "",
		}, labels),
		Offset: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_reader_offset",
			Help: "",
		}, labels),
		Lag: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_reader_lag",
			Help: "",
		}, labels),
		MinBytes: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_reader_bytes_min",
			Help: "",
		}, labels),
		MaxBytes: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_reader_bytes_max",
			Help: "",
		}, labels),
		MaxWait: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_reader_fetch_wait_max",
			Help: "",
		}, labels),
		QueueLength: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_reader_queue_length",
			Help: "",
		}, labels),
		QueueCapacity: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_reader_queue_capacity",
			Help: "",
		}, labels),
		DialTime: otkafka.ThreeStats{
			Min: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_dial_seconds_min",
				Help: "",
			}, labels),
			Max: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_dial_seconds_max",
				Help: "",
			}, labels),
			Avg: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_dial_seconds_avg",
				Help: "",
			}, labels),
		},
		ReadTime: otkafka.ThreeStats{
			Min: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_read_seconds_min",
				Help: "",
			}, labels),
			Max: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_read_seconds_max",
				Help: "",
			}, labels),
			Avg: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_read_seconds_avg",
				Help: "",
			}, labels),
		},
		WaitTime: otkafka.ThreeStats{
			Min: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_wait_seconds_min",
				Help: "",
			}, labels),
			Max: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_wait_seconds_max",
				Help: "",
			}, labels),
			Avg: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_wait_seconds_avg",
				Help: "",
			}, labels),
		},
		FetchSize: otkafka.ThreeStats{
			Min: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_fetch_size_min",
				Help: "",
			}, labels),
			Max: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_fetch_size_max",
				Help: "",
			}, labels),
			Avg: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_fetch_size_avg",
				Help: "",
			}, labels),
		},
		FetchBytes: otkafka.ThreeStats{
			Min: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_fetch_bytes_min",
				Help: "",
			}, labels),
			Max: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_fetch_bytes_max",
				Help: "",
			}, labels),
			Avg: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_fetch_bytes_avg",
				Help: "",
			}, labels),
		},
	}
}

// ProvideKafkaWriterMetrics returns a *otkafka.WriterStats that measures the writer info in kafka.
// It is meant to be consumed by the otkafka.Providers.
func ProvideKafkaWriterMetrics() *otkafka.WriterStats {
	labels := []string{"writer", "topic"}
	return &otkafka.WriterStats{
		Writes: prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_writer_write_count",
			Help: "",
		}, labels),
		Messages: prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_writer_message_count",
			Help: "",
		}, labels),
		Bytes: prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_writer_message_bytes",
			Help: "",
		}, labels),
		Errors: prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_writer_error_count",
			Help: "",
		}, labels),
		MaxAttempts: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_writer_attempts_max",
			Help: "",
		}, labels),
		MaxBatchSize: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_writer_batch_max",
			Help: "",
		}, labels),
		BatchTimeout: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_writer_batch_timeout",
			Help: "",
		}, labels),
		ReadTimeout: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_writer_read_timeout",
			Help: "",
		}, labels),
		WriteTimeout: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_writer_write_timeout",
			Help: "",
		}, labels),
		RequiredAcks: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_writer_acks_required",
			Help: "",
		}, labels),
		Async: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_writer_async",
			Help: "",
		}, labels),
		BatchTime: otkafka.ThreeStats{
			Min: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_batch_seconds_min",
				Help: "",
			}, labels),
			Max: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_batch_seconds_max",
				Help: "",
			}, labels),
			Avg: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_batch_seconds_avg",
				Help: "",
			}, labels),
		},
		WriteTime: otkafka.ThreeStats{
			Min: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_write_seconds_min",
				Help: "",
			}, labels),
			Max: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_write_seconds_max",
				Help: "",
			}, labels),
			Avg: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_write_seconds_avg",
				Help: "",
			}, labels),
		},
		WaitTime: otkafka.ThreeStats{
			Min: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_wait_seconds_min",
				Help: "",
			}, labels),
			Max: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_wait_seconds_max",
				Help: "",
			}, labels),
			Avg: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_wait_seconds_avg",
				Help: "",
			}, labels),
		},
		Retries: otkafka.ThreeStats{
			Min: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_retries_count_min",
				Help: "",
			}, labels),
			Max: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_retries_count_max",
				Help: "",
			}, labels),
			Avg: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_retries_count_avg",
				Help: "",
			}, labels),
		},
		BatchSize: otkafka.ThreeStats{
			Min: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_batch_size_min",
				Help: "",
			}, labels),
			Max: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_batch_size_max",
				Help: "",
			}, labels),
			Avg: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_batch_size_avg",
				Help: "",
			}, labels),
		},
		BatchBytes: otkafka.ThreeStats{
			Min: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_batch_bytes_min",
				Help: "",
			}, labels),
			Max: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_batch_bytes_max",
				Help: "",
			}, labels),
			Avg: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_batch_bytes_avg",
				Help: "",
			}, labels),
		},
	}
}
