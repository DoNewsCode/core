package config

import (
	"context"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/events"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func setup() *cobra.Command {
	config, _ := NewConfig()
	mod := Module{
		conf: config,
		exportedConfigs: []ExportedConfig{
			{
				"foo",
				map[string]any{
					"foo": "bar",
				},
				"A mock config",
				func(data map[string]any) error {
					if _, ok := data["foo"]; !ok {
						return errors.New("bad config")
					}
					return nil
				},
			},
			{
				"baz",
				map[string]any{
					"baz": "qux",
				},
				"Other mock config",
				nil,
			},
		},
		dispatcher: nil,
	}
	rootCmd := &cobra.Command{
		Use: "root",
	}
	mod.ProvideCommand(rootCmd)
	return rootCmd
}

func tearDown() {
	os.Remove("./testdata/module_test.yaml")
	os.Remove("./testdata/module_test.json")
	ioutil.WriteFile("./testdata/module_test_partial.json", []byte("{\n  \"foo\": \"bar\"\n}"), os.ModePerm)
	ioutil.WriteFile("./testdata/module_test_partial.yaml", []byte("# A mock config\nfoo: bar\n"), os.ModePerm)
}

func TestModule_ProvideCommand_initCmd(t *testing.T) {
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
			"old json",
			"./testdata/module_test.json",
			[]string{"config", "init", "--outputFile", "./testdata/module_test.json", "--style", "json"},
			"./testdata/module_test_expected.json",
		},
		{
			"partial json",
			"./testdata/module_test_partial.json",
			[]string{"config", "init", "--outputFile", "./testdata/module_test_partial.json", "--style", "json"},
			"./testdata/module_test_partial_expected.json",
		},
		{
			"partial yaml",
			"./testdata/module_test_partial.yaml",
			[]string{"config", "init", "baz", "--outputFile", "./testdata/module_test_partial.yaml"},
			"./testdata/module_test_partial_expected.yaml",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rootCmd := setup()
			defer tearDown()
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

func TestModule_ProvideCommand_verifyCmd(t *testing.T) {
	rootCmd := setup()
	cases := []struct {
		name  string
		args  []string
		isErr bool
	}{
		{
			"bad config",
			[]string{"config", "verify", "--targetFile", "./testdata/module_test_empty.yaml"},
			true,
		},
		{
			"bad config with module",
			[]string{"config", "verify", "foo", "--targetFile", "./testdata/module_test_empty.yaml"},
			true,
		},
		{
			"bad config with good module selected",
			[]string{"config", "verify", "baz", "--targetFile", "./testdata/module_test_empty.yaml"},
			false,
		},
		{
			"good config",
			[]string{"config", "verify", "--targetFile", "./testdata/module_test_gold.yaml"},
			false,
		},
		{
			"good config with module",
			[]string{"config", "verify", "foo", "--targetFile", "./testdata/module_test_gold.yaml"},
			false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rootCmd.SetArgs(c.args)
			err := rootCmd.Execute()
			if c.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestModule_Watch(t *testing.T) {
	t.Run("test without module", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		dispatcher := &events.Event[contract.ConfigUnmarshaler]{}
		dispatcher.On(func(ctx context.Context, event contract.ConfigUnmarshaler) error {
			data := event.(*KoanfAdapter)
			assert.Equal(t, "bar", data.String("foo"))
			cancel()
			return nil
		})
		conf, _ := NewConfig(WithDispatcher(dispatcher), WithProviderLayer(confmap.Provider(map[string]any{"foo": "bar"}, "."), nil))
		conf.Watch(ctx)
	})

	t.Run("test with module", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		dispatcher := &events.Event[contract.ConfigUnmarshaler]{}
		dispatcher.On(func(ctx context.Context, event contract.ConfigUnmarshaler) error {
			data := event.(*KoanfAdapter)
			assert.Equal(t, "bar", data.String("foo"))
			cancel()
			return nil
		})

		conf, _ := NewConfig(
			WithDispatcher(dispatcher),
			WithProviderLayer(confmap.Provider(map[string]any{"foo": "bar"}, "."), nil),
			WithWatcher(&MockWatcher{}),
		)
		module, _ := New(ConfigIn{
			Conf: conf,
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
