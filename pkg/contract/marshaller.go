package contract

type Marshaller interface {
	Marshal() ([]byte, error)
}

type UnMarshaller interface {
	Unmarshal() ([]byte, error)
}
