package config

// ExportedConfig is a struct that outlines a set of configuration.
// Each module is supposed to emit ExportedConfig into DI, and Package config should collect them.
type ExportedConfig struct {
	Owner    string
	Data     map[string]interface{}
	Comment  string
	Validate Validator
}

// Validator is a method to verify if config is valid. If it is not valid, the
// returned error should contain a human readable description of why.
type Validator func(data map[string]interface{}) error
