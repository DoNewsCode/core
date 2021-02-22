package di

import (
	"testing"

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
