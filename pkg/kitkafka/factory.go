package kitkafka

import (
	"fmt"

	"github.com/DoNewsCode/std/pkg/async"
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/di"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

// KafkaReaderFactory is a *async.Factory that creates *kafka.Reader.
type KafkaReaderFactory struct {
	*async.Factory
}

// Make returns a *kafka.Reader under the provided configuration entry.
func (k KafkaReaderFactory) Make(name string) (*kafka.Reader, error) {
	client, err := k.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(*kafka.Reader), nil
}

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

// ProvideKafka creates the KafkaReaderFactory and KafkaWriterFactory. It is valid
// dependency option for package core.
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

// KafkaWriterFactory is a *async.Factory that creates *kafka.Writer.
type KafkaWriterFactory struct {
	*async.Factory
}

// Make returns a *kafka.Writer under the provided configuration entry.
func (k KafkaWriterFactory) Make(name string) (*kafka.Writer, error) {
	client, err := k.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(*kafka.Writer), nil
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

// MakeClient creates an Handler. This handler can write *kafka.Message to
// kafka broker. The Handler is mean to be consumed by NewPublisher.
func (k KafkaWriterFactory) MakeClient(name string) (*writerHandle, error) {
	writer, err := k.Make(name)
	if err != nil {
		return nil, err
	}
	return &writerHandle{
		Writer: writer,
	}, nil
}

type subscriberConfig struct {
	parallelism int
	syncCommit  bool
}

// ReaderOpt are options that configures the kafka reader.
type ReaderOpt func(config *subscriberConfig)

// WithParallelism configures the parallelism of fan out workers.
func WithParallelism(parallelism int) ReaderOpt {
	return func(config *subscriberConfig) {
		config.parallelism = parallelism
	}
}

// WithSyncCommit is an kafka option that when enabled, only commit the message
// synchronously if no error is returned from the endpoint.
func WithSyncCommit() ReaderOpt {
	return func(config *subscriberConfig) {
		config.syncCommit = true
	}
}

// MakeSubscriberServer creates a *SubscriberServer.
//     name: the key of the configuration entry.
//     subscriber: the Handler (go kit transport layer)
func (k KafkaReaderFactory) MakeSubscriberServer(name string, subscriber Handler, opt ...ReaderOpt) (*SubscriberServer, error) {
	var config = subscriberConfig{
		parallelism: 1,
	}
	for _, o := range opt {
		o(&config)
	}
	reader, err := k.Make(name)
	if err != nil {
		return nil, errors.Wrap(err, "unable to make subscriber")
	}
	return &SubscriberServer{
		reader:      reader,
		handler:     subscriber,
		parallelism: config.parallelism,
		syncCommit:  config.syncCommit,
	}, nil
}

type publisherConfig struct{}

// A PublisherOpt is an option that configures publisher.
type PublisherOpt func(config *publisherConfig)

// MakePublisherService returns a *PublisherService that can publish user-domain messages to kafka brokers.
// In go kit analog, this is a service with one method, publish.
func MakePublisherService(endpoint endpoint.Endpoint, opt ...PublisherOpt) *PublisherService {
	return &PublisherService{endpoint: endpoint}
}

func provideConfig() []contract.ExportedConfig {
	return []contract.ExportedConfig{
		{
			Owner: "kitkafka",
			Data: map[string]interface{}{
				"kafka": map[string]interface{}{
					"reader": ReaderConfig{
						Brokers: []string{"127.0.0.1:9092"},
					},
					"writer": WriterConfig{
						Brokers: []string{"127.0.0.1:9092"},
					},
				},
			},
			Comment: "",
		},
	}
}

func fromWriterConfig(config WriterConfig) kafka.Writer {
	return kafka.Writer{
		Addr:         kafka.TCP(config.Brokers...),
		Topic:        config.Topic,
		MaxAttempts:  config.MaxAttempts,
		BatchSize:    config.BatchSize,
		BatchBytes:   int64(config.BatchBytes),
		BatchTimeout: config.BatchTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		RequiredAcks: kafka.RequiredAcks(config.RequiredAcks),
		Async:        config.Async,
	}
}

func fromReaderConfig(config ReaderConfig) kafka.ReaderConfig {
	return kafka.ReaderConfig{
		Brokers:                config.Brokers,
		GroupID:                config.GroupID,
		Topic:                  config.Topic,
		Partition:              config.MaxAttempts,
		MinBytes:               config.MinBytes,
		MaxBytes:               config.MaxBytes,
		MaxWait:                config.MaxWait,
		ReadLagInterval:        config.ReadLagInterval,
		HeartbeatInterval:      config.HeartbeatInterval,
		CommitInterval:         config.CommitInterval,
		PartitionWatchInterval: config.PartitionWatchInterval,
		WatchPartitionChanges:  config.WatchPartitionChanges,
		SessionTimeout:         config.SessionTimeout,
		RebalanceTimeout:       config.RebalanceTimeout,
		JoinGroupBackoff:       config.JoinGroupBackoff,
		RetentionTime:          config.RetentionTime,
		StartOffset:            config.StartOffset,
		ReadBackoffMin:         config.ReadBackoffMin,
		ReadBackoffMax:         config.ReadBackoffMax,
		MaxAttempts:            config.MaxAttempts,
	}
}
