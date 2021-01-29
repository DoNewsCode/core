package otgorm

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"gorm.io/gorm/logger"
)

type GormLogAdapter struct {
	Logging log.Logger
}

func (g GormLogAdapter) LogMode(logLevel logger.LogLevel) logger.Interface {
	panic("Setting GORM LogMode is not allowed for kit log")
}

func (g GormLogAdapter) Info(ctx context.Context, s string, i ...interface{}) {
	level.Info(g.Logging).Log("msg", fmt.Sprintf(s, i...))
}

func (g GormLogAdapter) Warn(ctx context.Context, s string, i ...interface{}) {
	level.Warn(g.Logging).Log("msg", fmt.Sprintf(s, i...))
}

func (g GormLogAdapter) Error(ctx context.Context, s string, i ...interface{}) {
	level.Error(g.Logging).Log("msg", fmt.Sprintf(s, i...))
}

func (g GormLogAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, rows := fc()
	elapsed := time.Since(begin)

	var l log.Logger
	if err == nil {
		l = level.Debug(g.Logging)
	} else {
		l = level.Warn(g.Logging)
	}
	if rows == -1 {
		l.Log("sql", sql, "duration", elapsed, "rows", "-", "err", err)
	} else {
		l.Log("sql", sql, "duration", elapsed, "rows", rows, "err", err)
	}
}
