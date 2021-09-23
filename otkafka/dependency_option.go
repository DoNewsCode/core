package otkafka

import "github.com/segmentio/kafka-go"

// ReaderInterceptor is an interceptor that makes last minute change to a *kafka.ReaderConfig
// during kafka.Reader's creation
type ReaderInterceptor func(name string, reader *kafka.ReaderConfig)

// WriterInterceptor is an interceptor that makes last minute change to a
// *kafka.Writer during its creation
type WriterInterceptor func(name string, writer *kafka.Writer)

type providersOption struct {
	readerReloadable  bool
	readerInterceptor ReaderInterceptor
	writerReloadable  bool
	writerInterceptor WriterInterceptor
}

// ProvidersOptionFunc is the type of functional providersOption for Providers. Use this type to change how Providers work.
type ProvidersOptionFunc func(options *providersOption)

// WithReaderInterceptor instructs the Providers to accept the
// ReaderInterceptor so that users can change reader config during runtime. This can
// be useful when some dynamic computations on configs are required.
func WithReaderInterceptor(interceptor ReaderInterceptor) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.readerInterceptor = interceptor
	}
}

// WithWriterInterceptor instructs the Providers to accept the
// WriterInterceptor so that users can change reader config during runtime. This can
// be useful when some dynamic computations on configs are required.
func WithWriterInterceptor(interceptor WriterInterceptor) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.writerInterceptor = interceptor
	}
}

// WithReaderReload toggles whether the reader factory should reload cached instances upon
// OnReload event.
func WithReaderReload(shouldReload bool) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.readerReloadable = shouldReload
	}
}

// WithWriterReload toggles whether the writer factory should reload cached instances upon
// OnReload event.
func WithWriterReload(shouldReload bool) ProvidersOptionFunc {
	return func(options *providersOption) {
		options.writerReloadable = shouldReload
	}
}
