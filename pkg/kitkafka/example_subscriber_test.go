package kitkafka_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/DoNewsCode/std/pkg/config"
	"github.com/DoNewsCode/std/pkg/kitkafka"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/segmentio/kafka-go"
)

// StringService provides operations on strings.
type StringService interface {
	Uppercase(string) (string, error)
	Count(string) int
}

// stringService is a concrete implementation of StringService
type stringService struct{}

func (stringService) Uppercase(s string) (string, error) {
	if s == "" {
		return "", ErrEmpty
	}
	fmt.Println(strings.ToUpper(s))
	return strings.ToUpper(s), nil
}

func (stringService) Count(s string) int {
	return len(s)
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
		v, err := svc.Uppercase(req.S)
		if err != nil {
			return uppercaseResponse{v, err.Error()}, nil
		}
		return uppercaseResponse{v, ""}, nil
	}
}

func makeCountEndpoint(svc StringService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		req := request.(countRequest)
		v := svc.Count(req.S)
		return countResponse{v}, nil
	}
}

func Example_subscriber() {

	sendTestData()

	svc := stringService{}

	uppercaseHandler := kitkafka.NewSubscriber(
		makeUppercaseEndpoint(svc),
		decodeUppercaseRequest,
	)

	countHandler := kitkafka.NewSubscriber(
		makeCountEndpoint(svc),
		decodeCountRequest,
	)

	factory, cleanup := kitkafka.ProvideKafkaReaderFactory(kitkafka.KafkaIn{
		Conf: config.MapAdapter{"kafka.reader": map[string]kitkafka.ReaderConfig{
			"uppercase": {
				Brokers: []string{"127.0.0.1:9092"},
				Topic:   "uppercase",
				GroupID: "kitkafka",
			},
			"count": {
				Brokers: []string{"127.0.0.1:9092"},
				Topic:   "count",
				GroupID: "kitkafka",
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

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
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
	writer := kafka.Writer{
		Addr:  kafka.TCP("127.0.0.1:9092"),
		Topic: "uppercase",
	}

	writer.WriteMessages(context.Background(), kafka.Message{
		Value: []byte(`{"s":"kitkafka"}`),
	})
}
