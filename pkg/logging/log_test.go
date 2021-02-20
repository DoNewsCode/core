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
	assert.Contains(t, buf.String(), "level=debug caller=log_test.go")
}
