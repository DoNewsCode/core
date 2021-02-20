package ginmw

import (
	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/DoNewsCode/std/pkg/key"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/opentracing/opentracing-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

func ExampleWithContext() {
	g := gin.New()
	g.Use(WithContext())
	g.Handle("GET", "/", func(context *gin.Context) {
		context.String(200, "the request path is %s",
			context.Request.Context().Value(contract.RequestUrlKey))
	})
}

func ExampleWithTrace() {
	g := gin.New()
	g.Use(WithTrace(opentracing.GlobalTracer(), key.New("module", "foo")))
	g.Handle("GET", "/", func(context *gin.Context) {
		// Do stuff
	})
}

func ExampleWithMetrics() {
	g := gin.New()
	g.Use(WithMetrics(prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
		Name: "request_duration_seconds",
		Help: "Total time spent serving requests.",
	}, []string{"module", "method"}), key.New("module", "foo"), false))
	g.Handle("GET", "/", func(context *gin.Context) {
		// Do stuff
	})
}
