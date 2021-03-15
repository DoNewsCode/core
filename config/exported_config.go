package config

// ExportedConfig is a struct that outlines a set of configuration.
// Each module is supposed to emit ExportedConfig into DI, and Package config should collect them.
type ExportedConfig struct {
	Owner   string
	Data    map[string]interface{}
	Comment string
	Order int
}
