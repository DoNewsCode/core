package config

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	gotesting "testing"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/stretchr/testify/assert"
)

func TestKoanfAdapter_Route(t *gotesting.T) {
	ka := prepareTestSubject(t)
	assert.Implements(t, MapAdapter{}, ka.Route("foo"))
	assert.Implements(t, MapAdapter{}, ka.Route("foo"))
}

func TestKoanfAdapter_Unmarshal(t *gotesting.T) {
	ka := prepareTestSubject(t)
	var target string
	err := ka.Unmarshal("foo.bar", &target)
	fmt.Println(err)
	assert.NoError(t, err)
	assert.Equal(t, "baz", target)
}

func TestKoanfAdapter_Watch(t *gotesting.T) {
	ka, err := NewConfig(WithFilePath("mock/watch.yaml"))
	assert.NoError(t, err)

	var reach = false

	go func() {
		time.Sleep(10 * time.Millisecond)
		err = ioutil.WriteFile("mock/watch.yaml", []byte("foo: bar"), 0644)
	}()
	e := ka.Watch(context.Background(), func() error {
		reach = true
		return errors.New("test")
	})
	assert.NotNil(t, e)
	assert.True(t, reach)
}

func prepareTestSubject(t *gotesting.T) KoanfAdapter {
	k := koanf.New(".")
	if err := k.Load(file.Provider("mock/mock.json"), json.Parser()); err != nil {
		t.Fatalf("error loading config: %v", err)
	}
	ka := KoanfAdapter{K: k}
	return ka
}

func TestMapAdapter_Route(t *gotesting.T) {
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
