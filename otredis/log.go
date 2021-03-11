package otredis

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
)

// redisLogAdapter is an adapter between kitlog and redis logger interface
type RedisLogAdapter struct {
	Logging log.Logger
}

// Printf implements internal.Logging
func (r RedisLogAdapter) Printf(ctx context.Context, s string, i ...interface{}) {
	_ = r.Logging.Log("msg", fmt.Sprintf(s, i...))
}
