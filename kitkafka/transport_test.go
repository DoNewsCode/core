// +build integration

package kitkafka

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/otkafka"
	"github.com/stretchr/testify/assert"

	"github.com/DoNewsCode/core/logging"
	"github.com/segmentio/kafka-go"
)

func TestTransport(t *testing.T) {

	kafka.DialLeader(context.Background(), "tcp", "localhost:9092", "Test", 0)

	writerFactory, cleanupWriter := otkafka.provideWriterFactory(otkafka.in{
		Conf: config.MapAdapter{"kafka.writer": map[string]otkafka.WriterConfig{
			"default": {
				Brokers: []string{"127.0.0.1:9092"},
				Topic:   "test",
			},
		}},
		Logger: logging.NewLogger("logfmt"),
	})
	defer cleanupWriter()

	readerFactory, cleanupReader := otkafka.provideReaderFactory(otkafka.in{
		Conf: config.MapAdapter{"kafka.reader": map[string]otkafka.ReaderConfig{
			"default": {
				Brokers: []string{"127.0.0.1:9092"},
				Topic:   "test",
			},
		}},
		Logger: logging.NewLogger("logfmt"),
	})
	defer cleanupReader()

	// write test data
	h, err := writerFactory.MakeClient("default")
	assert.NoError(t, err)

	err = h.Handle(context.Background(), kafka.Message{
		Value: []byte("hello"),
	})
	assert.NoError(t, err)

	// consume test data
	var consumed = false

	endpoint := func(ctx context.Context, message interface{}) (interface{}, error) {
		if message.(string) != "hello" {
			t.Fatalf("want hello, got %s", message)
		}
		consumed = true
		return nil, nil
	}
	sub, err := readerFactory.MakeSubscriberServer("default", NewSubscriber(endpoint, func(ctx context.Context, message *kafka.Message) (request interface{}, err error) {
		return string(message.Value), nil
	}))

	assert.NoError(t, err)

	err = sub.serveOnce(context.Background())

	assert.NoError(t, err)

	if !consumed {
		t.Fatal("failed to consume the message")
	}
}

type mockReader struct {
	fetchCount int
	hasCommit  bool
	commitFunc func(ctx context.Context, msgs ...kafka.Message) error
}

func (r *mockReader) Close() error {
	return nil
}

func (r *mockReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	panic("implement me")
}

func (r *mockReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	select {
	case <-ctx.Done():
		return kafka.Message{}, ctx.Err()
	default:
		r.fetchCount++
		return kafka.Message{}, nil
	}
}

func (r *mockReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	err := r.commitFunc(ctx, msgs...)
	if err == nil {
		r.hasCommit = true
	}
	return err
}

func TestServeSync(t *testing.T) {
	var i = 0
	for _, c := range []struct {
		name    string
		handler Handler
		reader  Reader
		asserts func(t *testing.T, reader *mockReader)
	}{
		{
			"failed",
			HandleFunc(func(ctx context.Context, msg kafka.Message) error {
				return errors.New("false")
			}),
			&mockReader{},
			func(t *testing.T, reader *mockReader) {
				assert.Equal(t, false, reader.hasCommit)
				assert.Equal(t, 1, reader.fetchCount)
			},
		}, {
			"retrying",
			HandleFunc(func(ctx context.Context, msg kafka.Message) error {
				return errors.New("false")
			}),
			&mockReader{},
			func(t *testing.T, reader *mockReader) {
				assert.Equal(t, false, reader.hasCommit)
				assert.Less(t, 1, reader.fetchCount)
			},
		},
		{
			"commit failed",
			HandleFunc(func(ctx context.Context, msg kafka.Message) error {
				return nil
			}),
			&mockReader{
				commitFunc: func(ctx context.Context, msgs ...kafka.Message) error {
					return errors.New("foo")
				},
			},
			func(t *testing.T, reader *mockReader) {
				assert.Equal(t, false, reader.hasCommit)
				assert.Less(t, 1, reader.fetchCount)
			},
		},
		{
			"commit fixed",
			HandleFunc(func(ctx context.Context, msg kafka.Message) error {
				return nil
			}),
			&mockReader{
				commitFunc: func(ctx context.Context, msgs ...kafka.Message) error {
					if i == 0 {
						i++
						return errors.New("foo")
					}
					return nil
				},
			},

			func(t *testing.T, reader *mockReader) {
				assert.Equal(t, false, reader.hasCommit)
				assert.Less(t, 1, reader.fetchCount)
			},
		},
	} {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			server := SubscriberServer{
				reader:      c.reader,
				handler:     c.handler,
				parallelism: 1,
				syncCommit:  true,
			}
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

			defer cancel()
			_ = server.Serve(ctx)
		})
	}
}

func TestGetRetryDuration(t *testing.T) {
	assert.Equal(t, 10*time.Second, getRetryDuration(time.Hour))
	assert.Equal(t, time.Second, getRetryDuration(time.Microsecond))
	assert.LessOrEqual(t, time.Second, getRetryDuration(time.Second))
	assert.GreaterOrEqual(t, 10*time.Second, getRetryDuration(time.Second))
}
