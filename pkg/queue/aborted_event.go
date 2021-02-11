package queue

type AbortedEvent struct {
	Err error
	Msg *PersistedEvent
}
