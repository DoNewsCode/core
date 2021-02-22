package kitkafka

import (
	"context"

	"github.com/DoNewsCode/core/contract"
	"github.com/segmentio/kafka-go"
)

// DecodeRequestFunc extracts a user-domain request object from
// a *kafka.Message. It is designed to be used in Kafka Subscribers.
type DecodeRequestFunc func(context.Context, *kafka.Message) (request interface{}, err error)

// EncodeRequestFunc encodes the passed request object into
// a *kafka.Message. It is designed to be used in Kafka Publishers.
type EncodeRequestFunc func(context.Context, *kafka.Message, interface{}) error

// EncodeResponseFunc encodes the passed response object to
// a *kafka.Message. It is designed to be used in Kafka Subscribers.
type EncodeResponseFunc func(context.Context, *kafka.Message, interface{}) error

// DecodeResponseFunc extracts a user-domain response object from
// an *kafka.Message. It is designed to be used in Kafka Publishers.
type DecodeResponseFunc func(context.Context, *kafka.Message) (response interface{}, err error)

// EncodeMarshaller encodes the user-domain request object into a *kafka.Message.
// The request object must implement contract.Marshaller. Protobuf objects
// implemented this interface out of box.
func EncodeMarshaller(ctx context.Context, msg *kafka.Message, request interface{}) error {
	byt, err := request.(contract.Marshaller).Marshal()
	if err != nil {
		return err
	}
	msg.Value = byt
	return nil
}
