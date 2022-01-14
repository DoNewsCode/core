package observability

import (
	"github.com/DoNewsCode/core/cronopts"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/otgorm"
	"github.com/DoNewsCode/core/otkafka"
	"github.com/DoNewsCode/core/otredis"
	"github.com/DoNewsCode/core/srvgrpc"
	"github.com/DoNewsCode/core/srvhttp"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

// MetricsIn is the injection parameter of most metrics constructors in the observability package.
type MetricsIn struct {
	di.In

	Registerer stdprometheus.Registerer `optional:"true"`
}

// ProvideHTTPRequestDurationSeconds returns a *srvhttp.RequestDurationSeconds
// that is designed to measure incoming HTTP requests to the system. Note it has
// three labels: "module", "service", "route".
func ProvideHTTPRequestDurationSeconds(in MetricsIn) *srvhttp.RequestDurationSeconds {
	http := stdprometheus.NewHistogramVec(stdprometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "Total time spent serving requests.",
	}, []string{"module", "service", "route", "status"})

	if in.Registerer == nil {
		in.Registerer = stdprometheus.DefaultRegisterer
	}
	in.Registerer.MustRegister(http)

	return srvhttp.NewRequestDurationSeconds(prometheus.NewHistogram(http))
}

// ProvideGRPCRequestDurationSeconds returns a *srvgrpc.RequestDurationSeconds
// that is designed to measure incoming GRPC requests to the system. Note it has
// three labels: "module", "service", "route".
func ProvideGRPCRequestDurationSeconds(in MetricsIn) *srvgrpc.RequestDurationSeconds {
	grpc := stdprometheus.NewHistogramVec(stdprometheus.HistogramOpts{
		Name: "grpc_request_duration_seconds",
		Help: "Total time spent serving requests.",
	}, []string{"module", "service", "route", "status"})

	if in.Registerer == nil {
		in.Registerer = stdprometheus.DefaultRegisterer
	}
	in.Registerer.MustRegister(grpc)

	return srvgrpc.NewRequestDurationSeconds(prometheus.NewHistogram(grpc))
}

// ProvideCronJobMetrics returns a *cronopts.CronJobMetrics that is designed to
// measure cron job metrics. The returned metrics can be used like this:
//  metrics := cronopts.NewCronJobMetrics(...)
//  job := cron.NewChain(
//  	cron.Recover(logger),
//  	cronopts.Measure(metrics),
//	).Then(job)
func ProvideCronJobMetrics(in MetricsIn) *cronopts.CronJobMetrics {
	histogram := stdprometheus.NewHistogramVec(stdprometheus.HistogramOpts{
		Name: "cronjob_duration_seconds",
		Help: "Total time spent running cron jobs.",
	}, []string{"module", "job"})

	counter := stdprometheus.NewCounterVec(stdprometheus.CounterOpts{
		Name: "cronjob_failures_total",
		Help: "Total number of cron jobs that failed.",
	}, []string{"module", "job"})

	if in.Registerer == nil {
		in.Registerer = stdprometheus.DefaultRegisterer
	}

	in.Registerer.MustRegister(histogram)
	in.Registerer.MustRegister(counter)

	return cronopts.NewCronJobMetrics(prometheus.NewHistogram(histogram), prometheus.NewCounter(counter))
}

// ProvideGORMMetrics returns a *otgorm.Gauges that measures the connection info
// in databases. It is meant to be consumed by the otgorm.Providers.
func ProvideGORMMetrics(in MetricsIn) *otgorm.Gauges {
	if in.Registerer == nil {
		in.Registerer = stdprometheus.DefaultRegisterer
	}
	return otgorm.NewGauges(
		newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "gorm_idle_connections",
			Help: "number of idle connections",
		}, []string{"dbname", "driver"}, in.Registerer),
		newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "gorm_open_connections",
			Help: "number of open connections",
		}, []string{"dbname", "driver"}, in.Registerer),
		newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "gorm_in_use_connections",
			Help: "number of in use connections",
		}, []string{"dbname", "driver"}, in.Registerer),
	)
}

