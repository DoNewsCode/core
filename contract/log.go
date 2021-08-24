package contract

import "github.com/go-kit/kit/log"

// Logger is an alias of go kit logger
type Logger = log.Logger

// LevelLogger is plaintext logger with level.
type LevelLogger interface {
	Logger
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Err(args ...interface{})
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errf(template string, args ...interface{})
	Debugw(msg string, fields ...interface{})
	Infow(msg string, fields ...interface{})
	Warnw(msg string, fields ...interface{})
	Errw(msg string, fields ...interface{})
}
