package contract

// Marshaller is an interface for the data type that knows how to marshal itself.
type Marshaller interface {
	Marshal() ([]byte, error)
}
