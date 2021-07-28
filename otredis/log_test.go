package otredis

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
)

func TestRedisLogAdapter_Printf(t *testing.T) {
	var buf bytes.Buffer
	l := RedisLogAdapter{Logging: log.NewLogfmtLogger(&buf)}
	l.Printf(context.Background(), "test %s", "logger")
	assert.Equal(t, "msg=\"test logger\"\n", buf.String())
}
