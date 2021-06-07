package processor

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/otkafka"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

type testData struct {
	ID int `json:"id"`
}

type testHandlerA struct {
	data chan *testData
}

func (h *testHandlerA) Info() *Info {
	return &Info{
		Name:      "A",
		BatchSize: 3,
	}
}

func (h *testHandlerA) Handle(ctx context.Context, msg *kafka.Message) (interface{}, error) {
	e := &testData{}
	if err := json.Unmarshal(msg.Value, &e); err != nil {
		return nil, err
	}
	return e, nil
}

func (h *testHandlerA) Batch(ctx context.Context, data []interface{}) error {
	for _, e := range data {
		h.data <- e.(*testData)
	}
	return nil
}

type testHandlerB struct {
	data chan *testData
}

func (h *testHandlerB) Info() *Info {
	return &Info{
		Name:      "B",
		BatchSize: 3,
	}
}

func (h *testHandlerB) Handle(ctx context.Context, msg *kafka.Message) (interface{}, error) {
	e := &testData{}
	if err := json.Unmarshal(msg.Value, &e); err != nil {
		return nil, err
	}
	return e, nil
}

func (h *testHandlerB) Batch(ctx context.Context, data []interface{}) error {
	for _, e := range data {
		h.data <- e.(*testData)
	}
	return nil
}

type testHandlerC struct {
	data chan *testData
}

func (h *testHandlerC) Info() *Info {
	return &Info{
		Name:      "C",
		BatchSize: 3,
	}
}

func (h *testHandlerC) Handle(ctx context.Context, msg *kafka.Message) (interface{}, error) {
	e := &testData{}
	if err := json.Unmarshal(msg.Value, &e); err != nil {
		return nil, err
	}
	h.data <- e
	return nil, nil
}

type testHandlerD struct {
	data chan *testData
}

func (h *testHandlerD) Info() *Info {
	return &Info{
		Name:      "D",
		BatchSize: 3,
	}
}

func (h *testHandlerD) Handle(ctx context.Context, msg *kafka.Message) (interface{}, error) {
	e := &testData{}
	if err := json.Unmarshal(msg.Value, &e); err != nil {
		return nil, err
	}
	h.data <- e
	return nil, nil
}

type testHandlerE struct {
	data chan *testData
}

func (h *testHandlerE) Info() *Info {
	return &Info{
		Name:              "default",
		BatchSize:         3,
		AutoBatchInterval: 1 * time.Second,
	}
}

func (h *testHandlerE) Handle(ctx context.Context, msg *kafka.Message) (interface{}, error) {
	e := &testData{}
	if err := json.Unmarshal(msg.Value, &e); err != nil {
		return nil, err
	}
	return e, nil
}

func (h *testHandlerE) Batch(ctx context.Context, data []interface{}) error {
	for _, e := range data {
		h.data <- e.(*testData)
	}
	return nil
}

type testHandlerF struct {
	data chan *testData
}

func (h *testHandlerF) Info() *Info {
	return &Info{
		Name:      "default",
		BatchSize: 3,
	}
}

func (h *testHandlerF) Handle(ctx context.Context, msg *kafka.Message) (interface{}, error) {
	e := &testData{}
	if err := json.Unmarshal(msg.Value, &e); err != nil {
		return nil, err
	}
	return e, nil
}

func (h *testHandlerF) Batch(ctx context.Context, data []interface{}) error {
	return errors.New("test error")
}

