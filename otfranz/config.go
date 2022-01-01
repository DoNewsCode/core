package otfranz

import (
	"context"
	"net"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
	"github.com/twmb/franz-go/pkg/kversion"
	"github.com/twmb/franz-go/pkg/sasl"
)

// Config is a configuration object used to create new instances of *kgo.Client.
type Config struct {
	ID                     string                                                  `json:"id" yaml:"id"` // client ID
	DialFn                 func(context.Context, string, string) (net.Conn, error) `json:"-" yaml:"-"`
	RequestTimeoutOverhead time.Duration                                           `json:"request_timeout_overhead" yaml:"request_timeout_overhead"`
	ConnIdleTimeout        time.Duration                                           `json:"conn_idle_timeout" yaml:"conn_idle_timeout"`

	SoftwareName    string `json:"software_name" yaml:"software_name"`       // KIP-511
	SoftwareVersion string `json:"software_version" yaml:"software_version"` // KIP-511

	Logger kgo.Logger `json:"logger" yaml:"logger"`

	SeedBrokers []string           `json:"seed_brokers" yaml:"seed_brokers"`
	MaxVersions *kversion.Versions `json:"-" yaml:"-"`
	MinVersions *kversion.Versions `json:"-" yaml:"-"`

	RetryBackoff func(int) time.Duration `json:"-" yaml:"-"`
	Retries      int                     `json:"retries" yaml:"retries"`
	RetryTimeout time.Duration           `json:"retry_timeout" yaml:"retry_timeout"`

	MaxBrokerWriteBytes int32 `json:"max_broker_write_bytes" yaml:"max_broker_write_bytes"`
	MaxBrokerReadBytes  int32 `json:"max_broker_read_bytes" yaml:"max_broker_read_bytes"`

	AllowAutoTopicCreation bool `json:"allow_auto_topic_creation" yaml:"allow_auto_topic_creation"`

	MetadataMaxAge time.Duration `json:"metadata_max_age" yaml:"metadata_max_age"`
	MetadataMinAge time.Duration `json:"metadata_min_age" yaml:"metadata_min_age"`

	Sasls []sasl.Mechanism `json:"-" yaml:"-"`

	Hooks []kgo.Hook `json:"-" yaml:"-"`

	//////////////////////
	// PRODUCER SECTION //
	//////////////////////

	TxnID               string                 `json:"txn_id" yaml:"txn_id"`
	TxnTimeout          time.Duration          `json:"txn_timeout" yaml:"txn_timeout"`
	Acks                int16                  `json:"acks" yaml:"acks"`
	DisableIdempotency  bool                   `json:"disable_idempotency" yaml:"disable_idempotency"`
	Compression         []kgo.CompressionCodec `json:"-" yaml:"-"` // order of preference
	DefaultProduceTopic string                 `json:"default_produce_topic" yaml:"default_produce_topic"`
	MaxRecordBatchBytes int32                  `json:"max_record_batch_bytes" yaml:"max_record_batch_bytes"`
	MaxBufferedRecords  int                    `json:"max_buffered_records" yaml:"max_buffered_records"`
	ProduceTimeout      time.Duration          `json:"produce_timeout" yaml:"produce_timeout"`
	RecordRetries       int                    `json:"record_retries" yaml:"record_retries"`
	Linger              time.Duration          `json:"linger" yaml:"linger"`
	RecordTimeout       time.Duration          `json:"record_timeout" yaml:"record_timeout"`
	ManualFlushing      bool                   `json:"manual_flushing" yaml:"manual_flushing"`
	Partitioner         kgo.Partitioner        `json:"-" yaml:"-"`
	StopOnDataLoss      bool                   `json:"stop_on_data_loss" yaml:"stop_on_data_loss"`
	OnDataLoss          func(string, int32)    `json:"-" yaml:"-"`

	//////////////////////
	// CONSUMER SECTION //
	//////////////////////

	MaxWait      time.Duration `json:"max_wait" yaml:"max_wait"`
	MinBytes     int32         `json:"min_bytes" yaml:"min_bytes"`
	MaxBytes     int32         `json:"max_bytes" yaml:"max_bytes"`
	MaxPartBytes int32         `json:"max_part_bytes" yaml:"max_part_bytes"`

	ResetOffset struct {
		At       int64 `json:"at" yaml:"at"`
		Relative int64 `json:"relative" yaml:"relative"`
		Epoch    int32 `json:"epoch" yaml:"epoch"`
	} `json:"reset_offset" yaml:"reset_offset"`
	IsolationLevel       int8                            `json:"isolation_level" yaml:"isolation_level"`
	KeepControl          bool                            `json:"keep_control" yaml:"keep_control"`
	Rack                 string                          `json:"rack" yaml:"rack"`
	MaxConcurrentFetches int                             `json:"max_concurrent_fetches" yaml:"max_concurrent_fetches"`
	DisableFetchSessions bool                            `json:"disable_fetch_sessions" yaml:"disable_fetch_sessions"`
	Topics               []string                        `json:"topics" yaml:"topics"` // topics to consume; if regex is true, values are compiled regular expressions
	Partitions           map[string]map[int32]kgo.Offset `json:"-" yaml:"-"`           // partitions to directly consume from
	Regex                bool                            `json:"regex" yaml:"regex"`

	////////////////////////////
	// CONSUMER GROUP SECTION //
	////////////////////////////

	Group              string                                                                          `json:"group" yaml:"group"`             // group we are in
	InstanceID         string                                                                          `json:"instance_id" yaml:"instance_id"` // optional group instance ID
	Balancers          []kgo.GroupBalancer                                                             `json:"-" yaml:"-"`                     // balancers we can use
	Protocol           string                                                                          `json:"protocol" yaml:"protocol"`       // "consumer" by default, expected to never be overridden
	SessionTimeout     time.Duration                                                                   `json:"session_timeout" yaml:"session_timeout"`
	RebalanceTimeout   time.Duration                                                                   `json:"rebalance_timeout" yaml:"rebalance_timeout"`
	HeartbeatInterval  time.Duration                                                                   `json:"heartbeat_interval" yaml:"heartbeat_interval"`
	RequireStable      bool                                                                            `json:"require_stable" yaml:"require_stable"`
	OnAssigned         func(context.Context, *kgo.Client, map[string][]int32)                          `json:"-" yaml:"-"`
	OnRevoked          func(context.Context, *kgo.Client, map[string][]int32)                          `json:"-" yaml:"-"`
	OnLost             func(context.Context, *kgo.Client, map[string][]int32)                          `json:"-" yaml:"-"`
	AutocommitDisable  bool                                                                            `json:"autocommit_disable" yaml:"autocommit_disable"` // true if autocommit was disabled or we are transactional
	AutocommitGreedy   bool                                                                            `json:"autocommit_greedy" yaml:"autocommit_greedy"`
	AutocommitMarks    bool                                                                            `json:"autocommit_marks" yaml:"autocommit_marks"`
	AutocommitInterval time.Duration                                                                   `json:"autocommit_interval" yaml:"autocommit_interval"`
	CommitCallback     func(*kgo.Client, *kmsg.OffsetCommitRequest, *kmsg.OffsetCommitResponse, error) `json:"-" yaml:"-"`
}

