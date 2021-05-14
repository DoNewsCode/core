package internal

import (
	"os"
	"strings"
)

// GetDefaultAddrsFromEnv return addrs
func GetDefaultAddrsFromEnv(env, defaultVal string) ([]string, bool) {
	if v := os.Getenv(env); v != "" {
		return strings.Split(v, ","), true
	}
	return []string{defaultVal}, false
}

// GetDefaultAddrFromEnv return addr/dsn
func GetDefaultAddrFromEnv(env, defaultVal string) (string, bool) {
	if v := os.Getenv(env); v != "" {
		return v, true
	}
	return defaultVal, false
}
