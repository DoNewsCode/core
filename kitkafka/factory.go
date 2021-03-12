package kitkafka

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/segmentio/kafka-go"
)

// MakeClient creates an Handler. This handler can write *kafka.Message to
// kafka broker. The Handler is mean to be consumed by NewPublisher.
func MakeClient(writer *kafka.Writer) (*writerHandle, error) {
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

// MakeSubscriberServer creates a *SubscriberServer. Subscriber is the go kit transport layer equivalent.
func MakeSubscriberServer(reader *kafka.Reader, subscriber Handler, opt ...ReaderOpt) (*SubscriberServer, error) {
	var config = subscriberConfig{
		parallelism: 1,
	}
	for _, o := range opt {
		o(&config)
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
// in go kit analog, this is a service with one method, publish.
func MakePublisherService(endpoint endpoint.Endpoint, opt ...PublisherOpt) *PublisherService {
	return &PublisherService{endpoint: endpoint}
}
