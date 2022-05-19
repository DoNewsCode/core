package config

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/DoNewsCode/core/config/watcher"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/stretchr/testify/assert"
)

func TestKoanfAdapter_Route(t *testing.T) {
	t.Parallel()
	ka := prepareJSONTestSubject(t)
	assert.Implements(t, MapAdapter{}, ka.Route("foo"))
	assert.Implements(t, MapAdapter{}, ka.Route("foo"))
}

func TestKoanfAdapter_race(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.True(t, false, "shouldn't reach here")
		}
	}()
	t.Parallel()
	ka := prepareJSONTestSubject(t)
	for i := 0; i < 100; i++ {
		go ka.Reload()
		ka.String("string")
	}
}

func TestKoanfAdapter_Watch(t *testing.T) {
	f, _ := ioutil.TempFile(os.TempDir(), "*")
	defer os.Remove(f.Name())

	ioutil.WriteFile(f.Name(), []byte("foo: baz"), 0o644)

	ka, err := NewConfig(
		WithProviderLayer(file.Provider(f.Name()), yaml.Parser()),
		WithWatcher(watcher.File{Path: f.Name()}),
	)
	assert.NoError(t, err)
	assert.Equal(t, "baz", ka.String("foo"))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := make(chan struct{})
	go func() {
		ka.watcher.Watch(ctx, func() error {
			assert.NoError(t, ka.Reload(), "reload should be successful")
			err := ka.Reload()
			fmt.Println(err)
			ch <- struct{}{}
			return nil
		})
	}()
	time.Sleep(time.Second)
	ioutil.WriteFile(f.Name(), []byte("foo: bar"), 0o644)
	ioutil.WriteFile(f.Name(), []byte("foo: bar"), 0o644)
	<-ch

	// The following test is flaky on CI. Unable to reproduce locally.
	/*
		time.Sleep(time.Second)
		assert.Equal(
			t,
			"bar",
			ka.String("foo"),
			"configAccessor should always return the latest value.",
		) */
}

func TestKoanfAdapter_Bool(t *testing.T) {
	t.Parallel()
	k := prepareJSONTestSubject(t)
	assert.True(t, k.Bool("bool"))
}

func TestKoanfAdapter_String(t *testing.T) {
	t.Parallel()
	k := prepareJSONTestSubject(t)
	assert.Equal(t, "string", k.String("string"))
}

func TestKoanfAdapter_Strings(t *testing.T) {
	t.Parallel()
	k := prepareJSONTestSubject(t)
	assert.Equal(t, []string{"foo", "bar"}, k.Strings("strings"))
}

func TestKoanfAdapter_Float64(t *testing.T) {
	t.Parallel()
	k := prepareJSONTestSubject(t)
	assert.Equal(t, 1.0, k.Float64("float"))
}

func TestKoanfAdapter_Get(t *testing.T) {
	t.Parallel()
	k := prepareJSONTestSubject(t)
	assert.Equal(t, 1.0, k.Get("float"))
}

func TestKoanfAdapter_Duration(t *testing.T) {
	t.Parallel()
	k := prepareJSONTestSubject(t)
	assert.Equal(t, time.Second, k.Duration("duration_string"))
}

func TestKoanfAdapter_Unmarshal_Json(t *testing.T) {
	t.Parallel()
	ka := prepareJSONTestSubject(t)
	var target string
	err := ka.Unmarshal("foo.bar", &target)
	assert.NoError(t, err)
	assert.Equal(t, "baz", target)

	var r Duration
	err = ka.Unmarshal("duration_string", &r)
	assert.NoError(t, err)
	assert.Equal(t, r, Duration{1 * time.Second})

	err = ka.Unmarshal("duration_number", &r)
	assert.NoError(t, err)
	assert.Equal(t, r, Duration{1 * time.Nanosecond})
}

func TestKoanfAdapter_Unmarshal_Yaml(t *testing.T) {
	t.Parallel()
	ka := prepareYamlTestSubject(t)
	var target string
	err := ka.Unmarshal("foo.bar", &target)
	assert.NoError(t, err)
	assert.Equal(t, "baz", target)

	var r Duration
	err = ka.Unmarshal("duration_string", &r)
	assert.NoError(t, err)
	assert.Equal(t, r, Duration{1 * time.Second})

	err = ka.Unmarshal("duration_number", &r)
	assert.NoError(t, err)
	assert.Equal(t, r, Duration{1 * time.Nanosecond})
}

func TestMapAdapter_Route(t *testing.T) {
	t.Parallel()
	m := MapAdapter(
		map[string]any{
			"foo": map[string]any{
				"bar": "baz",
			},
		},
	)
	assert.Equal(t, MapAdapter(map[string]any{
		"bar": "baz",
	}), m.Route("foo"))
	assert.Panics(t, func() {
		m.Route("foo2")
	})
}

func TestMapAdapter_Unmarshal(t *testing.T) {
	t.Parallel()
	m := MapAdapter(
		map[string]any{
			"foo": map[string]any{
				"bar": "baz",
			},
		},
	)
	var target map[string]any
	err := m.Unmarshal("foo", &target)
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{
		"bar": "baz",
	}, target)
}

func TestKoanfAdapter_Reload(t *testing.T) {
	t.Parallel()
	conf, err := NewConfig(
		WithValidators(func(data map[string]any) error {
			return errors.New("bad config")
		}),
	)
	assert.Error(t, err)
	assert.Nil(t, conf)
}

func TestUpgrade(t *testing.T) {
	var m MapAdapter = map[string]any{"foo": "bar"}
	upgraded := WithAccessor(m)

	assert.Equal(t, float64(0), upgraded.Float64("foo"))
	assert.Equal(t, 0, upgraded.Int("foo"))
	assert.Equal(t, "bar", upgraded.String("foo"))
	assert.Equal(t, false, upgraded.Bool("foo"))
	assert.Equal(t, "bar", upgraded.Get("foo"))
	assert.Equal(t, []string{"bar"}, upgraded.Strings("foo"))
	assert.Equal(t, time.Duration(0), upgraded.Duration("foo"))
}

func prepareJSONTestSubject(t *testing.T) *KoanfAdapter {
	k := koanf.New(".")
	if err := k.Load(file.Provider("testdata/mock.json"), json.Parser()); err != nil {
		t.Fatalf("error loading config: %v", err)
	}
	ka := KoanfAdapter{K: k}
	return &ka
}

func prepareYamlTestSubject(t *testing.T) *KoanfAdapter {
	k := koanf.New(".")
	if err := k.Load(file.Provider("testdata/mock.yaml"), yaml.Parser()); err != nil {
		t.Fatalf("error loading config: %v", err)
	}
	ka := KoanfAdapter{K: k}
	return &ka
}
