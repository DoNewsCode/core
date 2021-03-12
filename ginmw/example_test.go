package ginmw

import (
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/key"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/opentracing/opentracing-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

func ExampleWithContext() {
	g := gin.New()
	g.Use(Context())
	g.Handle("GET", "/", func(context *gin.Context) {
		context.String(200, "the request path is %s",
			context.Request.Context().Value(contract.RequestUrlKey))
	})
}

func ExampleWithTrace() {
	g := gin.New()
	g.Use(Trace(opentracing.GlobalTracer(), key.New("module", "foo")))
	g.Handle("GET", "/", func(context *gin.Context) {
		// Do stuff
	})
}

func ExampleWithMetrics() {
	g := gin.New()
	g.Use(Metrics(prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
		Name: "request_duration_seconds",
		Help: "Total time spent serving requests.",
	}, []string{"module", "method"}), key.New("module", "foo"), false))
	g.Handle("GET", "/", func(context *gin.Context) {
		// Do stuff
	})
}
