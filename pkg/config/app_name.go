package config

import "github.com/DoNewsCode/std/pkg/contract"

type AppName string

func (a AppName) String() string {
	return string(a)
}

func ProvideAppName(conf contract.ConfigAccessor) AppName {
	return AppName(conf.String("name"))
}
