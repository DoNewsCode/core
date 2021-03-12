package ginmw

import (
	"github.com/DoNewsCode/core/key"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestTrace(t *testing.T) {
	t.Parallel()
	tracer := mocktracer.New()
	g := gin.New()
	g.Use(Trace(tracer, key.New("module", "foo")))
	g.Handle("GET", "/", func(context *gin.Context) {
		context.String(200, "%s", "ok")
	})
	req := httptest.NewRequest("GET", "/", nil)
	writer := httptest.NewRecorder()
	g.ServeHTTP(writer, req)
	assert.NotZero(t, tracer.FinishedSpans())
}
