package observability

import (
	"os"
	"strings"
	"testing"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/otgorm"
	"github.com/DoNewsCode/core/otkafka"
	"github.com/DoNewsCode/core/otredis"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v8"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestProvideOpentracing(t *testing.T) {
	conf, _ := config.NewConfig(config.WithProviderLayer(rawbytes.Provider([]byte(sample)), yaml.Parser()))
	Out, cleanup, err := ProvideOpentracing(
		config.AppName("foo"),
		config.EnvTesting,
		ProvideJaegerLogAdapter(log.NewNopLogger()),
		conf,
	)
	assert.NoError(t, err)
	assert.NotNil(t, Out)
	cleanup()
}

func TestProvideHistogramMetrics(t *testing.T) {
	Out := ProvideHistogramMetrics(
		config.AppName("foo"),
		config.EnvTesting,
	)
	assert.NotNil(t, Out)
}

func TestProvideGORMMetrics(t *testing.T) {
	c := core.New()
	c.ProvideEssentials()
	c.Provide(Providers())
	c.Provide(otgorm.Providers())
	c.Invoke(func(db *gorm.DB, g *otgorm.Gauges) {
		d, err := db.DB()
		if err != nil {
			t.Error(err)
		}
		stats := d.Stats()
		withValues := []string{"dbname", "default", "driver", db.Name()}

		g.Idle.With(withValues...).Set(float64(stats.Idle))
		g.InUse.With(withValues...).Set(float64(stats.InUse))
		g.Open.With(withValues...).Set(float64(stats.OpenConnections))
	})
}

func TestProvideRedisMetrics(t *testing.T) {
	c := core.New()
	c.ProvideEssentials()
	c.Provide(Providers())
	c.Provide(otredis.Providers())
	c.Invoke(func(cli redis.UniversalClient, g *otredis.Gauges) {
		stats := cli.PoolStats()
		withValues := []string{"dbname", "default"}

		g.Hits.With(withValues...).Set(float64(stats.Hits))
		g.Misses.With(withValues...).Set(float64(stats.Misses))
		g.Timeouts.With(withValues...).Set(float64(stats.Timeouts))
		g.TotalConns.With(withValues...).Set(float64(stats.TotalConns))
		g.IdleConns.With(withValues...).Set(float64(stats.IdleConns))
		g.StaleConns.With(withValues...).Set(float64(stats.StaleConns))
	})
}

