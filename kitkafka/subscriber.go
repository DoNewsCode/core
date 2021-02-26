package kitkafka

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	"github.com/segmentio/kafka-go"
)

// Subscriber is go kit transport layer that wraps an endpoint. It is a handler
// for SubscriberServer.
type Subscriber struct {
	e            endpoint.Endpoint
	dec          DecodeRequestFunc
	before       []RequestResponseFunc
	after        []RequestResponseFunc
	errorHandler transport.ErrorHandler
}

// NewSubscriber constructs a new subscriber, which provides a handler
// for Kafka messages.
func NewSubscriber(
	e endpoint.Endpoint,
	dec DecodeRequestFunc,
	options ...SubscriberOption,
) *Subscriber {
	s := &Subscriber{
		e:            e,
		dec:          dec,
		errorHandler: transport.NewLogErrorHandler(log.NewNopLogger()),
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// SubscriberOption sets an optional parameter for subscribers.
type SubscriberOption func(*Subscriber)

// SubscriberBefore functions are executed on the publisher delivery object
// before the request is decoded.
func SubscriberBefore(before ...RequestResponseFunc) SubscriberOption {
	return func(s *Subscriber) { s.before = append(s.before, before...) }
}

// SubscriberAfter functions are executed on the subscriber reply after the
// endpoint is invoked, but before anything is published to the reply.
func SubscriberAfter(after ...RequestResponseFunc) SubscriberOption {
	return func(s *Subscriber) { s.after = append(s.after, after...) }
}

// SubscriberErrorLogger is used to log non-terminal errors. By default, no errors
// are logged. This is intended as a diagnostic measure. Finer-grained control
// of error handling, including logging in more detail, should be performed in a
// custom SubscriberErrorEncoder which has access to the context.
// Deprecated: Use SubscriberErrorHandler instead.
func SubscriberErrorLogger(logger log.Logger) SubscriberOption {
	return func(s *Subscriber) { s.errorHandler = transport.NewLogErrorHandler(logger) }
}

// SubscriberErrorHandler is used to handle non-terminal errors. By default, non-terminal errors
// are ignored. This is intended as a diagnostic measure. Finer-grained control
// of error handling, including logging in more detail, should be performed in a
// custom SubscriberErrorEncoder which has access to the context.
func SubscriberErrorHandler(errorHandler transport.ErrorHandler) SubscriberOption {
	return func(s *Subscriber) { s.errorHandler = errorHandler }
}

// Handle handles kafka messages.
func (s Subscriber) Handle(ctx context.Context, incoming kafka.Message) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, f := range s.before {
		ctx = f(ctx, &incoming)
	}

	request, err := s.dec(ctx, &incoming)
	if err != nil {
		s.errorHandler.Handle(ctx, err)
		return err
	}

	_, err = s.e(ctx, request)
	if err != nil {
		s.errorHandler.Handle(ctx, err)
		return err
	}

	for _, f := range s.after {
		ctx = f(ctx, &incoming)
	}

	return nil
}

type Reader interface {
	Close() error
	ReadMessage(ctx context.Context) (kafka.Message, error)
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
}
