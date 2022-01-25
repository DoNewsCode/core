package di

import (
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFactory(t *testing.T) {
	t.Parallel()
	var closed []string

	f := NewFactory[*string](func(name string) (Pair[*string], error) {
		nameCopy := name
		return Pair[*string]{
			Conn: &nameCopy,
			Closer: func() {
				closed = append(closed, nameCopy)
			},
		}, nil
	})

	foo, err := f.Make("foo")
	assert.NoError(t, err)
	assert.Equal(t, *foo, "foo")

	bar, err := f.Make("bar")
	assert.NoError(t, err)
	assert.Equal(t, *bar, "bar")

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

	f := NewFactory[*string](func(name string) (Pair[*string], error) {
		nameCopy := name
		return Pair[*string]{
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

	f := NewFactory[any](func(name string) (Pair[any], error) {
		return Pair[any]{}, errors.New("failed")
	})

	_, err := f.Make("foo")
	assert.Error(t, err)
}

func TestFactory_Watch(t *testing.T) {
	t.Parallel()

	mockConf := "foo"
	f := NewFactory[*string](func(_ string) (Pair[*string], error) {
		return Pair[*string]{
			Conn:   &mockConf,
			Closer: func() {},
		}, nil
	})

	foo, err := f.Make("default")
	assert.NoError(t, err)
	assert.Equal(t, "foo", *foo)

	mockConf = "bar"
	f.Reload()

	foo, err = f.Make("default")
	assert.NoError(t, err)
	assert.Equal(t, "bar", *foo)
}

func BenchmarkFactory_slowConn(b *testing.B) {
	f := NewFactory[*string](func(name string) (Pair[*string], error) {
		// Simulate a slow construction
		time.Sleep(100 * time.Millisecond)
		return Pair[*string]{
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
