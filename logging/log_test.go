package logging

import (
	"bytes"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
)

func TestWithLevel(t *testing.T) {
	var buf bytes.Buffer
	l := log.NewLogfmtLogger(&buf)
	WithLevel(l).Debug("hi")
	// ensure the caller depth is correct
	assert.Contains(t, buf.String(), "caller=log_test.go")
}

func TestNewLogger(t *testing.T) {
	_ = NewLogger("json")
}