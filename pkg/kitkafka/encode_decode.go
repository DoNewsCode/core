package kitkafka

import (
	"context"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/segmentio/kafka-go"
)

// DecodeRequestFunc extracts a user-domain request object from
// an AMQP Delivery object. It is designed to be used in AMQP Subscribers.
type DecodeRequestFunc func(context.Context, *kafka.Message) (request interface{}, err error)

// EncodeRequestFunc encodes the passed request object into
// an AMQP Publishing object. It is designed to be used in AMQP Publishers.
type EncodeRequestFunc func(context.Context, *kafka.Message, interface{}) error

// EncodeResponseFunc encodes the passed response object to
// an AMQP Publishing object. It is designed to be used in AMQP Subscribers.
type EncodeResponseFunc func(context.Context, *kafka.Message, interface{}) error

// DecodeResponseFunc extracts a user-domain response object from
// an AMQP Delivery object. It is designed to be used in AMQP Publishers.
type DecodeResponseFunc func(context.Context, *kafka.Message) (response interface{}, err error)

func EncodeMarshaller(ctx context.Context, msg *kafka.Message, request interface{}) error {
	byt, err := request.(contract.Marshaller).Marshal()
	if err != nil {
		return err
	}
	msg.Value = byt
	return nil
}
