package otkafka

import (
	"context"
	"fmt"
	"time"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/oklog/run"
	"github.com/opentracing/opentracing-go"
	"github.com/segmentio/kafka-go"
)

/*
Providers is a set of dependencies including ReaderMaker, WriterMaker and exported configs.
	Depends On:
		contract.ConfigAccessor
		log.Logger
	Provide:
		ReaderFactory
		WriterFactory
		ReaderMaker
		WriterMaker
		*kafka.Reader
		*kafka.Writer
		*readerCollector
		*writerCollector
*/
func Providers(optionFunc ...ProvidersOptionFunc) di.Deps {
	option := providersOption{
		readerInterceptor: func(name string, reader *kafka.ReaderConfig) {},
		writerInterceptor: func(name string, writer *kafka.Writer) {},
	}
	for _, f := range optionFunc {
		f(&option)
	}
	return di.Deps{
		provideKafkaFactory(&option),
		provideConfig,
		di.Bind(new(WriterFactory), new(WriterMaker)),
		di.Bind(new(ReaderFactory), new(ReaderMaker)),
	}
}

// WriterMaker models a WriterFactory
type WriterMaker interface {
	Make(name string) (*kafka.Writer, error)
}

// ReaderMaker models a ReaderFactory
type ReaderMaker interface {
	Make(name string) (*kafka.Reader, error)
}

// factoryIn is a injection parameter for provideReaderFactory.
type factoryIn struct {
	di.In

	Tracer      opentracing.Tracer `optional:"true"`
	Conf        contract.ConfigUnmarshaler
	Logger      log.Logger
	ReaderStats *ReaderStats        `optional:"true"`
	WriterStats *WriterStats        `optional:"true"`
	Dispatcher  contract.Dispatcher `optional:"true"`
}

// factoryOut is the result of provideKafkaFactory.
type factoryOut struct {
	di.Out

	ReaderFactory   ReaderFactory
	WriterFactory   WriterFactory
	Reader          *kafka.Reader
	Writer          *kafka.Writer
	ReaderCollector *readerCollector
	WriterCollector *writerCollector
}

// Module implements di.Modular
func (f factoryOut) Module() interface{} {
	return f
}

// ProvideRunGroup add a goroutine to periodically scan kafka's reader&writer info and
// report them to metrics collector such as prometheus.
func (f factoryOut) ProvideRunGroup(group *run.Group) {
	if f.ReaderCollector != nil {
		ctx, cancel := context.WithCancel(context.Background())
		ticker := time.NewTicker(f.ReaderCollector.interval)
		group.Add(func() error {
			for {
				select {
				case <-ticker.C:
					f.ReaderCollector.collectConnectionStats()
				case <-ctx.Done():
					ticker.Stop()
					return nil
				}
			}
		}, func(err error) {
			cancel()
		})
	}
	if f.WriterCollector != nil {
		ctx, cancel := context.WithCancel(context.Background())
		ticker := time.NewTicker(f.WriterCollector.interval)
		group.Add(func() error {
			for {
				select {
				case <-ticker.C:
					f.WriterCollector.collectConnectionStats()
				case <-ctx.Done():
					ticker.Stop()
					return nil
				}
			}
		}, func(err error) {
			cancel()
		})
	}
}

