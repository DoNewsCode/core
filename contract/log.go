package contract

import "github.com/go-kit/log"

// Logger is an alias of go kit logger
type Logger = log.Logger

// LevelLogger is plaintext logger with level.
type LevelLogger interface {
	Logger
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Err(args ...any)
	Debugf(template string, args ...any)
	Infof(template string, args ...any)
	Warnf(template string, args ...any)
	Errf(template string, args ...any)
	Debugw(msg string, fields ...any)
	Infow(msg string, fields ...any)
	Warnw(msg string, fields ...any)
	Errw(msg string, fields ...any)
}
