package ginmw

import (
	"context"

	"github.com/DoNewsCode/core/contract"
	"github.com/gin-gonic/gin"
)

// Context is a gin middleware that adds request context to contract keys.
func Context() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c, contract.TransportKey, "HTTP")
		ctx = context.WithValue(ctx, contract.RequestUrlKey, c.Request.URL.Path)
		ctx = context.WithValue(ctx, contract.IpKey, c.ClientIP())
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
