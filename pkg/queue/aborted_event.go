package queue

// AbortedEvent is a contract.Event that triggers when a event is timeout or failed.
// If the event still has retry attempts remaining, this event won't be triggered.
type AbortedEvent struct {
	Err error
	Msg *PersistedEvent
}
