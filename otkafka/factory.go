package otkafka

import (
	"github.com/DoNewsCode/core/di"
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
