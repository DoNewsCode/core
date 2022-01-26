package config

import (
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/eventsv2"
)

type OnReloadEvent = eventsv2.Event[contract.ConfigUnmarshaler]

var _ contract.ConfigReloadDispatcher = (*OnReloadEvent)(nil)
