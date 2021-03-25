package otes

import (
	"bytes"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
)

func TestEsLogAdapter_Printf(t *testing.T) {
	var buf bytes.Buffer
	l := ElasticLogAdapter{Logging: log.NewLogfmtLogger(&buf)}
	l.Printf("test %s", "logger")
	assert.Equal(t, "msg=\"test logger\"\n", buf.String())
}
