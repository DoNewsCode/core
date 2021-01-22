package kitkafka

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/segmentio/kafka-go"
)

// Publisher wraps an AMQP channel and queue, and provides a method that
// implements endpoint.Endpoint.
type Publisher struct {
	handler Handler
	enc     EncodeRequestFunc
	before  []RequestResponseFunc
	after   []RequestResponseFunc
	timeout time.Duration
}

// NewPublisher constructs a usable Publisher for a single remote method.
func NewPublisher(
	handler Handler,
	enc EncodeRequestFunc,
	options ...PublisherOption,
) *Publisher {
	p := &Publisher{
		handler: handler,
		enc:     enc,
		timeout: 10 * time.Second,
	}
	for _, option := range options {
		option(p)
	}
	return p
}

// PublisherOption sets an optional parameter for clients.
type PublisherOption func(*Publisher)

// PublisherBefore sets the RequestFuncs that are applied to the outgoing AMQP
// request before it's invoked.
func PublisherBefore(before ...RequestResponseFunc) PublisherOption {
	return func(p *Publisher) { p.before = append(p.before, before...) }
}

// PublisherAfter sets the ClientResponseFuncs applied to the incoming AMQP
// request prior to it being decoded. This is useful for obtaining anything off
// of the response and adding onto the context prior to decoding.
func PublisherAfter(after ...RequestResponseFunc) PublisherOption {
	return func(p *Publisher) { p.after = append(p.after, after...) }
}

// PublisherTimeout sets the available timeout for an AMQP request.
func PublisherTimeout(timeout time.Duration) PublisherOption {
	return func(p *Publisher) { p.timeout = timeout }
}

// Endpoint returns a usable endpoint that invokes the remote endpoint.
func (p Publisher) Endpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, p.timeout)
		defer cancel()

		outgoing := kafka.Message{}

		if err := p.enc(ctx, &outgoing, request); err != nil {
			return nil, err
		}

		for _, f := range p.before {
			ctx = f(ctx, &outgoing)
		}

		err := p.handler.Handle(ctx, outgoing)
		if err != nil {
			return nil, err
		}

		for _, f := range p.after {
			ctx = f(ctx, &outgoing)
		}

		return nil, nil
	}
}
