package config

import "github.com/DoNewsCode/std/pkg/contract"

// AppName represents the name the application. It is primarily used as a symbol for dependency injection.
type AppName string

// String gives the string form of AppName.
func (a AppName) String() string {
	return string(a)
}

// NewAppNameFromConf reads the name of application from configuration's "name" entry.
func NewAppNameFromConf(conf contract.ConfigAccessor) AppName {
	return AppName(conf.String("name"))
}
