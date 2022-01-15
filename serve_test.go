package core

import (
	"bytes"
	"context"
	"errors"
	"github.com/DoNewsCode/core/cron"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/observability"
	deprecatedcron "github.com/robfig/cron/v3"
	"os"
	"runtime"
	"sync/atomic"
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
		if runtime.GOOS == "windows" {
			t.Skip("TestServeIn_signalWatch/stop_when_signal_received only works on unix")
		}
		var group run.Group
		group.Add(do, cancel)
		group.Add(func() error {
			time.Sleep(time.Second)
			p, err := os.FindProcess(os.Getpid())
			if err != nil {
				return err
			}
			if err := p.Signal(os.Interrupt); err != nil {
				return err
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

type OldCronModule struct {
	CanRun uint32
}

func (module *OldCronModule) ProvideCron(crontab *deprecatedcron.Cron) {
	crontab.AddFunc("* * * * * *", func() {
		atomic.StoreUint32(&module.CanRun, 1)
	})
}

type NewCronModule struct {
	CanRun uint32
}

func (module *NewCronModule) ProvideCron(crontab *cron.Cron) {
	crontab.Add("* * * * * *", func(ctx context.Context) error {
		atomic.StoreUint32(&module.CanRun, 1)
		return nil
	})
}

func TestServeIn_cron_deprecation(t *testing.T) {
	c := Default(WithInline("grpc.disable", true), WithInline("http.disable", true))
	c.Provide(observability.Providers())
	c.Provide(
		di.Deps{func() *deprecatedcron.Cron {
			return deprecatedcron.New(deprecatedcron.WithSeconds())
		}, func() *cron.Cron {
			return cron.New(cron.Config{EnableSeconds: true})
		}},
	)

	mOld := OldCronModule{}
	mNew := NewCronModule{}
	c.AddModule(&mOld)
	c.AddModule(&mNew)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	c.Serve(ctx)
	assert.True(t, mOld.CanRun == 1)
	assert.True(t, mNew.CanRun == 1)
}
