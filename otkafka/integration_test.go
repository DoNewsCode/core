package otkafka

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/di"
	mock_metrics "github.com/DoNewsCode/core/otkafka/mocks"
	"github.com/golang/mock/gomock"
	knoaf_json "github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/oklog/run"
	"github.com/stretchr/testify/assert"
)

func TestFactoryOut_ProvideRunGroup(t *testing.T) {
	if os.Getenv("KAFKA_ADDR") == "" {
		t.Skip("set KAFKA_ADDR to run TestFactoryOut_ProvideRunGroup")
		return
	}
	addrs := strings.Split(os.Getenv("KAFKA_ADDR"), ",")
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock_metrics.NewMockGauge(ctrl)
	m.EXPECT().With(gomock.Any()).Return(m).MinTimes(1)
	m.EXPECT().Set(gomock.Any()).MinTimes(1)
	m.EXPECT().Add(gomock.Any()).AnyTimes()

	mc := mock_metrics.NewMockCounter(ctrl)
	mc.EXPECT().With(gomock.Any()).Return(mc).MinTimes(1)
	mc.EXPECT().Add(gomock.Any()).MinTimes(1)

	c := core.New(
		core.WithInline("kafka.writer.default.brokers", addrs),
		core.WithInline("kafka.reader.default.brokers", addrs),
		core.WithInline("kafka.writer.default.topic", "test"),
		core.WithInline("kafka.reader.default.topic", "test"),
		core.WithInline("kafkaMetrics.interval", "10ms"),
		core.WithInline("log.level", "none"),
	)
	c.ProvideEssentials()
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
			DialTime: AggStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			ReadTime: AggStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			WaitTime: AggStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			FetchSize: AggStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			FetchBytes: AggStats{
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
			BatchTime: AggStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			WriteTime: AggStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			WaitTime: AggStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			Retries: AggStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			BatchSize: AggStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			BatchBytes: AggStats{
				Min: m,
				Max: m,
				Avg: m,
			},
		}
	}})
	c.Provide(Providers())
	c.Invoke(func(reader *ReaderFactory, writer *WriterFactory) {})
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	c.Serve(ctx)
}

func TestCollector(t *testing.T) {
	if os.Getenv("KAFKA_ADDR") == "" {
		t.Skip("set KAFKA_ADDR to run TestCollector")
		return
	}
	addrs := strings.Split(os.Getenv("KAFKA_ADDR"), ",")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock_metrics.NewMockGauge(ctrl)
	m.EXPECT().With(gomock.Any()).Return(m).MinTimes(1)
	m.EXPECT().Set(gomock.Any()).MinTimes(1)
	m.EXPECT().Add(gomock.Any()).AnyTimes()

	mc := mock_metrics.NewMockCounter(ctrl)
	mc.EXPECT().With(gomock.Any()).Return(mc).MinTimes(1)
	mc.EXPECT().Add(gomock.Any()).MinTimes(1)

	c := core.New(
		core.WithInline("kafka.writer.default.brokers", addrs),
		core.WithInline("kafka.reader.default.brokers", addrs),
		core.WithInline("kafka.reader.default.topic", "test"),
		core.WithInline("kafkaMetrics.interval", "1ms"),
		core.WithInline("log.level", "none"),
	)
	c.ProvideEssentials()
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
			DialTime: AggStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			ReadTime: AggStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			WaitTime: AggStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			FetchSize: AggStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			FetchBytes: AggStats{
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
			BatchTime: AggStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			WriteTime: AggStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			WaitTime: AggStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			Retries: AggStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			BatchSize: AggStats{
				Min: m,
				Max: m,
				Avg: m,
			},
			BatchBytes: AggStats{
				Min: m,
				Max: m,
				Avg: m,
			},
		}
	}})
	c.Provide(Providers())

	c.Invoke(func(rf *ReaderFactory, s *ReaderStats) {
		rc := newReaderCollector(rf, s, time.Nanosecond)
		rc.collectConnectionStats()
	})

	c.Invoke(func(wf *WriterFactory, s *WriterStats) {
		wc := newWriterCollector(wf, s, time.Nanosecond)
		wc.collectConnectionStats()
	})
}

type channelWatcher struct {
	ch          chan struct{}
	afterReload chan struct{}
}

func (c *channelWatcher) Watch(ctx context.Context, reload func() error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.ch:
			reload()
			c.afterReload <- struct{}{}
		}
	}
}

func TestModule_hotReload(t *testing.T) {
	if os.Getenv("KAFKA_ADDR") == "" {
		t.Skip("set KAFKA_ADDR to run TestModule_ProvideRunGroup")
		return
	}
	addrs := strings.Split(os.Getenv("KAFKA_ADDR"), ",")
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cw := &channelWatcher{}
	cw.ch = make(chan struct{})
	cw.afterReload = make(chan struct{})

	conf := map[string]interface{}{
		"http": map[string]bool{
			"disable": true,
		},
		"grpc": map[string]bool{
			"disable": true,
		},
		"cron": map[string]bool{
			"disable": true,
		},
		"kafka": map[string]interface{}{
			"reader": map[string]interface{}{
				"default": map[string]interface{}{
					"brokers": addrs,
					"topic":   "foo",
				},
			},
			"writer": map[string]interface{}{
				"default": map[string]interface{}{
					"brokers": addrs,
					"topic":   "foo",
				},
			},
		},
	}
	path := createFile(conf)
	c := core.Default(core.WithConfigStack(file.Provider(path), knoaf_json.Parser()), core.WithConfigWatcher(cw))
	c.Provide(Providers(WithReaderReload(true), WithWriterReload(true)))
	c.AddModuleFunc(config.New)

	var group run.Group
	for _, m := range c.Modules() {
		if p, ok := m.(core.RunProvider); ok {
			p.ProvideRunGroup(&group)
		}
	}
	group.Add(func() error {
		<-ctx.Done()
		return ctx.Err()
	}, func(err error) {
		cancel()
	})
	go group.Run()

	// Test initial value
	c.Invoke(func(f *ReaderFactory) {
		reader, err := f.Make("default")
		assert.NoError(t, err)
		assert.Equal(t, "foo", reader.Config().Topic)
	})
	c.Invoke(func(f *WriterFactory) {
		writer, err := f.Make("default")
		assert.NoError(t, err)
		assert.Equal(t, "foo", writer.Topic)
	})

	// Close config
	conf["kafka"].(map[string]interface{})["writer"].(map[string]interface{})["default"].(map[string]interface{})["topic"] = "bar"
	conf["kafka"].(map[string]interface{})["reader"].(map[string]interface{})["default"].(map[string]interface{})["topic"] = "bar"
	overwriteFile(path, conf)
	cw.ch <- struct{}{}
	<-cw.afterReload

	// Test reloaded values
	c.Invoke(func(f *ReaderFactory) {
		reader, err := f.Make("default")
		assert.NoError(t, err)
		assert.Equal(t, "bar", reader.Config().Topic)
	})
	c.Invoke(func(f *WriterFactory) {
		writer, err := f.Make("default")
		assert.NoError(t, err)
		assert.Equal(t, "bar", writer.Topic)
	})
}

func createFile(content map[string]interface{}) string {
	f, _ := ioutil.TempFile("", "*")
	data, _ := json.Marshal(content)
	ioutil.WriteFile(f.Name(), data, os.ModePerm)
	return f.Name()
}

func overwriteFile(path string, content map[string]interface{}) {
	data, _ := json.Marshal(content)
	ioutil.WriteFile(path, data, os.ModePerm)
}
