package kitkafka

import (
	"fmt"
	"github.com/DoNewsCode/std/pkg/async"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/pkg/errors"
	"go.uber.org/dig"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/segmentio/kafka-go"
)

type KafkaReaderFactory struct {
	*async.Factory
}

func (k KafkaReaderFactory) Make(name string) (*kafka.Reader, error) {
	client, err := k.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(*kafka.Reader), nil
}

type KafkaParam struct {
	dig.In

	Conf   contract.ConfigAccessor
	Logger log.Logger
}

func ProvideKafkaReaderFactory(p KafkaParam) (KafkaReaderFactory, func()) {
	var err error
	var dbConfs map[string]kafka.ReaderConfig
	err = p.Conf.Unmarshal("kafka.reader", &dbConfs)
	if err != nil {
		_ = level.Warn(p.Logger).Log("err", err)
	}
	factory := async.NewFactory(func(name string) (async.Pair, error) {
		var (
			ok   bool
			conf kafka.ReaderConfig
		)
		if conf, ok = dbConfs[name]; !ok {
			return async.Pair{}, fmt.Errorf("kafka reader configuration %s not valid", name)
		}
		conf.Logger = KafkaLogAdapter{Logging: level.Debug(p.Logger)}
		conf.ErrorLogger = KafkaLogAdapter{Logging: level.Warn(p.Logger)}
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

type KafkaWriterFactory struct {
	*async.Factory
}

func (k KafkaWriterFactory) Make(name string) (*kafka.Writer, error) {
	client, err := k.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(*kafka.Writer), nil
}

func ProvideKafkaWriterFactory(p KafkaParam) (KafkaWriterFactory, func()) {
	var err error
	var dbConfs map[string]kafka.Writer
	err = p.Conf.Unmarshal("kafka.writer", &dbConfs)
	if err != nil {
		_ = level.Warn(p.Logger).Log("err", err)
	}
	factory := async.NewFactory(func(name string) (async.Pair, error) {
		var (
			ok     bool
			writer kafka.Writer
		)
		if writer, ok = dbConfs[name]; !ok {
			return async.Pair{}, fmt.Errorf("kafka writer configuration %s not valid", name)
		}
		writer.Logger = KafkaLogAdapter{Logging: level.Debug(p.Logger)}
		writer.ErrorLogger = KafkaLogAdapter{Logging: level.Warn(p.Logger)}

		return async.Pair{
			Conn: &writer,
			Closer: func() {
				_ = writer.Close()
			},
		}, nil
	})
	return KafkaWriterFactory{factory}, factory.Close
}

func (k KafkaWriterFactory) MakeWriterHandle(name string) (*writerHandle, error) {
	writer, err := k.Make(name)
	if err != nil {
		return nil, err
	}
	return &writerHandle{
		Writer: writer,
	}, nil
}

type readerConfig struct {
	parallelism int
}

type readerOpt func(config *readerConfig)

func WithParallelism(parallelism int) readerOpt {
	return func(config *readerConfig) {
		config.parallelism = parallelism
	}
}

func (k KafkaReaderFactory) MakeSubscriberClient(name string, subscriber Handler, opt ...readerOpt) (*SubscriberClient, error) {
	var config = readerConfig{
		parallelism: 1,
	}
	for _, o := range opt {
		o(&config)
	}
	reader, err := k.Make(name)
	if err != nil {
		return nil, errors.Wrap(err, "unable to make subscriber")
	}
	return &SubscriberClient{
		reader:      reader,
		handler:     subscriber,
		parallelism: config.parallelism,
	}, nil
}

type writerConfig struct{}

type writerOpt func(config *writerConfig)

func MakePublisherClient(endpoint endpoint.Endpoint, opt ...writerOpt) *PublisherClient {
	return &PublisherClient{endpoint: endpoint}
}
