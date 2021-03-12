package config

import (
	"strings"

	"github.com/DoNewsCode/core/contract"
)

const (
	EnvLocal       Env = "local"
	EnvTesting     Env = "testing"
	EnvDevelopment Env = "development"
	EnvStaging     Env = "staging"
	EnvProduction  Env = "production"
	EnvUnknown     Env = "unknown"
)

// Env is the environment of the application. It is primarily used as dependency injection symbol
type Env string

// IsLocal returns true if the environment is local
func (e Env) IsLocal() bool {
	return e == EnvLocal
}

// IsTesting returns true if the environment is testing
func (e Env) IsTesting() bool {
	return e == EnvTesting
}

// IsDevelopment returns true if the environment is development
func (e Env) IsDevelopment() bool {
	return e == EnvDevelopment
}

// IsStaging returns true if the environment is staging
func (e Env) IsStaging() bool {
	return e == EnvStaging
}

// IsProduction returns true if the environment is production
func (e Env) IsProduction() bool {
	return e == EnvProduction
}

// String returns the string form of the environment. This is a lowercase full word, such as production.
func (e Env) String() string {
	return string(e)
}

// NewEnv takes in environment string and returns a Env type. It does some "best-effort" normalization internally.
// For example, prod, PROD, production and PRODUCTION produces the same type. It is recommended to use one of
// "production", "staging", "development", "local", or "testing" as output to avoid unexpected outcome.
func NewEnv(env string) Env {
	switch strings.ToLower(env) {
	case "production", "prod", "online":
		return EnvProduction
	case "pre-prod", "staging":
		return EnvStaging
	case "development", "develop", "dev":
		return EnvDevelopment
	case "local":
		return EnvLocal
	case "testing", "test":
		return EnvTesting
	default:
		return EnvUnknown
	}
}

// NewEnvFromConf reads the name of application from configuration's "env" entry.
func NewEnvFromConf(conf contract.ConfigAccessor) Env {
	envStr := conf.String("env")
	return NewEnv(envStr)
}
