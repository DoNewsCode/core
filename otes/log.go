package otes

import (
	"fmt"

	"github.com/go-kit/kit/log"
)

//info logs informational messages
type esLogAdapter struct {
	logger log.Logger
}

// Printf implements elastic.Logger
func (l esLogAdapter) Printf(msg string, v ...interface{}) {
	l.logger.Log("msg", fmt.Sprintf(msg, v...))
}
