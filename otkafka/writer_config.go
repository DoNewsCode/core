package otkafka

import (
	"time"

	"github.com/segmentio/kafka-go"
)

// WriterInterceptor is an interceptor that makes last minute change to a
// *kafka.Writer during its creation
type WriterInterceptor func(name string, writer *kafka.Writer)

// WriterConfig is a configuration type used to create new instances of Writer.
type WriterConfig struct {
	// The list of brokers used to discover the partitions available on the
	// kafka cluster.
	//
	// This field is required, attempting to create a writer with an empty list
	// of brokers will panic.
	Brokers []string `json:"brokers" yaml:"brokers"`

	// The topic that the writer will produce messages to.
	//
	// If provided, this will be used to set the topic for all produced messages.
	// If not provided, each Message must specify a topic for itself. This must be
	// mutually exclusive, otherwise the Writer will return an error.
	Topic string `json:"topic" yaml:"topic"`

	// Limit on how many attempts will be made to deliver a message.
	//
	// The default is to try at most 10 times.
	MaxAttempts int `json:"maxAttempts" yaml:"maxAttempts"`

	// Limit on how many messages will be buffered before being sent to a
	// partition.
	//
	// The default is to use a target batch size of 100 messages.
	BatchSize int `json:"batchSize" yaml:"batchSize"`

	// Limit the maximum size of a request factoryIn bytes before being sent to
	// a partition.
	//
	// The default is to use a kafka default value of 1048576.
	BatchBytes int `json:"batchBytes" yaml:"batchBytes"`

	// Time limit on how often incomplete message batches will be flushed to
	// kafka.
	//
	// The default is to flush at least every second.
	BatchTimeout time.Duration `json:"batchTimeout" yaml:"batchTimeout"`

	// Timeout for read operations performed by the Writer.
	//
	// Defaults to 10 seconds.
	ReadTimeout time.Duration `json:"readTimeout" yaml:"readTimeout"`

	// Timeout for write operation performed by the Writer.
	//
	// Defaults to 10 seconds.
	WriteTimeout time.Duration `json:"writeTimeout" yaml:"writeTimeout"`

	// DEPRECATED: factoryIn versions prior to 0.4, the writer used to maintain a cache
	// the topic layout. With the change to use a transport to manage connections,
	// the responsibility of syncing the cluster layout has been delegated to the
	// transport.
	RebalanceInterval time.Duration `json:"rebalanceInterval" yaml:"rebalanceInterval"`

	// Number of acknowledges from partition replicas required before receiving
	// a response to a produce request. The default is -1, which means to wait for
	// all replicas, and a value above 0 is required to indicate how many replicas
	// should acknowledge a message to be considered successful.
	//
	// This version of kafka-go (v0.3) does not support 0 required acks, due to
	// some internal complexity implementing this with the Kafka protocol. If you
	// need that functionality specifically, you'll need to upgrade to v0.4.
	RequiredAcks int `json:"requiredAcks" yaml:"requiredAcks"`

	// Setting this flag to true causes the WriteMessages method to never block.
	// It also means that errors are ignored since the caller will not receive
	// the returned value. Use this only if you don't care about guarantees of
	// whether the messages were written to kafka.
	Async bool `json:"async" yaml:"async"`
}

func fromWriterConfig(conf WriterConfig) kafka.Writer {
	if len(conf.Brokers) == 0 {
		conf.Brokers = []string{"127.0.0.1:9092"}
	}
	return kafka.Writer{
		Addr:         kafka.TCP(conf.Brokers...),
		Topic:        conf.Topic,
		MaxAttempts:  conf.MaxAttempts,
		BatchSize:    conf.BatchSize,
		BatchBytes:   int64(conf.BatchBytes),
		BatchTimeout: conf.BatchTimeout,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
		RequiredAcks: kafka.RequiredAcks(conf.RequiredAcks),
		Async:        conf.Async,
	}
}
