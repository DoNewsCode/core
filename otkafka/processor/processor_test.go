package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/otkafka"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

type testData struct {
	Id int
}

type testHandlerA struct {
	data []*testData
	lock sync.Mutex
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
	h.lock.Lock()
	defer h.lock.Unlock()
	for _, e := range data {
		h.data = append(h.data, e.(*testData))
	}
	return nil
}

type testHandlerB struct {
	data []*testData
	lock sync.Mutex
}

func (h *testHandlerB) Info() *Info {
	return &Info{
		Name:       "B",
		BatchSize:  3,
		AutoCommit: true,
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
	h.lock.Lock()
	defer h.lock.Unlock()
	for _, e := range data {
		h.data = append(h.data, e.(*testData))
	}
	return nil
}

type testHandlerC struct {
	data []*testData
	lock sync.Mutex
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
	h.lock.Lock()
	defer h.lock.Unlock()
	h.data = append(h.data, e)
	return nil, nil
}

type testHandlerD struct {
	data []*testData
	lock sync.Mutex
}

func (h *testHandlerD) Info() *Info {
	return &Info{
		Name:       "D",
		BatchSize:  3,
		AutoCommit: true,
	}
}

func (h *testHandlerD) Handle(ctx context.Context, msg *kafka.Message) (interface{}, error) {
	e := &testData{}
	if err := json.Unmarshal(msg.Value, &e); err != nil {
		return nil, err
	}
	h.lock.Lock()
	defer h.lock.Unlock()
	h.data = append(h.data, e)
	return nil, nil
}

func TestProcessor(t *testing.T) {
	c := core.New(
		core.WithInline("kafka.reader.A.brokers", envDefaultKafkaAddrs),
		core.WithInline("kafka.reader.A.topic", "processor1"),
		core.WithInline("kafka.reader.A.groupID", "testA"),
		core.WithInline("kafka.reader.A.startOffset", -1),

		core.WithInline("kafka.reader.B.brokers", envDefaultKafkaAddrs),
		core.WithInline("kafka.reader.B.topic", "processor1"),
		core.WithInline("kafka.reader.B.groupID", "testB"),
		core.WithInline("kafka.reader.B.startOffset", -1),

		core.WithInline("kafka.reader.C.brokers", envDefaultKafkaAddrs),
		core.WithInline("kafka.reader.C.topic", "processor1"),
		core.WithInline("kafka.reader.C.groupID", "testC"),
		core.WithInline("kafka.reader.C.startOffset", -1),

		core.WithInline("kafka.reader.D.brokers", envDefaultKafkaAddrs),
		core.WithInline("kafka.reader.D.topic", "processor1"),
		core.WithInline("kafka.reader.D.groupID", "testD"),
		core.WithInline("kafka.reader.D.startOffset", -1),

		core.WithInline("kafka.writer.default.brokers", envDefaultKafkaAddrs),
		core.WithInline("kafka.writer.default.topic", "processor1"),
		core.WithInline("http.disable", "true"),
		core.WithInline("grpc.disable", "true"),
		core.WithInline("cron.disable", "true"),
		core.WithInline("log.level", "none"),
	)
	c.ProvideEssentials()
	c.Provide(otkafka.Providers())
	handlerA := &testHandlerA{[]*testData{}, sync.Mutex{}}
	handlerB := &testHandlerB{[]*testData{}, sync.Mutex{}}
	handlerC := &testHandlerC{[]*testData{}, sync.Mutex{}}
	handlerD := &testHandlerD{[]*testData{}, sync.Mutex{}}
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

	var messageCount = 4

	c.Invoke(func(w *kafka.Writer) {
		testMessages := make([]kafka.Message, 0)
		for i := 0; i < messageCount; i++ {
			testMessages = append(testMessages, kafka.Message{Value: []byte(fmt.Sprintf(`{"id":%d}`, i))})
		}
		err := w.WriteMessages(context.Background(), testMessages...)
		if err != nil {
			t.Fatal(err)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	g := sync.WaitGroup{}
	g.Add(1)
	go func() {
		err := c.Serve(ctx)
		if err != nil {
			t.Error(err)
		}
		g.Done()
	}()

	time.Sleep(1 * time.Second)
	cancel()

	g.Wait()

	assert.Equal(t, messageCount, len(handlerA.data))
	assert.Equal(t, messageCount, len(handlerB.data))
	assert.Equal(t, messageCount, len(handlerC.data))
	assert.Equal(t, messageCount, len(handlerD.data))
}

func TestProcessorGracefulShutdown(t *testing.T) {
	c := core.New(
		core.WithInline("kafka.reader.A.brokers", envDefaultKafkaAddrs),
		core.WithInline("kafka.reader.A.topic", "processor2"),
		core.WithInline("kafka.reader.A.groupID", "testE"),
		core.WithInline("kafka.reader.A.startOffset", -1),

		core.WithInline("kafka.writer.default.brokers", envDefaultKafkaAddrs),
		core.WithInline("kafka.writer.default.topic", "processor2"),
		core.WithInline("http.disable", "true"),
		core.WithInline("grpc.disable", "true"),
		core.WithInline("cron.disable", "true"),
		core.WithInline("log.level", "none"),
	)
	c.ProvideEssentials()
	c.Provide(otkafka.Providers())

	handlerA := &testHandlerA{[]*testData{}, sync.Mutex{}}
	c.Provide(di.Deps{
		func() Out {
			return NewOut(
				handlerA,
			)
		},
	})

	c.AddModuleFunc(New)

	var messageCount = 4

	c.Invoke(func(w *kafka.Writer) {
		testMessages := make([]kafka.Message, 0)
		for i := 0; i < messageCount; i++ {
			testMessages = append(testMessages, kafka.Message{Value: []byte(fmt.Sprintf(`{"id":%d}`, i))})
		}
		err := w.WriteMessages(context.Background(), testMessages...)
		if err != nil {
			t.Fatal(err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
	go func() {
		defer g.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				handlerA.lock.Lock()
				length := len(handlerA.data)
				handlerA.lock.Unlock()
				if length >= handlerA.Info().batchSize() {
					cancel()
					return
				}
			}
		}
	}()
	g.Wait()

	assert.Equal(t, messageCount, len(handlerA.data))
}
