package config_test

import (
	"flag"
	"fmt"
	"github.com/DoNewsCode/std/pkg/config"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/basicflag"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/rawbytes"
)

func Example_minimum() {
	conf, _ := config.NewConfig(
		config.WithProviderLayer(
			rawbytes.Provider([]byte(`{"foo": "bar"}`)), json.Parser(),
		),
	)
	fmt.Println(conf.String("foo"))
	// Output:
	// bar
}

func Example_configurationStack() {
	var fs = flag.NewFlagSet("config", flag.ContinueOnError)
	fs.String("foo", "", "foo value")
	fs.Parse([]string{"-foo", "bar"})
	conf, _ := config.NewConfig(
		config.WithProviderLayer(
			basicflag.Provider(fs, "."), nil,
		),
		config.WithProviderLayer(
			confmap.Provider(map[string]interface{}{"foo": "baz"}, "."), nil,
		),
	)
	// We have two layers of configuration, the first one from flags and the second one from a map.
	// Both of them defined "foo". The first one should take precedence.
	fmt.Println(conf.String("foo"))
	// Output:
	// bar
}
