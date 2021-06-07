package config

import (
	"github.com/DoNewsCode/core/contract"
)

type ReloadedEvent struct {
	NewConf contract.ConfigAccessor
}

func (r ReloadedEvent) Type() string {
	return "core.internal.reloaded_event"
}

func (r ReloadedEvent) Data() interface{} {
	return r.NewConf
}
