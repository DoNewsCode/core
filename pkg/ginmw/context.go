package ginmw

import (
	"context"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/gin-gonic/gin"
)

func WithContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c, contract.TransportKey, "HTTPJSON")
		ctx = context.WithValue(ctx, contract.RequestUrlKey, c.Request.URL.Path)
		ctx = context.WithValue(ctx, contract.IpKey, c.ClientIP())
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