// ProvideRedisMetrics returns a RedisMetrics that measures the connection info in redis.
// It is meant to be consumed by the otredis.Providers.
func ProvideRedisMetrics(in MetricsIn) *otredis.Gauges {
	if in.Registerer == nil {
		in.Registerer = stdprometheus.DefaultRegisterer
	}
	return &otredis.Gauges{
		Hits: newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "redis_hit_connections",
			Help: "number of times free connection was found in the pool",
		}, []string{"dbname"}, in.Registerer),
		Misses: newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "redis_miss_connections",
			Help: "number of times free connection was NOT found in the pool",
		}, []string{"dbname"}, in.Registerer),
		Timeouts: newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "redis_timeout_connections",
			Help: "number of times a wait timeout occurred",
		}, []string{"dbname"}, in.Registerer),
		TotalConns: newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "redis_total_connections",
			Help: "number of total connections in the pool",
		}, []string{"dbname"}, in.Registerer),
		IdleConns: newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "redis_idle_connections",
			Help: "number of idle connections in the pool",
		}, []string{"dbname"}, in.Registerer),
		StaleConns: newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "redis_stale_connections",
			Help: "number of stale connections removed from the pool",
		}, []string{"dbname"}, in.Registerer),
	}
}

// ProvideKafkaReaderMetrics returns a *otkafka.ReaderStats that measures the reader info in kafka.
// It is meant to be consumed by the otkafka.Providers.
func ProvideKafkaReaderMetrics(in MetricsIn) *otkafka.ReaderStats {
	labels := []string{"reader", "client_id", "topic", "partition"}

	if in.Registerer == nil {
		in.Registerer = stdprometheus.DefaultRegisterer
	}

	return &otkafka.ReaderStats{
		Dials: newCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_reader_dial_count",
			Help: "",
		}, labels, in.Registerer),
		Fetches: newCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_reader_fetch_count",
			Help: "",
		}, labels, in.Registerer),
		Messages: newCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_reader_message_count",
			Help: "",
		}, labels, in.Registerer),
		Bytes: newCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_reader_message_bytes",
			Help: "",
		}, labels, in.Registerer),
		Rebalances: newCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_reader_rebalance_count",
			Help: "",
		}, labels, in.Registerer),
		Timeouts: newCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_reader_timeout_count",
			Help: "",
		}, labels, in.Registerer),
		Errors: newCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_reader_error_count",
			Help: "",
		}, labels, in.Registerer),
		Offset: newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_reader_offset",
			Help: "",
		}, labels, in.Registerer),
		Lag: newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_reader_lag",
			Help: "",
		}, labels, in.Registerer),
		MinBytes: newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_reader_bytes_min",
			Help: "",
		}, labels, in.Registerer),
		MaxBytes: newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_reader_bytes_max",
			Help: "",
		}, labels, in.Registerer),
		MaxWait: newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_reader_fetch_wait_max",
			Help: "",
		}, labels, in.Registerer),
		QueueLength: newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_reader_queue_length",
			Help: "",
		}, labels, in.Registerer),
		QueueCapacity: newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_reader_queue_capacity",
			Help: "",
		}, labels, in.Registerer),
		DialTime: otkafka.AggStats{
			Min: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_dial_seconds_min",
				Help: "",
			}, labels, in.Registerer),
			Max: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_dial_seconds_max",
				Help: "",
			}, labels, in.Registerer),
			Avg: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_dial_seconds_avg",
				Help: "",
			}, labels, in.Registerer),
		},
		ReadTime: otkafka.AggStats{
			Min: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_read_seconds_min",
				Help: "",
			}, labels, in.Registerer),
			Max: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_read_seconds_max",
				Help: "",
			}, labels, in.Registerer),
			Avg: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_read_seconds_avg",
				Help: "",
			}, labels, in.Registerer),
		},
		WaitTime: otkafka.AggStats{
			Min: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_wait_seconds_min",
				Help: "",
			}, labels, in.Registerer),
			Max: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_wait_seconds_max",
				Help: "",
			}, labels, in.Registerer),
			Avg: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_wait_seconds_avg",
				Help: "",
			}, labels, in.Registerer),
		},
		FetchSize: otkafka.AggStats{
			Min: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_fetch_size_min",
				Help: "",
			}, labels, in.Registerer),
			Max: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_fetch_size_max",
				Help: "",
			}, labels, in.Registerer),
			Avg: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_fetch_size_avg",
				Help: "",
			}, labels, in.Registerer),
		},
		FetchBytes: otkafka.AggStats{
			Min: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_fetch_bytes_min",
				Help: "",
			}, labels, in.Registerer),
			Max: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_fetch_bytes_max",
				Help: "",
			}, labels, in.Registerer),
			Avg: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_reader_fetch_bytes_avg",
				Help: "",
			}, labels, in.Registerer),
		},
	}
}