func TestProcessor(t *testing.T) {
	c := core.New(
		core.WithInline("kafka.reader.A.brokers", envDefaultKafkaAddrs),
		core.WithInline("kafka.reader.A.topic", "processor"),
		core.WithInline("kafka.reader.A.groupID", "testA"),
		core.WithInline("kafka.reader.A.startOffset", kafka.FirstOffset),

		core.WithInline("kafka.reader.B.brokers", envDefaultKafkaAddrs),
		core.WithInline("kafka.reader.B.topic", "processor"),
		core.WithInline("kafka.reader.B.groupID", "testB"),
		core.WithInline("kafka.reader.B.startOffset", kafka.FirstOffset),

		core.WithInline("kafka.reader.C.brokers", envDefaultKafkaAddrs),
		core.WithInline("kafka.reader.C.topic", "processor"),
		core.WithInline("kafka.reader.C.groupID", "testC"),
		core.WithInline("kafka.reader.C.startOffset", kafka.FirstOffset),

		core.WithInline("kafka.reader.D.brokers", envDefaultKafkaAddrs),
		core.WithInline("kafka.reader.D.topic", "processor"),
		core.WithInline("kafka.reader.D.groupID", "testD"),
		core.WithInline("kafka.reader.D.startOffset", kafka.FirstOffset),

		core.WithInline("http.disable", "true"),
		core.WithInline("grpc.disable", "true"),
		core.WithInline("cron.disable", "true"),
		core.WithInline("log.level", "none"),
	)
	defer c.Shutdown()

	c.ProvideEssentials()
	c.Provide(otkafka.Providers())
	handlerA := &testHandlerA{make(chan *testData, 100)}
	handlerB := &testHandlerB{make(chan *testData, 100)}
	handlerC := &testHandlerC{make(chan *testData, 100)}
	handlerD := &testHandlerD{make(chan *testData, 100)}
	defer func() {
		close(handlerA.data)
		close(handlerB.data)
		close(handlerC.data)
		close(handlerD.data)
	}()

	c.Provide(di.Deps{
		func() Out {
			return NewOut(
				handlerA,
				handlerB,
				handlerC,
				handlerD,
			)
		},
	})
	c.AddModuleFunc(New)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := c.Serve(ctx)
	if err != nil {
		t.Error(err)
	}
	assert.NotZero(t, len(handlerA.data))
	assert.NotZero(t, len(handlerB.data))
	assert.Equal(t, 0, len(handlerA.data)%3)
	assert.Equal(t, 0, len(handlerB.data)%3)
	assert.Equal(t, 4, len(handlerC.data))
	assert.Equal(t, 4, len(handlerD.data))
}

func TestProcessorBatchInterval(t *testing.T) {
	c := core.New(
		core.WithInline("kafka.reader.default.brokers", envDefaultKafkaAddrs),
		core.WithInline("kafka.reader.default.topic", "processor"),
		core.WithInline("kafka.reader.default.groupID", "testE"),
		core.WithInline("kafka.reader.default.startOffset", kafka.FirstOffset),

		core.WithInline("http.disable", "true"),
		core.WithInline("grpc.disable", "true"),
		core.WithInline("cron.disable", "true"),
		core.WithInline("log.level", "none"),
	)
	defer c.Shutdown()
	c.ProvideEssentials()
	c.Provide(otkafka.Providers())

	handler := &testHandlerE{make(chan *testData, 100)}
	defer func() {
		close(handler.data)
	}()

	c.Provide(di.Deps{
		func() Out {
			return NewOut(
				handler,
			)
		},
	})

	c.AddModuleFunc(New)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	g := sync.WaitGroup{}
	g.Add(1)
	go func() {
		err := c.Serve(ctx)
		if err != nil {
			t.Error(err)
		}
		g.Done()
	}()
	g.Add(1)
	var count = 0
	go func() {
		defer g.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-handler.data:
				count++
				if count >= 4 {
					return
				}
			}
		}
	}()

	g.Wait()
	assert.Equal(t, 4, count)
}

func TestProcessorBatchError(t *testing.T) {
	c := core.New(
		core.WithInline("kafka.reader.default.brokers", envDefaultKafkaAddrs),
		core.WithInline("kafka.reader.default.topic", "processor"),
		core.WithInline("kafka.reader.default.groupID", "testF"),
		core.WithInline("kafka.reader.default.startOffset", kafka.FirstOffset),

		core.WithInline("http.disable", "true"),
		core.WithInline("grpc.disable", "true"),
		core.WithInline("cron.disable", "true"),
		core.WithInline("log.level", "none"),
	)
	defer c.Shutdown()
	c.ProvideEssentials()
	c.Provide(otkafka.Providers())

	handler := &testHandlerF{make(chan *testData, 100)}
	defer func() {
		close(handler.data)
	}()
	c.Provide(di.Deps{
		func() Out {
			return NewOut(
				handler,
			)
		},
	})

	c.AddModuleFunc(New)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := c.Serve(ctx)
	assert.Error(t, err)
	assert.Equal(t, "test error", err.Error())
}
