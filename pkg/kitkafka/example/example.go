package example

import (
	"github.com/DoNewsCode/std/pkg/kitkafka"
	"github.com/go-kit/kit/endpoint"
)

func Publisher(writerFactory kitkafka.KafkaWriterFactory) (*kitkafka.PublisherClient, error) {
	// make a handler
	handler, err := writerFactory.MakeWriterHandle("default")
	if err != nil {
		return nil, err
	}

	// convert a handler to endpoint
	publisher := kitkafka.NewPublisher(handler, kitkafka.EncodeMarshaller)
	ep := publisher.Endpoint()

	// TODO: add middleware to endpoint

	// convert an endpoint to a service
	return kitkafka.MakePublisherClient(ep), nil
}

func Subscriber(readerFactory kitkafka.KafkaReaderFactory, ep endpoint.Endpoint, dec kitkafka.DecodeRequestFunc) (*kitkafka.SubscriberClient, error) {
	// first convert service to endpoint

	// TODO: add middleware to endpoint

	// convert endpoint to handler
	subscriber := kitkafka.NewSubscriber(ep, dec)

	// connect handler to a server
	return readerFactory.MakeSubscriberClient("default", subscriber)
}
