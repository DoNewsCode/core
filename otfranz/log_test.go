package otfranz

import (
	"bytes"
	"testing"

	"github.com/go-kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/twmb/franz-go/pkg/kgo"
)

func Test_logAdapter(t *testing.T) {
	cases := []struct {
		name   string
		lvlCfg string
		level  kgo.LogLevel
		want   string
	}{
		{"debug-debug", "debug", kgo.LogLevelDebug, "msg=foo\n"},
		{"debug-info", "debug", kgo.LogLevelInfo, "msg=foo\n"},

		{"info-debug", "info", kgo.LogLevelDebug, ""},
		{"info-info", "info", kgo.LogLevelInfo, "msg=foo\n"},

		{"warn", "warn", kgo.LogLevelWarn, "msg=foo\n"},
		{"error", "error", kgo.LogLevelError, "msg=foo\n"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			buf := bytes.NewBuffer(nil)
			logger := FranzLogAdapter(c.lvlCfg, log.NewLogfmtLogger(buf))
			logger.Log(c.level, "foo")
			assert.Equal(t, c.want, buf.String())
		})
	}
}
