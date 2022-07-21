package core

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DoNewsCode/core/cron"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/logging"
	"github.com/DoNewsCode/core/observability"
	"github.com/gorilla/mux"

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

type NewCronModule struct {
	CanRun uint32
}

func (module *NewCronModule) ProvideCron(crontab *cron.Cron) {
	crontab.Add("* * * * * *", func(ctx context.Context) error {
		atomic.StoreUint32(&module.CanRun, 1)
		return nil
	})
}

func TestServeIn_cron(t *testing.T) {
	c := Default(
		WithInline("grpc.disable", true),
		WithInline("http.disable", true),
		WithInline("log.level", "none"),
	)
	c.Provide(observability.Providers())
	c.Provide(
		di.Deps{func() *cron.Cron {
			return cron.New(cron.Config{EnableSeconds: true})
		}},
	)

	m := NewCronModule{}
	c.AddModule(&m)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	c.Serve(ctx)
	assert.True(t, m.CanRun == 1)
}

func TestServeIn_inject_HTTPRouter(t *testing.T) {
	c := Default(
		WithInline("grpc.disable", true),
		WithInline("cron.disable", true),
		WithInline("http.addr", ":8080"),
		WithInline("log.level", "none"),
	)
	c.Provide(
		di.Deps{func() *mux.Router {
			r := mux.NewRouter()
			r.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("test"))
			}))
			return r
		}},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go c.Serve(ctx)

	time.Sleep(time.Second)

	resp, err := http.Get("http://localhost:8080/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, string(bs) == "test")
}

type SimpleRunModule struct {
	RunCount uint32
}

func (s *SimpleRunModule) Run(ctx context.Context) error {
	atomic.AddUint32(&s.RunCount, 1)
	return nil
}

func TestSimpleRun(t *testing.T) {
	c := Default(
		WithInline("grpc.disable", true),
		WithInline("http.disable", true),
		WithInline("cron.disable", true),
		WithInline("log.level", "none"),
	)
	m := SimpleRunModule{}
	c.AddModule(&m)
	c.Serve(context.Background())
	assert.Equal(t, uint32(1), m.RunCount)
}
