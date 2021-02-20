package kitkafka

import (
	"fmt"

	"github.com/DoNewsCode/std/pkg/async"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/di"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/segmentio/kafka-go"
)

// KafkaIn is a injection parameter for ProvideKafkaReaderFactory.
type KafkaIn struct {
	di.In

	ReaderInterceptor ReaderInterceptor
	WriterInterceptor WriterInterceptor
	Conf              contract.ConfigAccessor
	Logger            log.Logger
}

// KafkaOut is the result of ProvideKafka.
type KafkaOut struct {
	di.In

	ReaderFactory   KafkaReaderFactory
	WriterFactory   KafkaWriterFactory
	ExportedConfigs []contract.ExportedConfig `group:"config,flatten"`
}

// ProvideKafka creates the KafkaReaderFactory and KafkaWriterFactory. It is
// valid dependency option for package core. Note: when working with package
// core's DI container, use ProvideKafka over ProvideKafkaReaderFactory and
// ProvideKafkaWriterFactory. Not only ProvideKafka provides both reader and
// writer, but also only ProvideKafka exports default Kafka configuration.
func ProvideKafka(p KafkaIn) (KafkaOut, func(), func(), error) {
	rf, rc := ProvideKafkaReaderFactory(p)
	wf, wc := ProvideKafkaWriterFactory(p)
	return KafkaOut{
		ReaderFactory:   rf,
		WriterFactory:   wf,
		ExportedConfigs: provideConfig(),
	}, wc, rc, nil
}

// ProvideKafkaReaderFactory creates the KafkaReaderFactory. It is valid
// dependency option for package core.
func ProvideKafkaReaderFactory(p KafkaIn) (KafkaReaderFactory, func()) {
	var err error
	var dbConfs map[string]ReaderConfig
	err = p.Conf.Unmarshal("kafka.reader", &dbConfs)
	if err != nil {
		_ = level.Warn(p.Logger).Log("err", err)
	}
	factory := async.NewFactory(func(name string) (async.Pair, error) {
		var (
			ok           bool
			readerConfig ReaderConfig
		)
		if readerConfig, ok = dbConfs[name]; !ok {
			return async.Pair{}, fmt.Errorf("kafka reader configuration %s not valid", name)
		}

		// converts to the kafka.ReaderConfig from github.com/segmentio/kafka-go
		conf := fromReaderConfig(readerConfig)
		conf.Logger = KafkaLogAdapter{Logging: level.Debug(p.Logger)}
		conf.ErrorLogger = KafkaLogAdapter{Logging: level.Warn(p.Logger)}
		if p.WriterInterceptor != nil {
			p.ReaderInterceptor(name, &conf)
		}
		client := kafka.NewReader(conf)
		return async.Pair{
			Conn: client,
			Closer: func() {
				_ = client.Close()
			},
		}, nil
	})
	return KafkaReaderFactory{factory}, factory.Close
}

// ProvideKafkaWriterFactory creates KafkaWriterFactory. It is a valid injection
// option for pacakge core.
func ProvideKafkaWriterFactory(p KafkaIn) (KafkaWriterFactory, func()) {
	var err error
	var dbConfs map[string]WriterConfig
	err = p.Conf.Unmarshal("kafka.writer", &dbConfs)
	if err != nil {
		_ = level.Warn(p.Logger).Log("err", err)
	}
	factory := async.NewFactory(func(name string) (async.Pair, error) {
		var (
			ok           bool
			writerConfig WriterConfig
		)
		if writerConfig, ok = dbConfs[name]; !ok {
			return async.Pair{}, fmt.Errorf("kafka writer configuration %s not valid", name)
		}
		writer := fromWriterConfig(writerConfig)
		writer.Logger = KafkaLogAdapter{Logging: level.Debug(p.Logger)}
		writer.ErrorLogger = KafkaLogAdapter{Logging: level.Warn(p.Logger)}
		if p.WriterInterceptor != nil {
			p.WriterInterceptor(name, &writer)
		}

		return async.Pair{
			Conn: &writer,
			Closer: func() {
				_ = writer.Close()
			},
		}, nil
	})
	return KafkaWriterFactory{factory}, factory.Close
}
