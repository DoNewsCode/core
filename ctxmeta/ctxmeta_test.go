package ctxmeta

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextMeta_crud(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	metadata := New()
	baggage, _ := metadata.Inject(ctx)

	baggage.Set("foo", "bar")
	result, err := baggage.Get("foo")
	assert.NoError(t, err)
	assert.Equal(t, "bar", result)

	result = baggage.Slice()
	assert.ElementsMatch(t, []KeyVal{{Key: "foo", Val: "bar"}}, result)

	resultMap := baggage.Map()
	assert.Equal(t, "bar", resultMap["foo"])

	var s string
	baggage.Unmarshal("foo", &s)
	assert.Equal(t, "bar", s)

	baggage.Update("foo", func(value interface{}) interface{} { return "baz" })
	result, err = baggage.Get("foo")
	assert.NoError(t, err)
	assert.Equal(t, "baz", result)

	baggage.Delete("foo")
	_, err = baggage.Get("foo")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestContextMeta_ErrNoBaggage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	metadata := New()
	baggage := metadata.GetBaggage(ctx)

	err := baggage.Set("foo", "bar")
	assert.ErrorIs(t, err, ErrNoBaggage)

	_, err = baggage.Get("foo")
	assert.ErrorIs(t, err, ErrNoBaggage)

	var s string
	err = baggage.Unmarshal("foo", &s)
	assert.ErrorIs(t, err, ErrNoBaggage)

	err = baggage.Update("foo", func(value interface{}) interface{} { return "baz" })
	assert.ErrorIs(t, err, ErrNoBaggage)

	err = baggage.Delete("foo")
	assert.ErrorIs(t, err, ErrNoBaggage)

	assert.Nil(t, baggage.Slice())
	assert.Nil(t, baggage.Map())
}

func TestContextMeta_ErrNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	metadata := New()
	baggage, _ := metadata.Inject(ctx)

	_, err := baggage.Get("foo")
	assert.ErrorIs(t, err, ErrNotFound)

	var s string
	err = baggage.Unmarshal("foo", &s)
	assert.ErrorIs(t, err, ErrNotFound)

	err = baggage.Update("foo", func(value interface{}) interface{} { return "baz" })
	assert.ErrorIs(t, err, ErrNotFound)

	err = baggage.Delete("foo")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestContextMeta_ErrIncompatibleType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	metadata := New()
	baggage, _ := metadata.Inject(ctx)

	baggage.Set("foo", "bar")

	var s int
	err := baggage.Unmarshal("foo", &s)
	assert.ErrorIs(t, err, ErrIncompatibleType)
}

func TestContextMeta_parallel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		meta  *MetadataSet
		key   string
		value string
	}{
		{
			"first",
			New(),
			"foo",
			"bar",
		},
		{
			"second",
			New(),
			"foo",
			"baz",
		},
		{
			"default",
			&DefaultMetadata,
			"foo",
			"qux",
		},
	}
	ctx := context.Background()
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			b, ctx := c.meta.Inject(ctx)
			b.Set(c.key, c.value)
			value, err := c.meta.GetBaggage(ctx).Get(c.key)
			assert.NoError(t, err)
			assert.Equal(t, c.value, value)
		})
	}
}

func TestMetadata_global(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	baggage1, ctx := Inject(ctx)
	baggage1.Set("hello", "world")

	baggage2 := GetBaggage(ctx)
	world, _ := baggage2.Get("hello")
	assert.Equal(t, "world", world)

	baggage3 := DefaultMetadata.GetBaggage(ctx)
	world, _ = baggage3.Get("hello")
	assert.Equal(t, "world", world)
}

func TestContextMeta_StartBaggage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	metadata := New()
	baggage1, ctx := metadata.GetOrInjectBaggage(ctx)
	assert.NotNil(t, baggage1)

	baggage1.Set("hello", "world")

	baggage2 := metadata.GetBaggage(ctx)
	world, _ := baggage2.Get("hello")
	assert.Equal(t, "world", world)

}
