package eventsv2

import "github.com/DoNewsCode/core/contract"

type OnReloadEvent = Event[OnReload]

type OnReload struct {
	Accessor contract.ConfigAccessor
}
