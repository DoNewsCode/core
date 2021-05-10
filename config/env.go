package config

import (
	"os"
	"strings"

	"github.com/DoNewsCode/core/contract"
)

// global static variables for Env
const (
	// local
	EnvLocal Env = "local"
	// testing
	EnvTesting Env = "testing"
	// development
	EnvDevelopment Env = "development"
	// staging
	EnvStaging Env = "staging"
	// production
	EnvProduction Env = "production"
	// unknown
	EnvUnknown Env = "unknown"
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

func getDefaultAddrsFromEnv(env, defaultVal string) []string {
	if v := os.Getenv(env); v != "" {
		return strings.Split(v, ",")
	}
	return []string{defaultVal}
}

// Default multiple addresses from env
var (
	ENV_DEFAULT_ELASTICSEARCH_ADDRS = getDefaultAddrsFromEnv("ELASTICSEARCH_ADDR", "http://127.0.0.1:9200")
	ENV_DEFAULT_ETCD_ADDRS          = getDefaultAddrsFromEnv("ETCD_ADDR", "127.0.0.1:2379")
	ENV_DEFAULT_KAFKA_ADDRS         = getDefaultAddrsFromEnv("KAFKA_ADDR", "127.0.0.1:9092")
	ENV_DEFAULT_REDIS_ADDRS         = getDefaultAddrsFromEnv("REDIS_ADDR", "127.0.0.1:6379")
)
