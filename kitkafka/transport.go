package kitkafka

import (
	"context"
	"math/rand"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/oklog/run"
	"github.com/segmentio/kafka-go"
)

// HandleFunc is a functional Handler.
type HandleFunc func(ctx context.Context, msg kafka.Message) error

// Handle deals with the kafka.Message in some way.
func (h HandleFunc) Handle(ctx context.Context, msg kafka.Message) error {
	return h(ctx, msg)
}

// Handler is a symmetric interface for both kafka publication and subscription.
// As a publisher handler, it is responsible to writes the kafka message to kafka brokers.
// As a subscriber handler, it is responsible to pipe kafka message to endpoints layer.
// in go kit analog, this is a go kit transport.
type Handler interface {
	Handle(ctx context.Context, msg kafka.Message) error
}

// Server models a kafka server. It will block until context canceled. Server usually start
// serving when the application boot up.
type Server interface {
	Serve(ctx context.Context) error
}

// SubscriberServer is a kafka server that continuously consumes messages from
// kafka. It implements Server. The SubscriberServer internally uses a fan out
// model, where only one goroutine communicate with kafka, but distribute
// messages to many parallel worker goroutines. However, this means manual offset
// commit is also impossible, making it not suitable for tasks that demands
// strict consistency. An option, WithSyncCommit is provided, for such high
// consistency tasks. in Sync Commit mode, Server synchronously commit offset to
// kafka when the error returned by the Handler is Nil.
type SubscriberServer struct {
	reader      Reader
	handler     Handler
	parallelism int
	syncCommit  bool
}

func (s *SubscriberServer) serveOnce(ctx context.Context) error {
	msg, err := s.reader.ReadMessage(ctx)
	if err != nil {
		return err
	}
	// User space error will not result in a transport error
	_ = s.handler.Handle(ctx, msg)
	return nil
}

func (s *SubscriberServer) serveAsync(ctx context.Context) error {
	var (
		g  run.Group
		ch chan kafka.Message
	)
	ch = make(chan kafka.Message)
	ctx, cancel := context.WithCancel(ctx)
	g.Add(func() error {
		for {
			msg, err := s.reader.ReadMessage(ctx)
			if err != nil {
				return err
			}
			ch <- msg
		}
	}, func(err error) {
		cancel()
		_ = s.reader.Close()
	})

	for i := 0; i < s.parallelism; i++ {
		g.Add(func() error {
			for {
				select {
				case msg := <-ch:
					_ = s.handler.Handle(ctx, msg)
				case <-ctx.Done():
					return nil
				}
			}
		}, func(err error) {
			cancel()
		})
	}
	return g.Run()
}

func (s *SubscriberServer) serveSync(ctx context.Context) error {
	var g run.Group
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i := 0; i < s.parallelism; i++ {
		g.Add(func() error {
			var d time.Duration
		loop:
			for {
				msg, err := s.reader.FetchMessage(ctx)
				if err != nil {
					return err
				}
				err = s.handler.Handle(ctx, msg)
				if err != nil {
					d = getRetryDuration(d)
					select {
					case <-time.After(d):
						continue loop
					case <-ctx.Done():
						return ctx.Err()
					}
				}

				// when using sync commit, the commit cannot be cancelled by original context.
				// Intentionally creates a new context here.
				err = s.reader.CommitMessages(context.Background(), msg)

				// retry commit
				for err != nil {
					d = getRetryDuration(d)
					<-time.After(d)
					select {
					case <-time.After(d):
						err = s.reader.CommitMessages(context.Background(), msg)
					case <-ctx.Done():
						return ctx.Err()
					}
				}
				d = 0
			}
		}, func(err error) {
			cancel()
		})
	}

	return g.Run()
}

// Serve starts the Server. *SubscriberServer will connect to kafka immediately
// and continuously consuming messages from it. Note Serve uses consumer groups,
// so Serve can be called on multiple node for the same topic without manually
// balancing partitions.
func (s *SubscriberServer) Serve(ctx context.Context) error {
	if s.syncCommit {
		return s.serveSync(ctx)
	}
	return s.serveAsync(ctx)
}

// SubscriberClientMux is a group of kafka Server. Useful when consuming from multiple topics.
type SubscriberClientMux struct {
	servers []Server
}

// NewMux creates a SubscriberClientMux, which is a group of kafka servers.
func NewMux(servers ...Server) SubscriberClientMux {
	return SubscriberClientMux{servers}
}

// Serve calls the Serve method in parallel for each server in the
// SubscriberClientMux. It blocks until any of the servers returns.
func (m SubscriberClientMux) Serve(ctx context.Context) error {
	var g run.Group
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for _, server := range m.servers {
		s := server
		g.Add(func() error {
			return s.Serve(ctx)
		}, func(err error) {
			cancel()
		})
	}
	return g.Run()
}

// PublisherService is a go kit service with one method, publish.
type PublisherService struct {
	endpoint endpoint.Endpoint
}

// Publish sends the request to kafka.
func (p PublisherService) Publish(ctx context.Context, request interface{}) error {
	_, err := p.endpoint(ctx, request)
	return err
}

type writerHandle struct {
	*kafka.Writer
}

func (p *writerHandle) Handle(ctx context.Context, msg kafka.Message) error {
	return p.Writer.WriteMessages(ctx, msg)
}

func getRetryDuration(d time.Duration) time.Duration {
	d *= 2
	jitter := rand.Float64() + 0.5
	d = time.Duration(int64(float64(d.Nanoseconds()) * jitter))
	if d > 10*time.Second {
		d = 10 * time.Second
	}
	if d < time.Second {
		d = time.Second
	}
	return d
}
