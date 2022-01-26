package eventsv2

import "github.com/DoNewsCode/core/contract"

type OnReloadEvent = Event[OnReloadPayload]

type OnReloadPayload struct {
	Unmarshaler contract.ConfigUnmarshaler
}
