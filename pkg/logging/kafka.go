package logging

import (
	"fmt"

	"github.com/go-kit/kit/log"
)

type KafkaLogAdapter struct {
	Logging log.Logger
}

func (k KafkaLogAdapter) Printf(s string, i ...interface{}) {
	k.Logging.Log("msg", fmt.Sprintf(s, i...))
}

