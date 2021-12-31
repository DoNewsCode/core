package franz_go

import (
	"context"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
	"github.com/twmb/franz-go/pkg/kversion"
	"github.com/twmb/franz-go/pkg/sasl"
	"net"
	"time"
)

// Config is a configuration object used to create new instances of
// kgo.Opt.
type Config struct {
	Id                     string // client ID
	DialFn                 func(context.Context, string, string) (net.Conn, error)
	RequestTimeoutOverhead time.Duration
	ConnIdleTimeout        time.Duration

	SoftwareName    string // KIP-511
	SoftwareVersion string // KIP-511

	Logger kgo.Logger

	SeedBrokers []string
	MaxVersions *kversion.Versions
	MinVersions *kversion.Versions

	RetryBackoff func(int) time.Duration
	Retries      int
	RetryTimeout time.Duration

	MaxBrokerWriteBytes int32
	MaxBrokerReadBytes  int32

	AllowAutoTopicCreation bool

	MetadataMaxAge time.Duration
	MetadataMinAge time.Duration

	Sasls []sasl.Mechanism

	Hooks []kgo.Hook

	//////////////////////
	// PRODUCER SECTION //
	//////////////////////

	TxnID               string
	TxnTimeout          time.Duration
	Acks                int16
	DisableIdempotency  bool
	Compression         []kgo.CompressionCodec // order of preference
	DefaultProduceTopic string
	MaxRecordBatchBytes int32
	MaxBufferedRecords  int
	ProduceTimeout      time.Duration
	RecordRetries       int
	Linger              time.Duration
	RecordTimeout       time.Duration
	ManualFlushing      bool
	Partitioner         kgo.Partitioner
	StopOnDataLoss      bool
	OnDataLoss          func(string, int32)

	//////////////////////
	// CONSUMER SECTION //
	//////////////////////

	MaxWait              time.Duration
	MinBytes             int32
	MaxBytes             int32
	MaxPartBytes         int32
	ResetOffset          kgo.Offset
	IsolationLevel       int8
	KeepControl          bool
	Rack                 string
	MaxConcurrentFetches int
	DisableFetchSessions bool
	Topics               []string                        // topics to consume; if regex is true, values are compiled regular expressions
	Partitions           map[string]map[int32]kgo.Offset // partitions to directly consume from
	Regex                bool

	////////////////////////////
	// CONSUMER GROUP SECTION //
	////////////////////////////

	Group              string              // group we are in
	InstanceID         string              // optional group instance ID
	Balancers          []kgo.GroupBalancer // balancers we can use
	Protocol           string              // "consumer" by default, expected to never be overridden
	SessionTimeout     time.Duration
	RebalanceTimeout   time.Duration
	HeartbeatInterval  time.Duration
	RequireStable      bool
	OnAssigned         func(context.Context, *kgo.Client, map[string][]int32)
	OnRevoked          func(context.Context, *kgo.Client, map[string][]int32)
	OnLost             func(context.Context, *kgo.Client, map[string][]int32)
	AutocommitDisable  bool // true if autocommit was disabled or we are transactional
	AutocommitGreedy   bool
	AutocommitMarks    bool
	AutocommitInterval time.Duration
	CommitCallback     func(*kgo.Client, *kmsg.OffsetCommitRequest, *kmsg.OffsetCommitResponse, error)
}

func newConfig() Config {
	return Config{
		Acks: -1,
	}
}

func fromConfig(conf Config) (opts []kgo.Opt) {
	if conf.Id != "" {
		opts = append(opts, kgo.ClientID(conf.Id))
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
	if conf.ResetOffset.String() != "" {
		opts = append(opts, kgo.ConsumeResetOffset(conf.ResetOffset))
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
