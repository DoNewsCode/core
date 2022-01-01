package franz_go

import (
	"github.com/go-kit/log"
	"github.com/twmb/franz-go/pkg/kgo"
)

// KafkaLogAdapter is an log adapter bridging kitlog and kgo.Logger.
type KafkaLogAdapter struct {
	Logging log.Logger
}

func (w *KafkaLogAdapter) Level() kgo.LogLevel {
	return kgo.LogLevelDebug
}

func (w *KafkaLogAdapter) Log(_ kgo.LogLevel, msg string, keyvals ...interface{}) {
	kvs := []interface{}{
		"msg", msg,
	}
	kvs = append(kvs, keyvals...)
	_ = w.Logging.Log(kvs...)
}
