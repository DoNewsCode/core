package kitkafka

import (
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/segmentio/kafka-go"
	"github.com/DoNewsCode/std/pkg/logging"
)

type KafkaFactory struct {
	mutex   sync.Mutex
	brokers []string
	closers []func() error
	logger  log.Logger
}

func NewKafkaFactory(brokers []string, logger log.Logger) *KafkaFactory {
	return &KafkaFactory{
		brokers: brokers,
		closers: []func() error{},
		logger:  logger,
	}
}

func (k *KafkaFactory) MakeHandler(topic string) Handler {
	k.mutex.Lock()
	defer k.mutex.Unlock()

	writer := &kafka.Writer{
		Addr:        kafka.TCP(k.brokers...),
		Topic:       topic,
		Balancer:    &kafka.LeastBytes{},
		Logger:      logging.KafkaLogAdapter{Logging: log.NewNopLogger()},
		ErrorLogger: logging.KafkaLogAdapter{Logging: level.Warn(k.logger)},
		BatchSize:   1,
	}

	k.closers = append(k.closers, writer.Close)
	return &pub{
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

func (k *KafkaFactory) MakeKafkaServer(topic string, handler Handler, opt ...readerOpt) *sub {
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
		Logger:      logging.KafkaLogAdapter{Logging: level.Debug(k.logger)},
		ErrorLogger: logging.KafkaLogAdapter{Logging: level.Warn(k.logger)},
		MinBytes:    1,
		MaxBytes:    10 * 1024 * 1024,
	})

	k.closers = append(k.closers, reader.Close)

	return &sub{
		reader:      reader,
		handler:     handler,
		parallelism: config.parallelism,
	}
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
