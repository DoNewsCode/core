package otkafka

import (
	"github.com/DoNewsCode/core/di"
	"github.com/segmentio/kafka-go"
)

// ReaderFactory is a di.Factory[*kafka.Reader] that creates *kafka.Reader.
//
// Unlike other database providers, the kafka factories don't bundle a default
// kafka reader/writer. It is suggested to use Topic name as the identifier of
// kafka config rather than an opaque name such as default.
type ReaderFactory = di.Factory[*kafka.Reader]

// WriterFactory is a di.Factory[*kafka.Writer] that creates *kafka.Writer.
//
// Unlike other database providers, the kafka factories don't bundle a default
// kafka reader/writer. It is suggested to use Topic name as the identifier of
// kafka config rather than an opaque name such as default.
type WriterFactory = di.Factory[*kafka.Writer]
