package processor

import (
	"context"

	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/otkafka"
	"github.com/go-kit/kit/log"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

// Processor dispatch Handler.
type Processor struct {
	maker    otkafka.ReaderMaker
	handlers map[string]SimpleHandler
	logger   log.Logger
}

// SimpleHandler only include Info and Handle func.
type SimpleHandler interface {
	// Info set the topic name and some config.
	Info() *Info
	// Handle for *kafka.Message.
	Handle(ctx context.Context, msg *kafka.Message) (interface{}, error)
}

// Handler one more Batch method than SimpleHandler.
type Handler interface {
	SimpleHandler
	// Batch processing the results returned by SimpleHandler.Handle.
	Batch(ctx context.Context, data []interface{}) error
}

// HandleFunc type for SimpleHandler.Handle Func.
type HandleFunc func(ctx context.Context, msg *kafka.Message) (interface{}, error)

// BatchFunc type for Handler.Batch Func.
type BatchFunc func(ctx context.Context, data []interface{}) error

type in struct {
	di.In

	Hs     []SimpleHandler `group:"H"`
	Maker  otkafka.ReaderMaker
	Logger log.Logger
}

// New create *Processor Module.
func New(i in) (*Processor, error) {
	e := &Processor{
		maker:    i.Maker,
		logger:   i.Logger,
		handlers: map[string]SimpleHandler{},
	}
	if len(i.Hs) == 0 {
		return nil, errors.New("empty handler list")
	}
	for _, hh := range i.Hs {
		if err := e.addHandler(hh); err != nil {
			return nil, err
		}
	}
	return e, nil
}

type Out struct {
	di.Out

	Hs []SimpleHandler `group:"H,flatten"`
}

// NewOut for create di.Out.
// 	Usage:
// 		func newHandler(logger log.Logger) processor.Out {
//			return processor.NewOut(
//				&HandlerA{logger: logger},
//			)
//		}
// 	Or
// 		func newHandlers(logger log.Logger) processor.Out {
//			return processor.NewOut(
//				&HandlerA{logger: logger},
//				&HandlerB{logger: logger},
//			)
//		}
func NewOut(handler ...SimpleHandler) Out {
	return Out{Hs: handler}
}

func (e *Processor) addHandler(h SimpleHandler) error {
	name := h.Info().name()
	_, err := e.maker.Make(name)
	if err != nil {
		return err
	}

	e.handlers[name] = h

	return nil
}

// batchInfo data is the result of message processed by SimpleHandler.Handle.
//
// When Handler.Batch is successfully called, then commit the message.
type batchInfo struct {
	message *kafka.Message
	data    interface{}
}

// ProvideRunGroup run workers:
// 	1. Fetch message from *kafka.Reader.
// 	2. Handle message by SimpleHandler.Handle.
// 	3. Batch data by Handler.Batch. If batch success then commit message.
func (e *Processor) ProvideRunGroup(group *run.Group) {
	if len(e.handlers) == 0 {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())

	var g run.Group
	msgChs := make([]chan *kafka.Message, 0)
	batchChs := make([]chan *batchInfo, 0)

	for name, ooo := range e.handlers {
		one := ooo

		msgCh := make(chan *kafka.Message, one.Info().chanSize())
		batchCh := make(chan *batchInfo, one.Info().chanSize())
		msgChs = append(msgChs, msgCh)
		batchChs = append(batchChs, batchCh)

		reader, _ := e.maker.Make(name)

		for i := 0; i < one.Info().readWorker(); i++ {
			g.Add(func() error {
				for {
					select {
					case <-ctx.Done():
						return nil
					default:
						message, err := reader.FetchMessage(ctx)
						if err != nil {
							return err
						}
						if len(message.Value) > 0 {
							msgCh <- &message
						}
					}
				}
			}, func(err error) {

			})
		}

		for i := 0; i < one.Info().handleWorker(); i++ {
			g.Add(func() error {
				for {
					select {
					case msg := <-msgCh:
						v, err := one.Handle(ctx, msg)
						if err != nil {
							return err
						}
						batchCh <- &batchInfo{message: msg, data: v}
					case <-ctx.Done():
						return nil
					}
				}
			}, func(err error) {

			})
		}

		if v, ok := one.(Handler); ok {
			for i := 0; i < one.Info().batchWorker(); i++ {
				g.Add(func() error {
					err := e.batch(ctx, reader, batchCh, v.Batch, one.Info().batchSize())
					if err != nil {
						return err
					}
					return nil
				}, func(err error) {

				})
			}
		}

	}

	group.Add(func() error {
		if err := g.Run(); err != nil {
			return err
		}
		return nil
	}, func(err error) {
		cancel()
		for _, ch := range msgChs {
			close(ch)
		}
		for _, ch := range batchChs {
			close(ch)
		}
	})

}

// batch Call Handler.Batch. It's graceful when shutdown.
func (e *Processor) batch(ctx context.Context, reader *kafka.Reader, ch chan *batchInfo, batchFunc BatchFunc, batchSize int) error {
	var data = make([]interface{}, 0)
	var messages = make([]kafka.Message, 0)

	doFunc := func() error {
		if err := batchFunc(ctx, data); err != nil {
			return err
		}
		if err := reader.CommitMessages(context.Background(), messages...); err != nil {
			return err
		}
		data = data[0:0]
		messages = messages[0:0]
		return nil
	}

	for {
		select {
		case v := <-ch:
			data = append(data, v.data)
			messages = append(messages, *v.message)
			if len(data) >= batchSize {
				if err := doFunc(); err != nil {
					return err
				}
			}
		case <-ctx.Done():
			for v := range ch {
				data = append(data, v.data)
				messages = append(messages, *v.message)
			}
			if len(data) > 0 {
				if err := doFunc(); err != nil {
					return err
				}
			}
			return nil
		}
	}
}
