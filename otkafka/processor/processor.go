package processor

import (
	"context"
	"sync"
	"time"

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
	handlers []*handler
	logger   log.Logger
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

	Handlers []Handler `group:"ProcessorHandler"`
	Maker    otkafka.ReaderMaker
	Logger   log.Logger
}

// New create *Processor Module.
func New(i in) (*Processor, error) {
	e := &Processor{
		maker:    i.Maker,
		logger:   i.Logger,
		handlers: []*handler{},
	}
	if len(i.Handlers) == 0 {
		return nil, errors.New("empty handler list")
	}
	for _, hh := range i.Handlers {
		if err := e.addHandler(hh); err != nil {
			return nil, err
		}
	}
	return e, nil
}

// Out to provide Handler to in.Handlers.
type Out struct {
	di.Out

	Handlers []Handler `group:"ProcessorHandler,flatten"`
}

// NewOut create Out to provide Handler to in.Handlers.
// 	Usage:
// 		func newHandlerA(logger log.Logger) processor.Out {
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
func NewOut(handlers ...Handler) Out {
	return Out{Handlers: handlers}
}

// addHandler create handler and add to Processor.handlers
func (e *Processor) addHandler(h Handler) error {
	name := h.Info().name()
	reader, err := e.maker.Make(name)
	if err != nil {
		return err
	}

	if  reader.Config().GroupID == ""{
		return errors.New("")
	}

	batchFunc := func(ctx context.Context, data []interface{}) error {
		return nil
	}
	batchHandler, isBatchHandler := h.(BatchHandler)
	if isBatchHandler {
		batchFunc = batchHandler.Batch
	}

	var hd = &handler{
		msgCh:      make(chan *kafka.Message, h.Info().chanSize()),
		reader:     reader,
		handleFunc: h.Handle,
		batchFunc:  batchFunc,
		info:       h.Info(),
		once:       sync.Once{},
		batchCh:    make(chan *batchInfo, h.Info().chanSize()),
		ticker:     time.NewTicker(h.Info().autoBatchInterval()),
	}

	e.handlers = append(e.handlers, hd)

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

	ctx, cancel := context.WithCancel(context.Background())

	for _, one := range e.handlers {
		handler := one
		for i := 0; i < handler.info.readWorker(); i++ {
			g.Add(func() error {
				return handler.read(ctx)
			}, func(err error) {
				cancel()
				handler.close()
			})
		}

		for i := 0; i < handler.info.handleWorker(); i++ {
			g.Add(func() error {
				return handler.handle(ctx)
			}, func(err error) {
				cancel()
				handler.close()
			})
		}

		if handler.batchFunc != nil {
			for i := 0; i < handler.info.batchWorker(); i++ {
				g.Add(func() error {
					return handler.batch(ctx)
				}, func(err error) {
					cancel()
					handler.close()
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
	})

}

// handler private processor
// todo It's a bit messy
type handler struct {
	reader     *kafka.Reader
	batchCh    chan *batchInfo
	msgCh      chan *kafka.Message
	handleFunc HandleFunc
	batchFunc  BatchFunc
	info       *Info
	ticker     *time.Ticker
	once       sync.Once
}

// read fetch message from kafka
func (h *handler) read(ctx context.Context) error {
	for {
		select {
		default:
			message, err := h.reader.FetchMessage(ctx)
			if err != nil {
				return err
			}
			if len(message.Value) > 0 {
				h.msgCh <- &message
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// handle call Handler.Handle
func (h *handler) handle(ctx context.Context) error {
	for {
		select {
		case msg := <-h.msgCh:
			v, err := h.handleFunc(ctx, msg)
			if err != nil {
				return err
			}
			h.batchCh <- &batchInfo{message: msg, data: v}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// batch Call BatchHandler.Batch and commit *kafka.Message.
func (h *handler) batch(ctx context.Context) error {
	var data = make([]interface{}, 0)
	var messages = make([]kafka.Message, 0)

	appendData := func(d *batchInfo) {
		if d == nil {
			return
		}
		if d.message != nil {
			messages = append(messages, *d.message)
		}
		if d.data != nil {
			data = append(data, d.data)
		}
	}

	doFunc := func() error {
		if len(data) == 0 {
			return nil
		}
		defer func() {
			data = data[0:0]
			messages = messages[0:0]
		}()
		if err := h.batchFunc(ctx, data); err != nil {
			return err
		}

		if err := h.commit(messages...); err != nil {
			return err
		}
		return nil
	}

	for {
		select {
		case v := <-h.batchCh:
			appendData(v)
			if len(data) < h.info.batchSize() {
				continue
			}
			if err := doFunc(); err != nil {
				return err
			}
		case <-h.ticker.C:
			if err := doFunc(); err != nil {
				return err
			}
		case <-ctx.Done():
			for v := range h.batchCh {
				appendData(v)
			}
			if err := doFunc(); err != nil {
				return err
			}
			return ctx.Err()
		}
	}
}

func (h *handler) close() {
	h.once.Do(func() {
		if h.msgCh != nil {
			close(h.msgCh)
		}
		if h.batchCh != nil {
			close(h.batchCh)
		}
		if h.ticker != nil {
			h.ticker.Stop()
		}
	})
}

func (h *handler) commit(messages ...kafka.Message) error {
	if len(messages) > 0 {
		if err := h.reader.CommitMessages(context.Background(), messages...); err != nil {
			return err
		}
	}
	return nil
}
