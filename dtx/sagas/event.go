package sagas

type event string

func onRollback(name string) event {
	return event(name)
}

type onRollbackPayload struct {
	request interface{}
}