// ProvideKafkaWriterMetrics returns a *otkafka.WriterStats that measures the writer info in kafka.
// It is meant to be consumed by the otkafka.Providers.
func ProvideKafkaWriterMetrics(in MetricsIn) *otkafka.WriterStats {
	labels := []string{"writer", "topic"}

	if in.Registerer == nil {
		in.Registerer = stdprometheus.DefaultRegisterer
	}

	return &otkafka.WriterStats{
		Writes: newCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_writer_write_count",
			Help: "",
		}, labels, in.Registerer),
		Messages: newCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_writer_message_count",
			Help: "",
		}, labels, in.Registerer),
		Bytes: newCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_writer_message_bytes",
			Help: "",
		}, labels, in.Registerer),
		Errors: newCounterFrom(stdprometheus.CounterOpts{
			Name: "kafka_writer_error_count",
			Help: "",
		}, labels, in.Registerer),
		MaxAttempts: newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_writer_attempts_max",
			Help: "",
		}, labels, in.Registerer),
		MaxBatchSize: newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_writer_batch_max",
			Help: "",
		}, labels, in.Registerer),
		BatchTimeout: newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_writer_batch_timeout",
			Help: "",
		}, labels, in.Registerer),
		ReadTimeout: newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_writer_read_timeout",
			Help: "",
		}, labels, in.Registerer),
		WriteTimeout: newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_writer_write_timeout",
			Help: "",
		}, labels, in.Registerer),
		RequiredAcks: newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_writer_acks_required",
			Help: "",
		}, labels, in.Registerer),
		Async: newGaugeFrom(stdprometheus.GaugeOpts{
			Name: "kafka_writer_async",
			Help: "",
		}, labels, in.Registerer),
		BatchTime: otkafka.AggStats{
			Min: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_batch_seconds_min",
				Help: "",
			}, labels, in.Registerer),
			Max: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_batch_seconds_max",
				Help: "",
			}, labels, in.Registerer),
			Avg: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_batch_seconds_avg",
				Help: "",
			}, labels, in.Registerer),
		},
		WriteTime: otkafka.AggStats{
			Min: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_write_seconds_min",
				Help: "",
			}, labels, in.Registerer),
			Max: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_write_seconds_max",
				Help: "",
			}, labels, in.Registerer),
			Avg: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_write_seconds_avg",
				Help: "",
			}, labels, in.Registerer),
		},
		WaitTime: otkafka.AggStats{
			Min: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_wait_seconds_min",
				Help: "",
			}, labels, in.Registerer),
			Max: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_wait_seconds_max",
				Help: "",
			}, labels, in.Registerer),
			Avg: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_wait_seconds_avg",
				Help: "",
			}, labels, in.Registerer),
		},
		Retries: otkafka.AggStats{
			Min: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_retries_count_min",
				Help: "",
			}, labels, in.Registerer),
			Max: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_retries_count_max",
				Help: "",
			}, labels, in.Registerer),
			Avg: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_retries_count_avg",
				Help: "",
			}, labels, in.Registerer),
		},
		BatchSize: otkafka.AggStats{
			Min: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_batch_size_min",
				Help: "",
			}, labels, in.Registerer),
			Max: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_batch_size_max",
				Help: "",
			}, labels, in.Registerer),
			Avg: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_batch_size_avg",
				Help: "",
			}, labels, in.Registerer),
		},
		BatchBytes: otkafka.AggStats{
			Min: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_batch_bytes_min",
				Help: "",
			}, labels, in.Registerer),
			Max: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_batch_bytes_max",
				Help: "",
			}, labels, in.Registerer),
			Avg: newGaugeFrom(stdprometheus.GaugeOpts{
				Name: "kafka_writer_batch_bytes_avg",
				Help: "",
			}, labels, in.Registerer),
		},
	}
}

func newCounterFrom(opts stdprometheus.CounterOpts, labelNames []string, registerer stdprometheus.Registerer) metrics.Counter {
	cv := stdprometheus.NewCounterVec(opts, labelNames)
	registerer.MustRegister(cv)
	return prometheus.NewCounter(cv)
}

func newGaugeFrom(opts stdprometheus.GaugeOpts, labelNames []string, registerer stdprometheus.Registerer) metrics.Gauge {
	cv := stdprometheus.NewGaugeVec(opts, labelNames)
	registerer.MustRegister(cv)
	return prometheus.NewGauge(cv)
}
