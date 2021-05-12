package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/otkafka"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

type testHandler struct {
	data []*testData
}
type testData struct {
	Id int
}

func (h *testHandler) Info() *Info {
	return &Info{
		Name:      "default",
		BatchSize: 3,
	}
}

func (h *testHandler) Handle(ctx context.Context, msg *kafka.Message) (interface{}, error) {
	e := &testData{}
	if err := json.Unmarshal(msg.Value, &e); err != nil {
		return nil, err
	}
	return e, nil
}

func (h *testHandler) Batch(ctx context.Context, data []interface{}) error {
	for _, e := range data {
		h.data = append(h.data, e.(*testData))
	}
	return nil
}

func TestNew(t *testing.T) {
	c := core.New(
		core.WithInline("kafka.reader.default.brokers", strings.Split(os.Getenv("KAFKA_ADDR"), ",")),
		core.WithInline("kafka.reader.default.topic", "processor"),
		core.WithInline("kafka.reader.default.groupID", "test"),
		core.WithInline("kafka.reader.default.startOffset", -1),
		core.WithInline("kafka.writer.default.brokers", strings.Split(os.Getenv("KAFKA_ADDR"), ",")),
		core.WithInline("kafka.writer.default.topic", "processor"),
		core.WithInline("http.disable", "true"),
		core.WithInline("grpc.disable", "true"),
		core.WithInline("cron.disable", "true"),
		core.WithInline("log.level", "none"),
	)
	c.ProvideEssentials()
	c.Provide(otkafka.Providers())
	handler := &testHandler{[]*testData{}}
	c.Provide(di.Deps{
		func() Out {
			return NewOut(handler)
		},
	})
	c.AddModuleFunc(New)

	var messageCount = 4

	assert.Greater(t, messageCount, handler.Info().batchSize())

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
	go func() {
		err := c.Serve(ctx)
		if err != nil {
			t.Error(err)
		}
	}()

	for {
		if len(handler.data) == handler.Info().batchSize() {
			cancel()
			break
		}
	}
	// wait graceful shutdown
	time.Sleep(2 * time.Second)

	assert.Equal(t, messageCount, len(handler.data))
}
