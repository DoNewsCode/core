package ginmw

import (
	"net/http/httptest"
	"testing"

	"github.com/DoNewsCode/std/pkg/contract"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestWithContext(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	g := gin.New()
	g.Use(WithContext())
	g.Handle("GET", "/foo", func(context *gin.Context) {
		assert.Equal(t, "/foo", context.Request.Context().Value(contract.RequestUrlKey))
		context.String(200, "%s", "ok")
	})
	w := httptest.NewRecorder()
	g.ServeHTTP(w, req)
	assert.Equal(t, "ok", w.Body.String())
}
