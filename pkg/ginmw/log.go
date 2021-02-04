package ginmw

import (
	"time"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/key"
	"github.com/DoNewsCode/std/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

func WithLogger(logger log.Logger, keyer contract.Keyer, notlogged ...string) gin.HandlerFunc {
	var skip map[string]struct{}

	if length := len(notlogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range notlogged {
			skip[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		l := logging.WithContext(level.Debug(logger), c.Request.Context())
		l = log.With(l, key.SpreadInterface(keyer))

		c.Next()

		if _, ok := skip[path]; !ok {
			end := time.Now()
			latency := end.Sub(start)

			method := c.Request.Method
			statusCode := c.Writer.Status()

			l.Log(
				"HTTPVerb", method,
				"statusCode", statusCode,
				"latency", latency,
			)
		}
	}
}
