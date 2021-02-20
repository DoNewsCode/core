package kitkafka

import (
	"fmt"

	"github.com/DoNewsCode/std/pkg/config"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/di"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/segmentio/kafka-go"
)

// WriterMaker models a WriterFactory
type WriterMaker interface {
	Make(name string) (*kafka.Writer, error)
}

// ReaderMaker models a ReaderFactory
type ReaderMaker interface {
	Make(name string) (*kafka.Reader, error)
}

// KafkaIn is a injection parameter for ProvideReaderFactory.
type KafkaIn struct {
	di.In

	ReaderInterceptor ReaderInterceptor `optional:"true"`
	WriterInterceptor WriterInterceptor `optional:"true"`
	Conf              contract.ConfigAccessor
	Logger            log.Logger
}

// KafkaOut is the result of ProvideKafka.
type KafkaOut struct {
	di.In

	ReaderFactory   ReaderFactory
	WriterFactory   WriterFactory
	ReaderMaker     ReaderMaker
	WriterMaker     WriterMaker
	ExportedConfigs []config.ExportedConfig `group:"config,flatten"`
}

// ProvideKafka creates the ReaderFactory and WriterFactory. It is
// valid dependency option for package core. Note: when working with package
// core's DI container, use ProvideKafka over ProvideReaderFactory and
// ProvideWriterFactory. Not only ProvideKafka provides both reader and
// writer, but also only ProvideKafka exports default Kafka configuration.
func ProvideKafka(p KafkaIn) (KafkaOut, func(), func(), error) {
	rf, rc := ProvideReaderFactory(p)
	wf, wc := ProvideWriterFactory(p)
	return KafkaOut{
		ReaderMaker:     rf,
		ReaderFactory:   rf,
		WriterMaker:     wf,
		WriterFactory:   wf,
		ExportedConfigs: provideConfig(),
	}, wc, rc, nil
}

// ProvideReaderFactory creates the ReaderFactory. It is valid
// dependency option for package core.
func ProvideReaderFactory(p KafkaIn) (ReaderFactory, func()) {
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

// ProvideWriterFactory creates WriterFactory. It is a valid injection
// option for package core.
func ProvideWriterFactory(p KafkaIn) (WriterFactory, func()) {
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
		writer.Logger = KafkaLogAdapter{Logging: level.Debug(p.Logger)}
		writer.ErrorLogger = KafkaLogAdapter{Logging: level.Warn(p.Logger)}
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
