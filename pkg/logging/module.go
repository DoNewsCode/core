package logging

import (
	"github.com/DoNewsCode/std/pkg/contract"
)

type Module struct {
}

func (m Module) ProvideConfig() []contract.ExportedConfig {
	return []contract.ExportedConfig{
		{
			Name: "log",
			Data: map[string]interface{}{
				"log": map[string]interface{}{"level": "debug", "format": "logfmt"},
			},
			Comment: "The global logging level and format",
		},
	}
}
