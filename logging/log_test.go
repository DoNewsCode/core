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
	received []any
}

func (m *mockSpan) LogKV(alternatingKeyValues ...any) {
	m.received = alternatingKeyValues
}

func TestSpanLogger(t *testing.T) {
	var mock mockSpan
	spanLogger{
		span: &mock,
		base: log.NewNopLogger(),
		kvs:  []any{"foo", log.Valuer(func() any { return "bar" })},
	}.Log("baz", "qux")

	assert.Equal(t, []any{"foo", "bar", "baz", "qux"}, mock.received)
}

type mockValue struct{}

func (m mockValue) String() string {
	// potentially expensive, but the cost is avoided because we have set a higher log level.
	panic("should not reach here")
}

func TestPerformanceOptimization(t *testing.T) {
	var buf bytes.Buffer
	for _, c := range []struct {
		name   string
		logger log.Logger
	}{
		{
			"json",
			log.NewJSONLogger(&buf),
		},
		{
			"logfmt",
			log.NewLogfmtLogger(&buf),
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			l := level.NewFilter(c.logger, LevelFilter("error"))
			ll := WithLevel(l)
			ll.Debug(mockValue{})
			ll.Debugw("bar", "foo", mockValue{})
			ll.Debugf("%s", mockValue{})
		})
	}
}
