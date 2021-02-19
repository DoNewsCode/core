package kitkafka_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DoNewsCode/std/pkg/config"
	"github.com/DoNewsCode/std/pkg/kitkafka"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/segmentio/kafka-go"
)

// stringService is a concrete implementation of StringService
type remoteStringService struct {
	uppercase endpoint.Endpoint
	count     endpoint.Endpoint
}

func (s remoteStringService) Uppercase(ctx context.Context, str string) (string, error) {
	return s.Uppercase(ctx, str)
}

func (s remoteStringService) Count(ctx context.Context, str string) (string, error) {
	return s.Uppercase(ctx, str)
}

func Example_publisher() {
	factory, cleanup := kitkafka.ProvideKafkaWriterFactory(kitkafka.KafkaIn{
		Conf: config.MapAdapter{"kafka.writer": map[string]kitkafka.WriterConfig{
			"uppercase": {
				Brokers: []string{"127.0.0.1:9092"},
				Topic:   "uppercase",
			},
			"count": {
				Brokers: []string{"127.0.0.1:9092"},
				Topic:   "count",
			},
		}},
		Logger: log.NewNopLogger(),
	})
	defer cleanup()

	uppercaseClient, _ := factory.MakeClient("uppercase")
	countClient, _ := factory.MakeClient("count")

	uppercaseEndpoint := kitkafka.NewPublisher(uppercaseClient, encodeJSONRequest).Endpoint()
	countEndpoint := kitkafka.NewPublisher(countClient, encodeJSONRequest).Endpoint()

	svc := remoteStringService{uppercaseEndpoint, countEndpoint}

	_, err := svc.count(context.Background(), "kitkafka")
	fmt.Println(err)
	// Output:
	// <nil>
}

func encodeJSONRequest(_ context.Context, message *kafka.Message, i interface{}) error {
	bytes, err := json.Marshal(i)
	if err != nil {
		return err
	}
	message.Value = bytes
	return nil
}
