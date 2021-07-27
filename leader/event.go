package leader

type event string

// OnStatusChanged is an event that triggers when the leadership has transited. It's payload is OnStatusChangedPayload.
const OnStatusChanged event = "onStatusChanged"

// OnStatusChangedPayload is the payload of OnStatusChanged.
type OnStatusChangedPayload struct {
	Status *Status
}
