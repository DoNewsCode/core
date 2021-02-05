package ginmw

import (
	"time"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/key"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/metrics"
)

func WithMetrics(hist metrics.Histogram, keyer contract.Keyer) gin.HandlerFunc {
	return func(c *gin.Context) {
		keyer = key.With(keyer, "method", "-")
		defer func(begin time.Time) {
			hist.With(keyer.Spread()...).Observe(time.Since(begin).Seconds())
		}(time.Now())
		c.Next()
	}
}
