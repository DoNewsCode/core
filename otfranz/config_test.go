package otfranz

import (
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/stretchr/testify/assert"
)

func Test_fromConfig(t *testing.T) {
	// Parameters that can be set through the yaml file
	conf := Config{
		ID:                     "1",
		RequestTimeoutOverhead: 1,
		ConnIdleTimeout:        1,
		SoftwareName:           "1",
		SoftwareVersion:        "1",
		SeedBrokers:            []string{"foo"},
		Retries:                1,
		RetryTimeout:           1,
		MaxBrokerWriteBytes:    1,
		MaxBrokerReadBytes:     1,
		AllowAutoTopicCreation: true,
		MetadataMaxAge:         1,
		MetadataMinAge:         1,
		TxnID:                  "1",
		TxnTimeout:             1,
		Acks:                   1,
		DisableIdempotency:     true,
		DefaultProduceTopic:    "",
		MaxRecordBatchBytes:    1,
		MaxBufferedRecords:     1,
		ProduceTimeout:         1,
		RecordRetries:          1,
		Linger:                 1,
		RecordTimeout:          1,
		ManualFlushing:         true,
		StopOnDataLoss:         true,
		MaxWait:                1,
		MinBytes:               1,
		MaxBytes:               1,
		MaxPartBytes:           1,
		ResetOffset: struct {
			At       int64 `json:"at" yaml:"at"`
			Relative int64 `json:"relative" yaml:"relative"`
			Epoch    int32 `json:"epoch" yaml:"epoch"`
		}{
			At:       1,
			Relative: 2,
			Epoch:    3,
		},
		IsolationLevel:       1,
		KeepControl:          true,
		Rack:                 "1",
		MaxConcurrentFetches: 1,
		DisableFetchSessions: true,
		Topics:               []string{"foo"},
		Regex:                true,
		Group:                "1",
		InstanceID:           "1",
		Protocol:             "1",
		SessionTimeout:       1,
		RebalanceTimeout:     1,
		HeartbeatInterval:    1,
		RequireStable:        true,
		AutocommitDisable:    true,
		AutocommitGreedy:     true,
		AutocommitMarks:      true,
		AutocommitInterval:   1,
	}
	opts := fromConfig(conf)
	assert.Len(t, opts, 47)
}

func Test_Config_Unmarshal(t *testing.T) {
	conf := Config{}
	kf := config.MapAdapter{"kafka": map[string]Config{
		"default": {
			SeedBrokers: []string{"foo"},
		},
	}}

	// There are many options that can not be decoded.
	// This test is necessary to prevent missing tags of "-".
	err := kf.Unmarshal("kafka.default", &conf)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []string{"foo"}, conf.SeedBrokers)
}
