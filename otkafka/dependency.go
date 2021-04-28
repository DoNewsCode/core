package otkafka

import (
	"fmt"
	"time"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/opentracing/opentracing-go"
	"github.com/segmentio/kafka-go"
)

/*
Providers is a set of dependencies including ReaderMaker, WriterMaker and exported configs.
	Depends On:
		ReaderInterceptor `optional:"true"`
		WriterInterceptor `optional:"true"`
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
func Providers() []interface{} {
	return []interface{}{provideKafkaFactory,  provideConfig}
}

// WriterMaker models a WriterFactory
type WriterMaker interface {
	Make(name string) (*kafka.Writer, error)
}

// ReaderMaker models a ReaderFactory
type ReaderMaker interface {
	Make(name string) (*kafka.Reader, error)
}

// in is a injection parameter for provideReaderFactory.
type in struct {
	di.In

	ReaderInterceptor ReaderInterceptor  `optional:"true"`
	WriterInterceptor WriterInterceptor  `optional:"true"`
	Tracer            opentracing.Tracer `optional:"true"`
	Conf              contract.ConfigAccessor
	Logger            log.Logger
	ReaderStats       *ReaderStats `optional:"true"`
	WriterStats       *WriterStats `optional:"true"`
}

// out is the result of provideKafkaFactory.
type out struct {
	di.Out

	ReaderFactory   ReaderFactory
	WriterFactory   WriterFactory
	ReaderMaker     ReaderMaker
	WriterMaker     WriterMaker
	Reader          *kafka.Reader
	Writer          *kafka.Writer
	ReaderCollector *readerCollector
	WriterCollector *writerCollector
}

// provideKafkaFactory creates the ReaderFactory and WriterFactory. It is
// valid dependency option for package core. Note: when working with package
// core's DI container, use provideKafkaFactory over provideReaderFactory and
// provideWriterFactory. Not only provideKafkaFactory provides both reader and
// writer, but also only provideKafkaFactory exports default Kafka configuration.
func provideKafkaFactory(p in) (out, func(), func(), error) {
	rf, rc := provideReaderFactory(p)
	wf, wc := provideWriterFactory(p)
	dr, err1 := rf.Make("default")
	if err1 != nil {
		level.Warn(p.Logger).Log("err", err1)
	}
	dw, err2 := wf.Make("default")
	if err2 != nil {
		level.Warn(p.Logger).Log("err", err2)
	}
	var readerCollector *readerCollector
	var writerCollector *writerCollector
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

	return out{
		ReaderMaker:     rf,
		ReaderFactory:   rf,
		WriterMaker:     wf,
		WriterFactory:   wf,
		Reader:          dr,
		Writer:          dw,
		ReaderCollector: readerCollector,
		WriterCollector: writerCollector,
	}, wc, rc, nil
}

// provideReaderFactory creates the ReaderFactory. It is valid
// dependency option for package core.
func provideReaderFactory(p in) (ReaderFactory, func()) {
	var err error
	var dbConfs map[string]ReaderConfig
	err = p.Conf.Unmarshal("kafka.reader", &dbConfs)
	if err != nil {
		_ = level.Warn(p.Logger).Log("err", err)
	}
	factory := di.NewFactory(func(name string) (di.Pair, error) {
		var (
			ok           bool
			readerConfig ReaderConfig
		)
		if readerConfig, ok = dbConfs[name]; !ok {
			return di.Pair{}, fmt.Errorf("kafka reader configuration %s not valid", name)
		}

		// converts to the kafka.ReaderConfig from github.com/segmentio/kafka-go
		conf := fromReaderConfig(readerConfig)
		conf.Logger = KafkaLogAdapter{Logging: level.Debug(p.Logger)}
		conf.ErrorLogger = KafkaLogAdapter{Logging: level.Warn(p.Logger)}
		if p.WriterInterceptor != nil {
			p.ReaderInterceptor(name, &conf)
		}
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
func provideWriterFactory(p in) (WriterFactory, func()) {
	var err error
	var dbConfs map[string]WriterConfig
	err = p.Conf.Unmarshal("kafka.writer", &dbConfs)
	if err != nil {
		_ = level.Warn(p.Logger).Log("err", err)
	}
	factory := di.NewFactory(func(name string) (di.Pair, error) {
		var (
			ok           bool
			writerConfig WriterConfig
		)
		if writerConfig, ok = dbConfs[name]; !ok {
			return di.Pair{}, fmt.Errorf("kafka writer configuration %s not valid", name)
		}
		writer := fromWriterConfig(writerConfig)
		logger := log.With(p.Logger, "tag", "kafka")
		writer.Logger = KafkaLogAdapter{Logging: level.Debug(logger)}
		writer.ErrorLogger = KafkaLogAdapter{Logging: level.Warn(logger)}
		writer.Transport = NewTransport(kafka.DefaultTransport, p.Tracer)
		if p.WriterInterceptor != nil {
			p.WriterInterceptor(name, &writer)
		}

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
