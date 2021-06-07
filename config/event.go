package config

import (
	"github.com/DoNewsCode/core/contract"
)

// ReloadedEvent is an event that triggers the configuration reloads
type ReloadedEvent struct {
	// NewConf is the latest configuration after the reload.
	NewConf contract.ConfigAccessor
}

// Type implements contract.Event
func (r ReloadedEvent) Type() string {
	return "core.internal.reloaded_event"
}

// Data implements contract.Event
func (r ReloadedEvent) Data() interface{} {
	return r.NewConf
}
