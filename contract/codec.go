package contract

// Codec is an interface for serialization and deserialization.
type Codec interface {
	Unmarshal(data []byte, value any) error
	Marshal(value any) ([]byte, error)
}
