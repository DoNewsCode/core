package otkafka

import (
	"time"

	"github.com/segmentio/kafka-go"
)

// ReaderConfig is a configuration object used to create new instances of
// Reader.
type ReaderConfig struct {
	// The list of broker addresses used to connect to the kafka cluster.
	Brokers []string `json:"brokers" yaml:"brokers"`

	// GroupID holds the optional consumer group id.  If GroupID is specified, then
	// Partition should NOT be specified e.g. 0
	GroupID string `json:"groupId" yaml:"groupID"`

	// The topic to read messages from.
	Topic string `json:"topic" yaml:"topic"`

	// Partition to read messages from.  Either Partition or GroupID may
	// be assigned, but not both
	Partition int `json:"partition" yaml:"partition"`

	// The capacity of the internal message queue, defaults to 100 if none is
	// set.
	QueueCapacity int `json:"queue_capacity" yaml:"queue_capacity"`

	// Min and max number of bytes to fetch from kafka factoryIn each request.
	MinBytes int `json:"minBytes" yaml:"minBytes"`
	MaxBytes int `json:"maxBytes" yaml:"maxBytes"`

	// Maximum amount of time to wait for new data to come when fetching batches
	// of messages from kafka.
	MaxWait time.Duration `json:"maxWait" yaml:"maxWait"`

	// ReadLagInterval sets the frequency at which the reader lag is updated.
	// Setting this field to a negative value disables lag reporting.
	ReadLagInterval time.Duration `json:"readLagInterval" yaml:"readLagInterval"`

	// HeartbeatInterval sets the optional frequency at which the reader sends the consumer
	// group heartbeat update.
	//
	// Default: 3s
	//
	// Only used when GroupID is set
	HeartbeatInterval time.Duration `json:"heartbeatInterval" yaml:"heartbeatInterval"`

	// CommitInterval indicates the interval at which offsets are committed to
	// the broker.  If 0, commits will be handled synchronously.
	//
	// Default: 0
	//
	// Only used when GroupID is set
	CommitInterval time.Duration `json:"commitInterval" yaml:"commitInterval"`

	// PartitionWatchInterval indicates how often a reader checks for partition changes.
	// If a reader sees a partition change (such as a partition add) it will rebalance the group
	// picking up new partitions.
	//
	// Default: 5s
	//
	// Only used when GroupID is set and WatchPartitionChanges is set.
	PartitionWatchInterval time.Duration `json:"partitionWatchInterval" yaml:"partitionWatchInterval"`

	// WatchForPartitionChanges is used to inform kafka-go that a consumer group should be
	// polling the brokers and rebalancing if any partition changes happen to the topic.
	WatchPartitionChanges bool `json:"watchPartitionChanges" yaml:"watchPartitionChanges"`

	// SessionTimeout optionally sets the length of time that may pass without a heartbeat
	// before the coordinator considers the consumer dead and initiates a rebalance.
	//
	// Default: 30s
	//
	// Only used when GroupID is set
	SessionTimeout time.Duration `json:"sessionTimeout" yaml:"sessionTimeout"`

	// RebalanceTimeout optionally sets the length of time the coordinator will wait
	// for members to join as part of a rebalance.  For kafka servers under higher
	// load, it may be useful to set this value higher.
	//
	// Default: 30s
	//
	// Only used when GroupID is set
	RebalanceTimeout time.Duration `json:"rebalanceTimeout" yaml:"rebalanceTimeout"`

	// JoinGroupBackoff optionally sets the length of time to wait between re-joining
	// the consumer group after an error.
	//
	// Default: 5s
	JoinGroupBackoff time.Duration `json:"joinGroupBackoff" yaml:"joinGroupBackoff"`

	// RetentionTime optionally sets the length of time the consumer group will be saved
	// by the broker
	//
	// Default: 24h
	//
	// Only used when GroupID is set
	RetentionTime time.Duration `json:"retentionTime" yaml:"retentionTime"`

	// StartOffset determines from whence the consumer group should begin
	// consuming when it finds a partition without a committed offset.  If
	// non-zero, it must be set to one of FirstOffset or LastOffset.
	//
	// Default: FirstOffset
	//
	// Only used when GroupID is set
	StartOffset int64 `json:"startOffset" yaml:"startOffset"`

	// BackoffDelayMin optionally sets the smallest amount of time the reader will wait before
	// polling for new messages
	//
	// Default: 100ms
	ReadBackoffMin time.Duration `json:"readBackoffMin" yaml:"readBackoffMin"`

	// BackoffDelayMax optionally sets the maximum amount of time the reader will wait before
	// polling for new messages
	//
	// Default: 1s
	ReadBackoffMax time.Duration `json:"readBackoffMax" yaml:"readBackoffMax"`

	// Limit of how many attempts will be made before delivering the error.
	//
	// The default is to try 3 times.
	MaxAttempts int `json:"maxAttempts" yaml:"maxAttempts"`
}

// ReaderInterceptor is an interceptor that makes last minute change to a *kafka.ReaderConfig
// during kafka.Reader's creation
type ReaderInterceptor func(name string, reader *kafka.ReaderConfig)

func fromReaderConfig(conf ReaderConfig) kafka.ReaderConfig {
	if len(conf.Brokers) == 0 {
		conf.Brokers = []string{"127.0.0.1:9092"}
	}
	if len(conf.Topic) == 0 {
		conf.Topic = "default"
	}
	return kafka.ReaderConfig{
		Brokers:                conf.Brokers,
		GroupID:                conf.GroupID,
		Topic:                  conf.Topic,
		Partition:              conf.MaxAttempts,
		MinBytes:               conf.MinBytes,
		MaxBytes:               conf.MaxBytes,
		MaxWait:                conf.MaxWait,
		ReadLagInterval:        conf.ReadLagInterval,
		HeartbeatInterval:      conf.HeartbeatInterval,
		CommitInterval:         conf.CommitInterval,
		PartitionWatchInterval: conf.PartitionWatchInterval,
		WatchPartitionChanges:  conf.WatchPartitionChanges,
		SessionTimeout:         conf.SessionTimeout,
		RebalanceTimeout:       conf.RebalanceTimeout,
		JoinGroupBackoff:       conf.JoinGroupBackoff,
		RetentionTime:          conf.RetentionTime,
		StartOffset:            conf.StartOffset,
		ReadBackoffMin:         conf.ReadBackoffMin,
		ReadBackoffMax:         conf.ReadBackoffMax,
		MaxAttempts:            conf.MaxAttempts,
	}
}
