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

// Processor dispatch BatchHandler.
type Processor struct {
	maker    otkafka.ReaderMaker
	handlers []*handler
	logger   log.Logger
	ctx      context.Context
	closers  []func()
}

// Handler only include Info and Handle func.
type Handler interface {
	// Info set the topic name and some config.
	Info() *Info
	// Handle for *kafka.Message.
	Handle(ctx context.Context, msg *kafka.Message) (interface{}, error)
}

// BatchHandler one more Batch method than Handler.
type BatchHandler interface {
	Handler
	// Batch processing the results returned by Handler.Handle.
	Batch(ctx context.Context, data []interface{}) error
}

// HandleFunc type for Handler.Handle Func.
type HandleFunc func(ctx context.Context, msg *kafka.Message) (interface{}, error)

// BatchFunc type for BatchHandler.Batch Func.
type BatchFunc func(ctx context.Context, data []interface{}) error

type in struct {
	di.In

	Hs     []Handler `group:"H"`
	Maker  otkafka.ReaderMaker
	Logger log.Logger
}

// New create *Processor Module.
func New(i in) (*Processor, error) {
	ctx, cancel := context.WithCancel(context.Background())

	e := &Processor{
		maker:    i.Maker,
		logger:   i.Logger,
		handlers: []*handler{},
		ctx:      ctx,
		closers:  []func(){cancel},
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

	Hs []Handler `group:"H,flatten"`
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
func NewOut(handler ...Handler) Out {
	return Out{Hs: handler}
}

// addHandler create handler and add to Processor.handlers
func (e *Processor) addHandler(h Handler) error {
	name := h.Info().name()
	reader, err := e.maker.Make(name)
	if err != nil {
		return err
	}

	var hd = &handler{
		msgCh:      make(chan *kafka.Message, h.Info().chanSize()),
		reader:     reader,
		handleFunc: h.Handle,
		info:       h.Info(),
		ctx:        e.ctx,
	}

	batchHandler, isBatchHandler := h.(BatchHandler)
	if isBatchHandler {
		hd.batchCh = make(chan *batchInfo, h.Info().chanSize())
		hd.batchFunc = batchHandler.Batch
	}

	e.handlers = append(e.handlers, hd)
	e.closers = append(e.closers, func() {
		if hd.msgCh != nil {
			close(hd.msgCh)
		}
		if hd.batchCh != nil {
			close(hd.batchCh)
		}
	})

	return nil
}

// batchInfo data is the result of message processed by Handler.Handle.
//
// When BatchHandler.Batch is successfully called, then commit the message.
type batchInfo struct {
	message *kafka.Message
	data    interface{}
}

// ProvideRunGroup run workers:
// 	1. Fetch message from *kafka.Reader.
// 	2. Handle message by Handler.Handle.
// 	3. Batch data by BatchHandler.Batch. If batch success then commit message.
func (e *Processor) ProvideRunGroup(group *run.Group) {
	if len(e.handlers) == 0 {
		return
	}

	var g run.Group

	for _, handler := range e.handlers {
		for i := 0; i < handler.info.readWorker(); i++ {
			g.Add(handler.read, func(err error) {})
		}

		for i := 0; i < handler.info.handleWorker(); i++ {
			g.Add(handler.handle, func(err error) {})
		}

		if handler.batchFunc != nil {
			batchFunc := handler.batch

			if handler.info.autoCommit() {
				batchFunc = handler.batchIgnoreCommit
			}
			for i := 0; i < handler.info.batchWorker(); i++ {
				g.Add(batchFunc, func(err error) {})
			}
		}

	}

	group.Add(func() error {
		if err := g.Run(); err != nil {
			return err
		}
		return nil
	}, func(err error) {
		for _, closer := range e.closers {
			closer()
		}
	})

}

// handler private processor
// todo It's a bit messy
type handler struct {
	reader *kafka.Reader

	batchCh chan *batchInfo
	msgCh   chan *kafka.Message

	handleFunc HandleFunc
	batchFunc  BatchFunc

	info *Info

	ctx context.Context
}

// read fetch message from kafka
func (h *handler) read() error {
	for {
		select {
		default:
			message, err := h.reader.FetchMessage(h.ctx)
			if err != nil {
				return err
			}

			if len(message.Value) > 0 {
				h.msgCh <- &message
			}
			if h.info.autoCommit() {
				if err := h.reader.CommitMessages(context.Background(), message); err != nil {
					return err
				}
			}
		case <-h.ctx.Done():
			return nil
		}
	}
}

// handle call Handler.Handle
func (h *handler) handle() error {
	for {
		select {
		case msg, ok := <-h.msgCh:
			if !ok {
				continue
			}
			v, err := h.handleFunc(h.ctx, msg)
			if err != nil {
				return err
			}
			if h.batchFunc != nil {
				if h.info.autoCommit() {
					h.batchCh <- &batchInfo{data: v}
				} else {
					h.batchCh <- &batchInfo{message: msg, data: v}
				}
			} else {
				if !h.info.autoCommit() {
					if err := h.reader.CommitMessages(context.Background(), *msg); err != nil {
						return err
					}
				}
			}
		case <-h.ctx.Done():
			return nil
		}
	}
}

// batch Call BatchHandler.Batch and commit *kafka.Message.
func (h *handler) batch() error {
	var data = make([]interface{}, 0)
	var messages = make([]kafka.Message, 0)

	doFunc := func() error {
		if err := h.batchFunc(h.ctx, data); err != nil {
			return err
		}

		if err := h.reader.CommitMessages(context.Background(), messages...); err != nil {
			return err
		}

		data = data[0:0]
		messages = messages[0:0]
		return nil
	}

	for {
		select {
		case v, ok := <-h.batchCh:
			if !ok {
				continue
			}
			data = append(data, v.data)
			messages = append(messages, *v.message)
			if len(data) >= h.info.batchSize() {
				if err := doFunc(); err != nil {
					return err
				}
			}
		case <-h.ctx.Done():
			for v := range h.batchCh {
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

// batchIgnoreCommit Call BatchHandler.Batch, don't need to commit *kafka.Message.
func (h *handler) batchIgnoreCommit() error {
	var data = make([]interface{}, 0)

	doFunc := func() error {
		if err := h.batchFunc(h.ctx, data); err != nil {
			return err
		}
		data = data[0:0]
		return nil
	}

	for {
		select {
		case v, ok := <-h.batchCh:
			if !ok {
				continue
			}
			data = append(data, v.data)
			if len(data) >= h.info.batchSize() {
				if err := doFunc(); err != nil {
					return err
				}
			}
		case <-h.ctx.Done():
			for v := range h.batchCh {
				data = append(data, v.data)
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
