package example

import (
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/kitkafka"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

func Publisher(conf contract.ConfigAccessor, logger log.Logger) *kitkafka.PublisherClient {
	brokers := conf.Strings("bootstrapServer")
	factory := kitkafka.NewKafkaFactory(brokers, logger)
	topic := conf.String("topic")
	publisher := kitkafka.NewPublisher(factory.MakeWriterHandle(topic), kitkafka.EncodeMarshaller)
	ep := publisher.Endpoint()

	// TODO: add middleware to endpoint

	return factory.MakePublisherClient(ep)
}

func Subscriber(conf contract.ConfigAccessor, logger log.Logger, ep endpoint.Endpoint, dec kitkafka.DecodeRequestFunc) *kitkafka.SubscriberClient {
	brokers := conf.Strings("bootstrapServer")
	factory := kitkafka.NewKafkaFactory(brokers, logger)
	subscriber := kitkafka.NewSubscriber(ep, dec)
	return factory.MakeSubscriberClient("topic", subscriber)
}
