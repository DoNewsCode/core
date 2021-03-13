package sagas

type rollbackEvent struct {
	name    string
	request interface{}
}

func (s rollbackEvent) Type() string {
	return s.name
}

func (s rollbackEvent) Data() interface{} {
	return s.request
}
