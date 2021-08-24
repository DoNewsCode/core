package contract

// Codec is an interface for serialization and deserialization.
type Codec interface {
	Unmarshal(data []byte, value interface{}) error
	Marshal(value interface{}) ([]byte, error)
}
