package kitkafka

import (
	"github.com/DoNewsCode/std/pkg/config"
	"github.com/DoNewsCode/std/pkg/di"
	"github.com/go-kit/kit/endpoint"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

// ReaderFactory is a *di.Factory that creates *kafka.Reader.
//
// Unlike other database providers, the kafka factories don't bundle a default
// kafka reader/writer. It is suggested to use Topic name as the identifier of
// kafka config rather than an opaque name such as default.
type ReaderFactory struct {
	*di.Factory
}

// Make returns a *kafka.Reader under the provided configuration entry.
func (k ReaderFactory) Make(name string) (*kafka.Reader, error) {
	client, err := k.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(*kafka.Reader), nil
}

// WriterFactory is a *di.Factory that creates *kafka.Writer.
//
// Unlike other database providers, the kafka factories don't bundle a default
// kafka reader/writer. It is suggested to use Topic name as the identifier of
// kafka config rather than an opaque name such as default.
type WriterFactory struct {
	*di.Factory
}

// Make returns a *kafka.Writer under the provided configuration entry.
func (k WriterFactory) Make(name string) (*kafka.Writer, error) {
	client, err := k.Factory.Make(name)
	if err != nil {
		return nil, err
	}
	return client.(*kafka.Writer), nil
}

// MakeClient creates an Handler. This handler can write *kafka.Message to
// kafka broker. The Handler is mean to be consumed by NewPublisher.
func (k WriterFactory) MakeClient(name string) (*writerHandle, error) {
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
func (k ReaderFactory) MakeSubscriberServer(name string, subscriber Handler, opt ...ReaderOpt) (*SubscriberServer, error) {
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

func provideConfig() []config.ExportedConfig {
	return []config.ExportedConfig{
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
