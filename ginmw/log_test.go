package ginmw

import (
	"net/http/httptest"
	"testing"

	"github.com/DoNewsCode/core/key"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type MockLogger struct {
	values []interface{}
}

func (m *MockLogger) Log(keyvals ...interface{}) error {
	m.values = keyvals
	return nil
}

func TestLog(t *testing.T) {
	cases := []struct {
		name   string
		ignore []string
		assert func(t *testing.T, logger MockLogger)
	}{
		{
			"normal",
			[]string{},
			func(t *testing.T, logger MockLogger) {
				assert.Contains(t, logger.values, 200)
			},
		},
		{
			"ignore",
			[]string{"/"},
			func(t *testing.T, logger MockLogger) {
				assert.NotContains(t, logger.values, 200)
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			g := gin.New()
			logger := MockLogger{}
			g.Use(Log(&logger, key.New("module", "foo"), c.ignore...))
			g.Handle("GET", "/", func(context *gin.Context) {
				context.String(200, "%s", "ok")
			})
			req := httptest.NewRequest("GET", "/", nil)
			writer := httptest.NewRecorder()
			g.ServeHTTP(writer, req)
			c.assert(t, logger)
		})
	}
}