func fromConfig(conf Config) (opts []kgo.Opt) {
	if conf.ID != "" {
		opts = append(opts, kgo.ClientID(conf.ID))
	}
	if conf.DialFn != nil {
		opts = append(opts, kgo.Dialer(conf.DialFn))
	}
	if conf.RequestTimeoutOverhead > 0 {
		opts = append(opts, kgo.RequestTimeoutOverhead(conf.RequestTimeoutOverhead))
	}
	if conf.ConnIdleTimeout > 0 {
		opts = append(opts, kgo.ConnIdleTimeout(conf.ConnIdleTimeout))
	}
	if conf.SoftwareName != "" && conf.SoftwareVersion != "" {
		opts = append(opts, kgo.SoftwareNameAndVersion(conf.SoftwareName, conf.SoftwareVersion))
	}

	if conf.Logger != nil {
		opts = append(opts, kgo.WithLogger(conf.Logger))
	}
	if len(conf.SeedBrokers) > 0 {
		opts = append(opts, kgo.SeedBrokers(conf.SeedBrokers...))
	}
	if conf.MaxVersions != nil {
		opts = append(opts, kgo.MaxVersions(conf.MaxVersions))
	}
	if conf.MinVersions != nil {
		opts = append(opts, kgo.MinVersions(conf.MinVersions))
	}
	if conf.RetryBackoff != nil {
		opts = append(opts, kgo.RetryBackoffFn(conf.RetryBackoff))
	}
	if conf.Retries > 0 {
		opts = append(opts, kgo.RequestRetries(conf.Retries))
	}
	if conf.RetryTimeout > 0 {
		opts = append(opts, kgo.RetryTimeout(conf.RetryTimeout))
	}
	if conf.MaxBrokerWriteBytes > 0 {
		opts = append(opts, kgo.ConnIdleTimeout(conf.ConnIdleTimeout))
	}
	if conf.MaxBrokerReadBytes > 0 {
		opts = append(opts, kgo.ConnIdleTimeout(conf.ConnIdleTimeout))
	}
	if conf.AllowAutoTopicCreation {
		opts = append(opts, kgo.AllowAutoTopicCreation())
	}
	if conf.MetadataMaxAge > 0 {
		opts = append(opts, kgo.MetadataMaxAge(conf.MetadataMaxAge))
	}
	if conf.MetadataMinAge > 0 {
		opts = append(opts, kgo.MetadataMinAge(conf.MetadataMinAge))
	}
	if len(conf.Sasls) > 0 {
		opts = append(opts, kgo.SASL(conf.Sasls...))
	}
	if len(conf.Hooks) > 0 {
		opts = append(opts, kgo.WithHooks(conf.Hooks))
	}

	if conf.TxnID != "" {
		opts = append(opts, kgo.TransactionalID(conf.TxnID))
	}
	if conf.TxnTimeout > 0 {
		opts = append(opts, kgo.TransactionTimeout(conf.TxnTimeout))
	}

	if conf.Acks == 0 {
		opts = append(opts, kgo.RequiredAcks(kgo.NoAck()))
	}
	if conf.Acks == 1 {
		opts = append(opts, kgo.RequiredAcks(kgo.LeaderAck()))
	}
	if conf.Acks == -1 {
		opts = append(opts, kgo.RequiredAcks(kgo.AllISRAcks()))
	}

	if conf.DisableIdempotency {
		opts = append(opts, kgo.DisableIdempotentWrite())
	} else {
		opts = append(opts, kgo.RequiredAcks(kgo.AllISRAcks()))
	}
	if len(conf.Compression) > 0 {
		opts = append(opts, kgo.ProducerBatchCompression(conf.Compression...))
	}
	if conf.DefaultProduceTopic != "" {
		opts = append(opts, kgo.DefaultProduceTopic(conf.DefaultProduceTopic))
	}
	if conf.MaxRecordBatchBytes > 0 {
		opts = append(opts, kgo.ProducerBatchMaxBytes(conf.MaxRecordBatchBytes))
	}
	if conf.MaxBufferedRecords > 0 {
		opts = append(opts, kgo.MaxBufferedRecords(conf.MaxBufferedRecords))
	}
	if conf.ProduceTimeout > 0 {
		opts = append(opts, kgo.ProduceRequestTimeout(conf.ProduceTimeout))
	}
	if conf.RecordRetries > 0 {
		opts = append(opts, kgo.RecordRetries(conf.RecordRetries))
	}
	if conf.Linger > 0 {
		opts = append(opts, kgo.ProducerLinger(conf.Linger))
	}
	if conf.RecordTimeout > 0 {
		opts = append(opts, kgo.RecordDeliveryTimeout(conf.RecordTimeout))
	}
	if conf.ManualFlushing {
		opts = append(opts, kgo.ManualFlushing())
	}
	if conf.Partitioner != nil {
		opts = append(opts, kgo.RecordPartitioner(conf.Partitioner))
	}
	if conf.StopOnDataLoss {
		opts = append(opts, kgo.StopProducerOnDataLossDetected())
	}
	if conf.OnDataLoss != nil {
		opts = append(opts, kgo.ProducerOnDataLossDetected(conf.OnDataLoss))
	}
	if conf.MaxWait > 0 {
		opts = append(opts, kgo.FetchMaxWait(conf.MaxWait))
	}
	if conf.MinBytes > 0 {
		opts = append(opts, kgo.FetchMinBytes(conf.MinBytes))
	}
	if conf.MaxBytes > 0 {
		opts = append(opts, kgo.FetchMaxBytes(conf.MaxBytes))
	}
	if conf.MaxPartBytes > 0 {
		opts = append(opts, kgo.FetchMaxPartitionBytes(conf.MaxPartBytes))
	}

	resetOffset := kgo.NewOffset()
	setResetOffset := false
	if conf.ResetOffset.At != 0 {
		resetOffset.At(conf.ResetOffset.At)
		setResetOffset = true
	}
	if conf.ResetOffset.Relative != 0 {
		resetOffset.Relative(conf.ResetOffset.Relative)
		setResetOffset = true
	}
	if conf.ResetOffset.Epoch != 0 {
		resetOffset.WithEpoch(conf.ResetOffset.Epoch)
		setResetOffset = true
	}
	if setResetOffset {
		opts = append(opts, kgo.ConsumeResetOffset(resetOffset))
	}

	if conf.IsolationLevel == 0 {
		opts = append(opts, kgo.FetchIsolationLevel(kgo.ReadUncommitted()))
	}
	if conf.IsolationLevel == 1 {
		opts = append(opts, kgo.FetchIsolationLevel(kgo.ReadCommitted()))
	}
	if conf.KeepControl {
		opts = append(opts, kgo.KeepControlRecords())
	}
	if conf.Rack != "" {
		opts = append(opts, kgo.Rack(conf.Rack))
	}
	if conf.MaxConcurrentFetches > 0 {
		opts = append(opts, kgo.MaxConcurrentFetches(conf.MaxConcurrentFetches))
	}
	if conf.DisableFetchSessions {
		opts = append(opts, kgo.DisableFetchSessions())
	}
	if len(conf.Topics) > 0 {
		opts = append(opts, kgo.ConsumeTopics(conf.Topics...))
	}
	if len(conf.Partitions) > 0 {
		opts = append(opts, kgo.ConsumePartitions(conf.Partitions))
	}
	if conf.Regex {
		opts = append(opts, kgo.ConsumeRegex())
	}
	if conf.Group != "" {
		opts = append(opts, kgo.ConsumerGroup(conf.Group))
	}
	if conf.InstanceID != "" {
		opts = append(opts, kgo.InstanceID(conf.InstanceID))
	}
	if len(conf.Balancers) > 0 {
		opts = append(opts, kgo.Balancers(conf.Balancers...))
	}
	if conf.Protocol != "" {
		opts = append(opts, kgo.GroupProtocol(conf.Protocol))
	}
	if conf.SessionTimeout > 0 {
		opts = append(opts, kgo.SessionTimeout(conf.SessionTimeout))
	}
	if conf.RebalanceTimeout > 0 {
		opts = append(opts, kgo.RebalanceTimeout(conf.RebalanceTimeout))
	}
	if conf.HeartbeatInterval > 0 {
		opts = append(opts, kgo.HeartbeatInterval(conf.HeartbeatInterval))
	}
	if conf.RequireStable {
		opts = append(opts, kgo.RequireStableFetchOffsets())
	}
	if conf.OnAssigned != nil {
		opts = append(opts, kgo.OnPartitionsAssigned(conf.OnAssigned))
	}
	if conf.OnRevoked != nil {
		opts = append(opts, kgo.OnPartitionsRevoked(conf.OnRevoked))
	}
	if conf.OnLost != nil {
		opts = append(opts, kgo.OnPartitionsLost(conf.OnLost))
	}

	if conf.AutocommitDisable {
		opts = append(opts, kgo.DisableAutoCommit())
	}
	if conf.AutocommitGreedy {
		opts = append(opts, kgo.GreedyAutoCommit())
	}
	if conf.AutocommitMarks {
		opts = append(opts, kgo.AutoCommitMarks())
	}
	if conf.AutocommitInterval > 0 {
		opts = append(opts, kgo.AutoCommitInterval(conf.AutocommitInterval))
	}
	if conf.CommitCallback != nil {
		opts = append(opts, kgo.AutoCommitCallback(conf.CommitCallback))
	}
	return
}
