package ginmw

import (
	"net/http"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/key"
	"github.com/gin-gonic/gin"
	tracegin "github.com/opentracing-contrib/go-gin/ginhttp"
	stdtracing "github.com/opentracing/opentracing-go"
)

// Trace is a gin middleware that adds opentracing support.
func Trace(tracer stdtracing.Tracer, keyer contract.Keyer) gin.HandlerFunc {
	return tracegin.Middleware(tracer, tracegin.OperationNameFunc(func(r *http.Request) string {
		return key.KeepOdd(keyer).Key(".", "http", r.Method)
	}))
}
