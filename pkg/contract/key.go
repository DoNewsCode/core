package contract

// Keyer is an interface for key passing.
type Keyer interface {
	Key(delimiter string, args ...string) string
	Spread() []string
}
