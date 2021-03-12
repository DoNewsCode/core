package contract

// Marshaller is an interface for the data type that knows how to marshal itself.
type Marshaller interface {
	Marshal() ([]byte, error)
}

// Codec is an interface for serialization and deserialization.
type Codec interface {
	Unmarshal(data []byte, value interface{}) error
	Marshal(value interface{}) ([]byte, error)
}
