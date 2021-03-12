package ginmw

import (
	"time"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/key"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/metrics"
)

// Metrics is a gin middleware that adds request histogram. Setting addPath
// to true will make histogram to use request path as a dimension. This is ok
// with few total number of paths, but incurs performance issue if the
// cardinality of request path is high.
func Metrics(hist metrics.Histogram, keyer contract.Keyer, addPath bool) gin.HandlerFunc {
	return func(c *gin.Context) {

		if addPath {
			keyer = key.With(keyer, "method", c.FullPath())
		} else {
			keyer = key.With(keyer, "method", "-")
		}

		defer func(begin time.Time) {
			hist.With(keyer.Spread()...).Observe(time.Since(begin).Seconds())
		}(time.Now())
		c.Next()
	}
}
