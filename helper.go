package core

import (
	"errors"
	"fmt"
	"strings"
)

func getString(data map[string]interface{}, key ...string) (string, error) {
	if len(key) <= 0 {
		panic("key must be provided at least once")
	}
	for i := 0; i < len(key)-1; i++ {
		value, ok := data[key[i]]
		if !ok {
			return "", fmt.Errorf("%s doesn't exist", strings.Join(key[0:i+1], "."))
		}
		data, ok = value.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("%s is not a map", strings.Join(key[0:i+1], "."))
		}
	}
	value, ok := data[key[len(key)-1]]
	if !ok {
		return "", fmt.Errorf("%s doesn't exist", strings.Join(key, "."))
	}
	str, ok := value.(string)
	if !ok {
		return str, errors.New("must be a string")
	}
	return str, nil
}

func getBool(data map[string]interface{}, key ...string) (bool, error) {
	if len(key) <= 0 {
		panic("key must be provided at least once")
	}
	for i := 0; i < len(key)-1; i++ {
		value, ok := data[key[i]]
		if !ok {
			return false, fmt.Errorf("%s doesn't exist", strings.Join(key[0:i+1], "."))
		}
		data, ok = value.(map[string]interface{})
		if !ok {
			return false, fmt.Errorf("%s is not a map", strings.Join(key[0:i+1], "."))
		}
	}
	value, ok := data[key[len(key)-1]]
	if !ok {
		return false, fmt.Errorf("%s doesn't exist", strings.Join(key, "."))
	}
	b, ok := value.(bool)
	if !ok {
		return b, errors.New("must be a bool")
	}
	return b, nil
}

// isValidLevel tests if the given input is valid level config.
func isValidLevel(levelCfg string) bool {
	validLevel := []string{"debug", "info", "warn", "error", "none"}
	for i := range validLevel {
		if validLevel[i] == levelCfg {
			return true
		}
	}
	return false
}

// isValidLevel tests if the given input is valid format config.
func isValidFormat(format string) bool {
	validFormat := []string{"json", "logfmt"}
	for i := range validFormat {
		if validFormat[i] == format {
			return true
		}
	}
	return false
}
