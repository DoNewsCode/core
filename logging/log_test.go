package logging

import (
	"bytes"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/stretchr/testify/assert"
)

func TestWithLevel(t *testing.T) {
	var buf bytes.Buffer
	l := log.NewLogfmtLogger(&buf)
	WithLevel(l).Debug("hi")
	// ensure the caller depth is correct
	assert.Contains(t, buf.String(), "caller=log_test.go")
}

func TestLevelFilter(t *testing.T) {
	var buf bytes.Buffer
	l := log.NewLogfmtLogger(&buf)
	l = level.NewFilter(l, LevelFilter("error"))
	WithLevel(l).Debug("hi")
	// ensure the caller depth is correct
	assert.NotContains(t, buf.String(), "caller=log_test.go")
}

func TestNewLogger(t *testing.T) {
	_ = NewLogger("logfmt")
}
