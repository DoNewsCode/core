package contract

import "github.com/go-kit/kit/log"

// LevelLogger is plaintext logger with level.
type LevelLogger interface {
	log.Logger
	Debug(string)
	Info(string)
	Warn(string)
	Err(string)
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errf(string, ...interface{})
}
