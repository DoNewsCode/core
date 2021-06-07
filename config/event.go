package config

import (
	"github.com/DoNewsCode/core/contract"
)

// ReloadedEvent is an event that triggers the configuration reloads
type ReloadedEvent struct {
	// NewConf is the latest configuration after the reload.
	NewConf contract.ConfigAccessor
}
