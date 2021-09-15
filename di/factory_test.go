package di

import (
	"context"
	"errors"
	"math/rand"
	"runtime"
	"testing"
	"time"

	"github.com/DoNewsCode/core/events"
	"github.com/stretchr/testify/assert"
)

func TestFactory(t *testing.T) {
	t.Parallel()
	var closed []string

	f := NewFactory(func(name string) (Pair, error) {
		nameCopy := name
		return Pair{
			Conn: &nameCopy,
			Closer: func() {
				closed = append(closed, name)
			},
		}, nil
	})

	foo, err := f.Make("foo")
	assert.NoError(t, err)
	assert.Equal(t, *(foo.(*string)), "foo")

	bar, err := f.Make("bar")
	assert.NoError(t, err)
	assert.Equal(t, *(bar.(*string)), "bar")

	bar2, err := f.Make("bar")
	assert.NoError(t, err)
	assert.Equal(t, bar, bar2)

	assert.Len(t, f.List(), 2)

	f.CloseConn("foo")
	assert.Contains(t, closed, "foo")

	f.Close()
	assert.Contains(t, closed, "foo", "bar")
}

func TestFactory_nilCloser(t *testing.T) {
	t.Parallel()

	f := NewFactory(func(name string) (Pair, error) {
		nameCopy := name
		return Pair{
			Conn:   &nameCopy,
			Closer: nil,
		}, nil
	})

	f.Make("foo")

	f.CloseConn("foo")

	f.Close()
}

func TestFactory_malfunctionConstructor(t *testing.T) {
	t.Parallel()

	f := NewFactory(func(name string) (Pair, error) {
		return Pair{}, errors.New("failed")
	})

	_, err := f.Make("foo")
	assert.Error(t, err)
}

func TestFactory_Watch(t *testing.T) {
	t.Parallel()

	var (
		mockConf = "foo"
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f := NewFactory(func(_ string) (Pair, error) {
		return Pair{
			Conn:   mockConf,
			Closer: func() {},
		}, nil
	})
	dispatcher := events.SyncDispatcher{}
	go func() {
		f.SubscribeReloadEventFrom(&dispatcher)
	}()

	foo, err := f.Make("default")
	assert.NoError(t, err)
	assert.Equal(t, "foo", foo.(string))

	mockConf = "bar"

	foo, err = f.Make("default")
	assert.NoError(t, err)
	assert.Equal(t, "foo", foo.(string))

	time.Sleep(3 * time.Second)
	_ = dispatcher.Dispatch(ctx, events.OnReload, events.OnReloadPayload{})

	time.Sleep(3 * time.Second)
	foo, err = f.Make("default")
	assert.NoError(t, err)
	assert.Equal(t, "bar", foo.(string))
}

func TestFactory_SubscribeReloadEventFrom(t *testing.T) {
	t.Parallel()

	var (
		ptr = &struct {
			Dummy string
		}{Dummy: "dummy"}
		closed = make(chan struct{})
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f := NewFactory(func(_ string) (Pair, error) {
		return Pair{
			Conn:   ptr,
			Closer: func() { close(closed) },
		}, nil
	})
	dispatcher := events.SyncDispatcher{}
	f.SubscribeReloadEventFrom(&dispatcher)

	foo, err := f.Make("default")
	assert.NoError(t, err)
	assert.Same(t, ptr, foo)

	_ = dispatcher.Dispatch(ctx, events.OnReload, events.OnReloadPayload{})

	// We don't want to interrupt ongoing request, so foo should not be closed by now
	select {
	case <-closed:
		t.Fatalf("foo should not be closed.")
	default:
	}

	// now that foo is garbage collected, we can safely close foo.
	ptr = nil //nolint
	foo = nil //nolint
	runtime.GC()
	cancel()
	select {
	case <-closed:
	case <-time.After(4 * time.Second):
		t.Fatalf("foo should be closed by now")
	}

}

func BenchmarkFactory_slowConn(b *testing.B) {
	f := NewFactory(func(name string) (Pair, error) {
		// Simulate a slow construction
		time.Sleep(100 * time.Millisecond)
		return Pair{
			Conn:   &name,
			Closer: func() {},
		}, nil
	})
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f.Make(randomString(10))
		}
	})
}

const (
	chars = "abcdefghijklmnopqrstuvwxyz0123456789"
)

func randomString(l uint) string {
	s := make([]byte, l)
	for i := 0; i < int(l); i++ {
		s[i] = chars[rand.Intn(len(chars))]
	}
	return string(s)
}
