package config

import (
	"strings"

	"github.com/DoNewsCode/std/pkg/contract"
)

type Env string

func (e Env) IsLocal() bool {
	return e == "local"
}

func (e Env) IsTesting() bool {
	return e == "testing"
}

func (e Env) IsDevelopment() bool {
	return e == "development"
}

func (e Env) IsProduction() bool {
	return e == "production"
}

func (e Env) String() string {
	return string(e)
}

func NewEnv(env string) Env {
	if strings.EqualFold("prod", env) || strings.EqualFold("production", env) {
		return "production"
	}
	if strings.EqualFold("development", env) || strings.EqualFold("dev", env) {
		return "development"
	}
	if strings.EqualFold("local", env) {
		return "local"
	}
	if strings.EqualFold("testing", env) {
		return "testing"
	}
	return "unknown"
}

func NewEnvFromConf(conf contract.ConfigAccessor) Env {
	envStr := conf.String("env")
	return NewEnv(envStr)
}
