package queue

// RetryingEvent is a contract.Event that triggers when a certain event failed to be processed, and it is up for retry.
// Note: if retry attempts are exhausted, this event won't be triggered.
type RetryingEvent struct {
	Err error
	Msg *PersistedEvent
}
