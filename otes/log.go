package otes

import (
	"fmt"

	"github.com/go-kit/log"
)

// ElasticLogAdapter is an adapter between kitlog and elastic logger interface
type ElasticLogAdapter struct {
	Logging log.Logger
}

// Printf implements elastic.Logger
func (l ElasticLogAdapter) Printf(msg string, v ...interface{}) {
	_ = l.Logging.Log("msg", fmt.Sprintf(msg, v...))
}
