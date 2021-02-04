package ginmw

import (
	"net/http"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/key"
	"github.com/gin-gonic/gin"
	tracegin "github.com/opentracing-contrib/go-gin/ginhttp"
	stdtracing "github.com/opentracing/opentracing-go"
)

func WithTrace(tracer stdtracing.Tracer, keyer contract.Keyer) gin.HandlerFunc {
	return tracegin.Middleware(tracer, tracegin.OperationNameFunc(func(r *http.Request) string {
		return key.KeepOdd(keyer).Key(".", "http", r.Method)
	}))
}
