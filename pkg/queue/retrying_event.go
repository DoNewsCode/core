package queue

type RetryingEvent struct {
	Err error
	Msg *PersistedEvent
}
