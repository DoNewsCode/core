package config

import (
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"

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
