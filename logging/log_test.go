package logging

import (
	"bytes"
	"context"
	"testing"

	"github.com/DoNewsCode/core/ctxmeta"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/stretchr/testify/assert"
)

func TestWithLevel(t *testing.T) {
	var buf bytes.Buffer
	l := log.NewLogfmtLogger(&buf)
	ll := WithLevel(l)
	ll.Debug("hi")
	// ensure the caller depth is correct
	assert.Contains(t, buf.String(), "caller=log_test.go")
	assert.Contains(t, buf.String(), "level=debug")

	ll.Debugw("foo", "bar", "baz")
	assert.Contains(t, buf.String(), "bar=baz")

	ll.Debugf("foo%d", 1)
	assert.Contains(t, buf.String(), "foo1")
}

func TestLevelFilter(t *testing.T) {
	var buf bytes.Buffer
	l := log.NewLogfmtLogger(&buf)
	l = level.NewFilter(l, LevelFilter("error"))
	WithLevel(l).Debug("hi")
	// ensure the caller depth is correct
	assert.NotContains(t, buf.String(), "caller=log_test.go")
}

func TestNewLogger(t *testing.T) {
	_ = NewLogger("logfmt")
}

func TestWithContext(t *testing.T) {
	ctx := context.Background()
	bag, ctx := ctxmeta.Inject(ctx)
	bag.Set("foo", "bar")

	var buf bytes.Buffer
	l := log.NewLogfmtLogger(&buf)
	ll := WithContext(l, ctx)
	ll.Log("baz", "qux")
	assert.Contains(t, buf.String(), "foo=bar baz=qux")
}

type mockSpan struct {
	received []interface{}
}

func (m *mockSpan) LogKV(alternatingKeyValues ...interface{}) {
	m.received = alternatingKeyValues
}

func TestSpanLogger(t *testing.T) {
	var mock mockSpan
	spanLogger{
		span: &mock,
		base: log.NewNopLogger(),
		kvs:  []interface{}{"foo", log.Valuer(func() interface{} { return "bar" })},
	}.Log("baz", "qux")

	assert.Equal(t, []interface{}{"foo", "bar", "baz", "qux"}, mock.received)
}
