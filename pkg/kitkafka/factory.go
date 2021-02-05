package kitkafka

import (
	"sync"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/segmentio/kafka-go"
)

type KafkaFactory struct {
	mutex   sync.Mutex
	brokers []string
	closers []func() error
	logger  log.Logger
}

func NewKafkaFactory(brokers []string, logger log.Logger) *KafkaFactory {
	logger = log.With(logger, "component", "kafka")
	return &KafkaFactory{
		brokers: brokers,
		closers: []func() error{},
		logger:  logger,
	}
}

func (k *KafkaFactory) MakeWriterHandle(topic string) *writerHandle {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	writer := &kafka.Writer{
		Addr:        kafka.TCP(k.brokers...),
		Topic:       topic,
		Balancer:    &kafka.LeastBytes{},
		Logger:      KafkaLogAdapter{Logging: level.Debug(k.logger)},
		ErrorLogger: KafkaLogAdapter{Logging: level.Warn(k.logger)},
		BatchSize:   1,
	}

	k.closers = append(k.closers, writer.Close)
	return &writerHandle{
		Writer: writer,
	}
}

type readerConfig struct {
	groupId     string
	parallelism int
}

type readerOpt func(config *readerConfig)

func WithGroup(group string) readerOpt {
	return func(config *readerConfig) {
		config.groupId = group
	}
}

func WithParallelism(parallelism int) readerOpt {
	return func(config *readerConfig) {
		config.parallelism = parallelism
	}
}

func (k *KafkaFactory) MakeSubscriberClient(topic string, subscriber *Subscriber, opt ...readerOpt) *SubscriberClient {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	var config = readerConfig{
		groupId:     "",
		parallelism: 1,
	}
	for _, o := range opt {
		o(&config)
	}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     k.brokers,
		Topic:       topic,
		GroupID:     config.groupId,
		Logger:      KafkaLogAdapter{Logging: level.Debug(k.logger)},
		ErrorLogger: KafkaLogAdapter{Logging: level.Warn(k.logger)},
		MinBytes:    1,
		MaxBytes:    10 * 1024 * 1024,
	})

	k.closers = append(k.closers, reader.Close)

	return &SubscriberClient{
		reader:      reader,
		handler:     subscriber,
		parallelism: config.parallelism,
	}
}

type writerConfig struct{}

type writerOpt func(config *writerConfig)

func (k *KafkaFactory) MakePublisherClient(endpoint endpoint.Endpoint, opt ...writerOpt) *PublisherClient {
	return &PublisherClient{endpoint: endpoint}
}

func (k *KafkaFactory) Close() error {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	for _, v := range k.closers {
		err := v()
		if err != nil {
			return err
		}
	}
	return nil
}
