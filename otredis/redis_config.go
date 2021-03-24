package otredis

import (
	"github.com/DoNewsCode/core/config"
)

// RedisUniversalOptions is the configuration options for redis.
type RedisUniversalOptions struct {
	// Either a single address or a seed list of host:port addresses
	// of cluster/sentinel nodes.
	Addrs []string `json:"addrs" yaml:"addrs"`

	// Database to be selected after connecting to the server.
	// Only single-node and failover clients.
	DB int `json:"db" yaml:"db"`

	// Common options.

	Username         string `json:"username" yaml:"username"`
	Password         string `json:"password" yaml:"password"`
	SentinelPassword string `json:"sentinelPassword" yaml:"sentinelPassword"`

	MaxRetries      int             `json:"maxRetries" yaml:"maxRetries"`
	MinRetryBackoff config.Duration `json:"minRetryBackoff" yaml:"minRetryBackoff"`
	MaxRetryBackoff config.Duration `json:"maxRetryBackoff" yaml:"maxRetryBackoff"`

	DialTimeout  config.Duration `json:"dialTimeout" yaml:"dialTimeout"`
	ReadTimeout  config.Duration `json:"readTimeout" yaml:"readTimeout"`
	WriteTimeout config.Duration `json:"writeTimeout" yaml:"writeTimeout"`

	PoolSize           int             `json:"poolSize" yaml:"poolSize"`
	MinIdleConns       int             `json:"minIdleConns" yaml:"minIdleConns"`
	MaxConnAge         config.Duration `json:"maxConnAge" yaml:"maxConnAge"`
	PoolTimeout        config.Duration `json:"poolTimeout" yaml:"poolTimeout"`
	IdleTimeout        config.Duration `json:"idleTimeout" yaml:"idleTimeout"`
	IdleCheckFrequency config.Duration `json:"idleCheckFrequency" yaml:"idleCheckFrequency"`

	// Only cluster clients.

	MaxRedirects   int  `json:"maxRedirects" yaml:"maxRedirects"`
	ReadOnly       bool `json:"readOnly" yaml:"readOnly"`
	RouteByLatency bool `json:"routeByLatency" yaml:"routeByLatency"`
	RouteRandomly  bool `json:"routeRandomly" yaml:"routeRandomly"`

	// The sentinel master name.
	// Only failover clients.
	MasterName string `json:"masterName" yaml:"masterName"`
}
