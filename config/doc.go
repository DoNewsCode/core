// Package config provides supports to a fully customizable layered configuration stack.
//
// Introduction
//
// Configuration in go lives in many ways, with benefit and drawbacks for each.
// Flags, for example, offer self-explanatory help messages, but it can be
// cumbersome if the size of non-default entries is large. Environmental variables, on the other hand,
// are recommended by twelve-factor app methodology for its language neutrality and secrecy, but it lacks type
// safety and configuration hierarchy.
//
// Package config doesn't hold strong opinion on how you should structure your configurations. Rather, it allows you to
// define a configuration stack. For instance, You can put flags at first, envs at second, and configuration files at
// the third place. You are free to adjust the order in any other way or left out the layer you don't need.
//
// Package config also supports hot reload. If the desired signal
// triggers, the whole configuration stack will be reloaded in the same sequence you bootstrap them. Thus,
// a third place configuration file will not overwrite your flags and envs in first and second place in the reload.
//
// Usage
//
// See the example for how to create a typical configuration stack. After the configuration is created, you can access
// it with the interface in contract.ConfigAccessor.
//
//  type ConfigAccessor interface {
//    String(string) string
//    Int(string) int
//    Strings(string) []string
//    Bool(string) bool
//    Get(string) interface{}
//    Float64(string) float64
//    Unmarshal(path string, o interface{}) error
//  }
//
// Package config uses koanf (https://github.com/knadh/koanf) to achieve many of its features. The configuration stack
// can be build with a rich set of already available provider and parsers in koanf. See
// https://github.com/knadh/koanf/blob/master/README.md for more info.
//
// Integrate
//
// Package config is part of the core. When using package core, the config is bootstrapped in the initialization
// phase. Package core's constructor inherits some of the options for configuration stack.
// See package core for more info.
//
// A command is provided to export the default configuration:
//
//  go run main.go config init -o ./config/config.yaml
//
// Best Practice
//
// In general you should not pass contract.ConfigAccessor or config.KoanfAdapter to your services. You should only
// pass unmarshalled strings and structs that matters. You don't want your service unnecessarily depend on package config.
//
// The only exception is when you need configuration hot reload. In this case, you have to pass the contract.ConfigAccessor to your
// service, and access the config repeatedly in each request/iteration of you service.
//
// Future scope
//
// Remote configuration store such as zookeeper, consul and ETCD can be adopted in this package in the same manner
// as the file provider. To avoid dependency bloat, they might live in their own subpackage.
package config
