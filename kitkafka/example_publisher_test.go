// +build integration

package kitkafka_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/kitkafka"
	"github.com/DoNewsCode/core/otkafka"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/segmentio/kafka-go"
)

// stringService is a concrete implementation of StringService
type remoteStringService struct {
	uppercase endpoint.Endpoint
	count     endpoint.Endpoint
}

type remoteUppercaseRequest struct {
	S string `json:"s"`
}

type remoteCountRequest struct {
	S string `json:"s"`
}

func (s remoteStringService) Uppercase(ctx context.Context, str string) error {
	request := remoteUppercaseRequest{S: str}
	_, err := s.uppercase(ctx, &request)
	if err != nil {
		return err
	}
	return nil
}

func (s remoteStringService) Count(ctx context.Context, str string) {
	request := remoteCountRequest{S: str}
	_, _ = s.count(ctx, &request)
}

func Example_publisher() {
	// Create topic
	kafka.DialLeader(context.Background(), "tcp", "127.0.0.1:9092", "count", 0)

	c := core.Default(core.SetConfigProvider(func(configStack []config.ProviderSet, configWatcher contract.ConfigWatcher) contract.ConfigAccessor {
		return config.MapAdapter{"kafka.writer": map[string]otkafka.WriterConfig{
			"uppercase": {
				Brokers: []string{"127.0.0.1:9092"},
				Topic:   "uppercase",
			},
			"count": {
				Brokers: []string{"127.0.0.1:9092"},
				Topic:   "count",
			},
		}}
	}), core.SetLoggerProvider(func(conf contract.ConfigAccessor, appName contract.AppName, env contract.Env) log.Logger {
		return log.NewNopLogger()
	}))
	defer c.Shutdown()
	c.Provide(otkafka.Providers())

	c.Invoke(func(maker otkafka.WriterFactory) {
		uppercaseWriter, _ := maker.Make("uppercase")
		countWriter, _ := maker.Make("count")
		uppercaseClient, _ := kitkafka.MakeClient(uppercaseWriter)
		countClient, _ := kitkafka.MakeClient(countWriter)

		uppercaseEndpoint := kitkafka.NewPublisher(uppercaseClient, encodeJSONRequest).Endpoint()
		countEndpoint := kitkafka.NewPublisher(countClient, encodeJSONRequest).Endpoint()

		svc := remoteStringService{uppercaseEndpoint, countEndpoint}

		svc.Count(context.Background(), "kitkafka")

		received := getLastMessage()
		fmt.Println(received)
	})

	// Output:
	// {"s":"kitkafka"}
}

func encodeJSONRequest(_ context.Context, message *kafka.Message, i interface{}) error {
	bytes, err := json.Marshal(i)
	if err != nil {
		return err
	}
	message.Value = bytes
	return nil
}

func getLastMessage() string {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"127.0.0.1:9092"},
		GroupID:  "kitkafka",
		Topic:    "count",
		MaxBytes: 1,
	})
	m, _ := r.FetchMessage(context.Background())
	_ = r.CommitMessages(context.Background(), m)
	return string(m.Value)
}
