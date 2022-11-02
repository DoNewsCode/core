package otes

import (
	"fmt"

	"github.com/go-kit/log"
)

// ElasticLogAdapter is an adapter between kitlog and elastic logger interface
type ElasticLogAdapter struct {
	Logging   log.Logger
	LimitSize int
}

// Printf implements elastic.Logger
func (l ElasticLogAdapter) Printf(msg string, v ...any) {
	m := fmt.Sprintf(msg, v...)
	if l.LimitSize > 0 && len(m) > l.LimitSize {
		_ = l.Logging.Log("msg", m[:l.LimitSize])
		return
	}
	_ = l.Logging.Log("msg", m)
}
