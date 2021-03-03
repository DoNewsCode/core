package otes

import (
	"fmt"
	"github.com/go-kit/kit/log"
)

// esLogAdapter is an adapter between kitlog and es logger interface
type esLogAdapter struct {
	prefix string
	logger log.Logger
}

// Printf implements elastic.Logger
func (l esLogAdapter) Printf(msg string, v ...interface{}) {
	l.logger.Log(l.prefix, fmt.Sprintf(msg, v...))
}
