package contract

type Keyer interface {
	Key(delimiter string, args ...string) string
	Spread() []string
}
