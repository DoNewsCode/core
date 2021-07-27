package leader

type event string

const OnStatusChanged event = "onStatusChanged"

type OnStatusChangedPayload struct {
	Status *Status
}
