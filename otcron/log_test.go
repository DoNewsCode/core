package otcron

import (
	"bytes"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCronLogAdapter_Info(t *testing.T) {
	var buf bytes.Buffer
	l := CronLogAdapter{Logging: log.NewLogfmtLogger(&buf)}
	l.Info("msg", "key","value")
	assert.Equal(t, "level=info msg=msg key=value\n", buf.String())
}

func TestCronLogAdapter_Error(t *testing.T) {
	var buf bytes.Buffer
	l := CronLogAdapter{Logging: log.NewLogfmtLogger(&buf)}
	l.Error(errors.New("err"),"msg", "key","value")
	assert.Equal(t, "level=error msg=msg err=err key=value\n",buf.String())
}