package core

import (
	"bytes"
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/DoNewsCode/core/logging"
	"github.com/go-kit/log"
	"github.com/oklog/run"
	"github.com/stretchr/testify/assert"
)

func TestServeIn_signalWatch(t *testing.T) {
	var in serveIn
	var buf bytes.Buffer
	do, cancel, err := in.signalWatch(context.Background(), logging.WithLevel(log.NewLogfmtLogger(&buf)))
	assert.NoError(t, err)

	t.Run("stop when signal received", func(t *testing.T) {
		var group run.Group
		group.Add(do, cancel)
		group.Add(func() error {
			time.Sleep(time.Second)
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
	})

	t.Run("cancel when cancel func is called", func(t *testing.T) {
		var group run.Group
		group.Add(do, cancel)
		group.Add(func() error {
			return errors.New("some err")
		}, func(err error) {

		})
		err = group.Run()
		assert.Contains(t, buf.String(), "context canceled")
	})

}
