package di

import (
	"math/rand"
	"testing"
	"time"

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
