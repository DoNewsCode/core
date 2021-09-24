package cronopts

import (
	"bytes"
	"testing"

	"github.com/go-kit/log"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestCronLogAdapter_Info(t *testing.T) {
	var buf bytes.Buffer
	l := CronLogAdapter{Logging: log.NewLogfmtLogger(&buf)}
	l.Info("msg", "key", "value")
	assert.Equal(t, "level=info msg=msg key=value\n", buf.String())
}

func TestCronLogAdapter_Error(t *testing.T) {
	var buf bytes.Buffer
	l := CronLogAdapter{Logging: log.NewLogfmtLogger(&buf)}
	l.Error(errors.New("err"), "msg", "key", "value")
	assert.Equal(t, "level=error msg=msg err=err key=value\n", buf.String())
}
