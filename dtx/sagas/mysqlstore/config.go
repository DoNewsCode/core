package mysqlstore

import (
	"time"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/di"
)

type configOut struct {
	di.Out

	Config []config.ExportedConfig `group:"config,flatten"`
}

var defaultConfig = configuration{
	Connection:      "default",
	Retention:       config.Duration{Duration: 168 * time.Hour},
	CleanupInterval: config.Duration{Duration: time.Hour},
}

func provideConfig() configOut {
	return configOut{
		Config: []config.ExportedConfig{
			{
				Owner: "sagas",
				Data: map[string]interface{}{
					"sagas-mysql": defaultConfig,
				},
				Comment: "The saga mysql store configuration.",
			},
		},
	}
}

type configuration struct {
	Connection      string          `json:"connection" yaml:"connection"`
	Retention       config.Duration `json:"retention" yaml:"retention"`
	CleanupInterval config.Duration `json:"cleanupInterval" yaml:"cleanupInterval"`
}

func (c configuration) getConnection() string {
	if c.Connection == "" {
		return defaultConfig.Connection
	}
	return c.Connection
}

func (c configuration) getRetention() config.Duration {
	if c.Retention.IsZero() {
		return defaultConfig.Retention
	}
	return c.Retention
}

func (c configuration) getCleanupInterval() config.Duration {
	if c.CleanupInterval.IsZero() {
		return defaultConfig.CleanupInterval
	}
	return c.CleanupInterval
}
