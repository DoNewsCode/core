// +build integration

package kitkafka_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/kitkafka"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/segmentio/kafka-go"
)

// StringService provides operations on strings.
type StringService interface {
	Uppercase(string) error
	Count(string)
}

// stringService is a concrete implementation of StringService
type stringService struct{}

func (stringService) Uppercase(s string) error {
	if s == "" {
		return ErrEmpty
	}
	fmt.Println(strings.ToUpper(s))
	return nil
}

func (stringService) Count(s string) {
	fmt.Println(len(s))
}

// ErrEmpty is returned when an input string is empty.
var ErrEmpty = errors.New("empty string")

// For each method, we define request and response structs
type uppercaseRequest struct {
	S string `json:"s"`
}

type uppercaseResponse struct {
	V   string `json:"v"`
	Err string `json:"err,omitempty"` // errors don't define JSON marshaling
}

type countRequest struct {
	S string `json:"s"`
}

type countResponse struct {
	V int `json:"v"`
}

// Endpoints are a primary abstraction in go-kit. An endpoint represents a single RPC (method in our service interface)
func makeUppercaseEndpoint(svc StringService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(uppercaseRequest)
		err := svc.Uppercase(req.S)
		// We return error here so that error handler can log/handle this error.
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
}

func makeCountEndpoint(svc StringService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(countRequest)
		svc.Count(req.S)
		return nil, nil
	}
}

func Example_subscriber() {

	sendTestData()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := stringService{}

	uppercaseHandler := kitkafka.NewSubscriber(
		makeUppercaseEndpoint(svc),
		decodeUppercaseRequest,
		kitkafka.SubscriberAfter(func(ctx context.Context, _ *kafka.Message) context.Context {
			cancel()
			return ctx
		}),
	)

	countHandler := kitkafka.NewSubscriber(
		makeCountEndpoint(svc),
		decodeCountRequest,
		kitkafka.SubscriberAfter(func(ctx context.Context, _ *kafka.Message) context.Context {
			cancel()
			return ctx
		}),
	)

	factory, cleanup := kitkafka.ProvideReaderFactory(kitkafka.KafkaIn{
		Conf: config.MapAdapter{"kafka.reader": map[string]kitkafka.ReaderConfig{
			"uppercase": {
				Brokers:     []string{"127.0.0.1:9092"},
				GroupID:     "kitkafka",
				Topic:       "uppercase",
				StartOffset: kafka.FirstOffset,
			},
			"count": {
				Brokers:     []string{"127.0.0.1:9092"},
				Topic:       "count",
				GroupID:     "kitkafka",
				StartOffset: kafka.FirstOffset,
			},
		}},
		Logger: log.NewNopLogger(),
	})
	defer cleanup()

	uppercaseServer, err := factory.MakeSubscriberServer("uppercase", uppercaseHandler)
	if err != nil {
		panic(err)
	}
	countServer, err := factory.MakeSubscriberServer("count", countHandler)
	if err != nil {
		panic(err)
	}

	mux := kitkafka.NewMux(uppercaseServer, countServer)
	mux.Serve(ctx)

	// Output:
	// KITKAFKA
}

func decodeUppercaseRequest(_ context.Context, r *kafka.Message) (interface{}, error) {
	var request uppercaseRequest
	if err := json.NewDecoder(bytes.NewReader(r.Value)).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeCountRequest(_ context.Context, r *kafka.Message) (interface{}, error) {
	var request countRequest
	if err := json.NewDecoder(bytes.NewReader(r.Value)).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func sendTestData() {
	// to create topics when auto.create.topics.enable='true'
	kafka.DialLeader(context.Background(), "tcp", "localhost:9092", "my-topic", 0)

	writer := kafka.Writer{
		Addr:      kafka.TCP("127.0.0.1:9092"),
		Topic:     "uppercase",
		BatchSize: 1,
	}
	err := writer.WriteMessages(context.Background(), kafka.Message{
		Value: []byte(`{"s":"kitkafka"}`),
	})
	if err != nil {
		panic(err)
	}
}
