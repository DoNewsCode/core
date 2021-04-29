package otkafka

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/di"
	mock_metrics "github.com/DoNewsCode/core/otkafka/mocks"
	"github.com/golang/mock/gomock"
)

func TestModule_ProvideRunGroup(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock_metrics.NewMockGauge(ctrl)
	m.EXPECT().With(gomock.Any()).Return(m).MinTimes(1)
	m.EXPECT().Set(gomock.Any()).MinTimes(1)

	mc := mock_metrics.NewMockCounter(ctrl)
	mc.EXPECT().With(gomock.Any()).Return(mc).MinTimes(1)
	mc.EXPECT().Add(gomock.Any()).MinTimes(1)

	c := core.New(
		core.WithInline("kafka.writer.default.brokers", []string{os.Getenv("KAFKA_ADDR")}),
		core.WithInline("kafka.reader.default.brokers", []string{os.Getenv("KAFKA_ADDR")}),
		core.WithInline("kafka.reader.default.topic", "test"),
		core.WithInline("kafkaMetrics.interval", "1ms"),
		core.WithInline("log.level", "none"),
		core.WithInline("http.disable", "true"),
		core.WithInline("grpc.disable", "true"),
		core.WithInline("cron.disable", "true"),
	)
	c.ProvideEssentials()
	c.Provide(Providers())
	c.AddModuleFunc(New)
	c.Provide(di.Deps{func() *ReaderStats {
		return &ReaderStats{
			Dials:         mc,
			Fetches:       mc,
			Messages:      mc,
			Bytes:         mc,
			Rebalances:    mc,
			Timeouts:      mc,
			Errors:        mc,
			Offset:        m,
			Lag:           m,
			MinBytes:      m,
			MaxBytes:      m,
			MaxWait:       m,
			QueueLength:   m,
			QueueCapacity: m,
			DialTime: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			ReadTime: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			WaitTime: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			FetchSize: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			FetchBytes: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
		}
	}})
	c.Provide(di.Deps{func() *WriterStats {
		return &WriterStats{
			Writes:       mc,
			Messages:     mc,
			Bytes:        mc,
			Errors:       mc,
			MaxAttempts:  m,
			MaxBatchSize: m,
			BatchTimeout: m,
			ReadTimeout:  m,
			WriteTimeout: m,
			RequiredAcks: m,
			Async:        m,
			BatchTime: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			WriteTime: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			WaitTime: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			Retries: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			BatchSize: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			BatchBytes: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
		}
	}})

	ctx, cancel := context.WithCancel(context.Background())

	go c.Serve(ctx)
	<-time.After(100 * time.Millisecond)
	cancel()
	<-time.After(1000 * time.Millisecond)

}

func TestCollector(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock_metrics.NewMockGauge(ctrl)
	m.EXPECT().With(gomock.Any()).Return(m).MinTimes(1)
	m.EXPECT().Set(gomock.Any()).MinTimes(1)

	mc := mock_metrics.NewMockCounter(ctrl)
	mc.EXPECT().With(gomock.Any()).Return(mc).MinTimes(1)
	mc.EXPECT().Add(gomock.Any()).MinTimes(1)

	c := core.New(
		core.WithInline("kafka.writer.default.brokers", []string{os.Getenv("KAFKA_ADDR")}),
		core.WithInline("kafka.reader.default.brokers", []string{os.Getenv("KAFKA_ADDR")}),
		core.WithInline("kafka.reader.default.topic", "test"),
		core.WithInline("kafkaMetrics.interval", "1ms"),
		core.WithInline("log.level", "none"),
		core.WithInline("http.disable", "true"),
		core.WithInline("grpc.disable", "true"),
		core.WithInline("cron.disable", "true"),
	)
	c.ProvideEssentials()
	c.Provide(Providers())
	c.Provide(di.Deps{func() *ReaderStats {
		return &ReaderStats{
			Dials:         mc,
			Fetches:       mc,
			Messages:      mc,
			Bytes:         mc,
			Rebalances:    mc,
			Timeouts:      mc,
			Errors:        mc,
			Offset:        m,
			Lag:           m,
			MinBytes:      m,
			MaxBytes:      m,
			MaxWait:       m,
			QueueLength:   m,
			QueueCapacity: m,
			DialTime: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			ReadTime: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			WaitTime: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			FetchSize: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			FetchBytes: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
		}
	}})
	c.Provide(di.Deps{func() *WriterStats {
		return &WriterStats{
			Writes:       mc,
			Messages:     mc,
			Bytes:        mc,
			Errors:       mc,
			MaxAttempts:  m,
			MaxBatchSize: m,
			BatchTimeout: m,
			ReadTimeout:  m,
			WriteTimeout: m,
			RequiredAcks: m,
			Async:        m,
			BatchTime: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			WriteTime: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			WaitTime: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			Retries: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			BatchSize: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			BatchBytes: ThreeStats{
				Min: m,
				Max: m,
				Avg: m,
			},
		}
	}})

	c.Invoke(func(rf ReaderFactory, s *ReaderStats) {
		rc := newReaderCollector(rf, s, time.Nanosecond)
		rc.collectConnectionStats()

	})

	c.Invoke(func(wf WriterFactory, s *WriterStats) {
		wc := newWriterCollector(wf, s, time.Nanosecond)
		wc.collectConnectionStats()
	})
}