// provideKafkaFactory creates the ReaderFactory and WriterFactory. It is
// valid dependency option for package core. Note: when working with package
// core's DI container, use provideKafkaFactory over provideReaderFactory and
// provideWriterFactory. Not only provideKafkaFactory provides both reader and
// writer, but also only provideKafkaFactory exports default Kafka configuration.
func provideKafkaFactory(option *providersOption) func(p factoryIn) (factoryOut, func(), func(), error) {
	if option.readerInterceptor == nil {
		option.readerInterceptor = func(name string, reader *kafka.ReaderConfig) {}
	}
	if option.writerInterceptor == nil {
		option.writerInterceptor = func(name string, writer *kafka.Writer) {}
	}
	return func(p factoryIn) (factoryOut, func(), func(), error) {
		var readerCollector *readerCollector
		var writerCollector *writerCollector
		rf, rc := provideReaderFactory(p, option.readerInterceptor)
		wf, wc := provideWriterFactory(p, option.writerInterceptor)
		dr, err1 := rf.Make("default")
		if err1 != nil {
			level.Warn(p.Logger).Log("err", err1)
		}
		dw, err2 := wf.Make("default")
		if err2 != nil {
			level.Warn(p.Logger).Log("err", err2)
		}

		if p.ReaderStats != nil || p.WriterStats != nil {
			var interval time.Duration
			p.Conf.Unmarshal("kafkaMetrics.interval", &interval)
			if p.ReaderStats != nil {
				readerCollector = newReaderCollector(rf, p.ReaderStats, interval)
			}
			if p.WriterStats != nil {
				writerCollector = newWriterCollector(wf, p.WriterStats, interval)
			}
		}

		if p.Dispatcher != nil {
			rf.SubscribeReloadEventFrom(p.Dispatcher)
			wf.SubscribeReloadEventFrom(p.Dispatcher)
		}

		return factoryOut{
			ReaderFactory:   rf,
			WriterFactory:   wf,
			Reader:          dr,
			Writer:          dw,
			ReaderCollector: readerCollector,
			WriterCollector: writerCollector,
		}, wc, rc, nil
	}
}

// provideReaderFactory creates the ReaderFactory. It is valid
// dependency option for package core.
func provideReaderFactory(p factoryIn, interceptor ReaderInterceptor) (ReaderFactory, func()) {
	factory := di.NewFactory(func(name string) (di.Pair, error) {
		var (
			err          error
			readerConfig ReaderConfig
		)
		err = p.Conf.Unmarshal(fmt.Sprintf("kafka.reader.%s", name), &readerConfig)
		if err != nil {
			return di.Pair{}, fmt.Errorf("kafka reader configuration %s not valid: %w", name, err)
		}

		// converts to the kafka.ReaderConfig from github.com/segmentio/kafka-go
		conf := fromReaderConfig(readerConfig)
		conf.Logger = KafkaLogAdapter{Logging: level.Debug(p.Logger)}
		conf.ErrorLogger = KafkaLogAdapter{Logging: level.Warn(p.Logger)}
		interceptor(name, &conf)
		client := kafka.NewReader(conf)
		return di.Pair{
			Conn: client,
			Closer: func() {
				_ = client.Close()
			},
		}, nil
	})
	return ReaderFactory{factory}, factory.Close
}

// provideWriterFactory creates WriterFactory. It is a valid injection
// option for package core.
func provideWriterFactory(p factoryIn, interceptor WriterInterceptor) (WriterFactory, func()) {

	factory := di.NewFactory(func(name string) (di.Pair, error) {
		var (
			err          error
			writerConfig WriterConfig
		)
		err = p.Conf.Unmarshal(fmt.Sprintf("kafka.writer.%s", name), &writerConfig)
		if err != nil {
			return di.Pair{}, fmt.Errorf("kafka writer configuration %s not valid: %w", name, err)
		}
		writer := fromWriterConfig(writerConfig)
		logger := log.With(p.Logger, "tag", "kafka")
		writer.Logger = KafkaLogAdapter{Logging: level.Debug(logger)}
		writer.ErrorLogger = KafkaLogAdapter{Logging: level.Warn(logger)}
		writer.Transport = NewTransport(kafka.DefaultTransport, p.Tracer)
		interceptor(name, &writer)

		return di.Pair{
			Conn: &writer,
			Closer: func() {
				_ = writer.Close()
			},
		}, nil
	})
	return WriterFactory{factory}, factory.Close
}

type metricsConf struct {
	Interval config.Duration `json:"interval" yaml:"interval"`
}

type configOut struct {
	di.Out

	Config []config.ExportedConfig `group:"config,flatten"`
}

func provideConfig() configOut {
	configs := []config.ExportedConfig{
		{
			Owner: "kitkafka",
			Data: map[string]interface{}{
				"kafka": map[string]interface{}{
					"reader": map[string]interface{}{
						"default": ReaderConfig{
							Brokers: []string{"127.0.0.1:9092"},
						},
					},
					"writer": map[string]interface{}{
						"default": WriterConfig{
							Brokers: []string{"127.0.0.1:9092"},
						},
					},
				},
				"kafkaMetrics": metricsConf{
					Interval: config.Duration{Duration: 15 * time.Second},
				},
			},
			Comment: "",
		},
	}
	return configOut{Config: configs}
}
