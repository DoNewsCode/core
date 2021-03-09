package otes

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

//LogAdapter is an adapter between kitlog and es Logger interface

//info logs informational messages
type esInfoLogAdapter struct {
	logger log.Logger
}

// Printf implements elastic.Logger
func (l esInfoLogAdapter) Printf(msg string, v ...interface{}) {
	level.Info(l.logger).Log("msg", fmt.Sprintf(msg, v...))
}

//error logs to the error log
type esErrorLogAdapter struct {
	logger log.Logger
}

// Printf implements elastic.Logger
func (l esErrorLogAdapter) Printf(msg string, v ...interface{}) {
	level.Error(l.logger).Log("msg", fmt.Sprintf(msg, v...))
}

//trace log for debugging
type esTraceLogAdapter struct {
	logger log.Logger
}

// Printf implements elastic.Logger
func (l esTraceLogAdapter) Printf(msg string, v ...interface{}) {
	level.Debug(l.logger).Log("msg", fmt.Sprintf(msg, v...))
}
