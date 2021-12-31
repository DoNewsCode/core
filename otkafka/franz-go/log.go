package franz_go

import (
	"github.com/go-kit/log"
	"github.com/twmb/franz-go/pkg/kgo"
)

// KafkaLogAdapter is an log adapter bridging kitlog and kafka.
type KafkaLogAdapter struct {
	Logging log.Logger
}

func (w *KafkaLogAdapter) Level() kgo.LogLevel {
	return kgo.LogLevelDebug
}

func (w *KafkaLogAdapter) Log(level kgo.LogLevel, msg string, keyvals ...interface{}) {
	if w.Level() > level {
		return
	}
	kvs := []interface{}{
		"msg", msg,
	}
	kvs = append(kvs, keyvals...)
	_ = w.Logging.Log(kvs...)
}
