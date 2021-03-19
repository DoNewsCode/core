package watcher

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
)

func TestWatch(t *testing.T) {
	t.Run("edit", func(t *testing.T) {
		t.Parallel()
		ch := make(chan struct{})
		f, _ := ioutil.TempFile(".", "*")
		defer os.Remove(f.Name())

		ioutil.WriteFile(f.Name(), []byte(`foo`), os.ModePerm)

		w := File{f.Name()}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go w.Watch(ctx, func() error {
			close(ch)
			return nil
		})
		time.Sleep(time.Second)
		ioutil.WriteFile(f.Name(), []byte(`bar`), os.ModePerm)
		<-ch
	})

	t.Run("delete", func(t *testing.T) {
		t.Parallel()
		var (
			ch     chan struct{}
			called atomic.Bool
		)
		ch = make(chan struct{})
		f, _ := ioutil.TempFile(".", "*")

		ioutil.WriteFile(f.Name(), []byte(`foo`), os.ModePerm)

		w := File{f.Name()}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			w.Watch(ctx, func() error {
				called.Store(true)
				ch <- struct{}{}
				return nil
			})
			ch <- struct{}{}
		}()
		time.Sleep(time.Second)
		os.Remove(f.Name())
		<-ch
		assert.False(t, called.Load())
	})

	t.Run("reload failed", func(t *testing.T) {
		t.Parallel()
		var (
			ch     chan struct{}
			called atomic.Bool
		)
		ch = make(chan struct{})
		f, _ := ioutil.TempFile(".", "*")

		ioutil.WriteFile(f.Name(), []byte(`foo`), os.ModePerm)
		defer os.Remove(f.Name())

		w := File{f.Name()}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			w.Watch(ctx, func() error {
				return errors.New("foo")
			})
			called.Store(true)
			ch <- struct{}{}
		}()
		time.Sleep(time.Second)
		ioutil.WriteFile(f.Name(), []byte(`bar`), os.ModePerm)
		<-ch
		assert.True(t, called.Load())
	})
}
