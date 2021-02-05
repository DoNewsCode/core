package contract

import "github.com/go-kit/kit/log"

type LevelLogger interface {
	log.Logger
	Debug(string)
	Info(string)
	Warn(string)
	Err(error)
	CheckErr(error)
}
