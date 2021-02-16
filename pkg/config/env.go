package config

import (
	"strings"

	"github.com/DoNewsCode/std/pkg/contract"
)

// Env is the environment of the application. It is primarily used as dependency injection symbol
type Env string

// IsLocal returns true if the environment is local
func (e Env) IsLocal() bool {
	return e == "local"
}

// IsTesting returns true if the environment is testing
func (e Env) IsTesting() bool {
	return e == "testing"
}

// IsDevelopment returns true if the environment is development
func (e Env) IsDevelopment() bool {
	return e == "development"
}

// IsStaging returns true if the environment is staging
func (e Env) IsStaging() bool {
	return e == "staging"
}

// IsProduction returns true if the environment is production
func (e Env) IsProduction() bool {
	return e == "production"
}

// String returns the string form of the environment. This is a lowercase full word, such as production.
func (e Env) String() string {
	return string(e)
}

// NewEnv takes in environment string and returns a Env type. It does some "best-effort" normalization internally.
// For example, prod, PROD, production and PRODUCTION produces the same type. It is recommended to use one of
// "production", "staging", "development", "local", or "testing" as input to avoid unexpected outcome.
func NewEnv(env string) Env {
	if strings.EqualFold("prod", env) || strings.EqualFold("production", env) {
		return "production"
	}
	if strings.EqualFold("pre-prod", env) || strings.EqualFold("staging", env) {
		return "staging"
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

// NewEnvFromConf reads the name of application from configuration's "env" entry.
func NewEnvFromConf(conf contract.ConfigAccessor) Env {
	envStr := conf.String("env")
	return NewEnv(envStr)
}