func TestProvideKafkaMetrics(t *testing.T) {
	addr := os.Getenv("KAFKA_ADDR")
	if addr == "" {
		t.Skip("set KAFKA_ADDR for run kafka metrics test")
	}
	addrs := strings.Split(addr, ",")
	c := core.New(
		core.WithInline("kafka.writer.default.brokers", addrs),
		core.WithInline("kafka.reader.default.brokers", addrs),
		core.WithInline("kafka.reader.default.topic", "test"),
		core.WithInline("log.level", "none"),
	)
	c.ProvideEssentials()
	c.Provide(Providers())
	c.Provide(otkafka.Providers())
	c.Invoke(func(w *kafka.Writer, ws *otkafka.WriterStats) {
		stats := w.Stats()
		withValues := []string{"writer", "default", "topic", stats.Topic}

		ws.Writes.With(withValues...).Add(float64(stats.Writes))
		ws.Messages.With(withValues...).Add(float64(stats.Messages))
		ws.Bytes.With(withValues...).Add(float64(stats.Bytes))
		ws.Errors.With(withValues...).Add(float64(stats.Errors))

		ws.BatchTime.Min.With(withValues...).Add(stats.BatchTime.Min.Seconds())
		ws.BatchTime.Max.With(withValues...).Add(stats.BatchTime.Max.Seconds())
		ws.BatchTime.Avg.With(withValues...).Add(stats.BatchTime.Avg.Seconds())

		ws.WriteTime.Min.With(withValues...).Add(stats.WriteTime.Min.Seconds())
		ws.WriteTime.Max.With(withValues...).Add(stats.WriteTime.Max.Seconds())
		ws.WriteTime.Avg.With(withValues...).Add(stats.WriteTime.Avg.Seconds())

		ws.WaitTime.Min.With(withValues...).Add(stats.WaitTime.Min.Seconds())
		ws.WaitTime.Max.With(withValues...).Add(stats.WaitTime.Max.Seconds())
		ws.WaitTime.Avg.With(withValues...).Add(stats.WaitTime.Avg.Seconds())

		ws.Retries.Min.With(withValues...).Add(float64(stats.Retries.Min))
		ws.Retries.Max.With(withValues...).Add(float64(stats.Retries.Max))
		ws.Retries.Avg.With(withValues...).Add(float64(stats.Retries.Avg))

		ws.BatchSize.Min.With(withValues...).Add(float64(stats.BatchSize.Min))
		ws.BatchSize.Max.With(withValues...).Add(float64(stats.BatchSize.Max))
		ws.BatchSize.Avg.With(withValues...).Add(float64(stats.BatchSize.Avg))

		ws.BatchBytes.Min.With(withValues...).Add(float64(stats.BatchBytes.Min))
		ws.BatchBytes.Max.With(withValues...).Add(float64(stats.BatchBytes.Max))
		ws.BatchBytes.Avg.With(withValues...).Add(float64(stats.BatchBytes.Avg))

		ws.MaxAttempts.With(withValues...).Set(float64(stats.MaxAttempts))
		ws.MaxBatchSize.With(withValues...).Set(float64(stats.MaxBatchSize))
		ws.BatchTimeout.With(withValues...).Set(stats.BatchTimeout.Seconds())
		ws.ReadTimeout.With(withValues...).Set(stats.ReadTimeout.Seconds())
		ws.WriteTimeout.With(withValues...).Set(stats.WriteTimeout.Seconds())
		ws.RequiredAcks.With(withValues...).Set(float64(stats.RequiredAcks))
		var async float64
		if stats.Async {
			async = 1.0
		}
		ws.Async.With(withValues...).Set(async)
	})

	c.Invoke(func(r *kafka.Reader, rs *otkafka.ReaderStats) {
		stats := r.Stats()
		withValues := []string{
			"reader", "default",
			"client_id", stats.ClientID,
			"topic", stats.Topic,
			"partition", stats.Partition,
		}

		rs.Dials.With(withValues...).Add(float64(stats.Dials))
		rs.Fetches.With(withValues...).Add(float64(stats.Fetches))
		rs.Messages.With(withValues...).Add(float64(stats.Messages))
		rs.Bytes.With(withValues...).Add(float64(stats.Bytes))
		rs.Rebalances.With(withValues...).Add(float64(stats.Rebalances))
		rs.Timeouts.With(withValues...).Add(float64(stats.Timeouts))
		rs.Errors.With(withValues...).Add(float64(stats.Errors))

		rs.Offset.With(withValues...).Set(float64(stats.Offset))
		rs.Lag.With(withValues...).Set(float64(stats.Lag))
		rs.MinBytes.With(withValues...).Set(float64(stats.MinBytes))
		rs.MaxBytes.With(withValues...).Set(float64(stats.MaxBytes))
		rs.MaxWait.With(withValues...).Set(stats.MaxWait.Seconds())
		rs.QueueLength.With(withValues...).Set(float64(stats.QueueLength))
		rs.QueueCapacity.With(withValues...).Set(float64(stats.QueueCapacity))

		rs.DialTime.Min.With(withValues...).Set(stats.DialTime.Min.Seconds())
		rs.DialTime.Max.With(withValues...).Set(stats.DialTime.Max.Seconds())
		rs.DialTime.Avg.With(withValues...).Set(stats.DialTime.Avg.Seconds())

		rs.ReadTime.Min.With(withValues...).Set(stats.ReadTime.Min.Seconds())
		rs.ReadTime.Max.With(withValues...).Set(stats.ReadTime.Max.Seconds())
		rs.ReadTime.Avg.With(withValues...).Set(stats.ReadTime.Avg.Seconds())

		rs.WaitTime.Min.With(withValues...).Set(stats.WaitTime.Min.Seconds())
		rs.WaitTime.Max.With(withValues...).Set(stats.WaitTime.Max.Seconds())
		rs.WaitTime.Avg.With(withValues...).Set(stats.WaitTime.Avg.Seconds())

		rs.FetchSize.Min.With(withValues...).Set(float64(stats.FetchSize.Min))
		rs.FetchSize.Max.With(withValues...).Set(float64(stats.FetchSize.Max))
		rs.FetchSize.Avg.With(withValues...).Set(float64(stats.FetchSize.Avg))

		rs.FetchBytes.Min.With(withValues...).Set(float64(stats.FetchBytes.Min))
		rs.FetchBytes.Max.With(withValues...).Set(float64(stats.FetchBytes.Max))
		rs.FetchBytes.Avg.With(withValues...).Set(float64(stats.FetchBytes.Avg))
	})
}

func Test_provideConfig(t *testing.T) {
	Conf := provideConfig()
	assert.NotEmpty(t, Conf.Config)
}
