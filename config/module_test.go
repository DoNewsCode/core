package config

import (
	"context"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/DoNewsCode/core/events"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func setup() *cobra.Command {
	os.Remove("./testdata/module_test.yaml")
	os.Remove("./testdata/module_test.json")
	var config, _ = NewConfig()
	var mod = Module{conf: config, exportedConfigs: []ExportedConfig{
		{
			"foo",
			map[string]interface{}{
				"foo": "bar",
			},
			"A mock config",
		},
		{
			"baz",
			map[string]interface{}{
				"baz": "qux",
			},
			"Other mock config",
		},
	}}
	rootCmd := &cobra.Command{
		Use: "root",
	}
	mod.ProvideCommand(rootCmd)
	return rootCmd
}

func TestModule_ProvideCommand(t *testing.T) {
	rootCmd := setup()
	cases := []struct {
		name     string
		output   string
		args     []string
		expected string
	}{
		{
			"foo yaml",
			"./testdata/module_test.yaml",
			[]string{"config", "init", "foo", "--outputFile", "./testdata/module_test.yaml"},
			"./testdata/module_test_foo_expected.yaml",
		},
		{
			"baz yaml",
			"./testdata/module_test.yaml",
			[]string{"config", "init", "baz", "--outputFile", "./testdata/module_test.yaml"},
			"./testdata/module_test_baz_expected.yaml",
		},
		{
			"old yaml",
			"./testdata/module_test.yaml",
			[]string{"config", "init", "--outputFile", "./testdata/module_test.yaml"},
			"./testdata/module_test_expected.yaml",
		},
		{
			"foo json",
			"./testdata/module_test.json",
			[]string{"config", "init", "foo", "--outputFile", "./testdata/module_test.json", "--style", "json"},
			"./testdata/module_test_foo_expected.json",
		},
		{
			"baz json",
			"./testdata/module_test.json",
			[]string{"config", "init", "baz", "--outputFile", "./testdata/module_test.json", "--style", "json"},
			"./testdata/module_test_baz_expected.json",
		},
		{
			"old json",
			"./testdata/module_test.json",
			[]string{"config", "init", "--outputFile", "./testdata/module_test.json", "--style", "json"},
			"./testdata/module_test_expected.json",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rootCmd.SetArgs(c.args)
			rootCmd.Execute()
			testTarget, _ := ioutil.ReadFile(c.output)
			expected, _ := ioutil.ReadFile(c.expected)
			expectedString := string(expected)
			if runtime.GOOS == "windows" {
				expectedString = strings.ReplaceAll(expectedString, "\r", "")
			}
			assert.Equal(t, expectedString, string(testTarget))
		})
	}
}

func TestModule_Watch(t *testing.T) {
	t.Run("test without module", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		dispatcher := &events.SyncDispatcher{}
		dispatcher.Subscribe(events.Listen(events.OnReload, func(ctx context.Context, event interface{}) error {
			data := event.(events.OnReloadPayload).NewConf.(*KoanfAdapter)
			assert.Equal(t, "bar", data.String("foo"))
			cancel()
			return nil
		}))
		conf, _ := NewConfig(WithDispatcher(dispatcher), WithProviderLayer(confmap.Provider(map[string]interface{}{"foo": "bar"}, "."), nil))
		conf.Watch(ctx)
	})

	t.Run("test with module", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		dispatcher := &events.SyncDispatcher{}
		dispatcher.Subscribe(events.Listen(events.OnReload, func(ctx context.Context, event interface{}) error {
			data := event.(events.OnReloadPayload).NewConf.(*KoanfAdapter)
			assert.Equal(t, "bar", data.String("foo"))
			cancel()
			return nil
		}))

		conf, _ := NewConfig(WithProviderLayer(confmap.Provider(map[string]interface{}{"foo": "bar"}, "."), nil), WithWatcher(&MockWatcher{}))
		module, _ := New(ConfigIn{
			Conf:       conf,
			Dispatcher: dispatcher,
		})
		var g run.Group
		g.Add(func() error {
			<-ctx.Done()
			return nil
		}, func(err error) {})
		module.ProvideRunGroup(&g)
		g.Run()
	})
}

type MockWatcher struct{}

func (m *MockWatcher) Watch(ctx context.Context, reload func() error) error {
	return reload()
}
