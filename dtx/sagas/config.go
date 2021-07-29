package sagas

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
	SagaTimeout:     config.Duration{Duration: 600 * time.Second},
	RecoverInterval: config.Duration{Duration: 60 * time.Second},
}

func provideConfig() configOut {
	return configOut{
		Config: []config.ExportedConfig{
			{
				Owner: "sagas",
				Data: map[string]interface{}{
					"sagas": defaultConfig,
				},
				Comment: "The saga configuration.",
			},
		},
	}
}

type configuration struct {
	SagaTimeout     config.Duration `json:"sagaTimeout" yaml:"sagaTimeout"`
	RecoverInterval config.Duration `json:"recoverInterval" yaml:"recoverInterval"`
}

func (c configuration) getSagaTimeout() config.Duration {
	if c.SagaTimeout.IsZero() {
		return defaultConfig.SagaTimeout
	}
	return c.SagaTimeout
}

func (c configuration) getRecoverInterval() config.Duration {
	if c.RecoverInterval.IsZero() {
		return defaultConfig.RecoverInterval
	}
	return c.RecoverInterval
}
