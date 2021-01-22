package kitkafka

import (
	"context"

	"github.com/oklog/run"
	"github.com/segmentio/kafka-go"
)

type HandleFunc func(ctx context.Context, msg kafka.Message) error

func (h HandleFunc) Handle(ctx context.Context, msg kafka.Message) error {
	return h(ctx, msg)
}

type Handler interface {
	Handle(ctx context.Context, msg kafka.Message) error
}

type Server interface {
	Serve(ctx context.Context) error
}

type sub struct {
	reader      *kafka.Reader
	handler     Handler
	parallelism int
}

func (s *sub) ServeOnce(ctx context.Context) error {
	msg, err := s.reader.ReadMessage(ctx)
	if err != nil {
		return err
	}
	// User space error will not result in a transport error
	_ = s.handler.Handle(ctx, msg)
	return nil
}

func (s *sub) Serve(ctx context.Context) error {
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

type Mux struct {
	servers []Server
}

func NewMux(servers ...Server) Mux {
	return Mux{servers}
}

func (m Mux) Serve(ctx context.Context) error {
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

type pub struct {
	*kafka.Writer
}

func (p *pub) Handle(ctx context.Context, msg kafka.Message) error {
	return p.Writer.WriteMessages(ctx, msg)
}
