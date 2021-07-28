package events

import (
	"github.com/DoNewsCode/core/contract"
)

type event string

// OnReload is an event that triggers the configuration reloads. The event payload is OnReloadPayload.
const OnReload event = "onReload"

// OnReload is an event that triggers the configuration reloads
type OnReloadPayload struct {
	// NewConf is the latest configuration after the reload.
	NewConf contract.ConfigAccessor
}
