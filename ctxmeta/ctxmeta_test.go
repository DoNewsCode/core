package ctxmeta

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextMeta_crud(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	metadata := New[string, string]()
	baggage, _ := metadata.Inject(ctx)

	baggage.Set("foo", "bar")
	result, err := baggage.Get("foo")
	assert.NoError(t, err)
	assert.Equal(t, "bar", result)

	slice := baggage.Slice()
	assert.ElementsMatch(t, []KeyVal[string, string]{{Key: "foo", Val: "bar"}}, slice)

	maps := baggage.Map()
	assert.Equal(t, "bar", maps["foo"])

	var s string
	baggage.Unmarshal("foo", &s)
	assert.Equal(t, "bar", s)

	baggage.Update("foo", func(value string) string { return "baz" })
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
	metadata := New[string, string]()
	baggage := metadata.GetBaggage(ctx)

	err := baggage.Set("foo", "bar")
	assert.ErrorIs(t, err, ErrNoBaggage)

	_, err = baggage.Get("foo")
	assert.ErrorIs(t, err, ErrNoBaggage)

	var s string
	err = baggage.Unmarshal("foo", &s)
	assert.ErrorIs(t, err, ErrNoBaggage)

	err = baggage.Update("foo", func(value string) string { return "baz" })
	assert.ErrorIs(t, err, ErrNoBaggage)

	err = baggage.Delete("foo")
	assert.ErrorIs(t, err, ErrNoBaggage)

	assert.Nil(t, baggage.Slice())
	assert.Nil(t, baggage.Map())
}

func TestContextMeta_ErrNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	metadata := New[string, string]()
	baggage, _ := metadata.Inject(ctx)

	_, err := baggage.Get("foo")
	assert.ErrorIs(t, err, ErrNotFound)

	var s string
	err = baggage.Unmarshal("foo", &s)
	assert.ErrorIs(t, err, ErrNotFound)

	err = baggage.Update("foo", func(value string) string { return "baz" })
	assert.ErrorIs(t, err, ErrNotFound)

	err = baggage.Delete("foo")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestContextMeta_ErrIncompatibleType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	metadata := New[string, string]()
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
		meta  *MetadataSet[string, any]
		key   string
		value string
	}{
		{
			"first",
			New[string, any](),
			"foo",
			"bar",
		},
		{
			"second",
			New[string, any](),
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

func TestMetadata_GetOrInjectBaggage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	baggage1, ctx := GetOrInjectBaggage(ctx)
	baggage1.Set("hello", "world")

	baggage2 := GetBaggage(ctx)
	world, _ := baggage2.Get("hello")
	assert.Equal(t, "world", world)

	baggage3 := DefaultMetadata.GetBaggage(ctx)
	world, _ = baggage3.Get("hello")
	assert.Equal(t, "world", world)

	baggage4, _ := GetOrInjectBaggage(ctx)
	world, _ = baggage4.Get("hello")
	assert.Equal(t, "world", world)
}
