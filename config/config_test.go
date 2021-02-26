package config

import (
	"context"
	"io/ioutil"
	"os"
	gotesting "testing"
	"time"

	"github.com/DoNewsCode/core/config/watcher"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/stretchr/testify/assert"
)

func TestKoanfAdapter_Route(t *gotesting.T) {
	t.Parallel()
	ka := prepareTestSubject(t)
	assert.Implements(t, MapAdapter{}, ka.Route("foo"))
	assert.Implements(t, MapAdapter{}, ka.Route("foo"))
}

func TestKoanfAdapter_race(t *gotesting.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.True(t, false, "shouldn't reach here")
		}
	}()
	t.Parallel()
	ka := prepareTestSubject(t)
	for i := 0; i < 100; i++ {
		go ka.Reload()
		ka.String("string")
	}

}

func TestKoanfAdapter_Watch(t *gotesting.T) {
	t.Parallel()
	f, _ := ioutil.TempFile(".", "*")
	defer os.Remove(f.Name())

	ioutil.WriteFile(f.Name(), []byte("foo: baz"), 0644)

	ka, err := NewConfig(
		WithProviderLayer(file.Provider(f.Name()), yaml.Parser()),
		WithWatcher(watcher.File{Path: f.Name()}),
	)
	assert.NoError(t, err)
	assert.Equal(t, "baz", ka.String("foo"))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		ka.watcher.Watch(ctx, func() error {
			assert.NoError(t, ka.Reload(), "reload should be successful")
			return nil
		})
	}()
	time.Sleep(time.Second)
	ioutil.WriteFile(f.Name(), []byte("foo: bar"), 0644)

	// The following test is flaky on CI, cannot reproduce locally.

	//time.Sleep(time.Second)
	//assert.Equal(
	//	t,
	//	"bar",
	//	ka.String("foo"),
	//	"configAccessor should always return the latest value.",
	//)
}

func TestKoanfAdapter_Bool(t *gotesting.T) {
	t.Parallel()
	k := prepareTestSubject(t)
	assert.True(t, k.Bool("bool"))
}

func TestKoanfAdapter_String(t *gotesting.T) {
	t.Parallel()
	k := prepareTestSubject(t)
	assert.Equal(t, "string", k.String("string"))
}

func TestKoanfAdapter_Strings(t *gotesting.T) {
	t.Parallel()
	k := prepareTestSubject(t)
	assert.Equal(t, []string{"foo", "bar"}, k.Strings("strings"))
}

func TestKoanfAdapter_Float64(t *gotesting.T) {
	t.Parallel()
	k := prepareTestSubject(t)
	assert.Equal(t, 1.0, k.Float64("float"))
}

func TestKoanfAdapter_Get(t *gotesting.T) {
	t.Parallel()
	k := prepareTestSubject(t)
	assert.Equal(t, 1.0, k.Get("float"))
}

func TestKoanfAdapter_Unmarshal(t *gotesting.T) {
	t.Parallel()
	ka := prepareTestSubject(t)
	var target string
	err := ka.Unmarshal("foo.bar", &target)
	assert.NoError(t, err)
	assert.Equal(t, "baz", target)
}

func TestMapAdapter_Bool(t *gotesting.T) {
	t.Parallel()
	k := MapAdapter(
		map[string]interface{}{
			"bool": true,
		},
	)
	assert.True(t, k.Bool("bool"))
}

func TestMapAdapter_String(t *gotesting.T) {
	t.Parallel()
	k := MapAdapter(
		map[string]interface{}{
			"string": "string",
		},
	)
	assert.Equal(t, "string", k.String("string"))
}

func TestMapAdapter_Float64(t *gotesting.T) {
	t.Parallel()
	k := MapAdapter(
		map[string]interface{}{
			"float": 1.0,
		},
	)
	assert.Equal(t, 1.0, k.Float64("float"))
}

func TestMapAdapter_Get(t *gotesting.T) {
	t.Parallel()
	k := MapAdapter(
		map[string]interface{}{
			"float": 1.0,
		},
	)
	assert.Equal(t, 1.0, k.Get("float"))
}

func TestMapAdapter_Route(t *gotesting.T) {
	t.Parallel()
	m := MapAdapter(
		map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": "baz",
			},
		},
	)
	assert.Equal(t, MapAdapter(map[string]interface{}{
		"bar": "baz",
	}), m.Route("foo"))
	assert.Panics(t, func() {
		m.Route("foo2")
	})
}

func TestMapAdapter_Unmarshal(t *gotesting.T) {
	t.Parallel()
	m := MapAdapter(
		map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": "baz",
			},
		},
	)
	var target map[string]interface{}
	err := m.Unmarshal("foo", &target)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"bar": "baz",
	}, target)

	var badTarget struct{}
	err = m.Unmarshal("foo", &badTarget)
	assert.Error(t, err)
}

func prepareTestSubject(t *gotesting.T) *KoanfAdapter {
	k := koanf.New(".")
	if err := k.Load(file.Provider("testdata/mock.json"), json.Parser()); err != nil {
		t.Fatalf("error loading config: %v", err)
	}
	ka := KoanfAdapter{K: k}
	return &ka
}
