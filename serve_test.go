package core

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/DoNewsCode/core/logging"
	"github.com/go-kit/kit/log"
	"github.com/oklog/run"
	"github.com/stretchr/testify/assert"
)

func TestServeIn_signalWatch(t *testing.T) {
	var in serveIn
	var buf bytes.Buffer
	do, cancel, err := in.signalWatch(context.Background(), logging.WithLevel(log.NewLogfmtLogger(&buf)))
	assert.NoError(t, err)

	var group run.Group
	group.Add(do, cancel)
	group.Add(func() error {
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			t.Skip("TestServeIn_signalWatch only works on unix")
		}
		if err := p.Signal(os.Interrupt); err != nil {
			t.Skip("TestServeIn_signalWatch only works on unix")
		}
		// trigger the signal twice should be ok.
		p.Signal(os.Interrupt)
		return nil
	}, func(err error) {})
	err = group.Run()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "signal received: interrupt")
}
